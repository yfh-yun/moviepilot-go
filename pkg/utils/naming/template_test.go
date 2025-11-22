package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"moviepilot-go/internal/models"
)

func TestParseTemplate(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		wantErr  bool
		wantVars []string
	}{
		{
			name:     "电影模板",
			template: "${title} (${year})/${title}.${year}${ext}",
			wantErr:  false,
			wantVars: []string{"title", "year", "title", "year", "ext"},
		},
		{
			name:     "电视剧模板",
			template: "${title}/Season ${season_num}/${title}.${season}${episode}${ext}",
			wantErr:  false,
			wantVars: []string{"title", "season_num", "title", "season", "episode", "ext"},
		},
		{
			name:     "空模板",
			template: "",
			wantErr:  true,
		},
		{
			name:     "无变量模板",
			template: "Movies/Test/test.mkv",
			wantErr:  false,
			wantVars: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tc.template)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tmpl)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tmpl)
				assert.Equal(t, tc.template, tmpl.Raw())
				assert.Equal(t, tc.wantVars, tmpl.Variables())
			}
		})
	}
}

func TestTemplate_Render(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		vars     TemplateVars
		want     string
	}{
		{
			name:     "电影基本渲染",
			template: "${title} (${year})/${title}.${year}${ext}",
			vars: TemplateVars{
				Title:     "The Matrix",
				Year:      "1999",
				Extension: ".mkv",
			},
			want: "The Matrix (1999)/The Matrix.1999.mkv",
		},
		{
			name:     "电视剧渲染",
			template: "${title}/Season ${season_num}/${title}.${season}${episode}${ext}",
			vars: TemplateVars{
				Title:     "Breaking Bad",
				Season:    "S01",
				SeasonNum: "1",
				Episode:   "E01",
				Extension: ".mkv",
			},
			want: "Breaking Bad/Season 1/Breaking Bad.S01E01.mkv",
		},
		{
			name:     "完整电影信息",
			template: "${title} (${year})/${title}.${year}.${resolution}.${source}.${codec}${ext}",
			vars: TemplateVars{
				Title:      "Inception",
				Year:       "2010",
				Resolution: "1080p",
				Source:     "BluRay",
				Codec:      "x264",
				Extension:  ".mkv",
			},
			want: "Inception (2010)/Inception.2010.1080p.BluRay.x264.mkv",
		},
		{
			name:     "动漫格式",
			template: "[${group}] ${title} - ${episode_num}${ext}",
			vars: TemplateVars{
				Title:      "Demon Slayer",
				Group:      "SubsPlease",
				EpisodeNum: "01",
				Extension:  ".mkv",
			},
			want: "[SubsPlease] Demon Slayer - 01.mkv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tc.template)
			assert.NoError(t, err)

			result, err := tmpl.Render(tc.vars)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "正常文件名",
			input: "The Matrix",
			want:  "The Matrix",
		},
		{
			name:  "包含非法字符",
			input: "Movie: The Title",
			want:  "Movie - The Title",
		},
		{
			name:  "包含多个非法字符",
			input: `Movie<>:"/\|?*`,
			want:  "Movie -'---", // 每个非法字符都会被替换
		},
		{
			name:  "前后空格",
			input: "  Movie Title  ",
			want:  "Movie Title",
		},
		{
			name:  "超长文件名",
			input: string(make([]byte, 300)),
			want:  string(make([]byte, 200)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeFileName(tc.input)
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestMediaToVars(t *testing.T) {
	year := "2023"
	season := 1
	episode := 5
	tmdbID := 12345
	imdbID := "tt1234567"

	media := &models.Media{
		Title:         "Test Movie",
		OriginalTitle: "Original Test Movie",
		Year:          &year,
		Type:          "movie",
		Season:        &season,
		Episode:       &episode,
		TMDBID:        &tmdbID,
		IMDBID:        &imdbID,
	}

	vars := MediaToVars(media, "/path/to/file.mkv")

	assert.Equal(t, "Test Movie", vars.Title)
	assert.Equal(t, "Original Test Movie", vars.OriginalTitle)
	assert.Equal(t, "2023", vars.Year)
	assert.Equal(t, "movie", vars.Type)
	assert.Equal(t, ".mkv", vars.Extension)
	assert.Equal(t, "S01", vars.Season)
	assert.Equal(t, "1", vars.SeasonNum)
	assert.Equal(t, "E05", vars.Episode)
	assert.Equal(t, "5", vars.EpisodeNum)
	assert.Equal(t, "12345", vars.TMDBID)
	assert.Equal(t, "tt1234567", vars.IMDBID)
}

func TestGetDefaultTemplate(t *testing.T) {
	testCases := []struct {
		mediaType string
		want      string
	}{
		{
			mediaType: "movie",
			want:      "${title} (${year})/${title}.${year}.${resolution}.${source}.${codec}${ext}",
		},
		{
			mediaType: "tv",
			want:      "${title}/Season ${season_num}/${title}.${season}${episode}.${episode_title}${ext}",
		},
		{
			mediaType: "anime",
			want:      "${title}/Season ${season_num}/[${group}] ${title} - ${episode_num}${ext}",
		},
		{
			mediaType: "unknown",
			want:      "${title} (${year})/${title}.${year}.${resolution}.${source}.${codec}${ext}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.mediaType, func(t *testing.T) {
			result := GetDefaultTemplate(tc.mediaType)
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestCleanPath(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "正常路径",
			input: "Movies/The Matrix/movie.mkv",
			want:  "Movies/The Matrix/movie.mkv",
		},
		{
			name:  "多余斜杠",
			input: "Movies//The Matrix///movie.mkv",
			want:  "Movies/The Matrix/movie.mkv",
		},
		{
			name:  "连续空格",
			input: "Movies/The  Matrix/movie.mkv",
			want:  "Movies/The Matrix/movie.mkv",
		},
		{
			name:  "包含..",
			input: "Movies/../The Matrix/movie.mkv",
			want:  "The Matrix/movie.mkv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanPath(tc.input)
			assert.Equal(t, tc.want, result)
		})
	}
}

// 基准测试
func BenchmarkParseTemplate(b *testing.B) {
	template := "${title} (${year})/${title}.${year}.${resolution}.${source}.${codec}${ext}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseTemplate(template)
	}
}

func BenchmarkTemplate_Render(b *testing.B) {
	template := "${title} (${year})/${title}.${year}.${resolution}.${source}.${codec}${ext}"
	tmpl, _ := ParseTemplate(template)

	vars := TemplateVars{
		Title:      "The Matrix",
		Year:       "1999",
		Resolution: "1080p",
		Source:     "BluRay",
		Codec:      "x264",
		Extension:  ".mkv",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tmpl.Render(vars)
	}
}

func BenchmarkSanitizeFileName(b *testing.B) {
	input := "Movie: The Title <with> special/characters"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sanitizeFileName(input)
	}
}
