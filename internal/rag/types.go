package rag

type RetrievedChunk struct {
	DocumentID int
	ChunkID    int
	ChunkIndex int
	Title      string
	Content    string
	Score      float64
}

type EmbeddingResult struct {
	Vector   []float64
	Provider string
	Model    string
}
