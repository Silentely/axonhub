package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
)

type RerankHandlersParams struct {
	fx.In

	RerankService *biz.RerankService
}

func NewRerankHandlers(params RerankHandlersParams) *RerankHandlers {
	return &RerankHandlers{
		rerankService: params.RerankService,
	}
}

type RerankHandlers struct {
	rerankService *biz.RerankService
}

// Rerank handles rerank requests.
func (h *RerankHandlers) Rerank(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse request body
	var req objects.RerankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn(ctx, "invalid rerank request", log.Cause(err))
		JSONError(c, http.StatusBadRequest, err)

		return
	}

	// Call business logic
	resp, err := h.rerankService.Rerank(ctx, &req)
	if err != nil {
		log.Error(ctx, "rerank request failed", log.Cause(err))
		JSONError(c, http.StatusInternalServerError, err)

		return
	}

	// Return successful response
	c.JSON(http.StatusOK, resp)
}
