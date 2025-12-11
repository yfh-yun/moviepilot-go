package metadata

import (
	"context"

	"go.uber.org/zap"

	metaIntegration "moviepilot-go/internal/integration/metadata"
	appLogger "moviepilot-go/pkg/logger"
)

// AggregatedMovie 聚合后的电影元数据
// 目前仅包含 TMDB 主数据，预留豆瓣等扩展字段

type AggregatedMovie struct {
	Primary *metaIntegration.MovieInfo `json:"primary"`
	Extra   map[string]any             `json:"extra,omitempty"`
}

// AggregatedTVShow 聚合后的剧集元数据（暂未实现）
type AggregatedTVShow struct {
	Primary *metaIntegration.TVShowInfo `json:"primary"`
}

// Service 元数据聚合服务接口
type Service interface {
	// AggregateMovieByTMDB 根据 TMDB ID 聚合电影信息（最小实现：仅 TMDB）
	AggregateMovieByTMDB(ctx context.Context, tmdbID int) (*AggregatedMovie, error)

	// SearchAndAggregateMovie 根据标题+年份搜索并聚合电影
	// 当前实现：仅使用 TMDB 搜索，取第一个结果，再走 AggregateMovieByTMDB
	SearchAndAggregateMovie(ctx context.Context, title string, year int) (*AggregatedMovie, error)

	// AggregateTVByTMDB 根据 TMDB ID 聚合剧集信息
	// 当前实现：TMDB 主数据 + TVDB 可选补充
	AggregateTVByTMDB(ctx context.Context, tmdbID int) (*AggregatedTVShow, error)
}

type service struct {
	logger *zap.Logger

	// 主数据源
	tmdb metaIntegration.MetadataProvider
	// 辅助数据源
	tvdb   metaIntegration.MetadataProvider
	douban metaIntegration.MetadataProvider
}

// NewService 使用 metadata.Factory 创建聚合服务
// 要求工厂中至少注册 TMDB Provider
func NewService(factory *metaIntegration.Factory) Service {
	log := appLogger.GetLogger()

	tmdb, ok := factory.Get(metaIntegration.ProviderTMDB)
	if !ok {
		log.Warn("TMDB Provider 未注册，Metadata 聚合服务将无法正常工作")
	}

	tvdb, _ := factory.Get(metaIntegration.ProviderTVDB)
	douban, _ := factory.Get(metaIntegration.ProviderDouban)

	return &service{
		logger: log,
		tmdb:   tmdb,
		tvdb:   tvdb,
		douban: douban,
	}
}

// AggregateMovieByTMDB 最小版本实现
// 仅调用 TMDB Provider 的 GetMovieByTMDB，并包装为 AggregatedMovie
func (s *service) AggregateMovieByTMDB(ctx context.Context, tmdbID int) (*AggregatedMovie, error) {
	log := s.logger.With(zap.Int("tmdb_id", tmdbID))

	if s.tmdb == nil {
		log.Error("TMDB Provider 未初始化")
		return nil, ErrProviderNotAvailable("tmdb")
	}

	if tmdbID <= 0 {
		log.Warn("无效的 TMDB ID")
		return nil, ErrInvalidTMDBID
	}

	// 目前默认语言使用空值，由 TMDB 决定默认语言
	movie, err := s.tmdb.GetMovieByTMDB(ctx, tmdbID, "")
	if err != nil {
		log.Error("从 TMDB 获取电影详情失败", zap.Error(err))
		return nil, err
	}

	if movie == nil {
		log.Warn("TMDB 返回空电影信息")
		return nil, ErrMovieNotFound
	}

	log.Info("电影元数据聚合完成(最小版)",
		zap.String("title", movie.Title),
		zap.Int("year", movie.Year))

	return &AggregatedMovie{
		Primary: movie,
		Extra:   map[string]any{},
	}, nil
}

