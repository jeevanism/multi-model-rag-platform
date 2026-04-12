package httpapi

type ChatRequest struct {
	Message  string  `json:"message"`
	Provider string  `json:"provider"`
	Model    *string `json:"model"`
	RAG      bool    `json:"rag"`
	TopK     int     `json:"top_k"`
	Debug    bool    `json:"debug"`
}

type RetrievedChunkPreview struct {
	DocumentID int     `json:"document_id"`
	ChunkID    int     `json:"chunk_id"`
	ChunkIndex int     `json:"chunk_index"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
}

type ChatResponse struct {
	Answer          string                  `json:"answer"`
	Provider        string                  `json:"provider"`
	Model           string                  `json:"model"`
	LatencyMS       int                     `json:"latency_ms"`
	TokensIn        *int                    `json:"tokens_in"`
	TokensOut       *int                    `json:"tokens_out"`
	CostUSD         *float64                `json:"cost_usd"`
	Citations       []string                `json:"citations"`
	RAGUsed         bool                    `json:"rag_used"`
	RetrievedChunks []RetrievedChunkPreview `json:"retrieved_chunks"`
}

type IngestTextRequest struct {
	Title        string `json:"title"`
	Content      string `json:"content"`
	ChunkSize    int    `json:"chunk_size"`
	ChunkOverlap int    `json:"chunk_overlap"`
}

type IngestTextResponse struct {
	DocumentID        int    `json:"document_id"`
	ChunkCount        int    `json:"chunk_count"`
	EmbeddingCount    int    `json:"embedding_count"`
	EmbeddingProvider string `json:"embedding_provider"`
	EmbeddingModel    string `json:"embedding_model"`
}

type EvalRunSummaryItem struct {
	ID           int      `json:"id"`
	DatasetName  string   `json:"dataset_name"`
	Provider     string   `json:"provider"`
	Model        *string  `json:"model"`
	TotalCases   int      `json:"total_cases"`
	PassedCases  int      `json:"passed_cases"`
	AvgLatencyMS *float64 `json:"avg_latency_ms"`
	CreatedAt    string   `json:"created_at"`
}

type EvalRunCaseItem struct {
	ID                 int      `json:"id"`
	CaseID             string   `json:"case_id"`
	Question           string   `json:"question"`
	Passed             bool     `json:"passed"`
	LatencyMS          int      `json:"latency_ms"`
	CorrectnessScore   *float64 `json:"correctness_score"`
	GroundednessScore  *float64 `json:"groundedness_score"`
	HallucinationScore *float64 `json:"hallucination_score"`
	Citations          []string `json:"citations"`
	Error              *string  `json:"error"`
}

type EvalRunDetail struct {
	Run   EvalRunSummaryItem `json:"run"`
	Cases []EvalRunCaseItem  `json:"cases"`
}

type DemoUnlockRequest struct {
	Password string `json:"password"`
}

type DemoUnlockStatusResponse struct {
	Unlocked      bool `json:"unlocked"`
	UnlockEnabled bool `json:"unlock_enabled"`
}
