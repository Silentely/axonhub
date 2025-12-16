package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/looplj/axonhub/internal/llm"
	"github.com/looplj/axonhub/internal/llm/transformer"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/httpclient"
	"github.com/looplj/axonhub/internal/pkg/streams"
)

// EmbeddingInboundTransformer implements transformer.Inbound for the OpenAI embeddings endpoint.
type EmbeddingInboundTransformer struct{}

// NewEmbeddingInboundTransformer creates a new EmbeddingInboundTransformer.
func NewEmbeddingInboundTransformer() *EmbeddingInboundTransformer {
	return &EmbeddingInboundTransformer{}
}

func (t *EmbeddingInboundTransformer) APIFormat() llm.APIFormat {
	return llm.APIFormatOpenAIEmbedding
}

// TransformRequest transforms an HTTP embedding request into the unified llm.Request format.
// Since embeddings don't use messages, we store the input as JSON in ExtraBody.
func (t *EmbeddingInboundTransformer) TransformRequest(
	ctx context.Context,
	httpReq *httpclient.Request,
) (*llm.Request, error) {
	if httpReq == nil {
		return nil, fmt.Errorf("%w: http request is nil", transformer.ErrInvalidRequest)
	}

	if len(httpReq.Body) == 0 {
		return nil, fmt.Errorf("%w: request body is empty", transformer.ErrInvalidRequest)
	}

	// Check content type
	contentType := httpReq.Headers.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return nil, fmt.Errorf("%w: unsupported content type: %s", transformer.ErrInvalidRequest, contentType)
	}

	var embReq objects.EmbeddingRequest

	err := json.Unmarshal(httpReq.Body, &embReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode embedding request: %w", transformer.ErrInvalidRequest, err)
	}

	// Validate required fields
	if embReq.Model == "" {
		return nil, fmt.Errorf("%w: model is required", transformer.ErrInvalidRequest)
	}

	if embReq.Input == nil {
		return nil, fmt.Errorf("%w: input is required", transformer.ErrInvalidRequest)
	}

	// Build unified request
	// Embeddings don't use chat messages, so store embedding params in ExtraBody
	extraBody, err := json.Marshal(embReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request to ExtraBody: %w", err)
	}

	llmReq := &llm.Request{
		Model:        embReq.Model,
		Messages:     []llm.Message{}, // Embeddings don't use messages
		RawRequest:   httpReq,
		RawAPIFormat: llm.APIFormatOpenAIEmbedding,
		ExtraBody:    extraBody,
		Stream:       nil, // Embeddings don't stream
	}

	if embReq.User != "" {
		llmReq.User = &embReq.User
	}

	return llmReq, nil
}

// TransformResponse transforms the unified llm.Response back to HTTP response.
func (t *EmbeddingInboundTransformer) TransformResponse(
	ctx context.Context,
	llmResp *llm.Response,
) (*httpclient.Response, error) {
	if llmResp == nil {
		return nil, fmt.Errorf("embedding response is nil")
	}

	// Extract the embedding response from ProviderData
	var body []byte
	if llmResp.ProviderData != nil {
		var embResp objects.EmbeddingResponse
		switch v := llmResp.ProviderData.(type) {
		case objects.EmbeddingResponse:
			embResp = v
		case *objects.EmbeddingResponse:
			if v == nil {
				return nil, fmt.Errorf("embedding response provider data is nil")
			}
			embResp = *v
		default:
			return nil, fmt.Errorf("invalid provider data for embedding response")
		}

		var err error
		body, err = json.Marshal(embResp)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embedding response: %w", err)
		}
	} else {
		return nil, fmt.Errorf("embedding response missing provider data")
	}

	return &httpclient.Response{
		StatusCode: http.StatusOK,
		Body:       body,
		Headers: http.Header{
			"Content-Type":  []string{"application/json"},
			"Cache-Control": []string{"no-cache"},
		},
	}, nil
}

// TransformStream is not supported for embeddings.
func (t *EmbeddingInboundTransformer) TransformStream(
	ctx context.Context,
	stream streams.Stream[*llm.Response],
) (streams.Stream[*httpclient.StreamEvent], error) {
	return nil, fmt.Errorf("%w: embeddings do not support streaming", transformer.ErrInvalidRequest)
}

// AggregateStreamChunks is not supported for embeddings.
func (t *EmbeddingInboundTransformer) AggregateStreamChunks(
	ctx context.Context,
	chunks []*httpclient.StreamEvent,
) ([]byte, llm.ResponseMeta, error) {
	return nil, llm.ResponseMeta{}, fmt.Errorf("embeddings do not support streaming")
}

// TransformError re-uses standard OpenAI error formatting.
func (t *EmbeddingInboundTransformer) TransformError(ctx context.Context, rawErr error) *httpclient.Error {
	// Delegate to the standard chat inbound transformer for consistent error handling
	chatInbound := NewInboundTransformer()
	return chatInbound.TransformError(ctx, rawErr)
}
