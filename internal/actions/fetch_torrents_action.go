package actions

import (
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
)

// FetchTorrentsParams 描述拉取种子的可选参数。
type FetchTorrentsParams struct {
	BaseParams
	Keywords []string `mapstructure:"keywords"`
	Sites    []string `mapstructure:"sites"`
	Limit    int      `mapstructure:"limit"`
}

// Validate 限制参数范围。
func (p *FetchTorrentsParams) Validate() error {
	if p.Limit < 0 {
		p.Limit = 0
	}
	return nil
}

// FetchTorrentsAction 为链路预留的拉取种子骨架。
type FetchTorrentsAction struct {
	BaseActionImpl
}

// NewFetchTorrentsAction 创建实例。
func NewFetchTorrentsAction(logger *zap.Logger) *FetchTorrentsAction {
	return &FetchTorrentsAction{
		BaseActionImpl: NewBaseActionImpl("fetch_torrents", "Fetch torrents from configured sites", logger),
	}
}

// Execute 目前仅作占位，待接入 torrents service 后实现真实逻辑。
func (a *FetchTorrentsAction) Execute(workflowID int, rawParams any, ctx *ActionContext) (*ActionContext, error) {
	if ctx == nil {
		ctx = &ActionContext{}
	}
	ctx.Ensure()

	params := &FetchTorrentsParams{}
	if err := DecodeParams(rawParams, params); err != nil {
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	if len(params.Keywords) == 0 && len(ctx.Medias) == 0 {
		err := errors.New("fetch_torrents requires keywords or medias in context")
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	keywords := params.Keywords
	if len(keywords) == 0 {
		keywords = extractTitles(ctx.Medias)
	}

	downloads := make([]models.DownloadHistory, 0, len(keywords))
	now := time.Now()

	for _, kw := range keywords {
		downloads = append(downloads, models.DownloadHistory{
			BaseModel: models.BaseModel{
				CreatedAt: now,
				UpdatedAt: now,
			},
			Title: kw,
			Type:  "unknown",
			Note:  `{"placeholder":true}`,
		})
	}

	ctx.Downloads = downloads
	ctx.UpdatedAt = time.Now()
	a.SaveCache(workflowID, downloads)
	a.SetResult(true, "fetch torrents placeholder executed", downloads)
	return ctx, nil
}

func extractTitles(medias []models.Media) []string {
	titles := make([]string, 0, len(medias))
	for _, media := range medias {
		title := strings.TrimSpace(media.Title)
		if title != "" {
			titles = append(titles, title)
		}
	}
	return titles
}
