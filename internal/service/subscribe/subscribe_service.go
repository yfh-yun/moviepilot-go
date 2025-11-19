package subscribe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/bangumi"
	"github.com/yfh-yun/moviepilot-go/internal/integration/douban"
	"github.com/yfh-yun/moviepilot-go/internal/integration/tmdb"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/repository/models"
)

var (
	// ErrSubscribeNotFound 订阅不存在
	ErrSubscribeNotFound = errors.New("订阅不存在")
	// ErrSubscribeExists 订阅已存在
	ErrSubscribeExists = errors.New("订阅已存在")
	// ErrInvalidState 无效的订阅状态
	ErrInvalidState = errors.New("无效的订阅状态")
	// ErrMediaInfoNotFound 媒体信息未找到
	ErrMediaInfoNotFound = errors.New("媒体信息未找到")
)

// SubscribeService 订阅服务
type SubscribeService struct {
	subscribeRepo interfaces.SubscribeRepository
	mediaRepo     interfaces.MediaRepository
	tmdbClient    *tmdb.Client
	doubanClient  *douban.Client
	bangumiClient *bangumi.Client
	logger        *logger.Logger
}

// NewSubscribeService 创建订阅服务
func NewSubscribeService(
	subscribeRepo interfaces.SubscribeRepository,
	mediaRepo interfaces.MediaRepository,
	tmdbClient *tmdb.Client,
	doubanClient *douban.Client,
	bangumiClient *bangumi.Client,
	log *logger.Logger,
) *SubscribeService {
	return &SubscribeService{
		subscribeRepo: subscribeRepo,
		mediaRepo:     mediaRepo,
		tmdbClient:    tmdbClient,
		doubanClient:  doubanClient,
		bangumiClient: bangumiClient,
		logger:        log,
	}
}

// CreateRequest 创建订阅请求
type CreateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Year        *string `json:"year"`
	Type        string  `json:"type" validate:"required,oneof=movie tv"`
	Keyword     string  `json:"keyword"`
	TMDBID      *int    `json:"tmdb_id"`
	DoubanID    *string `json:"douban_id"`
	BangumiID   *int    `json:"bangumi_id"`
	Season      *int    `json:"season"`
	Username    string  `json:"username"`
	Quality     string  `json:"quality"`
	Resolution  string  `json:"resolution"`
	Effect      string  `json:"effect"`
	Include     string  `json:"include"`
	Exclude     string  `json:"exclude"`
	Sites       []int   `json:"sites"`
	Downloader  string  `json:"downloader"`
	SavePath    string  `json:"save_path"`
	BestVersion int     `json:"best_version"`
}

// UpdateRequest 更新订阅请求
type UpdateRequest struct {
	ID           uint       `json:"id" validate:"required"`
	Name         *string    `json:"name"`
	State        *string    `json:"state" validate:"omitempty,oneof=N R P S"`
	Quality      *string    `json:"quality"`
	Resolution   *string    `json:"resolution"`
	Effect       *string    `json:"effect"`
	Include      *string    `json:"include"`
	Exclude      *string    `json:"exclude"`
	Sites        []int      `json:"sites"`
	Downloader   *string    `json:"downloader"`
	SavePath     *string    `json:"save_path"`
	TotalEpisode *int       `json:"total_episode"`
	LackEpisode  *int       `json:"lack_episode"`
	BestVersion  *int       `json:"best_version"`
	LastUpdate   *time.Time `json:"last_update"`
}

