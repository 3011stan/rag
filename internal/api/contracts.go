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

type TemporaryTokenRequest struct {
	Email string `json:"email"`
}

type TemporaryTokenResponse struct {
	Token     string   `json:"token"`
	ExpiresAt string   `json:"expires_at"`
	Scopes    []string `json:"scopes"`
}

type DocumentInfo struct {
	ID         string                 `json:"id"`
	Source     string                 `json:"source"`
	Title      string                 `json:"title"`
	Checksum   string                 `json:"checksum,omitempty"`
	ChunkCount int                    `json:"chunk_count"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  string                 `json:"created_at"`
}

type DocumentsResponse struct {
	Documents []DocumentInfo `json:"documents"`
}

type DeleteDocumentResponse struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
}

type DebugMetadataResponse struct {
	Environment         string `json:"environment"`
	AIProvider          string `json:"ai_provider"`
	EmbeddingModel      string `json:"embedding_model"`
	LLMModel            string `json:"llm_model"`
	EmbeddingDimensions int    `json:"embedding_dimensions"`
	TopK                int    `json:"top_k"`
	ChunkTokens         int    `json:"chunk_tokens"`
	OverlapTokens       int    `json:"overlap_tokens"`
	PublicUploadEnabled bool   `json:"public_upload_enabled"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}
