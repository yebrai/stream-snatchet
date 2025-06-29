# Stream Snatchet 🎬

Stream Snatchet is a powerful Go-based tool designed to download streaming videos from iframe URLs. It automatically extracts HLS manifests, downloads video segments concurrently, and merges them into a single MP4 file for offline viewing.

## Features ✨

- **HLS Manifest Extraction**: Automatically detects and extracts `.m3u8` manifest URLs from iframe content
- **Concurrent Downloads**: Downloads multiple video segments simultaneously for optimal speed
- **Video Merging**: Uses FFmpeg to seamlessly merge segments into a single MP4 file
- **Progress Tracking**: Real-time progress updates with download speed and ETA
- **GUI Interface**: User-friendly graphical interface built with Fyne
- **CLI Interface**: Command-line interface with extensive configuration options
- **Error Handling**: Robust retry logic and error recovery
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Prerequisites 📋

- **Go 1.21+**: Required for building the application
- **FFmpeg**: Must be installed and available in your system PATH
  - Windows: Download from [ffmpeg.org](https://ffmpeg.org/download.html)
  - macOS: `brew install ffmpeg`
  - Linux: `sudo apt-get install ffmpeg` or equivalent

## Installation 🚀

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/yebrai/stream-snatchet.git
cd stream-snatchet

# Download dependencies
go mod tidy

# Build the application
go build -o stream-snatchet cmd/stream-snatchet/main.go
```

### Option 2: Direct Installation

```bash
go install github.com/yebrai/stream-snatchet/cmd/stream-snatchet@latest
```

## Usage 🎯

### GUI Mode

Launch the graphical interface:

```bash
./stream-snatchet --gui
```

1. Paste the iframe URL in the input field
2. Select output directory (optional)
3. Click "Download Video"
4. Monitor progress in real-time
5. Access your downloaded video when complete

### CLI Mode

Basic usage:

```bash
./stream-snatchet "https://example.com/iframe/video"
```

Advanced usage with options:

```bash
./stream-snatchet \
  --output ./downloads \
  --concurrent 10 \
  --retries 5 \
  --timeout 60 \
  --verbose \
  "https://example.com/iframe/video"
```

### Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `./downloads` | Output directory for downloaded videos |
| `--quality` | `-q` | `best` | Video quality preference |
| `--concurrent` | `-c` | `5` | Maximum concurrent downloads |
| `--retries` | `-r` | `3` | Number of retry attempts |
| `--timeout` | `-t` | `30` | Timeout in seconds for HTTP requests |
| `--user-agent` | | Mozilla/5.0... | Custom User-Agent string |
| `--gui` | | `false` | Launch GUI mode |
| `--verbose` | `-v` | `false` | Enable verbose output |
| `--help` | `-h` | | Show help information |

## Examples 📝

### Download with custom settings

```bash
# High concurrency for faster downloads
./stream-snatchet --concurrent 15 --verbose "https://example.com/iframe/video"

# Custom output directory and retry settings
./stream-snatchet -o ~/Videos -r 10 "https://example.com/iframe/video"

# Launch GUI with verbose logging
./stream-snatchet --gui --verbose
```

### Batch Processing

```bash
# Download multiple videos
./stream-snatchet "https://site1.com/iframe/video1"
./stream-snatchet "https://site2.com/iframe/video2"
./stream-snatchet "https://site3.com/iframe/video3"
```

## Architecture 🏗️

```
stream-snatchet/
├── cmd/stream-snatchet/     # Main application entry point
├── internal/
│   ├── extractor/           # HLS manifest extraction logic
│   ├── downloader/          # Concurrent segment downloader
│   └── merger/              # Video merging with FFmpeg
├── pkg/models/              # Data structures and models
├── gui/                     # Fyne-based GUI implementation
└── README.md
```

### Core Components

1. **Extractor**: Analyzes iframe content to locate HLS manifest URLs
2. **Downloader**: Manages concurrent download of video segments
3. **Merger**: Uses FFmpeg to combine segments into final MP4 file
4. **GUI**: Provides user-friendly interface with progress tracking

## How It Works 🔧

1. **URL Analysis**: Parses the provided iframe URL to extract video source
2. **Manifest Discovery**: Uses regex patterns to locate `.m3u8` manifest files
3. **Segment Parsing**: Analyzes the HLS manifest to identify individual video segments
4. **Concurrent Download**: Downloads multiple segments simultaneously using goroutines
5. **Progress Tracking**: Monitors download progress and provides real-time updates
6. **Video Merging**: Uses FFmpeg to concatenate segments into a single MP4 file
7. **Cleanup**: Removes temporary segment files after successful merge

## Configuration ⚙️

### Default Configuration

```go
&Config{
    OutputDir:       "./downloads",
    Quality:         "best",
    MaxConcurrency:  5,
    RetryAttempts:   3,
    TimeoutSeconds:  30,
    UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    EnableGUI:       false,
    Verbose:         false,
}
```

### Environment Variables

You can also configure the application using environment variables:

```bash
export STREAM_SNATCHET_OUTPUT_DIR="./my-videos"
export STREAM_SNATCHET_CONCURRENT="10"
export STREAM_SNATCHET_VERBOSE="true"
```

## Testing 🧪

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/extractor
go test ./internal/merger
```

## Supported Formats 📺

- **Input**: HLS (HTTP Live Streaming) `.m3u8` manifests
- **Output**: MP4 video files
- **Segments**: `.ts` (Transport Stream) files

## Troubleshooting 🔍

### Common Issues

1. **FFmpeg not found**
   ```
   Error: ffmpeg not found in PATH
   ```
   **Solution**: Install FFmpeg and ensure it's in your system PATH

2. **No manifest URL found**
   ```
   Error: no manifest URL found in iframe content
   ```
   **Solution**: The iframe may use a different streaming protocol or have JavaScript-generated URLs

3. **Download failures**
   ```
   Error: failed to download segments
   ```
   **Solution**: Try increasing retry attempts and timeout values

4. **Permission denied**
   ```
   Error: failed to create output directory
   ```
   **Solution**: Check write permissions for the output directory

### Debug Mode

Enable verbose logging for detailed information:

```bash
./stream-snatchet --verbose "https://example.com/iframe/video"
```

## Performance Tips 🚀

- **Increase concurrency** for faster downloads (but respect server limits)
- **Use SSD storage** for better I/O performance during merging
- **Monitor bandwidth** to avoid overwhelming your connection
- **Adjust timeout values** based on your network conditions

## Security Considerations 🔒

- The tool respects robots.txt and rate limiting
- Uses standard HTTP headers to avoid detection as a bot
- Does not bypass DRM or authentication mechanisms
- Only downloads publicly accessible content

## Contributing 🤝

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

```bash
git clone https://github.com/yebrai/stream-snatchet.git
cd stream-snatchet
go mod tidy
go run cmd/stream-snatchet/main.go --help
```

## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer ⚠️

This tool is intended for downloading publicly accessible content for personal use only. Users are responsible for ensuring they have the right to download and store the content they access. Please respect copyright laws and website terms of service.

## Support 💬

If you encounter any issues or have questions:

1. Check the [troubleshooting section](#troubleshooting-)
2. Search existing [GitHub issues](https://github.com/yebrai/stream-snatchet/issues)
3. Create a new issue with detailed information about your problem

---

**Happy downloading! 🎉**