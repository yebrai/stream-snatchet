package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yebrai/stream-snatchet/pkg/models"
)

type Downloader struct {
	client   *http.Client
	config   *models.Config
	progress *models.DownloadProgress
}

type SegmentResult struct {
	Index    int
	Filename string
	Error    error
}

func New(config *models.Config) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
		config:   config,
		progress: &models.DownloadProgress{},
	}
}

func (d *Downloader) DownloadSegments(streamInfo *models.StreamInfo, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	d.progress = &models.DownloadProgress{
		TotalSegments: len(streamInfo.Segments),
		Status:        "Initializing download...",
	}

	semaphore := make(chan struct{}, d.config.MaxConcurrency)
	results := make(chan SegmentResult, len(streamInfo.Segments))
	var wg sync.WaitGroup

	startTime := time.Now()

	for _, segment := range streamInfo.Segments {
		wg.Add(1)
		go func(seg models.Segment) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := SegmentResult{
				Index:    seg.Index,
				Filename: seg.Filename,
			}

			filePath := filepath.Join(outputDir, seg.Filename)
			if err := d.downloadSegmentWithRetry(seg.URL, filePath, streamInfo.Headers); err != nil {
				result.Error = err
			}

			results <- result
		}(segment)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	completed := 0
	failed := 0
	segmentFiles := make([]string, len(streamInfo.Segments))

	for result := range results {
		if result.Error != nil {
			failed++
			if d.config.Verbose {
				fmt.Printf("Failed to download segment %d: %v\n", result.Index, result.Error)
			}
		} else {
			completed++
			segmentFiles[result.Index] = result.Filename
		}

		elapsed := time.Since(startTime)
		remaining := len(streamInfo.Segments) - completed - failed
		var eta time.Duration
		if completed > 0 {
			avgTime := elapsed / time.Duration(completed)
			eta = avgTime * time.Duration(remaining)
		}

		speed := fmt.Sprintf("%.1f seg/s", float64(completed)/elapsed.Seconds())
		status := fmt.Sprintf("Downloaded %d/%d segments (%.1f%%) - %s",
			completed, len(streamInfo.Segments),
			float64(completed)/float64(len(streamInfo.Segments))*100,
			speed)

		d.progress.Update(completed, completed+failed, status)

		if d.config.Verbose {
			fmt.Printf("\r%s - ETA: %v", status, eta.Round(time.Second))
		}
	}

	if d.config.Verbose {
		fmt.Println()
	}

	if failed > 0 {
		return fmt.Errorf("failed to download %d out of %d segments", failed, len(streamInfo.Segments))
	}

	streamInfo.Segments = streamInfo.Segments[:0]
	for i, filename := range segmentFiles {
		if filename != "" {
			streamInfo.Segments = append(streamInfo.Segments, models.Segment{
				Index:    i,
				Filename: filename,
			})
		}
	}

	return nil
}

func (d *Downloader) downloadSegmentWithRetry(url, filePath string, headers map[string]string) error {
	var lastErr error

	for attempt := 0; attempt < d.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		if err := d.downloadSegment(url, filePath, headers); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", d.config.RetryAttempts, lastErr)
}

func (d *Downloader) downloadSegment(url, filePath string, headers map[string]string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", d.config.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func (d *Downloader) GetProgress() *models.DownloadProgress {
	return d.progress
}
