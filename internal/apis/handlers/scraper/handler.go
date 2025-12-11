package scraper

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	scraperbiz "moviepilot-go/internal/business/services/scraper"
	"moviepilot-go/pkg/logger"
)

// Handler 刮削 API 处理器
type Handler struct {
	service scraperbiz.Service
	logger  *zap.Logger
}

// NewHandler 创建刮削 API 处理器
func NewHandler(service scraperbiz.Service) *Handler {
	return &Handler{
		service: service,
		logger:  logger.GetLogger(),
	}
}

// ScrapeMovie 刮削电影
// @Summary 刮削电影
// @Description 刮削电影元数据
// @Tags scraper
// @Accept json
// @Produce json
// @Param request body ScrapeMovieRequest true "刮削请求"
// @Success 200 {object} scraper.MovieMetadata
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/scraper/movie [post]
func (h *Handler) ScrapeMovie(c *gin.Context) {
	var req ScrapeMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opts := scraperbiz.ScrapeOptions{
		UseCache:       req.UseCache,
		DownloadImages: req.DownloadImages,
	}

	metadata, err := h.service.ScrapeMovie(c.Request.Context(), req.Title, req.Year, opts)
	if err != nil {
		h.logger.Error("刮削电影失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// ScrapeTV 刮削电视剧
// @Summary 刮削电视剧
// @Description 刮削电视剧元数据
// @Tags scraper
// @Accept json
// @Produce json
// @Param request body ScrapeTVRequest true "刮削请求"
// @Success 200 {object} scraper.TVMetadata
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/scraper/tv [post]
func (h *Handler) ScrapeTV(c *gin.Context) {
	var req ScrapeTVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opts := scraperbiz.ScrapeOptions{
		UseCache:       req.UseCache,
		DownloadImages: req.DownloadImages,
	}

	metadata, err := h.service.ScrapeTV(c.Request.Context(), req.Title, req.Year, opts)
	if err != nil {
		h.logger.Error("刮削电视剧失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// ScrapeSeason 刮削季
// @Summary 刮削季
// @Description 刮削季元数据
// @Tags scraper
// @Accept json
// @Produce json
// @Param request body ScrapeSeasonRequest true "刮削请求"
// @Success 200 {object} scraper.SeasonMetadata
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/scraper/season [post]
func (h *Handler) ScrapeSeason(c *gin.Context) {
	var req ScrapeSeasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opts := scraperbiz.ScrapeOptions{
		UseCache:       req.UseCache,
		DownloadImages: req.DownloadImages,
	}

	metadata, err := h.service.ScrapeSeason(c.Request.Context(), req.TVID, req.SeasonNumber, opts)
	if err != nil {
		h.logger.Error("刮削季失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// ScrapeEpisode 刮削集
// @Summary 刮削集
// @Description 刮削集元数据
// @Tags scraper
// @Accept json
// @Produce json
// @Param request body ScrapeEpisodeRequest true "刮削请求"
// @Success 200 {object} scraper.EpisodeMetadata
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/scraper/episode [post]
func (h *Handler) ScrapeEpisode(c *gin.Context) {
	var req ScrapeEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opts := scraperbiz.ScrapeOptions{
		UseCache:       req.UseCache,
		DownloadImages: req.DownloadImages,
	}

	metadata, err := h.service.ScrapeEpisode(c.Request.Context(), req.TVID, req.SeasonNumber, req.EpisodeNumber, opts)
	if err != nil {
		h.logger.Error("刮削集失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// BatchScrapeMovies 批量刮削电影
// @Summary 批量刮削电影
// @Description 批量刮削电影元数据
// @Tags scraper
// @Accept json
// @Produce json
// @Param request body BatchScrapeRequest true "批量刮削请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/scraper/batch/movies [post]
func (h *Handler) BatchScrapeMovies(c *gin.Context) {
	var req BatchScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换请求
	requests := make([]scraperbiz.ScrapeRequest, len(req.Items))
	for i, item := range req.Items {
		requests[i] = scraperbiz.ScrapeRequest{
			Title: item.Title,
			Year:  item.Year,
		}
	}

	results, err := h.service.BatchScrapeMovies(c.Request.Context(), requests)
	if err != nil {
		h.logger.Error("批量刮削电影失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 统计结果
	successCount := 0
	failCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			failCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results":       results,
		"total":         len(results),
		"success_count": successCount,
		"fail_count":    failCount,
	})
}

// BatchScrapeTVs 批量刮削电视剧
// @Summary 批量刮削电视剧
// @Description 批量刮削电视剧元数据
// @Tags scraper
// @Accept json
// @Produce json
// @Param request body BatchScrapeRequest true "批量刮削请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/scraper/batch/tvs [post]
func (h *Handler) BatchScrapeTVs(c *gin.Context) {
	var req BatchScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换请求
	requests := make([]scraperbiz.ScrapeRequest, len(req.Items))
	for i, item := range req.Items {
		requests[i] = scraperbiz.ScrapeRequest{
			Title: item.Title,
			Year:  item.Year,
		}
	}

	results, err := h.service.BatchScrapeTVs(c.Request.Context(), requests)
	if err != nil {
		h.logger.Error("批量刮削电视剧失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 统计结果
	successCount := 0
	failCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		} else {
			failCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results":       results,
		"total":         len(results),
		"success_count": successCount,
		"fail_count":    failCount,
	})
}

// ScrapeMovieRequest 刮削电影请求
type ScrapeMovieRequest struct {
	Title          string `json:"title" binding:"required"`
	Year           int    `json:"year"`
	UseCache       bool   `json:"use_cache"`
	DownloadImages bool   `json:"download_images"`
}

// ScrapeTVRequest 刮削电视剧请求
type ScrapeTVRequest struct {
	Title          string `json:"title" binding:"required"`
	Year           int    `json:"year"`
	UseCache       bool   `json:"use_cache"`
	DownloadImages bool   `json:"download_images"`
}

// ScrapeSeasonRequest 刮削季请求
type ScrapeSeasonRequest struct {
	TVID           int  `json:"tv_id" binding:"required"`
	SeasonNumber   int  `json:"season_number" binding:"required"`
	UseCache       bool `json:"use_cache"`
	DownloadImages bool `json:"download_images"`
}

// ScrapeEpisodeRequest 刮削集请求
type ScrapeEpisodeRequest struct {
	TVID           int  `json:"tv_id" binding:"required"`
	SeasonNumber   int  `json:"season_number" binding:"required"`
	EpisodeNumber  int  `json:"episode_number" binding:"required"`
	UseCache       bool `json:"use_cache"`
	DownloadImages bool `json:"download_images"`
}

// BatchScrapeRequest 批量刮削请求
type BatchScrapeRequest struct {
	Items []BatchScrapeItem `json:"items" binding:"required"`
}

// BatchScrapeItem 批量刮削项
type BatchScrapeItem struct {
	Title string `json:"title" binding:"required"`
	Year  int    `json:"year"`
}
