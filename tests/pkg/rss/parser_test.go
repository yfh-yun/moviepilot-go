package rss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"moviepilot-go/pkg/rss"
)

func TestNewParser(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := rss.Config{
		Logger: logger,
	}

	parser := rss.NewParser(config)
	assert.NotNil(t, parser)
}

func TestParseXML(t *testing.T) {
	logger := zaptest.NewLogger(t)
	parser := rss.NewParser(rss.Config{Logger: logger})

	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test RSS Feed</title>
    <link>https://example.com</link>
    <description>Test Description</description>
    <item>
      <title>Test.Movie.2023.1080p.BluRay.x264-GROUP</title>
      <link>https://example.com/download/1</link>
      <description>Test movie description</description>
      <pubDate>Mon, 01 Jan 2023 12:00:00 +0000</pubDate>
      <guid>test-guid-1</guid>
      <enclosure url="https://example.com/torrent/1.torrent" length="1073741824" type="application/x-bittorrent"/>
    </item>
  </channel>
</rss>`

	feed, err := parser.ParseXML([]byte(xmlData))

	assert.NoError(t, err)
	assert.NotNil(t, feed)
	assert.Equal(t, "Test RSS Feed", feed.Channel.Title)
	assert.Len(t, feed.Channel.Items, 1)
	assert.Equal(t, "Test.Movie.2023.1080p.BluRay.x264-GROUP", feed.Channel.Items[0].Title)
}

func TestParsePubDate(t *testing.T) {
	testCases := []struct {
		name    string
		pubDate string
		wantErr bool
	}{
		{
			name:    "RFC1123Z",
			pubDate: "Mon, 02 Jan 2006 15:04:05 -0700",
			wantErr: false,
		},
		{
			name:    "RFC1123",
			pubDate: "Mon, 02 Jan 2006 15:04:05 MST",
			wantErr: false,
		},
		{
			name:    "ISO8601",
			pubDate: "2006-01-02T15:04:05Z",
			wantErr: false,
		},
		{
			name:    "Invalid",
			pubDate: "invalid date",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := rss.ParsePubDate(tc.pubDate)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
