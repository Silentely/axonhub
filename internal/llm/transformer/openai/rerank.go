package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/httpclient"
)

// Rerank sends a rerank request to the OpenAI-compatible rerank endpoint.
func (t *OutboundTransformer) Rerank(ctx context.Context, req *objects.RerankRequest) (*objects.RerankResponse, error) {
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

	// Create HTTP client and send request
	client := httpclient.NewHttpClient()

	httpResp, err := client.Do(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send rerank request: %w", err)
	}

	// Check for HTTP error status codes
	if httpResp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(httpResp.Body))
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
