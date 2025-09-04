package newspaper4k

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// VideoExtractor extracts videos from HTML content
type VideoExtractor struct{}

// NewVideoExtractor creates a new VideoExtractor
func NewVideoExtractor() *VideoExtractor {
	return &VideoExtractor{}
}

// Parse extracts videos from the article
func (ve *VideoExtractor) Parse(a *newspaper.Article) error {
	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
		a.Doc = doc
	}
	videos := ve.getVideos(a.Doc, a.URL)
	if len(videos) > 0 {
		a.Movies = videos
	}
	return nil
}

// getVideos extracts all videos from the document
func (ve *VideoExtractor) getVideos(doc *goquery.Document, articleURL string) []string {
	var videos []string

	// Extract from video tags using parser's GetElementsByTagslist
	videoElements := parsers.GetElementsByTagslist(doc.Selection, []string{"video"})
	for _, element := range videoElements {
		src := parsers.GetAttribute(element, "src", nil, "")
		if srcStr, ok := src.(string); ok && srcStr != "" {
			videoURL := urls.JoinURL(articleURL, srcStr)
			if videoURL != "" {
				videos = append(videos, videoURL)
			}
		}
	}

	// Extract from iframe tags
	iframeElements := parsers.GetElementsByTagslist(doc.Selection, []string{"iframe"})
	for _, element := range iframeElements {
		src := parsers.GetAttribute(element, "src", nil, "")
		if srcStr, ok := src.(string); ok && srcStr != "" {
			videoURL := urls.JoinURL(articleURL, srcStr)
			if videoURL != "" && ve.isVideoProvider(ve.getProvider(srcStr)) {
				videos = append(videos, videoURL)
			}
		}
	}

	// Extract from embed tags
	embedElements := parsers.GetElementsByTagslist(doc.Selection, []string{"embed"})
	for _, element := range embedElements {
		src := parsers.GetAttribute(element, "src", nil, "")
		if srcStr, ok := src.(string); ok && srcStr != "" {
			videoURL := urls.JoinURL(articleURL, srcStr)
			if videoURL != "" && ve.isVideoProvider(ve.getProvider(srcStr)) {
				videos = append(videos, videoURL)
			}
		}
	}

	// Extract from object tags
	objectElements := parsers.GetElementsByTagslist(doc.Selection, []string{"object"})
	for _, element := range objectElements {
		data := parsers.GetAttribute(element, "data", nil, "")
		if dataStr, ok := data.(string); ok && dataStr != "" {
			videoURL := urls.JoinURL(articleURL, dataStr)
			if videoURL != "" && ve.isVideoProvider(ve.getProvider(dataStr)) {
				videos = append(videos, videoURL)
			}
		}
	}

	// Extract from JSON-LD VideoObject
	videos = append(videos, ve.getVideosFromJSONLD(doc, articleURL)...)

	return videos
}

// getVideosFromJSONLD extracts videos from JSON-LD structured data
func (ve *VideoExtractor) getVideosFromJSONLD(doc *goquery.Document, articleURL string) []string {
	var videos []string

	// Use parser's GetLdJsonObject method
	jsonObjects := parsers.GetLdJsonObject(doc.Selection)

	for _, data := range jsonObjects {
		// Handle both single object and array of objects
		var objects []map[string]any
		if obj, ok := data["@type"]; ok && obj == "VideoObject" {
			objects = []map[string]any{data}
		} else if graph, ok := data["@graph"].([]any); ok {
			for _, item := range graph {
				if obj, ok := item.(map[string]any); ok {
					if objType, ok := obj["@type"]; ok && objType == "VideoObject" {
						objects = append(objects, obj)
					}
				}
			}
		}

		for _, obj := range objects {
			if contentURL, ok := obj["contentUrl"].(string); ok && contentURL != "" {
				videoURL := urls.JoinURL(articleURL, contentURL)
				if videoURL != "" {
					videos = append(videos, videoURL)
				}
			}
		}
	}

	return videos
}

// getProvider determines the video provider from the URL
func (ve *VideoExtractor) getProvider(videoURL string) string {
	parsedURL, err := url.Parse(videoURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(parsedURL.Host)
	host = strings.TrimPrefix(host, "www.")

	// Check for known providers
	for _, provider := range constants.VIDEO_PROVIDERS {
		if strings.Contains(host, provider) {
			return provider
		}
	}

	return ""
}

// isVideoProvider checks if the provider is a known video provider
func (ve *VideoExtractor) isVideoProvider(provider string) bool {
	for _, p := range constants.VIDEO_PROVIDERS {
		if provider == p {
			return true
		}
	}
	return false
}
