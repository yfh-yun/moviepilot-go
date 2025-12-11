package servarr

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Handler Servarr API 处理器
type Handler struct {
	logger *zap.Logger
}

// NewHandler 创建 Servarr 处理器
func NewHandler() *Handler {
	return &Handler{
		logger: logger.GetLogger(),
	}
}

// RadarrMovie Radarr电影响应结构
type RadarrMovie struct {
	ID               uint   `json:"id"`
	Title            string `json:"title"`
	Year             int    `json:"year"`
	IsAvailable      bool   `json:"isAvailable"`
	Monitored        bool   `json:"monitored"`
	TmdbID           int    `json:"tmdbId"`
	ImdbID           string `json:"imdbId"`
	ProfileID        int    `json:"profileId"`
	QualityProfileID int    `json:"qualityProfileId"`
	HasFile          bool   `json:"hasFile"`
	TitleSlug        string `json:"titleSlug,omitempty"`
	FolderName       string `json:"folderName,omitempty"`
}

// SonarrSeries Sonarr电视剧响应结构
type SonarrSeries struct {
	ID                uint           `json:"id"`
	Title             string         `json:"title"`
	SeasonCount       int            `json:"seasonCount"`
	Seasons           []SonarrSeason `json:"seasons"`
	RemotePoster      string         `json:"remotePoster,omitempty"`
	Year              int            `json:"year"`
	TmdbID            int            `json:"tmdbId"`
	TvdbID            int            `json:"tvdbId"`
	ImdbID            string         `json:"imdbId"`
	ProfileID         int            `json:"profileId"`
	LanguageProfileID int            `json:"languageProfileId"`
	QualityProfileID  int            `json:"qualityProfileId"`
	IsAvailable       bool           `json:"isAvailable"`
	Monitored         bool           `json:"monitored"`
	HasFile           bool           `json:"hasFile"`
}

// SonarrSeason Sonarr季信息
type SonarrSeason struct {
	SeasonNumber int  `json:"seasonNumber"`
	Monitored    bool `json:"monitored"`
}

// SystemStatus 系统状态
// @Summary 系统状态
// @Description 模拟Radarr、Sonarr系统状态
// @Tags servarr
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /system/status [get]
func (h *Handler) SystemStatus(c *gin.Context) {
	h.logger.Debug("SystemStatus called")

	c.JSON(http.StatusOK, gin.H{
		"appName":           "MoviePilot",
		"instanceName":      "moviepilot",
		"version":           "2.8.1",
		"buildTime":         "",
		"isDebug":           false,
		"isProduction":      true,
		"isAdmin":           true,
		"isUserInteractive": true,
		"startupPath":       "/app",
		"appData":           "/config",
		"osName":            "debian",
		"osVersion":         "",
		"isNetCore":         true,
		"isLinux":           true,
		"isOsx":             false,
		"isWindows":         false,
		"isDocker":          true,
		"mode":              "console",
		"branch":            "main",
		"databaseType":      "sqlite",
		"databaseVersion": gin.H{
			"major":         0,
			"minor":         0,
			"build":         0,
			"revision":      0,
			"majorRevision": 0,
			"minorRevision": 0,
		},
		"authentication":   "none",
		"migrationVersion": 0,
		"urlBase":          "",
		"runtimeVersion": gin.H{
			"major":         0,
			"minor":         0,
			"build":         0,
			"revision":      0,
			"majorRevision": 0,
			"minorRevision": 0,
		},
		"runtimeName":                   "",
		"startTime":                     "",
		"packageVersion":                "",
		"packageAuthor":                 "jxxghp",
		"packageUpdateMechanism":        "builtIn",
		"packageUpdateMechanismMessage": "",
	})
}

// QualityProfile 质量配置
// @Summary 质量配置
// @Description 模拟Radarr、Sonarr质量配置
// @Tags servarr
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /qualityProfile [get]
func (h *Handler) QualityProfile(c *gin.Context) {
	h.logger.Debug("QualityProfile called")

	c.JSON(http.StatusOK, []gin.H{
		{
			"id":             1,
			"name":           "默认",
			"upgradeAllowed": true,
			"cutoff":         0,
			"items": []gin.H{
				{
					"id":   0,
					"name": "默认",
					"quality": gin.H{
						"id":         0,
						"name":       "默认",
						"source":     "0",
						"resolution": 0,
					},
					"items":   []string{"string"},
					"allowed": true,
				},
			},
			"minFormatScore":    0,
			"cutoffFormatScore": 0,
			"formatItems": []gin.H{
				{
					"id":     0,
					"format": 0,
					"name":   "默认",
					"score":  0,
				},
			},
		},
	})
}

