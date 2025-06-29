package extractor

import (
	"testing"

	"github.com/yebrai/stream-snatchet/pkg/models"
)

func TestNewExtractor(t *testing.T) {
	config := models.DefaultConfig()
	ext := New(config)

	if ext == nil {
		t.Fatal("Expected extractor to be created, got nil")
	}

	if ext.config != config {
		t.Error("Expected config to be set correctly")
	}

	if ext.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestGetBaseURL(t *testing.T) {
	config := models.DefaultConfig()
	ext := New(config)

	tests := []struct {
		manifestURL string
		expected    string
	}{
		{
			manifestURL: "https://example.com/video/playlist.m3u8",
			expected:    "https://example.com/video/",
		},
		{
			manifestURL: "https://example.com/video/subdir/playlist.m3u8?token=123",
			expected:    "https://example.com/video/subdir/",
		},
		{
			manifestURL: "https://example.com/playlist.m3u8",
			expected:    "https://example.com/",
		},
	}

	for _, test := range tests {
		result := ext.getBaseURL(test.manifestURL)
		if result != test.expected {
			t.Errorf("getBaseURL(%s) = %s, expected %s", test.manifestURL, result, test.expected)
		}
	}
}

func TestFindManifestURL(t *testing.T) {
	config := models.DefaultConfig()
	ext := New(config)

	tests := []struct {
		name        string
		content     string
		baseURL     string
		expectError bool
	}{
		{
			name:        "Valid manifest URL in quotes",
			content:     `<script>var videoUrl = "https://example.com/video.m3u8";</script>`,
			baseURL:     "https://example.com",
			expectError: false,
		},
		{
			name:        "Valid manifest URL with source property",
			content:     `source: "https://example.com/playlist.m3u8"`,
			baseURL:     "https://example.com",
			expectError: false,
		},
		{
			name:        "No manifest URL found",
			content:     `<div>Some random content without manifest URL</div>`,
			baseURL:     "https://example.com",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ext.findManifestURL(test.content, test.baseURL)

			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestParseManifest(t *testing.T) {
	config := models.DefaultConfig()
	ext := New(config)

	manifestContent := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
segment001.ts
#EXTINF:10.0,
segment002.ts
#EXTINF:5.0,
segment003.ts
#EXT-X-ENDLIST`

	streamInfo := &models.StreamInfo{
		BaseURL: "https://example.com/video/",
	}

	err := ext.parseManifest(manifestContent, streamInfo)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(streamInfo.Segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(streamInfo.Segments))
	}

	expectedURLs := []string{
		"https://example.com/video/segment001.ts",
		"https://example.com/video/segment002.ts",
		"https://example.com/video/segment003.ts",
	}

	for i, segment := range streamInfo.Segments {
		if segment.URL != expectedURLs[i] {
			t.Errorf("Segment %d URL = %s, expected %s", i, segment.URL, expectedURLs[i])
		}

		if segment.Index != i {
			t.Errorf("Segment %d Index = %d, expected %d", i, segment.Index, i)
		}
	}

	expectedDuration := 25.0
	if streamInfo.Duration.Seconds() != expectedDuration {
		t.Errorf("Total duration = %v seconds, expected %v", streamInfo.Duration.Seconds(), expectedDuration)
	}
}