// SearchAndAggregateMovie 使用标题+年份搜索并聚合电影
// 当前实现策略：
// 1. 使用 TMDB.SearchMovie 搜索，带年份和 Limit=1
// 2. 如果有结果，取第一个的 TMDBID，调用 AggregateMovieByTMDB
// 3. 后续可以在此处加入豆瓣/TVDB 信息融合
func (s *service) SearchAndAggregateMovie(ctx context.Context, title string, year int) (*AggregatedMovie, error) {
	log := s.logger.With(
		zap.String("title", title),
		zap.Int("year", year),
	)

	if s.tmdb == nil {
		log.Error("TMDB Provider 未初始化")
		return nil, ErrProviderNotAvailable("tmdb")
	}
	if title == "" {
		log.Warn("搜索标题为空")
		return nil, Err("empty_title")
	}

	opts := metaIntegration.SearchOptions{
		Year:  year,
		Limit: 1,
	}
	movies, err := s.tmdb.SearchMovie(ctx, title, opts)
	if err != nil {
		log.Error("TMDB 搜索电影失败", zap.Error(err))
		return nil, err
	}
	if len(movies) == 0 {
		log.Warn("TMDB 未找到匹配电影")
		return nil, ErrMovieNotFound
	}

	primary := movies[0]
	if primary.TMDBID == nil {
		log.Warn("TMDB 搜索结果缺少 TMDBID，直接返回搜索结果",
			zap.String("title", primary.Title))
		return &AggregatedMovie{Primary: primary, Extra: map[string]any{}}, nil
	}

	log.Info("TMDB 搜索命中，进入聚合流程",
		zap.Int("tmdb_id", *primary.TMDBID))
	return s.AggregateMovieByTMDB(ctx, *primary.TMDBID)
}

// AggregateTVByTMDB 根据 TMDB ID 聚合剧集信息
// 当前策略：
// 1. 使用 TMDB.GetTVByTMDB 获取主数据
// 2. 如果 TVDB 可用，尝试通过 TMDB ID 查询 TVDB 补充 TVDBID（当前 TVDB 未实现映射，仅占位）
func (s *service) AggregateTVByTMDB(ctx context.Context, tmdbID int) (*AggregatedTVShow, error) {
	log := s.logger.With(zap.Int("tmdb_id", tmdbID))

	if s.tmdb == nil {
		log.Error("TMDB Provider 未初始化")
		return nil, ErrProviderNotAvailable("tmdb")
	}
	if tmdbID <= 0 {
		log.Warn("无效的 TMDB ID")
		return nil, ErrInvalidTMDBID
	}

	// 主数据来自 TMDB
	show, err := s.tmdb.GetTVByTMDB(ctx, tmdbID, "")
	if err != nil {
		log.Error("从 TMDB 获取剧集详情失败", zap.Error(err))
		return nil, err
	}
	if show == nil {
		log.Warn("TMDB 返回空剧集信息")
		return nil, Err("tv_not_found")
	}

	// 尝试从 TVDB 补充 TVDBID（如果 TVDB 可用）
	if s.tvdb != nil {
		tvdbShow, err := s.tvdb.GetTVByTMDB(ctx, tmdbID, "")
		if err == nil && tvdbShow != nil && tvdbShow.TVDBID != nil {
			log.Info("TVDB 补充成功", zap.Int("tvdb_id", *tvdbShow.TVDBID))
			show.TVDBID = tvdbShow.TVDBID
		} else {
			// TVDB 失败不影响主流程，仅记录 Warn
			log.Warn("TVDB 补充失败或未实现", zap.Error(err))
		}
	}

	log.Info("剧集元数据聚合完成",
		zap.String("title", show.Title),
		zap.Int("year", show.Year))

	return &AggregatedTVShow{Primary: show}, nil
}

// 领域错误定义

// ErrInvalidTMDBID TMDB ID 非法
var ErrInvalidTMDBID = Err("invalid_tmdb_id")

// ErrMovieNotFound 未找到电影
var ErrMovieNotFound = Err("movie_not_found")

// ErrProviderNotAvailable 指定 Provider 不可用
func ErrProviderNotAvailable(name string) error {
	return Err("provider_not_available:" + name)
}

// Err 简单的字符串错误类型
type Err string

func (e Err) Error() string { return string(e) }
