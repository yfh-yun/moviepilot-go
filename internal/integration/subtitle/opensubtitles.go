package subtitle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// OpenSubtitlesProvider OpenSubtitles提供者
type OpenSubtitlesProvider struct {
	config   map[string]interface{}
	client   *http.Client
	baseURL  string
	userAgent string
}

// OpenSubtitlesConfig OpenSubtitles配置
type OpenSubtitlesConfig struct {
	APIKey    string `json:"api_key"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	UserAgent string `json:"user_agent"`
}

// OpenSubtitlesSearchResponse OpenSubtitles搜索响应
type OpenSubtitlesSearchResponse struct {
	Data []OpenSubtitlesSubtitle `json:"data"`
}

// OpenSubtitlesSubtitle OpenSubtitles字幕
type OpenSubtitlesSubtitle struct {
	ID         string    `json:"id"`
	Attributes struct {
		SubtitleID    string  `json:"subtitle_id"`
		Language      string  `json:"language"`
		DownloadCount int     `json:"download_count"`
		Rating        float64 `json:"rating"`
		UploadDate    string  `json:"upload_date"`
		Release       string  `json:"release"`
		Comments      string  `json:"comments"`
		HearingImpaired bool  `json:"hearing_impaired"`
		HD            bool    `json:"hd"`
		FPS           float64 `json:"fps"`
		FromTrusted   bool    `json:"from_trusted"`
		Uploader      struct {
			UploaderID int    `json:"uploader_id"`
			Name       string `json:"name"`
			Rank       string `json:"rank"`
		} `json:"uploader"`
		FeatureDetails struct {
			FeatureID    int    `json:"feature_id"`
			FeatureType  string `json:"feature_type"`
			Year         int    `json:"year"`
			Title        string `json:"title"`
			MovieName    string `json:"movie_name"`
			IMDBID       int    `json:"imdb_id"`
			IMDBIDString string `json:"imdb_id_string"`
			TMDbID       int    `json:"tmdb_id"`
			SeasonNumber int    `json:"season_number"`
			EpisodeNumber int   `json:"episode_number"`
			ParentIMDBID int    `json:"parent_imdb_id"`
			ParentTitle  string `json:"parent_title"`
			ParentTMDbID int    `json:"parent_tmdb_id"`
		} `json:"feature_details"`
		Files []struct {
			FileID   int    `json:"file_id"`
			FileName string `json:"file_name"`
			CDNumber int    `json:"cd_number"`
			FileSize int64  `json:"file_size"`
		} `json:"files"`
		Link      string `json:"link"`
		Downloads int    `json:"downloads"`
		Points    int    `json:"points"`
	} `json:"attributes"`
}

// OpenSubtitlesDownloadResponse OpenSubtitles下载响应
type OpenSubtitlesDownloadResponse struct {
	Data struct {
		Link           string `json:"link"`
		RemainingDownloads int `json:"remaining_downloads"`
		Message        string `json:"message"`
		ResetTime      string `json:"reset_time"`
		ResetTimeUTC   string `json:"reset_time_utc"`
	} `json:"data"`
}

// NewOpenSubtitlesProvider 创建OpenSubtitles提供者
func NewOpenSubtitlesProvider(config map[string]interface{}) *OpenSubtitlesProvider {
	userAgent := "MoviePilot v1.0"
	if ua, ok := config["user_agent"].(string); ok && ua != "" {
		userAgent = ua
	}

	return &OpenSubtitlesProvider{
		config:   config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  "https://api.opensubtitles.com/api/v1",
		userAgent: userAgent,
	}
}

// Name 返回提供者名称
func (p *OpenSubtitlesProvider) Name() string {
	return "opensubtitles"
}

// Search 搜索字幕
func (p *OpenSubtitlesProvider) Search(ctx context.Context, req *SearchRequest) ([]*Subtitle, error) {
	// 构建搜索参数
	params := url.Values{}
	
	if req.Title != "" {
		params.Set("query", req.Title)
	}
	
	if req.Year > 0 {
		params.Set("year", fmt.Sprintf("%d", req.Year))
	}
	
	if req.Season > 0 && req.Episode > 0 {
		params.Set("season_number", fmt.Sprintf("%d", req.Season))
		params.Set("episode_number", fmt.Sprintf("%d", req.Episode))
	}
	
	if req.Language != "" {
		params.Set("languages", req.Language)
	}
	
	if req.FileHash != "" {
		params.Set("moviehash", req.FileHash)
	}
	
	if req.FileSize > 0 {
		params.Set("moviebytesize", fmt.Sprintf("%d", req.FileSize))
	}

	// 发送搜索请求
	searchURL := fmt.Sprintf("%s/subtitles?%s", p.baseURL, params.Encode())
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	p.setHeaders(httpReq)
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search subtitles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenSubtitles API returned status %d", resp.StatusCode)
	}

	var searchResp OpenSubtitlesSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// 转换为通用格式
	var subtitles []*Subtitle
	for _, item := range searchResp.Data {
		subtitle := &Subtitle{
			ID:        item.Attributes.SubtitleID,
			Provider:  p.Name(),
			Title:     item.Attributes.FeatureDetails.Title,
			Year:      item.Attributes.FeatureDetails.Year,
			Season:    item.Attributes.FeatureDetails.SeasonNumber,
			Episode:   item.Attributes.FeatureDetails.EpisodeNumber,
			Language:  item.Attributes.Language,
			Rate:      item.Attributes.Rating,
			Downloads: item.Attributes.DownloadCount,
			FileSize:  0, // 需要下载时获取
			Encoding:  "utf-8",
			Extra: map[string]string{
				"release":      item.Attributes.Release,
				"comments":     item.Attributes.Comments,
				"uploader":     item.Attributes.Uploader.Name,
				"hd":           fmt.Sprintf("%t", item.Attributes.HD),
				"fps":          fmt.Sprintf("%.2f", item.Attributes.FPS),
				"from_trusted": fmt.Sprintf("%t", item.Attributes.FromTrusted),
			},
		}

		// 解析上传时间
		if item.Attributes.UploadDate != "" {
			if uploadTime, err := time.Parse(time.RFC3339, item.Attributes.UploadDate); err == nil {
				subtitle.UploadDate = uploadTime
			}
		}

		// 获取文件大小
		if len(item.Attributes.Files) > 0 {
			subtitle.FileSize = item.Attributes.Files[0].FileSize
		}

		subtitles = append(subtitles, subtitle)
	}

	return subtitles, nil
}

// Download 下载字幕
func (p *OpenSubtitlesProvider) Download(ctx context.Context, subtitle *Subtitle) ([]byte, error) {
	// 构建下载请求
	downloadURL := fmt.Sprintf("%s/subtitles/%s/download", p.baseURL, subtitle.ID)
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	p.setHeaders(httpReq)
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get download link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenSubtitles download API returned status %d", resp.StatusCode)
	}

	var downloadResp OpenSubtitlesDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&downloadResp); err != nil {
		return nil, fmt.Errorf("failed to decode download response: %w", err)
	}

	if downloadResp.Data.Link == "" {
		return nil, fmt.Errorf("no download link provided")
	}

	// 下载实际文件
	return p.downloadFile(ctx, downloadResp.Data.Link)
}

// downloadFile 下载文件
func (p *OpenSubtitlesProvider) downloadFile(ctx context.Context, fileURL string) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create file download request: %w", err)
	}

	p.setHeaders(httpReq)
	
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file download returned status %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return buf.Bytes(), nil
}

// Test 测试连接
func (p *OpenSubtitlesProvider) Test(ctx context.Context) error {
	// 简单的搜索测试
	req := &SearchRequest{
		Title:    "test",
		Language: "en",
	}

	_, err := p.Search(ctx, req)
	if err != nil {
		return fmt.Errorf("OpenSubtitles test failed: %w", err)
	}

	logger.Info("OpenSubtitles provider test successful")
	return nil
}

// setHeaders 设置请求头
func (p *OpenSubtitlesProvider) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Api-Key", p.getAPIKey())
}

// getAPIKey 获取API密钥
func (p *OpenSubtitlesProvider) getAPIKey() string {
	if apiKey, ok := p.config["api_key"].(string); ok {
		return apiKey
	}
	return ""
}

// getLanguageCode 获取语言代码
func (p *OpenSubtitlesProvider) getLanguageCode(language string) string {
	languageMap := map[string]string{
		"zh": "zh-cn",
		"zh-cn": "zh-cn",
		"zh-tw": "zh-tw",
		"en": "en",
		"ja": "ja",
		"ko": "ko",
		"fr": "fr",
		"de": "de",
		"es": "es",
		"it": "it",
		"pt": "pt",
		"ru": "ru",
		"ar": "ar",
		"hi": "hi",
		"th": "th",
		"vi": "vi",
	}
	
	if code, ok := languageMap[language]; ok {
		return code
	}
	return language
}