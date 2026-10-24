package ai

// Source is a citation exposed to the frontend. IDs are assigned by the
// server (S1, S2, ...) and are the only citation handles the model may use.
type Source struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Publisher    string `json:"publisher,omitempty"`
	URL          string `json:"url,omitempty"`
	Snippet      string `json:"snippet"`
	GarmentID    int64  `json:"garment_id,omitempty"`
	AnnotationID int64  `json:"annotation_id,omitempty"`
}

// Action is a server-validated UI instruction. The frontend must never trust
// IDs coming directly from the model.
type Action struct {
	Type         string `json:"type"`
	GarmentID    int64  `json:"garment_id,omitempty"`
	AnnotationID int64  `json:"annotation_id,omitempty"`
	Label        string `json:"label"`
}

// ChatContext is the page context supplied by the client. IDs are optional and
// always re-checked against the database before they are used.
type ChatContext struct {
	Page         string `json:"page,omitempty"`
	GarmentID    int64  `json:"garment_id,omitempty"`
	AnnotationID int64  `json:"annotation_id,omitempty"`
}

type ChatTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Message   string      `json:"message"`
	History   []ChatTurn  `json:"history"`
	Context   ChatContext `json:"context"`
	SessionID string      `json:"session_id,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Status is safe to expose publicly: it never contains the API key or base URL.
type Status struct {
	Enabled        bool   `json:"enabled"`
	Configured     bool   `json:"configured"`
	MockMode       bool   `json:"mock_mode"`
	Model          string `json:"model"`
	EmbeddingModel string `json:"embedding_model,omitempty"`
	Chunks         int    `json:"chunks"`
	Message        string `json:"message"`
}

type KnowledgeDocument struct {
	ID         int64
	DocKey     string
	Title      string
	SourceType string
	SourceURL  string
	Publisher  string
	License    string
	Content    string
	Checksum   string
	Status     string

	GarmentID    int64
	AnnotationID int64
	Ethnic       string
	Province     string
	Craft        string
}

type KnowledgeChunk struct {
	ID             int64
	DocumentID     int64
	ChunkIndex     int
	Content        string
	TokenCount     int
	Embedding      []float32
	EmbeddingModel string
	MetadataJSON   string
}

type RetrievedChunk struct {
	Chunk
	Title      string
	SourceType string
	SourceURL  string
	Publisher  string
	License    string
	Score      float64
	BM25       float64
	Vector     float64
}

type Chunk struct {
	ID             int64
	DocumentID     int64
	ChunkIndex     int
	Content        string
	TokenCount     int
	Embedding      []float32
	EmbeddingModel string
	MetadataJSON   string
}

type ToolEvent struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Elapsed int64  `json:"elapsed_ms"`
}

type UsageRecord struct {
	SessionHash      string
	Endpoint         string
	Model            string
	PromptTokens     int
	CompletionTokens int
	LatencyMS        int64
	ToolCalls        int
	Status           string
}

const retrievalVersion = "hybrid-rrf-v1"
