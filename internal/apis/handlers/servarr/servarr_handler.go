// Package servarr ServArr API处理器
package servarr

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/internal/apis/validators"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
)

// ServArrHandler ServArr API处理器
type ServArrHandler struct {
	logger        *zap.Logger
	subscribeRepo interfaces.SubscribeRepository
	// 这里可以添加其他需要的服务，如 MediaChain 等
}

// NewServArrHandler 创建ServArr处理器实例
func NewServArrHandler(
	logger *zap.Logger,
	subscribeRepo interfaces.SubscribeRepository,
) *ServArrHandler {
	return &ServArrHandler{
		logger:        logger,
		subscribeRepo: subscribeRepo,
	}
}

// SubscribeInfoResponse 订阅信息响应结构
type SubscribeInfoResponse struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Year            *string                `json:"year"`
	Type            string                 `json:"type"`
	Keyword         string                 `json:"keyword"`
	TMDBID          *int                   `json:"tmdb_id"`
	IMDBID          *string                `json:"imdb_id"`
	TVDBID          *int                   `json:"tvdb_id"`
	DoubanID        *string                `json:"douban_id"`
	BangumiID       *int                   `json:"bangumi_id"`
	MediaID         *string                `json:"media_id"`
	Season          *int                   `json:"season"`
	Poster          string                 `json:"poster"`
	Backdrop        string                 `json:"backdrop"`
	Vote            *float64               `json:"vote"`
	Description     string                 `json:"description"`
	Filter          string                 `json:"filter"`
	Include         string                 `json:"include"`
	Exclude         string                 `json:"exclude"`
	Quality         string                 `json:"quality"`
	Resolution      string                 `json:"resolution"`
	Effect          string                 `json:"effect"`
	TotalEpisode    *int                   `json:"total_episode"`
	StartEpisode    *int                   `json:"start_episode"`
	LackEpisode     *int                   `json:"lack_episode"`
	Note            string                 `json:"note"`
	State           string                 `json:"state"`
	LastUpdate      *time.Time             `json:"last_update"`
	Username        string                 `json:"username"`
	Sites           string                 `json:"sites"`
	Downloader      string                 `json:"downloader"`
	BestVersion     int                    `json:"best_version"`
	CurrentPriority *int                   `json:"current_priority"`
	MediaCategory   string                 `json:"media_category"`
	EpisodeGroup    string                 `json:"episode_group"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

// CreateSubscribeRequest 创建订阅请求结构
type CreateSubscribeRequest struct {
	Name        string  `json:"name" binding:"required"`
	Year        *string `json:"year"`
	Type        string  `json:"type" binding:"required"`
	Keyword     string  `json:"keyword"`
	TMDBID      *int    `json:"tmdb_id"`
	IMDBID      *string `json:"imdb_id"`
	TVDBID      *int    `json:"tvdb_id"`
	DoubanID    *string `json:"douban_id"`
	BangumiID   *int    `json:"bangumi_id"`
	MediaID     *string `json:"media_id"`
	Season      *int    `json:"season"`
	Poster      string  `json:"poster"`
	Backdrop    string  `json:"backdrop"`
	Vote        *float64 `json:"vote"`
	Description string  `json:"description"`
	Filter      string  `json:"filter"`
	Include     string  `json:"include"`
	Exclude     string  `json:"exclude"`
	Quality     string  `json:"quality"`
	Resolution  string  `json:"resolution"`
	Effect      string  `json:"effect"`
	TotalEpisode *int    `json:"total_episode"`
	StartEpisode *int    `json:"start_episode"`
	Note        string  `json:"note"`
	Sites       string  `json:"sites"`
	Downloader  string  `json:"downloader"`
}

// UpdateSubscribeRequest 更新订阅请求结构
type UpdateSubscribeRequest struct {
	Name            *string  `json:"name"`
	Year            *string  `json:"year"`
	Type            *string  `json:"type"`
	Keyword         *string  `json:"keyword"`
	TMDBID          *int     `json:"tmdb_id"`
	IMDBID          *string  `json:"imdb_id"`
	TVDBID          *int     `json:"tvdb_id"`
	DoubanID        *string  `json:"douban_id"`
	BangumiID       *int     `json:"bangumi_id"`
	MediaID         *string  `json:"media_id"`
	Season          *int     `json:"season"`
	Poster          *string  `json:"poster"`
	Backdrop        *string  `json:"backdrop"`
	Vote            *float64 `json:"vote"`
	Description     *string  `json:"description"`
	Filter          *string  `json:"filter"`
	Include         *string  `json:"include"`
	Exclude         *string  `json:"exclude"`
	Quality         *string  `json:"quality"`
	Resolution      *string  `json:"resolution"`
	Effect          *string  `json:"effect"`
	TotalEpisode     *int     `json:"total_episode"`
	StartEpisode     *int     `json:"start_episode"`
	Note            *string  `json:"note"`
	State           *string  `json:"state"`
	Sites           *string  `json:"sites"`
	Downloader      *string  `json:"downloader"`
	BestVersion     *int     `json:"best_version"`
	CurrentPriority *int     `json:"current_priority"`
	MediaCategory   *string  `json:"media_category"`
	EpisodeGroup    *string  `json:"episode_group"`
}

// GetSubscribes 获取订阅列表
// @Summary 获取订阅列表
// @Description 获取订阅列表，支持按类型过滤
// @Tags 订阅
// @Produce json
// @Param type query string false "订阅类型"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} response.APIResponse{data=[]SubscribeInfoResponse}
// @Router /subscribe [get]
func (h *ServArrHandler) GetSubscribes(c *gin.Context) {
	mtype := c.Query("type")
	pageParam := c.DefaultQuery("page", "1")
	limitParam := c.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageParam)
	limit, _ := strconv.Atoi(limitParam)

	// 规范化分页参数
	page, limit = validator.NormalizePage(page, limit)
	offset := (page - 1) * limit

	// 获取订阅列表
	var subscribes []*models.Subscribe
	var total int64
	var err error

	if mtype != "" {
		subscribes, err = h.subscribeRepo.GetByState(mtype)
		if err != nil {
			h.logger.Error("获取订阅列表失败", zap.Error(err))
			response.InternalServerError(c, "获取订阅列表失败")
			return
		}
		total = int64(len(subscribes))
	} else {
		subscribes, total, err = h.subscribeRepo.List(offset, limit)
		if err != nil {
			h.logger.Error("获取订阅列表失败", zap.Error(err))
			response.InternalServerError(c, "获取订阅列表失败")
			return
		}
	}

	// 转换为响应格式
	var responseSubscribes []SubscribeInfoResponse
	for _, sub := range subscribes {
		responseSubscribes = append(responseSubscribes, SubscribeInfoResponse{
			ID:              strconv.FormatUint(uint64(sub.ID), 10),
			Name:            sub.Name,
			Year:            sub.Year,
			Type:            sub.Type,
			Keyword:         sub.Keyword,
			TMDBID:          sub.TMDBID,
			IMDBID:          sub.IMDBID,
			TVDBID:          sub.TVDBID,
			DoubanID:        sub.DoubanID,
			BangumiID:       sub.BangumiID,
			MediaID:         sub.MediaID,
			Season:          sub.Season,
			Poster:          sub.Poster,
			Backdrop:        sub.Backdrop,
			Vote:            sub.Vote,
			Description:     sub.Description,
			Filter:          sub.Filter,
			Include:         sub.Include,
			Exclude:         sub.Exclude,
			Quality:         sub.Quality,
			Resolution:      sub.Resolution,
			Effect:          sub.Effect,
			TotalEpisode:    sub.TotalEpisode,
			StartEpisode:    sub.StartEpisode,
			LackEpisode:     sub.LackEpisode,
			Note:            sub.Note,
			State:           sub.State,
			LastUpdate:      sub.LastUpdate,
			Username:        sub.Username,
			Sites:           sub.Sites,
			Downloader:      sub.Downloader,
			BestVersion:     sub.BestVersion,
			CurrentPriority: sub.CurrentPriority,
			MediaCategory:   sub.MediaCategory,
			EpisodeGroup:    sub.EpisodeGroup,
			CreatedAt:       sub.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       sub.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.SuccessWithPagination(c, responseSubscribes, page, limit, total)
}

// GetSubscribe 获取订阅详情
// @Summary 获取订阅详情
// @Description 根据ID获取订阅详细信息
// @Tags 订阅
// @Produce json
// @Param id path string true "订阅ID"
// @Success 200 {object} response.APIResponse{data=SubscribeInfoResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /subscribe/{id} [get]
func (h *ServArrHandler) GetSubscribe(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "订阅ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订阅ID格式错误")
		return
	}

	subscribe, err := h.subscribeRepo.GetByID(uint(id))
	if err != nil {
		h.logger.Error("获取订阅详情失败", zap.Error(err))
		response.InternalServerError(c, "获取订阅详情失败")
		return
	}

	// 转换为响应格式
	responseSubscribe := SubscribeInfoResponse{
		ID:              strconv.FormatUint(uint64(subscribe.ID), 10),
		Name:            subscribe.Name,
		Year:            subscribe.Year,
		Type:            subscribe.Type,
		Keyword:         subscribe.Keyword,
		TMDBID:          subscribe.TMDBID,
		IMDBID:          subscribe.IMDBID,
		TVDBID:          subscribe.TVDBID,
		DoubanID:        subscribe.DoubanID,
		BangumiID:       subscribe.BangumiID,
		MediaID:         subscribe.MediaID,
		Season:          subscribe.Season,
		Poster:          subscribe.Poster,
		Backdrop:        subscribe.Backdrop,
		Vote:            subscribe.Vote,
		Description:     subscribe.Description,
		Filter:          subscribe.Filter,
		Include:         subscribe.Include,
		Exclude:         subscribe.Exclude,
		Quality:         subscribe.Quality,
		Resolution:      subscribe.Resolution,
		Effect:          subscribe.Effect,
		TotalEpisode:    subscribe.TotalEpisode,
		StartEpisode:    subscribe.StartEpisode,
		LackEpisode:     subscribe.LackEpisode,
		Note:            subscribe.Note,
		State:           subscribe.State,
		LastUpdate:      subscribe.LastUpdate,
		Username:        subscribe.Username,
		Sites:           subscribe.Sites,
		Downloader:      subscribe.Downloader,
		BestVersion:     subscribe.BestVersion,
		CurrentPriority: subscribe.CurrentPriority,
		MediaCategory:   subscribe.MediaCategory,
		EpisodeGroup:    subscribe.EpisodeGroup,
		CreatedAt:       subscribe.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       subscribe.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	response.Success(c, responseSubscribe)
}

// CreateSubscribe 创建订阅
// @Summary 创建订阅
// @Description 创建新的订阅
// @Tags 订阅
// @Accept json
// @Produce json
// @Param request body CreateSubscribeRequest true "创建订阅请求"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /subscribe [post]
func (h *ServArrHandler) CreateSubscribe(c *gin.Context) {
	var req CreateSubscribeRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("创建订阅请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.NewValidator(h.logger).Validate(req); err != nil {
		h.logger.Warn("创建订阅请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.NewValidator(h.logger).TranslateError(err))
		return
	}

	// 创建订阅对象
	subscribe := &models.Subscribe{
		Name:            req.Name,
		Year:            req.Year,
		Type:            req.Type,
		Keyword:         req.Keyword,
		TMDBID:          req.TMDBID,
		IMDBID:          req.IMDBID,
		TVDBID:          req.TVDBID,
		DoubanID:        req.DoubanID,
		BangumiID:       req.BangumiID,
		MediaID:         req.MediaID,
		Season:          req.Season,
		Poster:          req.Poster,
		Backdrop:        req.Backdrop,
		Vote:            req.Vote,
		Description:     req.Description,
		Filter:          req.Filter,
		Include:         req.Include,
		Exclude:         req.Exclude,
		Quality:         req.Quality,
		Resolution:      req.Resolution,
		Effect:          req.Effect,
		TotalEpisode:    req.TotalEpisode,
		StartEpisode:    req.StartEpisode,
		Note:            req.Note,
		State:           "N", // 新建状态
		Sites:           req.Sites,
		Downloader:      req.Downloader,
		BestVersion:     0,
		CurrentPriority: nil,
		MediaCategory:   "",
		EpisodeGroup:    "",
		Username:        "system", // 可以从JWT获取
	}

	// 保存订阅
	if err := h.subscribeRepo.Create(subscribe); err != nil {
		h.logger.Error("创建订阅失败", zap.Error(err))
		response.InternalServerError(c, "创建订阅失败")
		return
	}

	response.SuccessWithMessage(c, "订阅创建成功", map[string]interface{}{
		"id": strconv.FormatUint(uint64(subscribe.ID), 10),
	})
}

// UpdateSubscribe 更新订阅
// @Summary 更新订阅
// @Description 更新现有订阅
// @Tags 订阅
// @Accept json
// @Produce json
// @Param id path string true "订阅ID"
// @Param request body UpdateSubscribeRequest true "更新订阅请求"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /subscribe/{id} [put]
func (h *ServArrHandler) UpdateSubscribe(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "订阅ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订阅ID格式错误")
		return
	}

	var req UpdateSubscribeRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新订阅请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 获取现有订阅
	subscribe, err := h.subscribeRepo.GetByID(uint(id))
	if err != nil {
		h.logger.Error("获取订阅失败", zap.Error(err))
		response.InternalServerError(c, "获取订阅失败")
		return
	}

	// 更新字段（只更新非空字段）
	if req.Name != nil {
		subscribe.Name = *req.Name
	}
	if req.Year != nil {
		subscribe.Year = req.Year
	}
	if req.Type != nil {
		subscribe.Type = *req.Type
	}
	if req.Keyword != nil {
		subscribe.Keyword = *req.Keyword
	}
	if req.TMDBID != nil {
		subscribe.TMDBID = req.TMDBID
	}
	if req.IMDBID != nil {
		subscribe.IMDBID = req.IMDBID
	}
	if req.TVDBID != nil {
		subscribe.TVDBID = req.TVDBID
	}
	if req.DoubanID != nil {
		subscribe.DoubanID = req.DoubanID
	}
	if req.BangumiID != nil {
		subscribe.BangumiID = req.BangumiID
	}
	if req.MediaID != nil {
		subscribe.MediaID = req.MediaID
	}
	if req.Season != nil {
		subscribe.Season = req.Season
	}
	if req.Poster != nil {
		subscribe.Poster = *req.Poster
	}
	if req.Backdrop != nil {
		subscribe.Backdrop = *req.Backdrop
	}
	if req.Vote != nil {
		subscribe.Vote = req.Vote
	}
	if req.Description != nil {
		subscribe.Description = *req.Description
	}
	if req.Filter != nil {
		subscribe.Filter = *req.Filter
	}
	if req.Include != nil {
		subscribe.Include = *req.Include
	}
	if req.Exclude != nil {
		subscribe.Exclude = *req.Exclude
	}
	if req.Quality != nil {
		subscribe.Quality = *req.Quality
	}
	if req.Resolution != nil {
		subscribe.Resolution = *req.Resolution
	}
	if req.Effect != nil {
		subscribe.Effect = *req.Effect
	}
	if req.TotalEpisode != nil {
		subscribe.TotalEpisode = req.TotalEpisode
	}
	if req.StartEpisode != nil {
		subscribe.StartEpisode = req.StartEpisode
	}
	if req.Note != nil {
		subscribe.Note = *req.Note
	}
	if req.State != nil {
		subscribe.State = *req.State
	}
	if req.Sites != nil {
		subscribe.Sites = *req.Sites
	}
	if req.Downloader != nil {
		subscribe.Downloader = *req.Downloader
	}
	if req.BestVersion != nil {
		subscribe.BestVersion = *req.BestVersion
	}
	if req.CurrentPriority != nil {
		subscribe.CurrentPriority = req.CurrentPriority
	}
	if req.MediaCategory != nil {
		subscribe.MediaCategory = *req.MediaCategory
	}
	if req.EpisodeGroup != nil {
		subscribe.EpisodeGroup = *req.EpisodeGroup
	}

	// 保存更新
	if err := h.subscribeRepo.Update(subscribe); err != nil {
		h.logger.Error("更新订阅失败", zap.Error(err))
		response.InternalServerError(c, "更新订阅失败")
		return
	}

	response.SuccessWithMessage(c, "订阅更新成功", nil)
}

// DeleteSubscribe 删除订阅
// @Summary 删除订阅
// @Description 删除订阅
// @Tags 订阅
// @Produce json
// @Param id path string true "订阅ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /subscribe/{id} [delete]
func (h *ServArrHandler) DeleteSubscribe(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "订阅ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "订阅ID格式错误")
		return
	}

	// 删除订阅
	if err := h.subscribeRepo.Delete(uint(id)); err != nil {
		h.logger.Error("删除订阅失败", zap.Error(err))
		response.InternalServerError(c, "删除订阅失败")
		return
	}

	response.SuccessWithMessage(c, "订阅删除成功", nil)
}