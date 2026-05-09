package answering

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

type Service interface {
	Answer(ctx context.Context, question string, ret *retriever.Retriever) (*Response, error)
}

type QAService struct {
	client         *openai.Client
	model          string
	promptTemplate *PromptTemplate
	temperature    float32
	topK           int
}

type Response struct {
	Answer  string
	Sources []SourceReference
	Model   string
}

type SourceReference struct {
	DocumentID string  `json:"document_id"`
	ChunkIndex int     `json:"chunk_index"`
	Score      float64 `json:"score"`
	Preview    string  `json:"preview"`
}

func NewQAService(apiKey string) *QAService {
	return &QAService{
		client:         openai.NewClient(apiKey),
		model:          openai.GPT4o,
		promptTemplate: DefaultPromptTemplate(),
		temperature:    0.3,
		topK:           5,
	}
}

func (qs *QAService) WithModel(model string) *QAService {
	qs.model = model
	return qs
}

func (qs *QAService) WithPromptTemplate(template *PromptTemplate) *QAService {
	qs.promptTemplate = template
	return qs
}

func (qs *QAService) WithTemperature(temp float32) *QAService {
	qs.temperature = temp
	return qs
}

func (qs *QAService) WithTopK(k int) *QAService {
	qs.topK = k
	return qs
}

func (qs *QAService) Answer(ctx context.Context, question string, ret *retriever.Retriever) (*Response, error) {
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

	prompt, err := BuildPrompt(qs.promptTemplate, question, context, sources)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

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

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
