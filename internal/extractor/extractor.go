package extractor

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yebrai/stream-snatchet/pkg/models"
)

type Extractor struct {
	client *http.Client
	config *models.Config
}

func New(config *models.Config) *Extractor {
	return &Extractor{
		client: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
		config: config,
	}
}

func (e *Extractor) ExtractFromIframe(iframeURL string) (*models.StreamInfo, error) {
	streamInfo := &models.StreamInfo{
		IframeURL: iframeURL,
		Headers:   make(map[string]string),
	}

	iframeContent, err := e.fetchContent(iframeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch iframe content: %w", err)
	}

	manifestURL, err := e.findManifestURL(iframeContent, iframeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest URL: %w", err)
	}

	streamInfo.ManifestURL = manifestURL
	streamInfo.BaseURL = e.getBaseURL(manifestURL)

	manifestContent, err := e.fetchContent(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	if err := e.parseManifest(manifestContent, streamInfo); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return streamInfo, nil
}

func (e *Extractor) fetchContent(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", e.config.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", url)

	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (e *Extractor) findManifestURL(content, baseURL string) (string, error) {
	patterns := []string{
		`['"](https?://[^'"]*\.m3u8[^'"]*?)['"]`,
		`['"](https?://[^'"]*m3u8[^'"]*?)['"]`,
		`source:\s*['"](https?://[^'"]*\.m3u8[^'"]*?)['"]`,
		`src:\s*['"](https?://[^'"]*\.m3u8[^'"]*?)['"]`,
		`file:\s*['"](https?://[^'"]*\.m3u8[^'"]*?)['"]`,
		`url:\s*['"](https?://[^'"]*\.m3u8[^'"]*?)['"]`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			manifestURL := matches[1]
			if strings.HasPrefix(manifestURL, "http") {
				return manifestURL, nil
			}

			baseURLParsed, err := url.Parse(baseURL)
			if err != nil {
				continue
			}
			manifestURLParsed, err := url.Parse(manifestURL)
			if err != nil {
				continue
			}
			return baseURLParsed.ResolveReference(manifestURLParsed).String(), nil
		}
	}

	return "", fmt.Errorf("no manifest URL found in iframe content")
}

func (e *Extractor) getBaseURL(manifestURL string) string {
	u, err := url.Parse(manifestURL)
	if err != nil {
		return manifestURL
	}
	u.Path = u.Path[:strings.LastIndex(u.Path, "/")+1]
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func (e *Extractor) parseManifest(content string, streamInfo *models.StreamInfo) error {
	lines := strings.Split(content, "\n")
	var segments []models.Segment
	var currentDuration float64
	segmentIndex := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#EXTINF:") {
			durationStr := strings.TrimPrefix(line, "#EXTINF:")
			durationStr = strings.Split(durationStr, ",")[0]
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				currentDuration = 10.0
			} else {
				currentDuration = duration
			}
		} else if line != "" && !strings.HasPrefix(line, "#") {
			segmentURL := line
			if !strings.HasPrefix(segmentURL, "http") {
				segmentURL = streamInfo.BaseURL + segmentURL
			}

			segment := models.Segment{
				URL:      segmentURL,
				Index:    segmentIndex,
				Duration: currentDuration,
				Filename: fmt.Sprintf("segment_%04d.ts", segmentIndex),
			}
			segments = append(segments, segment)
			segmentIndex++
		}
	}

	if len(segments) == 0 {
		return fmt.Errorf("no segments found in manifest")
	}

	streamInfo.Segments = segments

	var totalDuration float64
	for _, segment := range segments {
		totalDuration += segment.Duration
	}
	streamInfo.Duration = time.Duration(totalDuration) * time.Second

	return nil
}
