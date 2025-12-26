package biz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/ent/enttest"
	"github.com/looplj/axonhub/internal/ent/privacy"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/httpclient"
)

// TestModelFetcher_CustomModelsEndpoint verifies that when a custom models endpoint is provided,
// it takes precedence over the default endpoint construction.
func TestModelFetcher_CustomModelsEndpoint(t *testing.T) {
	// Create a mock server that simulates an OpenAI-compatible models endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request hits our custom endpoint
		require.Equal(t, "/custom/path/models", r.URL.Path)
		require.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"gpt-4"},{"id":"gpt-3.5-turbo"}]}`))
	}))
	defer mockServer.Close()

	// Create test database and channel service
	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	channelService := &ChannelService{
		AbstractService: &AbstractService{
			db: client,
		},
	}
	httpClient := httpclient.NewHttpClient()
	modelFetcher := NewModelFetcher(httpClient, channelService)

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Test 1: Custom endpoint in input takes precedence
	t.Run("CustomEndpointInInput", func(t *testing.T) {
		customEndpoint := mockServer.URL + "/custom/path/models"
		result, err := modelFetcher.FetchModels(ctx, FetchModelsInput{
			ChannelType:          channel.TypeOpenai.String(),
			BaseURL:              "https://api.openai.com", // This should be ignored
			APIKey:               lo.ToPtr("test-api-key"),
			CustomModelsEndpoint: &customEndpoint,
		})

		require.NoError(t, err)
		require.Nil(t, result.Error)
		require.Len(t, result.Models, 2)
		require.Equal(t, "gpt-4", result.Models[0].ID)
		require.Equal(t, "gpt-3.5-turbo", result.Models[1].ID)
	})

	// Test 2: Custom endpoint from channel entity
	t.Run("CustomEndpointFromChannel", func(t *testing.T) {
		// Create a channel with custom models endpoint
		customEndpoint := mockServer.URL + "/custom/path/models"
		ch, err := client.Channel.Create().
			SetType(channel.TypeOpenai).
			SetName("Test Channel").
			SetBaseURL("https://api.openai.com").
			SetCredentials(&objects.ChannelCredentials{
				APIKey: "test-api-key",
			}).
			SetDefaultTestModel("gpt-4").
			SetSupportedModels([]string{"gpt-4"}).
			SetCustomModelsEndpoint(customEndpoint).
			Save(ctx)
		require.NoError(t, err)

		result, err := modelFetcher.FetchModels(ctx, FetchModelsInput{
			ChannelType: channel.TypeOpenai.String(),
			BaseURL:     "https://api.openai.com", // This should be ignored
			ChannelID:   lo.ToPtr(ch.ID),
		})

		require.NoError(t, err)
		require.Nil(t, result.Error)
		require.Len(t, result.Models, 2)
		require.Equal(t, "gpt-4", result.Models[0].ID)
		require.Equal(t, "gpt-3.5-turbo", result.Models[1].ID)
	})

	// Test 3: Input custom endpoint takes precedence over channel's custom endpoint
	t.Run("InputEndpointOverridesChannelEndpoint", func(t *testing.T) {
		// Create a channel with a different custom models endpoint
		ch, err := client.Channel.Create().
			SetType(channel.TypeOpenai).
			SetName("Test Channel 2").
			SetBaseURL("https://api.openai.com").
			SetCredentials(&objects.ChannelCredentials{
				APIKey: "test-api-key",
			}).
			SetDefaultTestModel("gpt-4").
			SetSupportedModels([]string{"gpt-4"}).
			SetCustomModelsEndpoint("https://wrong.endpoint.com/models"). // This should be ignored
			Save(ctx)
		require.NoError(t, err)

		customEndpoint := mockServer.URL + "/custom/path/models"
		result, err := modelFetcher.FetchModels(ctx, FetchModelsInput{
			ChannelType:          channel.TypeOpenai.String(),
			BaseURL:              "https://api.openai.com",
			ChannelID:            lo.ToPtr(ch.ID),
			CustomModelsEndpoint: &customEndpoint, // This should be used
		})

		require.NoError(t, err)
		require.Nil(t, result.Error)
		require.Len(t, result.Models, 2)
	})
}

// TestModelFetcher_CloudflareGateway tests the Cloudflare AI Gateway scenario.
func TestModelFetcher_CloudflareGateway(t *testing.T) {
	// Simulate Cloudflare AI Gateway's URL structure
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Cloudflare gateway URL: /v1/{account_id}/{gateway_id}/openai/v1/models
		require.Contains(t, r.URL.Path, "/abc123/my-gateway/openai/v1/models")
		require.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"gpt-4"},{"id":"gpt-3.5-turbo"}]}`))
	}))
	defer mockServer.Close()

	client := enttest.NewEntClient(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	channelService := &ChannelService{
		AbstractService: &AbstractService{
			db: client,
		},
	}
	httpClient := httpclient.NewHttpClient()
	modelFetcher := NewModelFetcher(httpClient, channelService)

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Test Cloudflare Gateway with custom endpoint
	customEndpoint := mockServer.URL + "/v1/abc123/my-gateway/openai/v1/models"
	result, err := modelFetcher.FetchModels(ctx, FetchModelsInput{
		ChannelType:          channel.TypeOpenai.String(),
		BaseURL:              mockServer.URL + "/v1/abc123/my-gateway/openai",
		APIKey:               lo.ToPtr("test-api-key"),
		CustomModelsEndpoint: &customEndpoint,
	})

	require.NoError(t, err)
	require.Nil(t, result.Error)
	require.Len(t, result.Models, 2)
	require.Equal(t, "gpt-4", result.Models[0].ID)
	require.Equal(t, "gpt-3.5-turbo", result.Models[1].ID)
}
