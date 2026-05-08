package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stan/Projects/studies/rag/internal/config"
	"github.com/stan/Projects/studies/rag/internal/logging"
	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/chunker"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
	"github.com/stan/Projects/studies/rag/internal/rag/loader"
	"github.com/stan/Projects/studies/rag/internal/rag/providers"
	"github.com/stan/Projects/studies/rag/internal/rag/qa"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

const (
	contentTypeJSON   = "application/json"
	contentTypeHeader = "Content-Type"
	maxFileSize       = 100 << 20 // 100MB
	maxQuestionLength = 5000
	maxTopK           = 20
	ingestTimeout     = 5 * time.Minute
	askTimeout        = 2 * time.Minute
)

// APIServer contém as dependências para os handlers
type APIServer struct {
	cfg       *config.Config
	vs        rag.VectorStore
	embedder  embeddings.Provider
	chunker   *chunker.Chunker
	pdfLoader *loader.PDFLoader
	retriever *retriever.Retriever
	qaService qa.Service
}

// NewAPIServer cria uma nova instância do servidor API
func NewAPIServer(cfg *config.Config) (*APIServer, error) {
	// Criar VectorStore
	vs, err := rag.NewPGVectorStore(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}
	if err := vs.EnsureSchema(context.Background(), cfg.EmbeddingDimensions); err != nil {
		return nil, fmt.Errorf("failed to ensure vector store schema: %w", err)
	}

	embProvider, err := providers.NewEmbeddingProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}
	qaService, err := providers.NewQAService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create QA service: %w", err)
	}

	// Criar Chunker
	chk, err := chunker.NewChunker(chunker.ChunkerConfig{
		ChunkTokens:   cfg.ChunkTokens,
		OverlapTokens: cfg.OverlapTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chunker: %w", err)
	}

	// Criar PDF Loader
	pdfLdr := loader.NewPDFLoader(0) // 0 = sem limite de páginas

	// Criar Retriever
	ret := retriever.NewRetriever(vs, embProvider, cfg.TopK)

	return &APIServer{
		cfg:       cfg,
		vs:        vs,
		embedder:  embProvider,
		chunker:   chk,
		pdfLoader: pdfLdr,
		retriever: ret,
		qaService: qaService,
	}, nil
}

// Request/Response types
type IngestRequest struct {
	DocumentID string `json:"document_id"` // opcional, gerado se não fornecido
}

type IngestResponse struct {
	DocumentID string `json:"document_id"`
	ChunkCount int    `json:"chunk_count"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

type AskRequest struct {
	Question string `json:"question"`
	TopK     int    `json:"top_k,omitempty"`
}

type SourceInfo struct {
	DocumentID string  `json:"document_id"`
	ChunkIndex int     `json:"chunk_index"`
	Score      float64 `json:"score"`
	Preview    string  `json:"preview"`
}

type AskResponse struct {
	Answer  string       `json:"answer"`
	Sources []SourceInfo `json:"sources"`
	Error   string       `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// IngestHandler processa o upload e ingestão de PDFs
func (srv *APIServer) IngestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), ingestTimeout)
		defer cancel()

		logger := logging.FromContext(ctx)

		// Parse e validar arquivo
		doc, fileData, fileName, err := srv.parseAndValidateFile(r)
		if err != nil {
			logger.Warn().Err(err).Msg("validation failed")
			respondError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}

		logger.Info().
			Str("filename", fileName).
			Int("size_bytes", len(fileData)).
			Msg("processing PDF")

		// Pipeline de ingestão
		chunkCount, err := srv.ingestPipeline(ctx, doc, fileData, fileName)
		if err != nil {
			logger.Error().Err(err).Msg("ingest pipeline failed")
			respondError(w, err.Error(), http.StatusInternalServerError, nil)
			return
		}

		// Responder com sucesso
		resp := IngestResponse{
			DocumentID: doc.ID,
			ChunkCount: chunkCount,
			Status:     "success",
			Message:    fmt.Sprintf("Successfully ingested PDF: %s", fileName),
		}

		logger.Info().
			Str("document_id", doc.ID).
			Msg("PDF ingested successfully")

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// AskHandler processa perguntas e retorna respostas
func (srv *APIServer) AskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), askTimeout)
		defer cancel()

		logger := logging.FromContext(ctx)

		// Parse e validar request
		req, err := srv.parseAndValidateQuestion(r)
		if err != nil {
			logger.Warn().Err(err).Msg("validation failed")
			respondError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}

		question := req.Question
		if len(question) > 50 {
			question = question[:50]
		}

		logger.Info().
			Str("question_preview", question).
			Int("top_k", req.TopK).
			Msg("processing question")

		// Pipeline de resposta
		resp, err := srv.answerPipeline(ctx, req)
		if err != nil {
			logger.Error().Err(err).Msg("answer pipeline failed")
			respondError(w, err.Error(), http.StatusInternalServerError, nil)
			return
		}

		logger.Info().
			Int("source_count", len(resp.Sources)).
			Msg("answer generated successfully")

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// Helper methods

