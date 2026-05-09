package api

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

type SeedDemoResponse struct {
	Seeded int    `json:"seeded"`
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}
