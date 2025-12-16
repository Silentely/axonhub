package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/llm"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/httpclient"
)

func TestEmbeddingInboundTransformer_TransformRequest(t *testing.T) {
	transformer := NewEmbeddingInboundTransformer()

	t.Run("valid string input", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "text-embedding-ada-002",
			"input": "The quick brown fox",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		httpReq := &httpclient.Request{
			Body: body,
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}

		llmReq, err := transformer.TransformRequest(context.Background(), httpReq)
		require.NoError(t, err)
		require.NotNil(t, llmReq)
		require.Equal(t, "text-embedding-ada-002", llmReq.Model)
		require.Equal(t, llm.APIFormatOpenAIEmbedding, llmReq.RawAPIFormat)
		require.Nil(t, llmReq.Stream)
		require.NotEmpty(t, llmReq.ExtraBody)
	})

	t.Run("valid array input", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "text-embedding-ada-002",
			"input": []string{"Hello", "World"},
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		httpReq := &httpclient.Request{
			Body: body,
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}

		llmReq, err := transformer.TransformRequest(context.Background(), httpReq)
		require.NoError(t, err)
		require.NotNil(t, llmReq)
	})

	t.Run("missing model", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"input": "test",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		httpReq := &httpclient.Request{
			Body: body,
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}

		_, err = transformer.TransformRequest(context.Background(), httpReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "model is required")
	})

	t.Run("missing input", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "text-embedding-ada-002",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		httpReq := &httpclient.Request{
			Body: body,
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}

		_, err = transformer.TransformRequest(context.Background(), httpReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "input is required")
	})
}

func TestEmbeddingOutboundTransformer_TransformRequest(t *testing.T) {
	transformer, err := NewEmbeddingOutboundTransformer("https://api.openai.com/v1", "test-key")
	require.NoError(t, err)

	t.Run("valid request", func(t *testing.T) {
		embReq := objects.EmbeddingRequest{
			Model: "text-embedding-ada-002",
			Input: "Hello world",
		}
		extraBody, err := json.Marshal(embReq)
		require.NoError(t, err)

		llmReq := &llm.Request{
			Model:     "text-embedding-ada-002",
			ExtraBody: extraBody,
		}

		httpReq, err := transformer.TransformRequest(context.Background(), llmReq)
		require.NoError(t, err)
		require.NotNil(t, httpReq)
		require.Equal(t, http.MethodPost, httpReq.Method)
		require.Equal(t, "https://api.openai.com/v1/embeddings", httpReq.URL)
		require.Equal(t, "application/json", httpReq.Headers.Get("Content-Type"))
		require.NotNil(t, httpReq.Auth)
		require.Equal(t, "bearer", httpReq.Auth.Type)
	})
}

func TestEmbeddingOutboundTransformer_TransformResponse(t *testing.T) {
	transformer, err := NewEmbeddingOutboundTransformer("https://api.openai.com/v1", "test-key")
	require.NoError(t, err)

	t.Run("valid response", func(t *testing.T) {
		embResp := objects.EmbeddingResponse{
			Object: "list",
			Model:  "text-embedding-ada-002",
			Data: []objects.Embedding{
				{
					Object:    "embedding",
					Index:     0,
					Embedding: []float64{0.1, 0.2, 0.3},
				},
			},
			Usage: objects.Usage{
				PromptTokens: 5,
				TotalTokens:  5,
			},
		}

		respBody, err := json.Marshal(embResp)
		require.NoError(t, err)

		httpResp := &httpclient.Response{
			StatusCode: http.StatusOK,
			Body:       respBody,
		}

		llmResp, err := transformer.TransformResponse(context.Background(), httpResp)
		require.NoError(t, err)
		require.NotNil(t, llmResp)
		require.Equal(t, "list", llmResp.Object)
		require.Equal(t, "text-embedding-ada-002", llmResp.Model)
		require.NotNil(t, llmResp.Usage)
		require.Equal(t, int64(5), llmResp.Usage.PromptTokens)
		require.Equal(t, int64(0), llmResp.Usage.CompletionTokens)
		require.Equal(t, int64(5), llmResp.Usage.TotalTokens)
		require.NotNil(t, llmResp.ProviderData)
	})
}

func TestEmbeddingInboundTransformer_TransformResponse(t *testing.T) {
	transformer := NewEmbeddingInboundTransformer()

	t.Run("valid response with provider data", func(t *testing.T) {
		embResp := objects.EmbeddingResponse{
			Object: "list",
			Model:  "text-embedding-ada-002",
			Data: []objects.Embedding{
				{
					Object:    "embedding",
					Index:     0,
					Embedding: []float64{0.1, 0.2, 0.3},
				},
			},
			Usage: objects.Usage{
				PromptTokens: 5,
				TotalTokens:  5,
			},
		}

		llmResp := &llm.Response{
			Object:       "list",
			Model:        "text-embedding-ada-002",
			ProviderData: embResp,
		}

		httpResp, err := transformer.TransformResponse(context.Background(), llmResp)
		require.NoError(t, err)
		require.NotNil(t, httpResp)
		require.Equal(t, http.StatusOK, httpResp.StatusCode)
		require.Equal(t, "application/json", httpResp.Headers.Get("Content-Type"))

		var returnedEmbResp objects.EmbeddingResponse
		err = json.Unmarshal(httpResp.Body, &returnedEmbResp)
		require.NoError(t, err)
		require.Equal(t, "list", returnedEmbResp.Object)
		require.Equal(t, "text-embedding-ada-002", returnedEmbResp.Model)
		require.Len(t, returnedEmbResp.Data, 1)
	})

	t.Run("missing provider data", func(t *testing.T) {
		llmResp := &llm.Response{
			Object: "list",
			Model:  "text-embedding-ada-002",
		}

		_, err := transformer.TransformResponse(context.Background(), llmResp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing provider data")
	})
}

func TestEmbeddingTransformers_APIFormat(t *testing.T) {
	inbound := NewEmbeddingInboundTransformer()
	require.Equal(t, llm.APIFormatOpenAIEmbedding, inbound.APIFormat())

	outbound, err := NewEmbeddingOutboundTransformer("https://api.openai.com/v1", "test-key")
	require.NoError(t, err)
	require.Equal(t, llm.APIFormatOpenAIEmbedding, outbound.APIFormat())
}
