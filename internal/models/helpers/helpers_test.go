package helpers

import (
	"testing"

	. "moviepilot-go/internal/models/dto"
	. "moviepilot-go/internal/models/types"
)

func TestIntPtr(t *testing.T) {
	val := 42
	ptr := IntPtr(val)
	if ptr == nil {
		t.Error("IntPtr returned nil")
	}
	if *ptr != val {
		t.Errorf("Expected %d, got %d", val, *ptr)
	}
}

func TestStringPtr(t *testing.T) {
	val := "test"
	ptr := StringPtr(val)
	if ptr == nil {
		t.Error("StringPtr returned nil")
	}
	if *ptr != val {
		t.Errorf("Expected %s, got %s", val, *ptr)
	}
}

func TestIntValue(t *testing.T) {
	// Test with non-nil pointer
	val := 42
	ptr := &val
	result := IntValue(ptr, 0)
	if result != val {
		t.Errorf("Expected %d, got %d", val, result)
	}

	// Test with nil pointer
	result = IntValue(nil, 99)
	if result != 99 {
		t.Errorf("Expected 99, got %d", result)
	}
}

func TestIsValidMediaType(t *testing.T) {
	tests := []struct {
		mediaType string
		expected  bool
	}{
		{string(MediaTypeMovie), true},
		{string(MediaTypeTV), true},
		{string(MediaTypeCollection), true},
		{string(MediaTypeUnknown), true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidMediaType(tt.mediaType)
		if result != tt.expected {
			t.Errorf("IsValidMediaType(%s) = %v, expected %v", tt.mediaType, result, tt.expected)
		}
	}
}

func TestIsValidEventType(t *testing.T) {
	tests := []struct {
		eventType string
		expected  bool
	}{
		{string(EventTypeDownloadAdded), true},
		{string(EventTypeSubscribeComplete), true},
		{string(EventTypeTransferComplete), true},
		{"invalid.event", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidEventType(tt.eventType)
		if result != tt.expected {
			t.Errorf("IsValidEventType(%s) = %v, expected %v", tt.eventType, result, tt.expected)
		}
	}
}

func TestGetEventTypeName(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventTypeDownloadAdded, "添加下载"},
		{EventTypeSubscribeComplete, "订阅已完成"},
		{EventTypeTransferComplete, "整理完成"},
		{EventType("invalid"), "未知事件"},
	}

	for _, tt := range tests {
		result := GetEventTypeName(tt.eventType)
		if result != tt.expected {
			t.Errorf("GetEventTypeName(%s) = %s, expected %s", tt.eventType, result, tt.expected)
		}
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := FormatFileSize(tt.size)
		if result != tt.expected {
			t.Errorf("FormatFileSize(%d) = %s, expected %s", tt.size, result, tt.expected)
		}
	}
}

func TestFormatRatio(t *testing.T) {
	tests := []struct {
		upload   int64
		download int64
		expected float64
	}{
		{1024, 1024, 1.0},
		{2048, 1024, 2.0},
		{1024, 2048, 0.5},
		{1024, 0, 999.99},
		{0, 0, 0.0},
	}

	for _, tt := range tests {
		result := FormatRatio(tt.upload, tt.download)
		if result != tt.expected {
			t.Errorf("FormatRatio(%d, %d) = %.2f, expected %.2f", tt.upload, tt.download, result, tt.expected)
		}
	}
}

func TestBuildSeasonEpisode(t *testing.T) {
	tests := []struct {
		season   int
		episode  int
		expected string
	}{
		{1, 1, "S01E01"},
		{2, 15, "S02E15"},
		{10, 99, "S10E99"},
	}

	for _, tt := range tests {
		result := BuildSeasonEpisode(tt.season, tt.episode)
		if result != tt.expected {
			t.Errorf("BuildSeasonEpisode(%d, %d) = %s, expected %s", tt.season, tt.episode, result, tt.expected)
		}
	}
}

func TestMergeEpisodeList(t *testing.T) {
	list1 := []int{1, 2, 3}
	list2 := []int{2, 3, 4}
	list3 := []int{5, 6}

	result := MergeEpisodeList(list1, list2, list3)
	expected := []int{1, 2, 3, 4, 5, 6}

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, result[i])
		}
	}
}

func TestValidateSubscribe(t *testing.T) {
	// Valid subscribe
	validSub := &Subscribe{
		Name: "Test Subscribe",
		Type: string(MediaTypeMovie),
	}
	err := ValidateSubscribe(validSub)
	if err != nil {
		t.Errorf("ValidateSubscribe failed for valid subscribe: %v", err)
	}

	// Nil subscribe
	err = ValidateSubscribe(nil)
	if err == nil {
		t.Error("ValidateSubscribe should fail for nil subscribe")
	}

	// Missing name
	invalidSub := &Subscribe{
		Type: string(MediaTypeMovie),
	}
	err = ValidateSubscribe(invalidSub)
	if err == nil {
		t.Error("ValidateSubscribe should fail for subscribe without name")
	}

	// Invalid type
	invalidSub = &Subscribe{
		Name: "Test",
		Type: "invalid",
	}
	err = ValidateSubscribe(invalidSub)
	if err == nil {
		t.Error("ValidateSubscribe should fail for invalid type")
	}
}

func TestValidateMediaInfo(t *testing.T) {
	// Valid media info
	validMedia := &MediaInfo{
		Title: "Test Movie",
		Type:  string(MediaTypeMovie),
	}
	err := ValidateMediaInfo(validMedia)
	if err != nil {
		t.Errorf("ValidateMediaInfo failed for valid media: %v", err)
	}

	// Nil media info
	err = ValidateMediaInfo(nil)
	if err == nil {
		t.Error("ValidateMediaInfo should fail for nil media")
	}

	// Missing title
	invalidMedia := &MediaInfo{
		Type: string(MediaTypeMovie),
	}
	err = ValidateMediaInfo(invalidMedia)
	if err == nil {
		t.Error("ValidateMediaInfo should fail for media without title")
	}
}

func TestToJSON(t *testing.T) {
	media := &MediaInfo{
		Title: "Test Movie",
		Type:  string(MediaTypeMovie),
		Year:  "2024",
	}

	jsonStr, err := ToJSON(media)
	if err != nil {
		t.Errorf("ToJSON failed: %v", err)
	}

	if jsonStr == "" {
		t.Error("ToJSON returned empty string")
	}
}

func TestFromJSON(t *testing.T) {
	jsonStr := `{"title":"Test Movie","type":"电影","year":"2024"}`

	var media MediaInfo
	err := FromJSON(jsonStr, &media)
	if err != nil {
		t.Errorf("FromJSON failed: %v", err)
	}

	if media.Title != "Test Movie" {
		t.Errorf("Expected title 'Test Movie', got '%s'", media.Title)
	}

	if media.Type != string(MediaTypeMovie) {
		t.Errorf("Expected type '%s', got '%s'", MediaTypeMovie, media.Type)
	}
}

func TestCloneMediaInfo(t *testing.T) {
	original := &MediaInfo{
		Title:  "Test Movie",
		Type:   string(MediaTypeMovie),
		Year:   "2024",
		TmdbID: IntPtr(12345),
	}

	cloned, err := CloneMediaInfo(original)
	if err != nil {
		t.Errorf("CloneMediaInfo failed: %v", err)
	}

	if cloned.Title != original.Title {
		t.Errorf("Cloned title doesn't match: %s != %s", cloned.Title, original.Title)
	}

	// Modify cloned and ensure original is not affected
	cloned.Title = "Modified"
	if original.Title == "Modified" {
		t.Error("Modifying clone affected original")
	}
}
