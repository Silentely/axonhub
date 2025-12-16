package objects

// EmbeddingRequest represents an OpenAI-compatible embedding request payload.
type EmbeddingRequest struct {
	// Input text to embed. Can be a string or an array of strings/tokens according to OpenAI spec.
	Input interface{} `json:"input"`
	// ID of the model to use.
	Model string `json:"model"`
	// Optional format of the returned embeddings: "float" (default) or "base64".
	EncodingFormat string `json:"encoding_format,omitempty"`
	// Optional number of dimensions for the resulting embeddings.
	Dimensions *int `json:"dimensions,omitempty"`
	// Optional unique identifier representing the end user.
	User string `json:"user,omitempty"`
}

// EmbeddingResponse represents an OpenAI-compatible embedding response payload.
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding is a single embedding vector entry in the response.
type Embedding struct {
	Object    string      `json:"object"`
	Embedding interface{} `json:"embedding"`
	Index     int         `json:"index"`
}

// Usage represents prompt/total token usage for embeddings.
type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
