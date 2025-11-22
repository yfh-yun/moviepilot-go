package rss

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"moviepilot-go/pkg/rss"
)

func TestExtractTorrentInfo(t *testing.T) {
	testCases := []struct {
		name     string
		title    string
		wantInfo *rss.TorrentInfo
	}{
		{
			name:  "电影标题",
			title: "The.Matrix.1999.1080p.BluRay.x264-GROUP",
			wantInfo: &rss.TorrentInfo{
				Year:       1999,
				Resolution: "1080p",
				Source:     "BluRay",
				Codec:      "x264",
				Group:      "GROUP",
			},
		},
		{
			name:  "电视剧标题",
			title: "Breaking.Bad.S01E01.1080p.WEB-DL.x265-GROUP",
			wantInfo: &rss.TorrentInfo{
				Season:     1,
				Episode:    1,
				Resolution: "1080p",
				Source:     "WEB-DL",
				Codec:      "x265",
				Group:      "GROUP",
			},
		},
		{
			name:  "4K标题",
			title: "Movie.2023.2160p.UHD.BluRay.x265-GROUP",
			wantInfo: &rss.TorrentInfo{
				Year:       2023,
				Resolution: "2160p",
				Source:     "BluRay",
				Codec:      "x265",
				Group:      "GROUP",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := rss.RSSItem{
				Title: tc.title,
			}

			info, err := rss.ExtractTorrentInfo(item)

			assert.NoError(t, err)
			assert.NotNil(t, info)
			assert.Equal(t, tc.title, info.Title)

			if tc.wantInfo.Year > 0 {
				assert.Equal(t, tc.wantInfo.Year, info.Year)
			}
			if tc.wantInfo.Season > 0 {
				assert.Equal(t, tc.wantInfo.Season, info.Season)
			}
			if tc.wantInfo.Episode > 0 {
				assert.Equal(t, tc.wantInfo.Episode, info.Episode)
			}
			if tc.wantInfo.Resolution != "" {
				assert.Equal(t, tc.wantInfo.Resolution, info.Resolution)
			}
			if tc.wantInfo.Source != "" {
				assert.Equal(t, tc.wantInfo.Source, info.Source)
			}
			if tc.wantInfo.Codec != "" {
				assert.Equal(t, tc.wantInfo.Codec, info.Codec)
			}
			if tc.wantInfo.Group != "" {
				assert.Equal(t, tc.wantInfo.Group, info.Group)
			}
		})
	}
}

func TestTorrentInfo_MatchesQuality(t *testing.T) {
	info := &rss.TorrentInfo{
		Quality: "1080p",
	}

	assert.True(t, info.MatchesQuality("1080p"))
	assert.False(t, info.MatchesQuality("720p"))
	assert.True(t, info.MatchesQuality("")) // 空字符串匹配任何
}

func TestTorrentInfo_MatchesSource(t *testing.T) {
	info := &rss.TorrentInfo{
		Source: "BluRay",
	}

	assert.True(t, info.MatchesSource("BluRay"))
	assert.True(t, info.MatchesSource("Blu")) // 部分匹配,不区分大小写
	assert.False(t, info.MatchesSource("WEB-DL"))
	assert.True(t, info.MatchesSource("")) // 空字符串匹配任何
}

func TestTorrentInfo_ContainsKeyword(t *testing.T) {
	info := &rss.TorrentInfo{
		Title: "The.Matrix.1999.1080p.BluRay.x264",
	}

	assert.True(t, info.ContainsKeyword("Matrix"))
	assert.True(t, info.ContainsKeyword("matrix")) // 不区分大小写
	assert.True(t, info.ContainsKeyword("1999"))
	assert.False(t, info.ContainsKeyword("Inception"))
}

func TestTorrentInfo_ExcludesKeyword(t *testing.T) {
	info := &rss.TorrentInfo{
		Title: "The.Matrix.1999.1080p.BluRay.x264",
	}

	assert.True(t, info.ExcludesKeyword("Inception")) // 不包含
	assert.False(t, info.ExcludesKeyword("Matrix"))   // 包含
	assert.False(t, info.ExcludesKeyword("1999"))
}
