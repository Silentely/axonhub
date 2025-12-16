package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/llm"
	"github.com/looplj/axonhub/internal/llm/transformer"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/httpclient"
	"github.com/looplj/axonhub/internal/pkg/xcontext"
)

type RerankServiceParams struct {
	fx.In

	ChannelService *ChannelService
	RequestService *RequestService
	SystemService  *SystemService
}

func NewRerankService(params RerankServiceParams) *RerankService {
	return &RerankService{
		channelService: params.ChannelService,
		requestService: params.RequestService,
		systemService:  params.SystemService,
	}
}

type RerankService struct {
	channelService *ChannelService
	requestService *RequestService
	systemService  *SystemService
}

// Rerank performs document reranking using the specified model and channel.
func (s *RerankService) Rerank(ctx context.Context, req *objects.RerankRequest) (*objects.RerankResponse, error) {
	startTime := time.Now()

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

	// Create request record for logging
	reqBody := mustMarshalJSON(req)
	llmRequest := &llm.Request{
		Model:  req.Model,
		Stream: boolPtr(false),
	}
	httpRequest := &httpclient.Request{
		Body: reqBody,
	}

	requestRecord, err := s.requestService.CreateRequest(ctx, llmRequest, httpRequest, llm.APIFormatOpenAIRerank)
	if err != nil {
		log.Warn(ctx, "failed to create request record", log.Cause(err))
		// 继续执行，不因为日志记录失败而中断请求
	}

	// 更新请求的 channel_id
	if requestRecord != nil {
		if err := s.requestService.UpdateRequestChannelID(ctx, requestRecord.ID, selectedChannel.ID); err != nil {
			log.Warn(ctx, "failed to update request channel_id", log.Cause(err))
		}
	}

	// Create execution record
	var executionRecord *ent.RequestExecution
	if requestRecord != nil {
		channelRequest := httpclient.Request{
			Body: reqBody,
		}
		executionRecord, err = s.requestService.CreateRequestExecution(
			ctx,
			selectedChannel,
			req.Model,
			requestRecord,
			channelRequest,
			llm.APIFormatOpenAIRerank,
		)
		if err != nil {
			log.Warn(ctx, "failed to create request execution record", log.Cause(err))
		}
	}

	// Call the transformer's Rerank method
	resp, err := rerankTransformer.Rerank(ctx, req)
	if err != nil {
		log.Error(ctx, "rerank request failed",
			log.String("model", req.Model),
			log.Int("channel_id", selectedChannel.ID),
			log.Cause(err))

		// 更新请求和执行状态为失败
		persistCtx, cancel := xcontext.DetachWithTimeout(ctx, time.Second*5)
		defer cancel()

		if executionRecord != nil {
			if updateErr := s.requestService.UpdateRequestExecutionStatusFromError(persistCtx, executionRecord.ID, err); updateErr != nil {
				log.Warn(persistCtx, "failed to update execution status", log.Cause(updateErr))
			}
		}

		if requestRecord != nil {
			if updateErr := s.requestService.UpdateRequestStatusFromError(persistCtx, requestRecord.ID, err); updateErr != nil {
				log.Warn(persistCtx, "failed to update request status", log.Cause(updateErr))
			}
		}

		return nil, fmt.Errorf("rerank request failed: %w", err)
	}

	// 更新请求和执行状态为完成
	persistCtx, cancel := xcontext.DetachWithTimeout(ctx, time.Second*5)
	defer cancel()

	if executionRecord != nil {
		if updateErr := s.requestService.UpdateRequestExecutionCompleted(persistCtx, executionRecord.ID, "", resp); updateErr != nil {
			log.Warn(persistCtx, "failed to update execution completed", log.Cause(updateErr))
		}
	}

	if requestRecord != nil {
		if updateErr := s.requestService.UpdateRequestCompleted(persistCtx, requestRecord.ID, "", resp); updateErr != nil {
			log.Warn(persistCtx, "failed to update request completed", log.Cause(updateErr))
		}
	}

	// Log successful rerank with latency
	latency := time.Since(startTime)
	log.Info(ctx, "rerank request completed",
		log.String("model", req.Model),
		log.Int("num_documents", len(req.Documents)),
		log.Int("num_results", len(resp.Results)),
		log.Duration("latency", latency))

	return resp, nil
}

// boolPtr returns a pointer to the bool value.
func boolPtr(b bool) *bool {
	return &b
}

// mustMarshalJSON marshals v to JSON bytes, returns empty JSON object on error.
func mustMarshalJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}

	return data
}