// RootFolder 根目录
// @Summary 根目录
// @Description 模拟Radarr、Sonarr根目录
// @Tags servarr
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /rootfolder [get]
func (h *Handler) RootFolder(c *gin.Context) {
	h.logger.Debug("RootFolder called")

	c.JSON(http.StatusOK, []gin.H{
		{
			"id":              1,
			"path":            "/",
			"accessible":      true,
			"freeSpace":       0,
			"unmappedFolders": []string{},
		},
	})
}

// Tag 标签
// @Summary 标签
// @Description 模拟Radarr、Sonarr标签
// @Tags servarr
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /tag [get]
func (h *Handler) Tag(c *gin.Context) {
	h.logger.Debug("Tag called")

	c.JSON(http.StatusOK, []gin.H{
		{
			"id":    1,
			"label": "默认",
		},
	})
}

// LanguageProfile 语言配置
// @Summary 语言配置
// @Description 模拟Radarr、Sonarr语言
// @Tags servarr
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /languageprofile [get]
func (h *Handler) LanguageProfile(c *gin.Context) {
	h.logger.Debug("LanguageProfile called")

	c.JSON(http.StatusOK, []gin.H{
		{
			"id":             1,
			"name":           "默认",
			"upgradeAllowed": true,
			"cutoff": gin.H{
				"id":   1,
				"name": "默认",
			},
			"languages": []gin.H{
				{
					"id": 1,
					"language": gin.H{
						"id":   1,
						"name": "默认",
					},
					"allowed": true,
				},
			},
		},
	})
}

// GetMovies 获取所有订阅电影
// @Summary 所有订阅电影
// @Description 查询所有电影订阅
// @Tags servarr
// @Produce json
// @Success 200 {array} RadarrMovie
// @Router /movie [get]
func (h *Handler) GetMovies(c *gin.Context) {
	h.logger.Debug("GetMovies called")

	// 简化实现，暂时返回空数组
	result := make([]RadarrMovie, 0)
	c.JSON(http.StatusOK, result)
}

// GetMovieLookup 查询电影
// @Summary 查询电影
// @Description 查询电影 term: `tmdb:${id}`
// @Tags servarr
// @Produce json
// @Param term query string true "查询条件"
// @Success 200 {array} RadarrMovie
// @Router /movie/lookup [get]
func (h *Handler) GetMovieLookup(c *gin.Context) {
	h.logger.Debug("GetMovieLookup called")

	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "term is required"})
		return
	}

	// 处理tmdb:id格式
	tmdbIDStr := term
	if strings.HasPrefix(tmdbIDStr, "tmdb:") {
		tmdbIDStr = tmdbIDStr[5:]
	}

	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		h.logger.Warn("Invalid tmdb id", zap.String("term", term))
		c.JSON(http.StatusOK, []RadarrMovie{{}})
		return
	}

	// 简化实现，暂时返回模拟数据
	radarrMovie := RadarrMovie{
		TmdbID:           tmdbID,
		IsAvailable:      true,
		Monitored:        false,
		ProfileID:        1,
		QualityProfileID: 1,
		HasFile:          false,
	}

	c.JSON(http.StatusOK, []RadarrMovie{radarrMovie})
}

// GetMovie 获取电影详情
// @Summary 电影订阅详情
// @Description 查询电影订阅详情
// @Tags servarr
// @Produce json
// @Param mid path int true "电影ID"
// @Success 200 {object} RadarrMovie
// @Failure 404 {object} map[string]interface{}
// @Router /movie/{mid} [get]
func (h *Handler) GetMovie(c *gin.Context) {
	h.logger.Debug("GetMovie called")

	midStr := c.Param("mid")
	mid, err := strconv.ParseUint(midStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie id"})
		return
	}

	// 简化实现，暂时返回模拟数据
	c.JSON(http.StatusOK, RadarrMovie{
		ID:               uint(mid),
		Title:            "Test Movie",
		Year:             2023,
		IsAvailable:      true,
		Monitored:        true,
		TmdbID:           12345,
		ImdbID:           "tt1234567",
		ProfileID:        1,
		QualityProfileID: 1,
		HasFile:          false,
	})
}

// AddMovie 新增电影订阅
// @Summary 新增电影订阅
// @Description 新增电影订阅
// @Tags servarr
// @Accept json
// @Produce json
// @Param movie body RadarrMovie true "电影信息"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /movie [post]
func (h *Handler) AddMovie(c *gin.Context) {
	h.logger.Debug("AddMovie called")

	var movie RadarrMovie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 简化实现，暂时返回成功
	c.JSON(http.StatusOK, gin.H{"id": 1})
}

