package merger

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yebrai/stream-snatchet/pkg/models"
)

type Merger struct {
	config *models.Config
}

func New(config *models.Config) *Merger {
	return &Merger{
		config: config,
	}
}

func (m *Merger) MergeSegments(streamInfo *models.StreamInfo, segmentsDir, outputPath string) error {
	if err := m.checkFFmpegInstalled(); err != nil {
		return fmt.Errorf("ffmpeg not available: %w", err)
	}

	listFile := filepath.Join(segmentsDir, "segments.txt")
	if err := m.createSegmentsList(streamInfo.Segments, segmentsDir, listFile); err != nil {
		return fmt.Errorf("failed to create segments list: %w", err)
	}
	defer os.Remove(listFile)

	if err := m.mergeWithFFmpeg(listFile, outputPath); err != nil {
		return fmt.Errorf("failed to merge segments: %w", err)
	}

	if err := m.cleanupSegments(streamInfo.Segments, segmentsDir); err != nil && m.config.Verbose {
		fmt.Printf("Warning: failed to cleanup segments: %v\n", err)
	}

	return nil
}

func (m *Merger) checkFFmpegInstalled() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH. Please install ffmpeg to merge video segments")
	}
	return nil
}

func (m *Merger) createSegmentsList(segments []models.Segment, segmentsDir, listFile string) error {
	file, err := os.Create(listFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, segment := range segments {
		segmentPath := filepath.Join(segmentsDir, segment.Filename)
		if _, err := os.Stat(segmentPath); err != nil {
			if m.config.Verbose {
				fmt.Printf("Warning: segment file not found: %s\n", segmentPath)
			}
			continue
		}

		// Convert to absolute path and escape for concat demuxer
		absPath, err := filepath.Abs(segmentPath)
		if err != nil {
			absPath = segmentPath
		}

		_, err = writer.WriteString(fmt.Sprintf("file '%s'\n", strings.ReplaceAll(absPath, "'", "'\\''")))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Merger) mergeWithFFmpeg(listFile, outputPath string) error {
	if m.config.Verbose {
		fmt.Printf("Merging segments with ffmpeg...\n")
	}

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		"-avoid_negative_ts", "make_zero",
		"-fflags", "+genpts",
		"-y",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)

	if !m.config.Verbose {
		cmd.Stderr = nil
		cmd.Stdout = nil
	} else {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %w", err)
	}

	if m.config.Verbose {
		fmt.Printf("Successfully merged video to: %s\n", outputPath)
	}

	return nil
}

func (m *Merger) cleanupSegments(segments []models.Segment, segmentsDir string) error {
	for _, segment := range segments {
		segmentPath := filepath.Join(segmentsDir, segment.Filename)
		if err := os.Remove(segmentPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (m *Merger) GenerateOutputFilename(streamInfo *models.StreamInfo, outputDir string) string {
	title := streamInfo.Title
	if title == "" {
		title = "video"
	}

	title = strings.ReplaceAll(title, " ", "_")
	title = strings.ReplaceAll(title, "/", "_")
	title = strings.ReplaceAll(title, "\\", "_")
	title = strings.ReplaceAll(title, ":", "_")
	title = strings.ReplaceAll(title, "*", "_")
	title = strings.ReplaceAll(title, "?", "_")
	title = strings.ReplaceAll(title, "\"", "_")
	title = strings.ReplaceAll(title, "<", "_")
	title = strings.ReplaceAll(title, ">", "_")
	title = strings.ReplaceAll(title, "|", "_")

	if len(title) > 100 {
		title = title[:100]
	}

	return filepath.Join(outputDir, fmt.Sprintf("%s.mp4", title))
}