// SubscribeResponse 订阅响应
type SubscribeResponse struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Year            *string    `json:"year"`
	Type            string     `json:"type"`
	Keyword         string     `json:"keyword"`
	TMDBID          *int       `json:"tmdb_id"`
	DoubanID        *string    `json:"douban_id"`
	BangumiID       *int       `json:"bangumi_id"`
	Season          *int       `json:"season"`
	Poster          string     `json:"poster"`
	Backdrop        string     `json:"backdrop"`
	Vote            *float64   `json:"vote"`
	Description     string     `json:"description"`
	Quality         string     `json:"quality"`
	Resolution      string     `json:"resolution"`
	Effect          string     `json:"effect"`
	Include         string     `json:"include"`
	Exclude         string     `json:"exclude"`
	TotalEpisode    *int       `json:"total_episode"`
	LackEpisode     *int       `json:"lack_episode"`
	State           string     `json:"state"`
	LastUpdate      *time.Time `json:"last_update"`
	Username        string     `json:"username"`
	Downloader      string     `json:"downloader"`
	SavePath        string     `json:"save_path"`
	BestVersion     int        `json:"best_version"`
	CurrentPriority *int       `json:"current_priority"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Create 创建订阅
func (s *SubscribeService) Create(ctx context.Context, req *CreateRequest) (*SubscribeResponse, error) {
	s.logger.Info("开始创建订阅", "name", req.Name)

	// 检查订阅是否已存在
	if req.TMDBID != nil {
		exists, err := s.subscribeRepo.Exists(req.TMDBID, nil, req.Season)
		if err != nil {
			s.logger.Error("检查订阅是否存在失败", "error", err)
			return nil, err
		}
		if exists {
			return nil, ErrSubscribeExists
		}
	}

	// 识别媒体信息
	mediaInfo, err := s.recognizeMedia(ctx, req)
	if err != nil {
		s.logger.Error("识别媒体信息失败", "error", err)
		return nil, err
	}

	// 获取总集数(电视剧)
	if req.Type == "tv" && req.Season != nil {
		totalEpisode, err := s.getTotalEpisode(ctx, req.TMDBID, req.DoubanID, *req.Season)
		if err != nil {
			s.logger.Warn("获取总集数失败", "error", err)
		} else if totalEpisode > 0 {
			mediaInfo.TotalEpisode = &totalEpisode
			mediaInfo.LackEpisode = &totalEpisode
		}
	}

	// 创建订阅记录
	subscribe := &models.Subscribe{
		Name:         req.Name,
		Year:         req.Year,
		Type:         req.Type,
		Keyword:      req.Keyword,
		TMDBID:       req.TMDBID,
		DoubanID:     req.DoubanID,
		BangumiID:    req.BangumiID,
		Season:       req.Season,
		Poster:       mediaInfo.Poster,
		Backdrop:     mediaInfo.Backdrop,
		Vote:         mediaInfo.Vote,
		Description:  mediaInfo.Description,
		Quality:      req.Quality,
		Resolution:   req.Resolution,
		Effect:       req.Effect,
		Include:      req.Include,
		Exclude:      req.Exclude,
		TotalEpisode: mediaInfo.TotalEpisode,
		LackEpisode:  mediaInfo.LackEpisode,
		State:        "N", // N-新建
		Username:     req.Username,
		Downloader:   req.Downloader,
		SavePath:     req.SavePath,
		BestVersion:  req.BestVersion,
	}

	if err := s.subscribeRepo.Create(subscribe); err != nil {
		s.logger.Error("创建订阅失败", "error", err)
		return nil, err
	}

	s.logger.Info("订阅创建成功", "id", subscribe.ID, "name", req.Name)

	return s.toResponse(subscribe), nil
}

// Update 更新订阅
func (s *SubscribeService) Update(ctx context.Context, req *UpdateRequest) error {
	s.logger.Info("更新订阅", "id", req.ID)

	// 获取订阅
	subscribe, err := s.subscribeRepo.GetByID(req.ID)
	if err != nil {
		return err
	}
	if subscribe == nil {
		return ErrSubscribeNotFound
	}

	// 更新字段
	if req.Name != nil {
		subscribe.Name = *req.Name
	}
	if req.State != nil {
		// 验证状态
		validStates := map[string]bool{"N": true, "R": true, "P": true, "S": true}
		if !validStates[*req.State] {
			return ErrInvalidState
		}
		subscribe.State = *req.State
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
	if req.Include != nil {
		subscribe.Include = *req.Include
	}
	if req.Exclude != nil {
		subscribe.Exclude = *req.Exclude
	}
	if req.Downloader != nil {
		subscribe.Downloader = *req.Downloader
	}
	if req.SavePath != nil {
		subscribe.SavePath = *req.SavePath
	}
	if req.TotalEpisode != nil {
		subscribe.TotalEpisode = req.TotalEpisode
	}
	if req.LackEpisode != nil {
		subscribe.LackEpisode = req.LackEpisode
	}
	if req.BestVersion != nil {
		subscribe.BestVersion = *req.BestVersion
	}
	if req.LastUpdate != nil {
		subscribe.LastUpdate = req.LastUpdate
	}

	if err := s.subscribeRepo.Update(subscribe); err != nil {
		s.logger.Error("更新订阅失败", "error", err)
		return err
	}

	s.logger.Info("订阅更新成功", "id", req.ID)
	return nil
}

// Delete 删除订阅
func (s *SubscribeService) Delete(ctx context.Context, id uint) error {
	s.logger.Info("删除订阅", "id", id)

	// 检查订阅是否存在
	subscribe, err := s.subscribeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if subscribe == nil {
		return ErrSubscribeNotFound
	}

	if err := s.subscribeRepo.Delete(id); err != nil {
		s.logger.Error("删除订阅失败", "error", err)
		return err
	}

	s.logger.Info("订阅删除成功", "id", id)
	return nil
}

// GetByID 根据ID获取订阅
func (s *SubscribeService) GetByID(ctx context.Context, id uint) (*SubscribeResponse, error) {
	subscribe, err := s.subscribeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if subscribe == nil {
		return nil, ErrSubscribeNotFound
	}
	return s.toResponse(subscribe), nil
}

// List 获取订阅列表
func (s *SubscribeService) List(ctx context.Context, offset, limit int) ([]*SubscribeResponse, int64, error) {
	subscribes, total, err := s.subscribeRepo.List(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*SubscribeResponse, 0, len(subscribes))
	for _, sub := range subscribes {
		responses = append(responses, s.toResponse(sub))
	}

	return responses, total, nil
}

// ListByUsername 根据用户名获取订阅列表
func (s *SubscribeService) ListByUsername(ctx context.Context, username, state, mtype string) ([]*SubscribeResponse, error) {
	subscribes, err := s.subscribeRepo.ListByUsername(username, state, mtype)
	if err != nil {
		return nil, err
	}

	responses := make([]*SubscribeResponse, 0, len(subscribes))
	for _, sub := range subscribes {
		responses = append(responses, s.toResponse(sub))
	}

	return responses, nil
}

// UpdateState 更新订阅状态
func (s *SubscribeService) UpdateState(ctx context.Context, id uint, state string) error {
	// 验证状态
	validStates := map[string]bool{"N": true, "R": true, "P": true, "S": true}
	if !validStates[state] {
		return ErrInvalidState
	}

	// 检查订阅是否存在
	subscribe, err := s.subscribeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if subscribe == nil {
		return ErrSubscribeNotFound
	}

	return s.subscribeRepo.UpdateState(id, state)
}

// Reset 重置订阅
func (s *SubscribeService) Reset(ctx context.Context, id uint) error {
	s.logger.Info("重置订阅", "id", id)

	subscribe, err := s.subscribeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if subscribe == nil {
		return ErrSubscribeNotFound
	}

	// 重置订阅状态
	subscribe.LackEpisode = subscribe.TotalEpisode
	subscribe.State = "R"
	subscribe.Note = ""

	if err := s.subscribeRepo.Update(subscribe); err != nil {
		s.logger.Error("重置订阅失败", "error", err)
		return err
	}

	s.logger.Info("订阅重置成功", "id", id)
	return nil
}

// GetByMediaID 根据媒体ID获取订阅
func (s *SubscribeService) GetByMediaID(ctx context.Context, mediaID string, season *int) (*SubscribeResponse, error) {
	var subscribe *models.Subscribe
	var err error

	// 解析媒体ID类型
	if len(mediaID) > 5 && mediaID[:5] == "tmdb:" {
		// TMDB ID
		var tmdbID int
		_, err := fmt.Sscanf(mediaID[5:], "%d", &tmdbID)
		if err != nil {
			return nil, errors.New("invalid tmdb id")
		}

		subscribes, err := s.subscribeRepo.GetByTMDBID(tmdbID, season)
		if err != nil {
			return nil, err
		}
		if len(subscribes) > 0 {
			subscribe = subscribes[0]
		}
	} else if len(mediaID) > 7 && mediaID[:7] == "douban:" {
		// 豆瓣 ID
		doubanID := mediaID[7:]
		subscribe, err = s.subscribeRepo.GetByDoubanID(doubanID)
		if err != nil {
			return nil, err
		}
	} else if len(mediaID) > 8 && mediaID[:8] == "bangumi:" {
		// Bangumi ID
		var bangumiID int
		_, err := fmt.Sscanf(mediaID[8:], "%d", &bangumiID)
		if err != nil {
			return nil, errors.New("invalid bangumi id")
		}

		subscribe, err = s.subscribeRepo.GetByBangumiID(bangumiID)
		if err != nil {
			return nil, err
		}
	} else {
		// 尝试作为媒体ID查询
		subscribe, err = s.subscribeRepo.GetByMediaID(mediaID)
		if err != nil {
			return nil, err
		}
	}

	if subscribe == nil {
		return nil, ErrSubscribeNotFound
	}

	return s.toResponse(subscribe), nil
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Poster       string
	Backdrop     string
	Vote         *float64
	Description  string
	TotalEpisode *int
	LackEpisode  *int
}

// recognizeMedia 识别媒体信息
func (s *SubscribeService) recognizeMedia(ctx context.Context, req *CreateRequest) (*MediaInfo, error) {
	info := &MediaInfo{}

	// 优先使用TMDB
	if req.TMDBID != nil {
		if req.Type == "movie" {
			movie, err := s.tmdbClient.GetMovieDetails(ctx, *req.TMDBID)
			if err != nil {
				return nil, err
			}
			if movie == nil {
				return nil, ErrMediaInfoNotFound
			}

			info.Poster = s.tmdbClient.GetPosterURL(movie.PosterPath, "w500")
			info.Backdrop = s.tmdbClient.GetBackdropURL(movie.BackdropPath, "w1280")
			info.Vote = &movie.VoteAverage
			info.Description = movie.Overview
		} else {
			tv, err := s.tmdbClient.GetTVDetails(ctx, *req.TMDBID)
			if err != nil {
				return nil, err
			}
			if tv == nil {
				return nil, ErrMediaInfoNotFound
			}

			info.Poster = s.tmdbClient.GetPosterURL(tv.PosterPath, "w500")
			info.Backdrop = s.tmdbClient.GetBackdropURL(tv.BackdropPath, "w1280")
			info.Vote = &tv.VoteAverage
			info.Description = tv.Overview
		}
		return info, nil
	}

	// 使用豆瓣
	if req.DoubanID != nil {
		movie, err := s.doubanClient.GetMovieDetails(ctx, *req.DoubanID)
		if err != nil {
			return nil, err
		}
		if movie == nil {
			return nil, ErrMediaInfoNotFound
		}

		info.Poster = movie.Cover
		info.Description = movie.Intro
		score := movie.Rating.Value
		info.Vote = &score
		return info, nil
	}

	// 使用Bangumi
	if req.BangumiID != nil {
		subject, err := s.bangumiClient.GetSubjectDetails(ctx, *req.BangumiID)
		if err != nil {
			return nil, err
		}
		if subject == nil {
			return nil, ErrMediaInfoNotFound
		}

		info.Poster = subject.Images.Large
		info.Description = subject.Summary
		score := subject.Rating.Score
		info.Vote = &score
		return info, nil
	}

	return nil, ErrMediaInfoNotFound
}

// getTotalEpisode 获取总集数
func (s *SubscribeService) getTotalEpisode(ctx context.Context, tmdbID *int, doubanID *string, season int) (int, error) {
	if tmdbID != nil {
		tv, err := s.tmdbClient.GetTVDetails(ctx, *tmdbID)
		if err != nil {
			return 0, err
		}

		// 查找对应季的集数
		for _, s := range tv.Seasons {
			if s.SeasonNumber == season {
				return s.EpisodeCount, nil
			}
		}
	}

	// TODO: 豆瓣和Bangumi的总集数获取

	return 0, errors.New("无法获取总集数")
}

// toResponse 转换为响应对象
func (s *SubscribeService) toResponse(subscribe *models.Subscribe) *SubscribeResponse {
	return &SubscribeResponse{
		ID:              subscribe.ID,
		Name:            subscribe.Name,
		Year:            subscribe.Year,
		Type:            subscribe.Type,
		Keyword:         subscribe.Keyword,
		TMDBID:          subscribe.TMDBID,
		DoubanID:        subscribe.DoubanID,
		BangumiID:       subscribe.BangumiID,
		Season:          subscribe.Season,
		Poster:          subscribe.Poster,
		Backdrop:        subscribe.Backdrop,
		Vote:            subscribe.Vote,
		Description:     subscribe.Description,
		Quality:         subscribe.Quality,
		Resolution:      subscribe.Resolution,
		Effect:          subscribe.Effect,
		Include:         subscribe.Include,
		Exclude:         subscribe.Exclude,
		TotalEpisode:    subscribe.TotalEpisode,
		LackEpisode:     subscribe.LackEpisode,
		State:           subscribe.State,
		LastUpdate:      subscribe.LastUpdate,
		Username:        subscribe.Username,
		Downloader:      subscribe.Downloader,
		SavePath:        subscribe.SavePath,
		BestVersion:     subscribe.BestVersion,
		CurrentPriority: subscribe.CurrentPriority,
		CreatedAt:       subscribe.CreatedAt,
		UpdatedAt:       subscribe.UpdatedAt,
	}
}
