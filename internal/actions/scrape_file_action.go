package actions

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/media"
	"moviepilot-go/internal/models"
)

// ScrapeFileParams 控制刮削行为。
type ScrapeFileParams struct {
	BaseParams
	ForceRefresh bool   `mapstructure:"force_refresh"`
	Source       string `mapstructure:"source"`
}

// Validate 保留扩展点。
func (p *ScrapeFileParams) Validate() error {
	return nil
}

// ScrapeFileAction 将扫描结果转为媒体元信息。
type ScrapeFileAction struct {
	BaseActionImpl
	mediaService media.Service
}

// NewScrapeFileAction 构造实例。
func NewScrapeFileAction(logger *zap.Logger, mediaSvc media.Service) *ScrapeFileAction {
	return &ScrapeFileAction{
		BaseActionImpl: NewBaseActionImpl("scrape_file", "Scrape metadata for files", logger),
		mediaService:   mediaSvc,
	}
}

// Execute 基于文件名生成基础媒体信息（后续可接入真正的媒体识别服务）。
func (a *ScrapeFileAction) Execute(workflowID int, rawParams any, ctx *ActionContext) (*ActionContext, error) {
	if ctx == nil {
		ctx = &ActionContext{}
	}
	ctx.Ensure()

	params := &ScrapeFileParams{}
	if err := DecodeParams(rawParams, params); err != nil {
		a.SetResult(false, err.Error(), nil)
		return ctx, err
	}

	if len(ctx.Files) == 0 {
		msg := "no files in context to scrape"
		a.SetResult(false, msg, nil)
		return ctx, errors.New(msg)
	}

	medias := make([]models.Media, 0, len(ctx.Files))

	if a.mediaService != nil {
		fileItems := make([]media.FileItem, 0, len(ctx.Files))
		for _, file := range ctx.Files {
			fileItems = append(fileItems, media.FileItem{Path: file.Path})
		}

		serviceMedias, err := a.mediaService.Identify(fileItems, media.IdentifyOptions{
			ForceRefresh: params.ForceRefresh,
			Source:       params.Source,
		})
		if err != nil {
			a.SetResult(false, err.Error(), nil)
			return ctx, err
		}
		medias = serviceMedias
	} else {
		now := time.Now()
		for _, file := range ctx.Files {
			guess := guessTitle(file.Path)
			mediaType := guessTypeFromPath(file.Path)

			medias = append(medias, models.Media{
				BaseModel: models.BaseModel{
					CreatedAt: now,
					UpdatedAt: now,
				},
				Title:       guess,
				Type:        mediaType,
				Description: "placeholder metadata — integrate real scraper",
				Year:        nil,
				Season:      nil,
				Episode:     nil,
			})
		}
	}

	ctx.Medias = medias
	ctx.UpdatedAt = time.Now()
	a.SaveCache(workflowID, medias)
	a.SetResult(true, "scrape completed", medias)
	return ctx, nil
}

var seasonRegex = regexp.MustCompile(`(?i)s(\d{1,2})e(\d{1,2})`)

func guessTitle(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if idx := strings.Index(base, "["); idx > 0 {
		base = base[:idx]
	}
	if idx := strings.Index(base, "("); idx > 0 {
		base = base[:idx]
	}
	base = strings.TrimSpace(strings.ReplaceAll(base, ".", " "))
	if base == "" {
		return filepath.Base(path)
	}
	return base
}

func guessTypeFromPath(path string) string {
	lower := strings.ToLower(path)
	if seasonRegex.MatchString(lower) {
		return "tv"
	}
	if strings.Contains(lower, "s0") && strings.Contains(lower, "e") {
		return "tv"
	}
	return "movie"
}
