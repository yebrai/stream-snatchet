package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/yebrai/stream-snatchet/gui"
	"github.com/yebrai/stream-snatchet/internal/downloader"
	"github.com/yebrai/stream-snatchet/internal/extractor"
	"github.com/yebrai/stream-snatchet/internal/merger"
	"github.com/yebrai/stream-snatchet/pkg/models"
)

var config *models.Config

var rootCmd = &cobra.Command{
	Use:   "stream-snatchet [iframe-url]",
	Short: "Download streaming videos from iframe URLs",
	Long: `Stream Snatchet is a tool to download streaming videos from iframe URLs.
It extracts HLS manifests, downloads segments concurrently, and merges them into a single video file.

Examples:
  stream-snatchet "https://example.com/iframe/video"
  stream-snatchet --gui
  stream-snatchet --output ./videos --quality best "https://example.com/iframe/video"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDownload,
}

func init() {
	config = models.DefaultConfig()

	rootCmd.Flags().StringVarP(&config.OutputDir, "output", "o", config.OutputDir, "Output directory for downloaded videos")
	rootCmd.Flags().StringVarP(&config.Quality, "quality", "q", config.Quality, "Video quality preference (best, worst, or specific)")
	rootCmd.Flags().IntVarP(&config.MaxConcurrency, "concurrent", "c", config.MaxConcurrency, "Maximum concurrent downloads")
	rootCmd.Flags().IntVarP(&config.RetryAttempts, "retries", "r", config.RetryAttempts, "Number of retry attempts for failed downloads")
	rootCmd.Flags().IntVarP(&config.TimeoutSeconds, "timeout", "t", config.TimeoutSeconds, "Timeout in seconds for HTTP requests")
	rootCmd.Flags().StringVar(&config.UserAgent, "user-agent", config.UserAgent, "User agent string for HTTP requests")
	rootCmd.Flags().BoolVar(&config.EnableGUI, "gui", config.EnableGUI, "Launch GUI mode")
	rootCmd.Flags().BoolVarP(&config.Verbose, "verbose", "v", config.Verbose, "Enable verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDownload(cmd *cobra.Command, args []string) error {
	if config.EnableGUI {
		return gui.LaunchGUI(config)
	}

	if len(args) == 0 {
		return fmt.Errorf("iframe URL is required when not using GUI mode")
	}

	iframeURL := args[0]

	if config.Verbose {
		fmt.Printf("Starting download from: %s\n", iframeURL)
		fmt.Printf("Output directory: %s\n", config.OutputDir)
		fmt.Printf("Max concurrency: %d\n", config.MaxConcurrency)
		fmt.Printf("Retry attempts: %d\n", config.RetryAttempts)
		fmt.Println()
	}

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	ext := extractor.New(config)

	if config.Verbose {
		fmt.Println("Extracting stream information...")
	}

	streamInfo, err := ext.ExtractFromIframe(iframeURL)
	if err != nil {
		return fmt.Errorf("failed to extract stream info: %w", err)
	}

	if config.Verbose {
		fmt.Printf("Found %d segments\n", len(streamInfo.Segments))
		fmt.Printf("Estimated duration: %v\n", streamInfo.Duration)
		fmt.Printf("Manifest URL: %s\n", streamInfo.ManifestURL)
		fmt.Println()
	}

	tempDir := filepath.Join(config.OutputDir, fmt.Sprintf("temp_%d", time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	dl := downloader.New(config)

	if config.Verbose {
		fmt.Println("Starting segment downloads...")
	}

	if err := dl.DownloadSegments(streamInfo, tempDir); err != nil {
		return fmt.Errorf("failed to download segments: %w", err)
	}

	mrg := merger.New(config)
	outputPath := mrg.GenerateOutputFilename(streamInfo, config.OutputDir)

	if config.Verbose {
		fmt.Printf("\nMerging segments into: %s\n", outputPath)
	}

	if err := mrg.MergeSegments(streamInfo, tempDir, outputPath); err != nil {
		return fmt.Errorf("failed to merge segments: %w", err)
	}

	fmt.Printf("✅ Download completed successfully!\n")
	fmt.Printf("📁 Output file: %s\n", outputPath)

	if fileInfo, err := os.Stat(outputPath); err == nil {
		fmt.Printf("📊 File size: %.2f MB\n", float64(fileInfo.Size())/1024/1024)
	}

	return nil
}
