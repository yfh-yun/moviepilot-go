package subscribe

import (
	"context"
	"strconv"

	"moviepilot-go/internal/models/database"
	repoif "moviepilot-go/internal/repositories/interfaces"
)

// DBSubscribeRepository 是对底层 interfaces.SubscribeRepository 的业务层适配器，
// 用于实现本包定义的 SubscribeRepository 接口。
type DBSubscribeRepository struct {
	repo repoif.SubscribeRepository
}

// NewDBSubscribeRepository 创建一个基于数据库仓储的业务层 SubscribeRepository 实现。
func NewDBSubscribeRepository(repo repoif.SubscribeRepository) SubscribeRepository {
	return &DBSubscribeRepository{repo: repo}
}

// mapDBSubscribeToService 将数据库订阅模型映射到业务层订阅模型
func mapDBSubscribeToService(dbSub *database.Subscribe) *Subscribe {
	if dbSub == nil {
		return nil
	}

	doubanID := ""
	if dbSub.DoubanID != nil {
		doubanID = *dbSub.DoubanID
	}

	season := 0
	if dbSub.Season != nil {
		season = *dbSub.Season
	}

	bestVersion := false
	if dbSub.BestVersion > 0 {
		bestVersion = true
	}

	searchIMDBID := false
	if dbSub.SearchIMDBID > 0 {
		searchIMDBID = true
	}

	return &Subscribe{
		ID:              int(dbSub.ID),
		Name:            dbSub.Name,
		Type:            dbSub.Type,
		Keyword:         dbSub.Keyword,
		TMDBID:          dbSub.TMDBID,
		DoubanID:        doubanID,
		BangumiID:       dbSub.BangumiID,
		Season:          season,
		Poster:          dbSub.Poster,
		Backdrop:        dbSub.Backdrop,
		Description:     dbSub.Description,
		Include:         dbSub.Include,
		Exclude:         dbSub.Exclude,
		Quality:         dbSub.Quality,
		Resolution:      dbSub.Resolution,
		Effect:          dbSub.Effect,
		State:           dbSub.State,
		Username:        dbSub.Username,
		BestVersion:     bestVersion,
		CurrentPriority: dbSub.CurrentPriority,
		SavePath:        dbSub.SavePath,
		SearchIMDBID:    searchIMDBID,
		CustomWords:     dbSub.CustomWords,
		MediaCategory:   dbSub.MediaCategory,
		EpisodeGroup:    dbSub.EpisodeGroup,
	}
}

// Add 添加订阅
func (r *DBSubscribeRepository) Add(ctx context.Context, sub *Subscribe) (int, error) {
	doubanID := ""
	if sub.DoubanID != "" {
		doubanID = sub.DoubanID
	}

	bestVersion := 0
	if sub.BestVersion {
		bestVersion = 1
	}

	searchIMDBID := 0
	if sub.SearchIMDBID {
		searchIMDBID = 1
	}

	season := sub.Season

	dbSub := &database.Subscribe{
		Name:            sub.Name,
		Type:            sub.Type,
		Keyword:         sub.Keyword,
		TMDBID:          sub.TMDBID,
		DoubanID:        &doubanID,
		BangumiID:       sub.BangumiID,
		Season:          &season,
		Poster:          sub.Poster,
		Backdrop:        sub.Backdrop,
		Description:     sub.Description,
		Include:         sub.Include,
		Exclude:         sub.Exclude,
		Quality:         sub.Quality,
		Resolution:      sub.Resolution,
		Effect:          sub.Effect,
		State:           sub.State,
		Username:        sub.Username,
		BestVersion:     bestVersion,
		CurrentPriority: sub.CurrentPriority,
		SavePath:        sub.SavePath,
		SearchIMDBID:    searchIMDBID,
		CustomWords:     sub.CustomWords,
		MediaCategory:   sub.MediaCategory,
		EpisodeGroup:    sub.EpisodeGroup,
	}

	err := r.repo.Create(ctx, dbSub)
	if err != nil {
		return 0, err
	}

	return int(dbSub.ID), nil
}

// Get 获取订阅
func (r *DBSubscribeRepository) Get(ctx context.Context, id int) (*Subscribe, error) {
	dbSub, err := r.repo.GetByID(ctx, strconv.Itoa(id))
	if err != nil {
		return nil, err
	}

	return mapDBSubscribeToService(dbSub), nil
}

// List 获取订阅列表
func (r *DBSubscribeRepository) List(ctx context.Context, states []string) ([]*Subscribe, error) {
	// For now, we'll just return all active subscriptions
	// We need to implement proper state filtering in the repository
	dbSubs, err := r.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	subs := make([]*Subscribe, len(dbSubs))
	for i, dbSub := range dbSubs {
		subs[i] = mapDBSubscribeToService(dbSub)
	}

	return subs, nil
}

// Update 更新订阅
func (r *DBSubscribeRepository) Update(ctx context.Context, id int, updates map[string]any) error {
	return r.repo.Update(ctx, &database.Subscribe{BaseModel: database.BaseModel{ID: uint(id)}, Name: ""}) // This is a temporary fix
}

// Delete 删除订阅
func (r *DBSubscribeRepository) Delete(ctx context.Context, id int) error {
	return r.repo.Delete(ctx, strconv.Itoa(id))
}

// Exists 判断订阅是否存在
func (r *DBSubscribeRepository) Exists(ctx context.Context, tmdbID *int, doubanID string, season int) (bool, error) {
	var dbDoubanID *string
	if doubanID != "" {
		dbDoubanID = &doubanID
	}

	return r.repo.Exists(ctx, tmdbID, dbDoubanID, &season)
}

// ExistHistory 判断订阅历史是否存在
func (r *DBSubscribeRepository) ExistHistory(ctx context.Context, tmdbID *int, doubanID string, season int) (bool, error) {
	// Implement this if needed
	return false, nil
}

// AddHistory 添加订阅历史
func (r *DBSubscribeRepository) AddHistory(ctx context.Context, sub *Subscribe) error {
	// Implement this if needed
	return nil
}
