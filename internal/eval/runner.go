package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Runner struct {
	apiURL string
	client *http.Client
}

type askRequest struct {
	Question    string       `json:"question"`
	TopK        int          `json:"top_k"`
	Preferences *Preferences `json:"preferences,omitempty"`
}

type askResponse struct {
	Answer  string       `json:"answer"`
	Sources []SourceInfo `json:"sources"`
}

type SourceInfo struct {
	DocumentID string                 `json:"document_id"`
	ChunkIndex int                    `json:"chunk_index"`
	Score      float64                `json:"score"`
	Preview    string                 `json:"preview"`
	Metadata   map[string]interface{} `json:"metadata"`
}

func NewRunner(apiURL string) *Runner {
	return &Runner{
		apiURL: strings.TrimRight(apiURL, "/"),
		client: &http.Client{Timeout: 2 * time.Minute},
	}
}

func (r *Runner) Run(ctx context.Context, dataset *Dataset) (*Report, error) {
	results := make([]QuestionResult, 0, len(dataset.Questions))
	for _, question := range dataset.Questions {
		response, err := r.ask(ctx, question)
		if err != nil {
			return nil, fmt.Errorf("question %s failed: %w", question.ID, err)
		}
		results = append(results, evaluateQuestion(question, response))
	}
	return summarize(dataset.Thresholds, results), nil
}

func (r *Runner) ask(ctx context.Context, question Question) (*askResponse, error) {
	payload, err := json.Marshal(askRequest{
		Question:    question.Question,
		TopK:        question.TopK,
		Preferences: preferencesPtr(question.Preferences),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ask request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.apiURL+"/rag/ask", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create ask request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ask endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ask endpoint returned status %d", resp.StatusCode)
	}

	var result askResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ask response: %w", err)
	}
	return &result, nil
}

func preferencesPtr(preferences Preferences) *Preferences {
	if len(preferences.Layers) == 0 &&
		len(preferences.Categories) == 0 &&
		len(preferences.Platforms) == 0 &&
		len(preferences.SourceKinds) == 0 &&
		len(preferences.SourceQuality) == 0 {
		return nil
	}
	return &preferences
}
