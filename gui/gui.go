package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/yebrai/stream-snatchet/internal/downloader"
	"github.com/yebrai/stream-snatchet/internal/extractor"
	"github.com/yebrai/stream-snatchet/internal/merger"
	"github.com/yebrai/stream-snatchet/pkg/models"
)

type GUI struct {
	app    fyne.App
	window fyne.Window
	config *models.Config

	urlEntry    *widget.Entry
	outputEntry *widget.Entry
	downloadBtn *widget.Button
	progressBar *widget.ProgressBar
	statusLabel *widget.Label
	logText     *widget.RichText

	isDownloading bool
}

func LaunchGUI(config *models.Config) error {
	gui := &GUI{
		app:    app.New(),
		config: config,
	}

	gui.app.SetIcon(nil)

	gui.setupWindow()
	gui.window.ShowAndRun()
	return nil
}

func (g *GUI) setupWindow() {
	g.window = g.app.NewWindow("Stream Snatchet - Video Downloader")
	g.window.Resize(fyne.NewSize(800, 600))
	g.window.CenterOnScreen()

	g.createWidgets()
	g.layoutWidgets()
}

func (g *GUI) createWidgets() {
	g.urlEntry = widget.NewEntry()
	g.urlEntry.SetPlaceHolder("Paste iframe URL here...")
	g.urlEntry.MultiLine = false

	g.outputEntry = widget.NewEntry()
	g.outputEntry.SetText(g.config.OutputDir)

	g.downloadBtn = widget.NewButton("Download Video", g.startDownload)
	g.downloadBtn.Importance = widget.HighImportance

	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()

	g.statusLabel = widget.NewLabel("Ready to download")

	g.logText = widget.NewRichText()
	g.logText.Wrapping = fyne.TextWrapWord
	g.addLog("Welcome to Stream Snatchet!")
	g.addLog("Paste an iframe URL and click Download to start.")
}

func (g *GUI) layoutWidgets() {
	urlContainer := container.NewBorder(
		widget.NewLabel("Iframe URL:"), nil, nil, nil,
		g.urlEntry,
	)

	outputContainer := container.NewBorder(
		widget.NewLabel("Output Directory:"), nil, nil,
		widget.NewButton("Browse", g.browseOutputDir),
		g.outputEntry,
	)

	buttonContainer := container.NewHBox(
		g.downloadBtn,
	)

	progressContainer := container.NewVBox(
		g.progressBar,
		g.statusLabel,
	)

	settingsBtn := widget.NewButton("Settings", g.showSettings)
	aboutBtn := widget.NewButton("About", g.showAbout)

	toolbar := container.NewHBox(
		settingsBtn,
		aboutBtn,
	)

	logContainer := container.NewBorder(
		widget.NewLabel("Log:"), nil, nil, nil,
		container.NewScroll(g.logText),
	)

	content := container.NewVBox(
		urlContainer,
		outputContainer,
		buttonContainer,
		progressContainer,
		logContainer,
		toolbar,
	)

	g.window.SetContent(content)
}

func (g *GUI) browseOutputDir() {
	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil || dir == nil {
			return
		}
		g.outputEntry.SetText(dir.Path())
		g.config.OutputDir = dir.Path()
	}, g.window)
}

func (g *GUI) startDownload() {
	if g.isDownloading {
		return
	}

	url := g.urlEntry.Text
	if url == "" {
		dialog.ShowError(fmt.Errorf("Please enter an iframe URL"), g.window)
		return
	}

	outputDir := g.outputEntry.Text
	if outputDir == "" {
		dialog.ShowError(fmt.Errorf("Please select an output directory"), g.window)
		return
	}

	g.config.OutputDir = outputDir
	g.isDownloading = true
	g.downloadBtn.SetText("Downloading...")
	g.downloadBtn.Disable()
	g.progressBar.Show()
	g.statusLabel.SetText("Starting download...")
	g.addLog(fmt.Sprintf("Starting download from: %s", url))

	go g.performDownload(url)
}

