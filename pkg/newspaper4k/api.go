package newspaper4k

import (
	"fmt"

	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/extractors/newspaper4k"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

func DefaultExtractors(config *configuration.Configuration) []newspaper.Extractor {
	return []newspaper.Extractor{
		newspaper4k.NewMetadataExtractor(config),
		newspaper4k.NewLanguageExtractor(config),
		newspaper4k.NewTitleExtractor(config),
		newspaper4k.NewAuthorsExtractor(config),
		newspaper4k.NewPubdateExtractor(config),
		newspaper4k.NewBodyExtractor(config),
		newspaper4k.NewLanguageExtractor(config), // Run twice to ensure language is set after text extraction
		newspaper4k.NewCategoryExtractor(config),
		newspaper4k.NewImageExtractor(config),
		newspaper4k.NewVideoExtractor(),
	}
}

func NewDefaultParseRequest(url string) newspaper.ParseRequest {
	config := configuration.NewConfiguration()
	return newspaper.ParseRequest{URL: url, Configuration: config, Extractors: DefaultExtractors(config)}
}

// NewArticleFromURL convenience: create and build an article from a URL using
// default configuration and extractors.
func NewArticleFromURL(url string) (*newspaper.Article, error) {
	return NewArticleFromRequest(NewDefaultParseRequest(url))
}

// NewArticleFromHTML creates and parses an Article from raw HTML only.
// It delegates to NewArticleFromRequest using a placeholder URL.
func NewArticleFromHTML(html string) (*newspaper.Article, error) {
	req := NewDefaultParseRequest("")
	req.InputHTML = html
	return NewArticleFromRequest(req)
}

// NewArticleFromRequest creates an Article from a ParseRequest. If Configuration
// or Extractors are nil they default to the package defaults.
func NewArticleFromRequest(req newspaper.ParseRequest) (*newspaper.Article, error) {
	// Allow HTML-only requests: if URL is empty but InputHTML is provided,
	// proceed using a fallback localhost URL so NewArticle can construct an Article.
	if req.URL == "" && req.InputHTML == "" {
		return nil, fmt.Errorf("empty URL and empty InputHTML")
	}

	// Create the base article
	art, err := NewArticle(req.URL)
	if err != nil {
		return nil, err
	}

	// Download if provided input HTML or no HTML
	art.Config.DownloadOptions.InputHTML = req.InputHTML
	art.Download()

	art.Parse(req.Extractors)

	art.NLP()

	return art, nil
}

// NewArticle constructs the article class. Will not download or parse the article.
func NewArticle(url string) (*newspaper.Article, error) {
	// Set source URL if not provided
	parsedURL, err := urls.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("input url bad format: %w", err)
	}
	sourceURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Domain)
	preparedURL := urls.PrepareURL(url, sourceURL)

	if sourceURL == "" {
		return nil, fmt.Errorf("input url bad format")
	}

	article := &newspaper.Article{
		Config:        configuration.NewConfiguration(),
		SourceURL:     sourceURL,
		URL:           preparedURL,
		Title:         "",
		DownloadState: newspaper.NotStarted,
		IsParsed:      false,
		Images:        []string{},
		Movies:        []string{},
		Keywords:      []string{},
		KeywordScores: map[string]float64{},
		MetaKeywords:  []string{},
		Tags:          map[string]string{},
		Authors:       []string{},
		MetaData:      map[string]string{},
		Categories:    []*urls.URL{},
	}

	return article, nil
}
