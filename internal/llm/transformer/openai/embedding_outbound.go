package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/looplj/axonhub/internal/llm"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/httpclient"
	"github.com/looplj/axonhub/internal/pkg/streams"
)

// EmbeddingOutboundTransformer implements transformer.Outbound for OpenAI embeddings.
type EmbeddingOutboundTransformer struct {
	config *Config
}

// NewEmbeddingOutboundTransformer creates a new EmbeddingOutboundTransformer.
func NewEmbeddingOutboundTransformer(baseURL, apiKey string) (*EmbeddingOutboundTransformer, error) {
	config := &Config{
		Type:    PlatformOpenAI,
		BaseURL: baseURL,
		APIKey:  apiKey,
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &EmbeddingOutboundTransformer{
		config: config,
	}, nil
}

// NewEmbeddingOutboundTransformerWithConfig creates a transformer with the given config.
func NewEmbeddingOutboundTransformerWithConfig(config *Config) (*EmbeddingOutboundTransformer, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &EmbeddingOutboundTransformer{
		config: config,
	}, nil
}

func (t *EmbeddingOutboundTransformer) APIFormat() llm.APIFormat {
	return llm.APIFormatOpenAIEmbedding
}

// TransformRequest converts the unified llm.Request into an HTTP embedding request.
func (t *EmbeddingOutboundTransformer) TransformRequest(
	ctx context.Context,
	llmReq *llm.Request,
) (*httpclient.Request, error) {
	if llmReq == nil {
		return nil, fmt.Errorf("llm request is nil")
	}

	// Unmarshal the embedding request from ExtraBody
	var embReq objects.EmbeddingRequest
	if len(llmReq.ExtraBody) > 0 {
		err := json.Unmarshal(llmReq.ExtraBody, &embReq)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding request from ExtraBody: %w", err)
		}
	} else {
		return nil, fmt.Errorf("embedding request missing in ExtraBody")
	}

	// Re-serialize to JSON (ensures clean output)
	body, err := json.Marshal(embReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	// Prepare headers
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	// Build URL
	url := t.config.BaseURL + "/embeddings"

	// Build auth config
	var auth *httpclient.AuthConfig

	switch t.config.Type {
	case PlatformAzure:
		auth = &httpclient.AuthConfig{
			Type:      "api_key",
			APIKey:    t.config.APIKey,
			HeaderKey: "Api-Key",
		}
	case PlatformOpenAI:
		auth = &httpclient.AuthConfig{
			Type:   "bearer",
			APIKey: t.config.APIKey,
		}
	}

	return &httpclient.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headers,
		Body:    body,
		Auth:    auth,
	}, nil
}

// TransformResponse converts the HTTP embedding response into the unified llm.Response.
func (t *EmbeddingOutboundTransformer) TransformResponse(
	ctx context.Context,
	httpResp *httpclient.Response,
) (*llm.Response, error) {
	if httpResp == nil {
		return nil, fmt.Errorf("http response is nil")
	}

	// Parse the OpenAI embedding response
	var embResp objects.EmbeddingResponse
	if err := json.Unmarshal(httpResp.Body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding response: %w", err)
	}

	// Build unified response
	llmResp := &llm.Response{
		ID:           fmt.Sprintf("emb-%d", len(embResp.Data)),
		Object:       embResp.Object,
		Model:        embResp.Model,
		Choices:      nil, // Embeddings don't have choices
		ProviderData: embResp,
	}

	// Map usage
	if embResp.Usage.PromptTokens > 0 || embResp.Usage.TotalTokens > 0 {
		llmResp.Usage = &llm.Usage{
			PromptTokens:     int64(embResp.Usage.PromptTokens),
			CompletionTokens: 0, // Embeddings don't have completion tokens
			TotalTokens:      int64(embResp.Usage.TotalTokens),
		}
	}

	return llmResp, nil
}

// TransformStream is not supported for embeddings.
func (t *EmbeddingOutboundTransformer) TransformStream(
	ctx context.Context,
	stream streams.Stream[*httpclient.StreamEvent],
) (streams.Stream[*llm.Response], error) {
	return nil, fmt.Errorf("embeddings do not support streaming")
}

// AggregateStreamChunks is not supported for embeddings.
func (t *EmbeddingOutboundTransformer) AggregateStreamChunks(
	ctx context.Context,
	chunks []*httpclient.StreamEvent,
) ([]byte, llm.ResponseMeta, error) {
	return nil, llm.ResponseMeta{}, fmt.Errorf("embeddings do not support streaming")
}

// TransformError re-uses the standard OpenAI outbound error transformer.
func (t *EmbeddingOutboundTransformer) TransformError(
	ctx context.Context,
	httpErr *httpclient.Error,
) *llm.ResponseError {
	// Delegate to the standard chat outbound transformer
	chatOutbound, _ := NewOutboundTransformer(t.config.BaseURL, t.config.APIKey)
	return chatOutbound.TransformError(ctx, httpErr)
}