func (g *GUI) performDownload(iframeURL string) {
	defer func() {
		g.isDownloading = false
		g.downloadBtn.SetText("Download Video")
		g.downloadBtn.Enable()
		g.progressBar.Hide()
	}()

	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		g.showError(fmt.Errorf("Failed to create output directory: %w", err))
		return
	}

	ext := extractor.New(g.config)
	g.updateStatus("Extracting stream information...")
	g.addLog("Extracting stream information...")

	streamInfo, err := ext.ExtractFromIframe(iframeURL)
	if err != nil {
		g.showError(fmt.Errorf("Failed to extract stream info: %w", err))
		return
	}

	g.addLog(fmt.Sprintf("Found %d segments", len(streamInfo.Segments)))
	g.addLog(fmt.Sprintf("Estimated duration: %v", streamInfo.Duration))

	tempDir := filepath.Join(g.config.OutputDir, fmt.Sprintf("temp_%d", time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		g.showError(fmt.Errorf("Failed to create temp directory: %w", err))
		return
	}
	defer os.RemoveAll(tempDir)

	dl := downloader.New(g.config)
	g.updateStatus("Downloading segments...")
	g.addLog("Starting segment downloads...")

	go g.trackProgress(dl)

	if err := dl.DownloadSegments(streamInfo, tempDir); err != nil {
		g.showError(fmt.Errorf("Failed to download segments: %w", err))
		return
	}

	mrg := merger.New(g.config)
	outputPath := mrg.GenerateOutputFilename(streamInfo, g.config.OutputDir)

	g.updateStatus("Merging video...")
	g.addLog(fmt.Sprintf("Merging segments into: %s", outputPath))

	if err := mrg.MergeSegments(streamInfo, tempDir, outputPath); err != nil {
		g.showError(fmt.Errorf("Failed to merge segments: %w", err))
		return
	}

	g.updateStatus("Download completed!")
	g.addLog("✅ Download completed successfully!")
	g.addLog(fmt.Sprintf("📁 Output file: %s", outputPath))

	if fileInfo, err := os.Stat(outputPath); err == nil {
		g.addLog(fmt.Sprintf("📊 File size: %.2f MB", float64(fileInfo.Size())/1024/1024))
	}

	dialog.ShowInformation("Download Complete",
		fmt.Sprintf("Video downloaded successfully!\n\nOutput: %s", outputPath),
		g.window)
}

func (g *GUI) trackProgress(dl *downloader.Downloader) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for g.isDownloading {
		select {
		case <-ticker.C:
			progress := dl.GetProgress()
			completed, total, status := progress.GetProgress()

			if total > 0 {
				percentage := float64(completed) / float64(total)
				g.progressBar.SetValue(percentage)
				g.updateStatus(status)
			}
		}
	}
}

func (g *GUI) updateStatus(status string) {
	g.statusLabel.SetText(status)
}

func (g *GUI) addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s", timestamp, message)
	g.logText.ParseMarkdown(logMessage)
	g.logText.Refresh()
}

func (g *GUI) showError(err error) {
	g.addLog(fmt.Sprintf("❌ Error: %v", err))
	g.updateStatus(fmt.Sprintf("Error: %v", err))
	dialog.ShowError(err, g.window)
}

func (g *GUI) showSettings() {
	settingsWindow := g.app.NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(400, 300))

	concurrencyEntry := widget.NewEntry()
	concurrencyEntry.SetText(fmt.Sprintf("%d", g.config.MaxConcurrency))

	retriesEntry := widget.NewEntry()
	retriesEntry.SetText(fmt.Sprintf("%d", g.config.RetryAttempts))

	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(fmt.Sprintf("%d", g.config.TimeoutSeconds))

	verboseCheck := widget.NewCheck("Verbose logging", func(checked bool) {
		g.config.Verbose = checked
	})
	verboseCheck.SetChecked(g.config.Verbose)

	saveBtn := widget.NewButton("Save", func() {
		g.config.MaxConcurrency = parseInt(concurrencyEntry.Text, g.config.MaxConcurrency)
		g.config.RetryAttempts = parseInt(retriesEntry.Text, g.config.RetryAttempts)
		g.config.TimeoutSeconds = parseInt(timeoutEntry.Text, g.config.TimeoutSeconds)
		settingsWindow.Close()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Max Concurrency", concurrencyEntry),
			widget.NewFormItem("Retry Attempts", retriesEntry),
			widget.NewFormItem("Timeout (seconds)", timeoutEntry),
		),
		verboseCheck,
		saveBtn,
	)

	settingsWindow.SetContent(form)
	settingsWindow.Show()
}

func (g *GUI) showAbout() {
	aboutText := `Stream Snatchet v1.0
	
A tool for downloading streaming videos from iframe URLs.

Features:
• HLS manifest extraction
• Concurrent segment downloading
• Video merging with ffmpeg
• Progress tracking
• GUI and CLI interfaces

Requirements:
• ffmpeg must be installed and available in PATH

Created with Go and Fyne framework.`

	dialog.ShowInformation("About Stream Snatchet", aboutText, g.window)
}

func parseInt(s string, defaultVal int) int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		return defaultVal
	}
	if val <= 0 {
		return defaultVal
	}
	return val
}
