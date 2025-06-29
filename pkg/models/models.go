package models

import (
	"sync"
	"time"
)

type StreamInfo struct {
	IframeURL   string
	ManifestURL string
	BaseURL     string
	Title       string
	Duration    time.Duration
	Quality     string
	Segments    []Segment
	Headers     map[string]string
}

type Segment struct {
	URL      string
	Index    int
	Duration float64
	Filename string
}

type DownloadProgress struct {
	TotalSegments     int
	CompletedSegments int
	CurrentSegment    int
	Speed             string
	ETA               time.Duration
	Status            string
	mu                sync.RWMutex
}

func (dp *DownloadProgress) Update(completed, current int, status string) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.CompletedSegments = completed
	dp.CurrentSegment = current
	dp.Status = status
}

func (dp *DownloadProgress) GetProgress() (int, int, string) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.CompletedSegments, dp.TotalSegments, dp.Status
}

type Config struct {
	OutputDir      string
	Quality        string
	MaxConcurrency int
	RetryAttempts  int
	TimeoutSeconds int
	UserAgent      string
	EnableGUI      bool
	Verbose        bool
}

func DefaultConfig() *Config {
	return &Config{
		OutputDir:      "./downloads",
		Quality:        "best",
		MaxConcurrency: 5,
		RetryAttempts:  3,
		TimeoutSeconds: 30,
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		EnableGUI:      false,
		Verbose:        false,
	}
}
