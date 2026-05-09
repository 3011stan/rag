package rag

type Chunk struct {
	ID         string
	DocumentID string
	ChunkIndex int
	Content    string
	TokenCount int
	Embedding  []float32
	Score      float64
	Metadata   map[string]interface{}
}

type Document struct {
	ID       string
	Source   string
	Title    string
	Checksum string
	Metadata map[string]interface{}
}
