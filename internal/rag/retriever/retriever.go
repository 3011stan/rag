package retriever

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
)

// SearchFilter define filtros para busca
type SearchFilter struct {
	DocumentID string
	MinScore   float64
}

// SearchResult representa um resultado de busca
type SearchResult struct {
	Chunk rag.Chunk
	Score float64
	Rank  int
}

// Retriever é responsável por buscar chunks relevantes para uma pergunta
type Retriever struct {
	vectorStore rag.VectorStore
	embeddings  embeddings.Provider
	defaultTopK int
	scoreScale  float64 // Para normalização de scores (0-1)
}

// NewRetriever cria uma nova instância do Retriever
func NewRetriever(vs rag.VectorStore, ep embeddings.Provider, defaultTopK int) *Retriever {
	return &Retriever{
		vectorStore: vs,
		embeddings:  ep,
		defaultTopK: defaultTopK,
		scoreScale:  1.0, // Normalizar para 0-1
	}
}

// Retrieve busca os chunks mais relevantes para uma pergunta
func (r *Retriever) Retrieve(ctx context.Context, question string, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = r.defaultTopK
	}

	// Gerar embedding da pergunta
	embedding, err := r.embeddings.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	// Buscar chunks similares
	chunks, err := r.vectorStore.Search(ctx, embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}

	// Converter para SearchResult com scores normalizados
	results := make([]SearchResult, len(chunks))
	for i, chunk := range chunks {
		results[i] = SearchResult{
			Chunk: chunk,
			Score: r.normalizeScore(chunk.Score),
			Rank:  i + 1,
		}
	}

	return results, nil
}

// RetrieveWithFilters busca chunks com filtros aplicados
func (r *Retriever) RetrieveWithFilters(ctx context.Context, question string, topK int, filters *SearchFilter) ([]SearchResult, error) {
	if topK <= 0 {
		topK = r.defaultTopK
	}

	// Gerar embedding da pergunta
	embedding, err := r.embeddings.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	// Preparar filtros para o VectorStore
	filterMap := make(map[string]interface{})
	if filters != nil && filters.DocumentID != "" {
		filterMap["document_id"] = filters.DocumentID
	}

	// Buscar chunks similares com filtros
	chunks, err := r.vectorStore.SearchWithFilters(ctx, embedding, topK*2, filterMap)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks with filters: %w", err)
	}

	// Converter para SearchResult e aplicar score mínimo
	var results []SearchResult
	for i, chunk := range chunks {
		normalizedScore := r.normalizeScore(chunk.Score)

		// Aplicar filtro de score mínimo
		if filters != nil && filters.MinScore > 0 && normalizedScore < filters.MinScore {
			continue
		}

		results = append(results, SearchResult{
			Chunk: chunk,
			Score: normalizedScore,
			Rank:  i + 1,
		})

		// Limitar ao topK solicitado
		if len(results) >= topK {
			break
		}
	}

	return results, nil
}

// RetrieveMultiple busca chunks para múltiplas perguntas
func (r *Retriever) RetrieveMultiple(ctx context.Context, questions []string, topK int) (map[string][]SearchResult, error) {
	if topK <= 0 {
		topK = r.defaultTopK
	}

	results := make(map[string][]SearchResult)

	for _, question := range questions {
		searchResults, err := r.Retrieve(ctx, question, topK)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve for question '%s': %w", question, err)
		}

		results[question] = searchResults
	}

	return results, nil
}

// RetrieveByDocumentID busca todos os chunks de um documento específico
func (r *Retriever) RetrieveByDocumentID(ctx context.Context, documentID string) ([]rag.Chunk, error) {
	chunks, err := r.vectorStore.GetChunksByDocumentID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve chunks for document: %w", err)
	}

	// Ordenar por índice de chunk
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkIndex < chunks[j].ChunkIndex
	})

	return chunks, nil
}

// normalizeScore normaliza um score de similaridade para 0-1
// Assume que scores de similaridade cosine estão entre -1 e 1
// Converte para 0-1 onde 1 é mais similar
func (r *Retriever) normalizeScore(score float64) float64 {
	// Clampar score entre -1 e 1 (caso haja erro)
	if score < -1 {
		score = -1
	} else if score > 1 {
		score = 1
	}

	// Converter de [-1, 1] para [0, 1]
	// score = (score + 1) / 2
	// Mas consideramos que scores cosine estão em [0, 1] para embeddings
	// Por isso retornamos o score já normalizado

	// Se score for > 1, normalizamos usando sigmoid aproximado
	if score > 1 {
		return 1.0 / (1.0 + math.Exp(-score))
	}

	return score
}

// RankResults ordena resultados por score (decrescente)
func RankResults(results []SearchResult) []SearchResult {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Rank < results[j].Rank
	})

	// Atualizar ranks após ordenação
	for i := range results {
		results[i].Rank = i + 1
	}

	return results
}

// FilterByScore filtra resultados por score mínimo
func FilterByScore(results []SearchResult, minScore float64) []SearchResult {
	var filtered []SearchResult
	for _, result := range results {
		if result.Score >= minScore {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// MergeResults mescla múltiplos resultados (ex: de múltiplas buscas)
func MergeResults(resultsList ...[]SearchResult) []SearchResult {
	// Usar um mapa para evitar duplicatas
	resultMap := make(map[string]SearchResult)

	for _, results := range resultsList {
		for _, result := range results {
			key := fmt.Sprintf("%s_%d", result.Chunk.ID, result.Chunk.ChunkIndex)
			if existing, exists := resultMap[key]; exists {
				// Manter o score mais alto
				if result.Score > existing.Score {
					resultMap[key] = result
				}
			} else {
				resultMap[key] = result
			}
		}
	}

	// Converter de volta para slice
	var merged []SearchResult
	for _, result := range resultMap {
		merged = append(merged, result)
	}

	return RankResults(merged)
}
