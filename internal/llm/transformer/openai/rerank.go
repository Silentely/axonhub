package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/httpclient"
)

// RerankError represents an error response from the rerank API.
type RerankError struct {
	StatusCode int
	Message    string
}

func (e *RerankError) Error() string {
	return fmt.Sprintf("rerank error (status %d): %s", e.StatusCode, e.Message)
}

// Rerank sends a rerank request to the OpenAI-compatible rerank endpoint.
// The httpClient parameter allows using a custom HTTP client with proxy/timeout configuration.
// If httpClient is nil, a default client will be used.
func (t *OutboundTransformer) Rerank(ctx context.Context, req *objects.RerankRequest, httpClient *httpclient.HttpClient) (*objects.RerankResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("rerank request is nil")
	}

	// Validate required fields
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if len(req.Documents) == 0 {
		return nil, fmt.Errorf("documents are required")
	}

	// Validate top_n if provided
	if req.TopN != nil {
		if *req.TopN <= 0 {
			return nil, fmt.Errorf("top_n must be a positive integer")
		}

		if *req.TopN > len(req.Documents) {
			return nil, fmt.Errorf("top_n (%d) cannot exceed the number of documents (%d)", *req.TopN, len(req.Documents))
		}
	}

	// Validate documents are not empty strings
	for i, doc := range req.Documents {
		if doc == "" {
			return nil, fmt.Errorf("document at index %d is empty", i)
		}
	}

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rerank request: %w", err)
	}

	// Prepare headers
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	var auth *httpclient.AuthConfig

	//nolint:exhaustive // Checked.
	switch t.config.Type {
	case PlatformAzure:
		auth = &httpclient.AuthConfig{
			Type:      "api_key",
			APIKey:    t.config.APIKey,
			HeaderKey: "Api-Key",
		}
	default:
		auth = &httpclient.AuthConfig{
			Type:   "bearer",
			APIKey: t.config.APIKey,
		}
	}

	// Build rerank URL
	url, err := t.buildRerankURL()
	if err != nil {
		return nil, fmt.Errorf("failed to build rerank URL: %w", err)
	}

	// Create HTTP request
	httpReq := &httpclient.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headers,
		Body:    body,
		Auth:    auth,
	}

	// Use provided HTTP client or create a default one
	client := httpClient
	if client == nil {
		client = httpclient.NewHttpClient()
	}

	httpResp, err := client.Do(ctx, httpReq)
	if err != nil {
		// 从 httpclient.Error 提取状态码和响应体，转换为 RerankError
		var httpErr *httpclient.Error
		if errors.As(err, &httpErr) {
			return nil, &RerankError{
				StatusCode: httpErr.StatusCode,
				Message:    string(httpErr.Body),
			}
		}

		return nil, fmt.Errorf("failed to send rerank request: %w", err)
	}

	// Check for empty response body
	if len(httpResp.Body) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}

	// Unmarshal response
	var rerankResp objects.RerankResponse

	err = json.Unmarshal(httpResp.Body, &rerankResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rerank response: %w", err)
	}

	return &rerankResp, nil
}

// buildRerankURL constructs the appropriate rerank URL based on the platform.
func (t *OutboundTransformer) buildRerankURL() (string, error) {
	//nolint:exhaustive // Checked.
	switch t.config.Type {
	case PlatformAzure:
		if t.config.APIVersion == "" {
			return "", fmt.Errorf("API version is required for Azure platform")
		}
		// Azure rerank endpoint pattern
		return fmt.Sprintf("%s/rerank?api-version=%s", t.config.BaseURL, t.config.APIVersion), nil
	default:
		// Standard OpenAI-compatible API
		if len(t.config.BaseURL) > 0 && t.config.BaseURL[len(t.config.BaseURL)-1:] == "/" {
			return t.config.BaseURL + "v1/rerank", nil
		}

		if len(t.config.BaseURL) > 3 && t.config.BaseURL[len(t.config.BaseURL)-3:] == "/v1" {
			return t.config.BaseURL + "/rerank", nil
		}

		return t.config.BaseURL + "/v1/rerank", nil
	}
}
