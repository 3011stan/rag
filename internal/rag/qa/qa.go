package qa

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

// Service define o contrato mínimo usado pelos handlers da API.
type Service interface {
	Answer(ctx context.Context, question string, ret *retriever.Retriever) (*Response, error)
}

// QAService é responsável por gerar respostas a perguntas
type QAService struct {
	client           *openai.Client
	model            string
	promptTemplate   *PromptTemplate
	maxContextTokens int
	temperature      float32
	topK             int
}

// Response representa uma resposta gerada pelo QA Service
type Response struct {
	Answer  string
	Sources []SourceReference
	Model   string
}

// SourceReference é uma referência de fonte
type SourceReference struct {
	DocumentID string  `json:"document_id"`
	ChunkIndex int     `json:"chunk_index"`
	Score      float64 `json:"score"`
	Preview    string  `json:"preview"`
}

// NewQAService cria uma nova instância do QA Service
func NewQAService(apiKey string) *QAService {
	return &QAService{
		client:           openai.NewClient(apiKey),
		model:            openai.GPT4o,
		promptTemplate:   DefaultPromptTemplate(),
		maxContextTokens: 2000,
		temperature:      0.3,
		topK:             5,
	}
}

// WithModel define o modelo a ser usado
func (qs *QAService) WithModel(model string) *QAService {
	qs.model = model
	return qs
}

// WithPromptTemplate define o template de prompt
func (qs *QAService) WithPromptTemplate(template *PromptTemplate) *QAService {
	qs.promptTemplate = template
	return qs
}

// WithTemperature define a temperature
func (qs *QAService) WithTemperature(temp float32) *QAService {
	qs.temperature = temp
	return qs
}

// WithTopK define o número padrão de chunks a recuperar
func (qs *QAService) WithTopK(k int) *QAService {
	qs.topK = k
	return qs
}

// Answer gera uma resposta para uma pergunta usando um retriever
func (qs *QAService) Answer(ctx context.Context, question string, ret *retriever.Retriever) (*Response, error) {
	// Recuperar chunks relevantes
	searchResults, err := ret.Retrieve(ctx, question, qs.topK)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve context: %w", err)
	}

	if len(searchResults) == 0 {
		return &Response{
			Answer: "I could not find any relevant information in the knowledge base to answer your question.",
			Model:  qs.model,
		}, nil
	}

	// Preparar contexto
	contextChunks := make([]string, len(searchResults))
	sources := make([]SourceInfo, len(searchResults))
	sourceRefs := make([]SourceReference, len(searchResults))

	for i, result := range searchResults {
		contextChunks[i] = result.Chunk.Content
		sources[i] = SourceInfo{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
		}
		sourceRefs[i] = SourceReference{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
			Preview:    truncateText(result.Chunk.Content, 100),
		}
	}

	context := FormatContext(contextChunks)

	// Construir prompt
	prompt, err := BuildPrompt(qs.promptTemplate, question, context, sources)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Chamar LLM
	answer, err := qs.callLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	return &Response{
		Answer:  answer,
		Sources: sourceRefs,
		Model:   qs.model,
	}, nil
}

// AnswerWithOptions gera uma resposta com opções customizadas
func (qs *QAService) AnswerWithOptions(ctx context.Context, question string, ret *retriever.Retriever, topK int) (*Response, error) {
	// Temporariamente alterar topK
	oldTopK := qs.topK
	qs.topK = topK
	defer func() { qs.topK = oldTopK }()

	return qs.Answer(ctx, question, ret)
}

// callLLM chama a API do LLM
func (qs *QAService) callLLM(ctx context.Context, prompt string) (string, error) {
	resp, err := qs.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       qs.model,
		Temperature: qs.temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: SystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}

// truncateText trunca um texto para um comprimento máximo
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// StreamAnswer gera uma resposta em streaming
func (qs *QAService) StreamAnswer(ctx context.Context, question string, ret *retriever.Retriever, streamChan chan string) error {
	defer close(streamChan)

	// Recuperar chunks relevantes
	searchResults, err := ret.Retrieve(ctx, question, qs.topK)
	if err != nil {
		return fmt.Errorf("failed to retrieve context: %w", err)
	}

	if len(searchResults) == 0 {
		streamChan <- "I could not find any relevant information in the knowledge base to answer your question."
		return nil
	}

	// Preparar contexto
	contextChunks := make([]string, len(searchResults))
	sources := make([]SourceInfo, len(searchResults))

	for i, result := range searchResults {
		contextChunks[i] = result.Chunk.Content
		sources[i] = SourceInfo{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
		}
	}

	context := FormatContext(contextChunks)

	// Construir prompt
	prompt, err := BuildPrompt(qs.promptTemplate, question, context, sources)
	if err != nil {
		return fmt.Errorf("failed to build prompt: %w", err)
	}

	// Chamar LLM em streaming
	return qs.streamLLM(ctx, prompt, streamChan)
}

// streamLLM chama a API do LLM em streaming
func (qs *QAService) streamLLM(ctx context.Context, prompt string, streamChan chan string) error {
	stream, err := qs.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       qs.model,
		Temperature: qs.temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: SystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}

		if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
			streamChan <- response.Choices[0].Delta.Content
		}
	}
}
