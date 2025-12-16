package biz

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/llm/transformer"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
)

type RerankServiceParams struct {
	fx.In

	ChannelService *ChannelService
	SystemService  *SystemService
}

func NewRerankService(params RerankServiceParams) *RerankService {
	return &RerankService{
		channelService: params.ChannelService,
		systemService:  params.SystemService,
	}
}

type RerankService struct {
	channelService *ChannelService
	systemService  *SystemService
}

// Rerank performs document reranking using the specified model and channel.
func (s *RerankService) Rerank(ctx context.Context, req *objects.RerankRequest) (*objects.RerankResponse, error) {
	// Validate request
	if req == nil {
		return nil, fmt.Errorf("rerank request is nil")
	}

	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if len(req.Documents) == 0 {
		return nil, fmt.Errorf("documents are required")
	}

	// Find a suitable channel for the model
	channels := s.channelService.EnabledChannels
	if len(channels) == 0 {
		return nil, fmt.Errorf("no enabled channels available")
	}

	var selectedChannel *Channel

	for _, ch := range channels {
		if ch.IsModelSupported(req.Model) {
			selectedChannel = ch
			break
		}
	}

	if selectedChannel == nil {
		return nil, fmt.Errorf("no channel supports model: %s", req.Model)
	}

	// Log channel selection
	log.Info(ctx, "selected channel for rerank",
		log.String("model", req.Model),
		log.Int("channel_id", selectedChannel.ID),
		log.String("channel_name", selectedChannel.Name))

	// Check if the transformer supports Rerank
	rerankTransformer, ok := selectedChannel.Outbound.(transformer.Transformer)
	if !ok {
		return nil, fmt.Errorf("channel transformer does not support rerank operation")
	}

	// Call the transformer's Rerank method
	resp, err := rerankTransformer.Rerank(ctx, req)
	if err != nil {
		log.Error(ctx, "rerank request failed",
			log.String("model", req.Model),
			log.Int("channel_id", selectedChannel.ID),
			log.Cause(err))

		return nil, fmt.Errorf("rerank request failed: %w", err)
	}

	// Log successful rerank
	log.Info(ctx, "rerank request completed",
		log.String("model", req.Model),
		log.Int("num_documents", len(req.Documents)),
		log.Int("num_results", len(resp.Results)))

	return resp, nil
}
