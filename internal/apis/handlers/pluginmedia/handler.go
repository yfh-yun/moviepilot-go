package pluginmedia

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pluginmediabiz "moviepilot-go/internal/business/services/pluginmedia"
	"moviepilot-go/pkg/logger"
)

type Handler struct {
	service pluginmediabiz.Service
	log     *zap.Logger
}

func NewHandler(service pluginmediabiz.Service) *Handler {
	return &Handler{
		service: service,
		log:     logger.GetLogger().With(zap.String("handler", "pluginmedia")),
	}
}

func (h *Handler) SearchTorrents(c *gin.Context) {
	var req pluginmediabiz.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("bind search request failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.service.SearchTorrents(c.Request.Context(), req)
	if err != nil {
		h.log.Error("search torrents via plugin failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *Handler) RecognizeMedia(c *gin.Context) {
	var req pluginmediabiz.RecognizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("bind recognize request failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.RecognizeMedia(c.Request.Context(), req)
	if err != nil {
		h.log.Error("recognize media via plugin failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