// DeleteMovie 删除电影订阅
// @Summary 删除电影订阅
// @Description 删除电影订阅
// @Tags servarr
// @Produce json
// @Param mid path int true "电影ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /movie/{mid} [delete]
func (h *Handler) DeleteMovie(c *gin.Context) {
	h.logger.Debug("DeleteMovie called")

	midStr := c.Param("mid")
	_, err := strconv.ParseUint(midStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie id"})
		return
	}

	// 简化实现，暂时返回成功
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetSeries 获取所有订阅剧集
// @Summary 所有订阅剧集
// @Description 查询所有电视剧订阅
// @Tags servarr
// @Produce json
// @Success 200 {array} SonarrSeries
// @Router /series [get]
func (h *Handler) GetSeries(c *gin.Context) {
	h.logger.Debug("GetSeries called")

	// 简化实现，暂时返回空数组
	result := make([]SonarrSeries, 0)
	c.JSON(http.StatusOK, result)
}

// GetSeriesLookup 查询剧集
// @Summary 查询剧集
// @Description 查询剧集 term: `tvdb:${id}` title
// @Tags servarr
// @Produce json
// @Param term query string true "查询条件"
// @Success 200 {array} SonarrSeries
// @Router /series/lookup [get]
func (h *Handler) GetSeriesLookup(c *gin.Context) {
	h.logger.Debug("GetSeriesLookup called")

	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "term is required"})
		return
	}

	// TODO: 实现剧集查询逻辑
	// 目前返回空数组，后续需要完善
	c.JSON(http.StatusOK, []SonarrSeries{{}})
}

// GetSerie 获取剧集详情
// @Summary 剧集详情
// @Description 查询剧集详情
// @Tags servarr
// @Produce json
// @Param tid path int true "剧集ID"
// @Success 200 {object} SonarrSeries
// @Failure 404 {object} map[string]interface{}
// @Router /series/{tid} [get]
func (h *Handler) GetSerie(c *gin.Context) {
	h.logger.Debug("GetSerie called")

	tidStr := c.Param("tid")
	tid, err := strconv.ParseUint(tidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series id"})
		return
	}

	// 简化实现，暂时返回模拟数据
	c.JSON(http.StatusOK, SonarrSeries{
		ID:          uint(tid),
		Title:       "Test TV Show",
		SeasonCount: 1,
		Seasons: []SonarrSeason{
			{
				SeasonNumber: 1,
				Monitored:    true,
			},
		},
		RemotePoster:      "",
		Year:              2023,
		TmdbID:            12345,
		TvdbID:            67890,
		ImdbID:            "tt1234567",
		ProfileID:         1,
		LanguageProfileID: 1,
		QualityProfileID:  1,
		IsAvailable:       true,
		Monitored:         true,
		HasFile:           false,
	})
}

// AddSeries 新增剧集订阅
// @Summary 新增剧集订阅
// @Description 新增剧集订阅
// @Tags servarr
// @Accept json
// @Produce json
// @Param tv body SonarrSeries true "剧集信息"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /series [post]
func (h *Handler) AddSeries(c *gin.Context) {
	h.logger.Debug("AddSeries called")

	var tv SonarrSeries
	if err := c.ShouldBindJSON(&tv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: 实现新增剧集订阅逻辑
	// 目前返回成功，后续需要完善
	c.JSON(http.StatusOK, gin.H{"id": 1})
}

// UpdateSeries 更新剧集订阅
// @Summary 更新剧集订阅
// @Description 更新剧集订阅
// @Tags servarr
// @Accept json
// @Produce json
// @Param tv body SonarrSeries true "剧集信息"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /series [put]
func (h *Handler) UpdateSeries(c *gin.Context) {
	h.logger.Debug("UpdateSeries called")

	// 复用AddSeries逻辑
	h.AddSeries(c)
}

// DeleteSeries 删除剧集订阅
// @Summary 删除剧集订阅
// @Description 删除剧集订阅
// @Tags servarr
// @Produce json
// @Param tid path int true "剧集ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /series/{tid} [delete]
func (h *Handler) DeleteSeries(c *gin.Context) {
	h.logger.Debug("DeleteSeries called")

	tidStr := c.Param("tid")
	_, err := strconv.ParseUint(tidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series id"})
		return
	}

	// 简化实现，暂时返回成功
	c.JSON(http.StatusOK, gin.H{"success": true})
}
