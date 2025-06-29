package merger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yebrai/stream-snatchet/pkg/models"
)

func TestNewMerger(t *testing.T) {
	config := models.DefaultConfig()
	merger := New(config)

	if merger == nil {
		t.Fatal("Expected merger to be created, got nil")
	}

	if merger.config != config {
		t.Error("Expected config to be set correctly")
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	config := models.DefaultConfig()
	merger := New(config)

	tests := []struct {
		name       string
		streamInfo *models.StreamInfo
		outputDir  string
		expected   string
	}{
		{
			name: "Basic filename",
			streamInfo: &models.StreamInfo{
				Title: "Test Video",
			},
			outputDir: "/tmp",
			expected:  "/tmp/Test_Video.mp4",
		},
		{
			name: "Filename with special characters",
			streamInfo: &models.StreamInfo{
				Title: "Test/Video:With*Special?Characters",
			},
			outputDir: "/tmp",
			expected:  "/tmp/Test_Video_With_Special_Characters.mp4",
		},
		{
			name: "Empty title",
			streamInfo: &models.StreamInfo{
				Title: "",
			},
			outputDir: "/tmp",
			expected:  "/tmp/video.mp4",
		},
		{
			name: "Very long title",
			streamInfo: &models.StreamInfo{
				Title: "This is a very long title that should be truncated because it exceeds the maximum length limit",
			},
			outputDir: "/tmp",
			expected:  "/tmp/This_is_a_very_long_title_that_should_be_truncated_because_it_exceeds_the_maximum_length_limit.mp4",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := merger.GenerateOutputFilename(test.streamInfo, test.outputDir)
			if result != test.expected {
				t.Errorf("GenerateOutputFilename() = %s, expected %s", result, test.expected)
			}
		})
	}
}

func TestCreateSegmentsList(t *testing.T) {
	config := models.DefaultConfig()
	merger := New(config)

	tempDir := t.TempDir()

	segments := []models.Segment{
		{Index: 0, Filename: "segment_0000.ts"},
		{Index: 1, Filename: "segment_0001.ts"},
		{Index: 2, Filename: "segment_0002.ts"},
	}

	for _, segment := range segments {
		segmentPath := filepath.Join(tempDir, segment.Filename)
		file, err := os.Create(segmentPath)
		if err != nil {
			t.Fatalf("Failed to create test segment file: %v", err)
		}
		file.Close()
	}

	listFile := filepath.Join(tempDir, "segments.txt")

	err := merger.createSegmentsList(segments, tempDir, listFile)
	if err != nil {
		t.Fatalf("createSegmentsList failed: %v", err)
	}

	if _, err := os.Stat(listFile); os.IsNotExist(err) {
		t.Error("Expected segments list file to be created")
	}

	content, err := os.ReadFile(listFile)
	if err != nil {
		t.Fatalf("Failed to read segments list file: %v", err)
	}

	expectedContent := "file '" + filepath.Join(tempDir, "segment_0000.ts") + "'\n" +
		"file '" + filepath.Join(tempDir, "segment_0001.ts") + "'\n" +
		"file '" + filepath.Join(tempDir, "segment_0002.ts") + "'\n"

	if string(content) != expectedContent {
		t.Errorf("Segments list content mismatch.\nExpected:\n%s\nGot:\n%s", expectedContent, string(content))
	}
}