func (srv *APIServer) parseAndValidateFile(r *http.Request) (*rag.Document, []byte, string, error) {
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		return nil, nil, "", fmt.Errorf("failed to parse form: %v", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, nil, "", fmt.Errorf("no file provided")
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to read file: %v", err)
	}

	if len(fileData) == 0 {
		return nil, nil, "", fmt.Errorf("file is empty")
	}

	// Validar que é um PDF
	if len(fileData) < 4 || string(fileData[:4]) != "%PDF" {
		return nil, nil, "", fmt.Errorf("file is not a valid PDF")
	}

	doc := &rag.Document{} // Será populado pelo PDF Loader
	return doc, fileData, header.Filename, nil
}

func (srv *APIServer) ingestPipeline(ctx context.Context, doc *rag.Document, fileData []byte, fileName string) (int, error) {
	logger := logging.FromContext(ctx)

	// 1. Extrair documento e texto usando PDF Loader
	loadedDoc, text, err := srv.pdfLoader.LoadPDF(fileData, fileName)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PDF: %v", err)
	}

	*doc = *loadedDoc
	logger.Info().
		Int("char_count", len(text)).
		Str("filename", fileName).
		Msg("PDF text extracted")

	// 2. Inserir documento no banco
	if err := srv.vs.InsertDocument(ctx, *doc); err != nil {
		return 0, fmt.Errorf("failed to insert document: %v", err)
	}

	// 3. Chunking
	chunks, err := srv.chunker.ChunkText(doc.ID, text, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to chunk text: %v", err)
	}

	if len(chunks) == 0 {
		return 0, fmt.Errorf("no chunks generated from text")
	}

	logger.Info().
		Int("chunk_count", len(chunks)).
		Msg("text chunking completed")

	// 4. Gerar embeddings
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Content
	}

	embeddings, err := srv.embedder.Embed(ctx, texts)
	if err != nil {
		return 0, fmt.Errorf("failed to generate embeddings: %v", err)
	}

	if len(embeddings) != len(chunks) {
		return 0, fmt.Errorf("embedding count mismatch")
	}

	// Atribuir embeddings aos chunks
	for i, emb := range embeddings {
		chunks[i].Embedding = emb
	}

	logger.Info().
		Int("embedding_count", len(embeddings)).
		Msg("embeddings generated")

	// 5. Inserir chunks no VectorStore
	if err := srv.vs.InsertBatch(ctx, chunks); err != nil {
		return 0, fmt.Errorf("failed to insert chunks: %v", err)
	}

	logger.Info().
		Str("document_id", doc.ID).
		Int("chunk_count", len(chunks)).
		Msg("document ingestion pipeline completed")

	return len(chunks), nil
}

func (srv *APIServer) parseAndValidateQuestion(r *http.Request) (*AskRequest, error) {
	var req AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request format: %v", err)
	}

	if req.Question == "" {
		return nil, fmt.Errorf("question is required")
	}

	if len(req.Question) > maxQuestionLength {
		return nil, fmt.Errorf("question is too long (max %d chars)", maxQuestionLength)
	}

	// Usar topK default se não fornecido
	if req.TopK <= 0 {
		req.TopK = srv.cfg.TopK
	}
	if req.TopK > maxTopK {
		req.TopK = maxTopK // Limitar para não sobrecarregar
	}

	return &req, nil
}

func (srv *APIServer) answerPipeline(ctx context.Context, req *AskRequest) (*AskResponse, error) {
	logger := logging.FromContext(ctx)

	// 1. Usar Retriever para buscar chunks relevantes
	searchResults, err := srv.retriever.Retrieve(ctx, req.Question, req.TopK)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve context: %v", err)
	}

	if len(searchResults) == 0 {
		logger.Warn().Msg("no relevant chunks found for question")
		return &AskResponse{
			Answer:  "I could not find any relevant information in the knowledge base to answer your question.",
			Sources: []SourceInfo{},
		}, nil
	}

	logger.Info().
		Int("chunk_count", len(searchResults)).
		Msg("relevant chunks retrieved")

	// 2. Preparar dados para resposta
	sourceRefs := make([]SourceInfo, len(searchResults))
	for i, result := range searchResults {
		sourceRefs[i] = SourceInfo{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
			Preview:    truncateText(result.Chunk.Content, 100),
		}
	}

	// 3. Usar QA Service para gerar resposta
	qaResponse, err := srv.qaService.Answer(ctx, req.Question, srv.retriever)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %v", err)
	}

	logger.Info().
		Int("source_count", len(sourceRefs)).
		Msg("answer generation completed")

	return &AskResponse{
		Answer:  qaResponse.Answer,
		Sources: sourceRefs,
	}, nil
}

func respondError(w http.ResponseWriter, message string, statusCode int, err error) {
	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}

	resp := ErrorResponse{
		Error:   message,
		Message: errorMsg,
		Code:    statusCode,
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
