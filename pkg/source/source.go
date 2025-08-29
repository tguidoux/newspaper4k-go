package source

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/helpers"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// Source interface defines the methods for a news source
type Source interface {
	Build(inputHTML string, onlyHomepage bool, onlyInPath bool)
	Download()
	Parse()
	SearchCategories()
	GetFeeds(limitFeeds int)
	extractDescription()
	DownloadCategories()
	BuildCategories()
	feedsToArticles() []newspaper.Article
	categoriesToArticles() []newspaper.Article
	GetArticles(limit int, onlyInPath bool)
	Size() int
	FeedURLs() []string
	CategoryURLs() []string
	ArticleURLs() []string
	PrintSummary()
	String() string
}

// DefaultSource is the default implementation of the Source interface
type DefaultSource struct {
	URL          string
	ParsedURL    *urls.URL
	Config       *configuration.Configuration
	Categories   []newspaper.Category
	Feeds        []newspaper.Feed
	Articles     []newspaper.Article
	HTML         string
	Doc          *goquery.Document
	LogoURL      string
	Favicon      string
	Description  string
	IsParsed     bool
	IsDownloaded bool
}

type SourceRequest struct {
	URL    string
	Config configuration.Configuration
}

type BuildParams struct {
	InputHTML       string
	OnlyHomepage    bool
	OnlyInPath      bool
	LimitCategories int
	LimitArticles   int
	LimitFeeds      int
}

// NewDefaultSource creates a new DefaultSource
func NewDefaultSource(request SourceRequest) (*DefaultSource, error) {

	url := urls.PrepareURL(request.URL, request.URL)
	config := request.Config

	preparedURL, err := urls.New(url)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare URL: %v", err)
	}

	source := &DefaultSource{
		URL:          url,
		ParsedURL:    preparedURL,
		Config:       &config,
		Categories:   []newspaper.Category{},
		Feeds:        []newspaper.Feed{},
		Articles:     []newspaper.Article{},
		IsParsed:     false,
		IsDownloaded: false,
	}

	return source, nil
}

// -----------------------------------------------------------------
// Source methods
// -----------------------------------------------------------------

// Build encapsulates download and basic parsing
func (s *DefaultSource) Build() error {
	return s.BuildWithParams(BuildParams{InputHTML: "", OnlyHomepage: false, OnlyInPath: false, LimitArticles: 5000})
}

// Build encapsulates download and basic parsing
func (s *DefaultSource) BuildWithParams(params BuildParams) error {

	inputHTML := params.InputHTML
	onlyHomepage := params.OnlyHomepage
	limitFeeds := params.LimitFeeds
	if limitFeeds <= 0 {
		limitFeeds = 100
	}
	limitCategories := params.LimitCategories
	if limitCategories <= 0 {
		limitCategories = 100
	}

	// Step 1: Download and parse homepage
	// if InputHTML is provided, use it instead of downloading
	if inputHTML != "" {
		s.HTML = inputHTML
	} else {
		err := s.Download()
		if err != nil {
			return fmt.Errorf("failed to download source: %v", err)
		}
	}
	err := s.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse source: %v", err)
	}

	// Step 2: Set categories and feeds, download and parse them
	// if onlyHomepage is true, skip categories and feeds
	if onlyHomepage {
		s.Categories = []newspaper.Category{{URL: s.URL, HTML: s.HTML, Doc: s.Doc}}
	} else {
		err := s.SearchCategories()
		if err != nil {
			return fmt.Errorf("failed to set categories: %v", err)
		}
		s.DownloadCategories()
	}
	s.BuildCategories()

	if len(s.Categories) > limitCategories {
		s.Categories = s.Categories[:limitCategories]
	}

	// Step 3: Download and parse feeds, generate articles
	// we skip feeds if onlyHomepage is true
	if !onlyHomepage {
		s.GetFeeds(limitFeeds)
	}

	return nil
}

// Download downloads the HTML of the source
func (s *DefaultSource) Download() error {
	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)
	resp, err := client.Get(s.URL)
	if err != nil {
		// Handle error - could log or set a flag
		return fmt.Errorf("failed to download URL: %v", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode >= 400 {
		// Handle HTTP error
		return fmt.Errorf("received HTTP status %d", resp.StatusCode)
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		return fmt.Errorf("failed to read response body: %v", err)
	}
	s.HTML = string(htmlBytes)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s.HTML))
	if err != nil {
		// Handle error
		return fmt.Errorf("failed to parse HTML: %v", err)
	}
	s.Doc = doc
	s.IsDownloaded = true

	return nil
}

// Parse sets the goquery document and sets description
func (s *DefaultSource) Parse() error {

	// Ensure Doc is set
	if s.Doc == nil {
		doc, err := parsers.FromString(s.HTML)
		if err != nil {
			// Handle error
			return fmt.Errorf("failed to parse HTML: %v", err)
		}
		s.Doc = doc
	}
	s.extractDescription()
	s.IsParsed = true

	return nil
}

// -----------------------------------------------------------------
// Categories
// -----------------------------------------------------------------
// SearchCategories sets the categories for the source
// Only includes categories from the same domain as the source URL
func (s *DefaultSource) SearchCategories() error {
	// Simple implementation: extract categories from links on the homepage
	if s.Doc == nil {
		return fmt.Errorf("document not parsed")
	}

	categoryURLs := []string{}
	sourceDomain := s.ParsedURL.Domain

	s.Doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if exists && href != "" && href != "/" && href != "#" {
			fullURL := urls.PrepareURL(href, s.URL)
			if newspaper.IsValidCategoryURL(fullURL) && fullURL != s.URL {
				// Only include URLs from the same domain as the source
				categoryUrl, err := urls.Parse(fullURL)
				if err != nil {
					return
				}
				if categoryUrl.Domain == sourceDomain {
					categoryURLs = append(categoryURLs, categoryUrl.String())
				}
			}
		}
	})

	// Add main page as a category for feed extraction
	// categoryURLs = append(categoryURLs, s.URL)

	// Remove duplicates
	uniqueURLs := helpers.UniqueStringsSimple(categoryURLs)

	s.Categories = make([]newspaper.Category, len(uniqueURLs))
	for i, u := range uniqueURLs {
		s.Categories[i] = newspaper.Category{URL: u}
	}

	return nil
}

// DownloadCategories downloads HTML for all categories
func (s *DefaultSource) DownloadCategories() {
	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)
	for i, cat := range s.Categories {
		resp, err := client.Get(cat.URL)
		if err != nil || resp.StatusCode >= 400 {
			continue
		}

		htmlBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			continue
		}
		s.Categories[i].HTML = string(htmlBytes)
	}
}

// BuildCategories parses the HTML into goquery documents
func (s *DefaultSource) BuildCategories() {
	for i, cat := range s.Categories {
		if cat.HTML != "" {
			doc, err := parsers.FromString(cat.HTML)
			if err == nil {
				s.Categories[i].Doc = doc
			}
		}
	}
}

// -----------------------------------------------------------------
// Feeds
// -----------------------------------------------------------------
func (s *DefaultSource) GetFeeds(limitFeeds int) {
	commonFeedSuffixes := constants.COMMON_FEED_SUFFIXES
	commonFeedURLs := []string{}

	for _, suffix := range commonFeedSuffixes {
		commonFeedURLs = append(commonFeedURLs, s.URL+suffix)
	}

	// Check for medium.com specific feeds
	if strings.Contains(s.URL, "medium.com") {
		parsed, _ := url.Parse(s.URL)
		if strings.HasPrefix(parsed.Path, "/@") {
			newPath := "/feed/" + strings.Split(parsed.Path, "/")[1]
			newURL := s.ParsedURL.Scheme + "://" + s.ParsedURL.Domain + newPath
			commonFeedURLs = append(commonFeedURLs, newURL)
		}
	}

	// Add feeds from categories
	for _, cat := range s.Categories {
		pathChunks := strings.Split(strings.Trim(cat.URL, "/"), "/")
		if len(pathChunks) > 0 && strings.Contains(pathChunks[len(pathChunks)-1], ".") {
			continue // skip files
		}
		for _, suffix := range commonFeedSuffixes {
			commonFeedURLs = append(commonFeedURLs, cat.URL+suffix)
		}
	}

	// Download and check feeds
	validFeeds := []newspaper.Feed{}
	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)
	// Limit feed URLs to check
	if limitFeeds > 0 && len(commonFeedURLs) > limitFeeds {
		commonFeedURLs = commonFeedURLs[:limitFeeds]
	}
	for _, feedURL := range commonFeedURLs {
		resp, err := client.Get(feedURL)
		if err != nil || resp.StatusCode >= 300 {
			continue
		}

		rssBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			continue
		}
		rss := string(rssBytes)

		feed := newspaper.Feed{URL: urls.PrepareURL(feedURL, feedURL), RSS: rss}
		validFeeds = append(validFeeds, feed)
	}

	// Extract feed URLs from categories
	feedURLs := s.extractFeedURLs(s.Categories)
	for _, feedURL := range feedURLs {
		validFeeds = append(validFeeds, newspaper.Feed{URL: urls.PrepareURL(feedURL, feedURL)})
	}

	validFeeds = helpers.UniqueStructByKey(
		validFeeds,
		func(f newspaper.Feed) string {
			return f.URL
		},
		helpers.UniqueOptions{CaseSensitive: true, PreserveOrder: true},
	)

	s.Feeds = validFeeds
}

// -----------------------------------------------------------------
// Articles
// -----------------------------------------------------------------

// feedsToArticles returns articles from RSS feeds
func (s *DefaultSource) feedsToArticles() []newspaper.Article {
	articles := []newspaper.Article{}

	for _, feed := range s.Feeds {
		if feed.RSS == "" {

			continue
		}

		parsedFeedURL, err := url.Parse(feed.URL)
		if err != nil {

			continue
		}

		// Parse RSS XML properly instead of using regex
		doc, err := parsers.FromString(feed.RSS)
		if err != nil {
			continue
		}

		// Extract URLs from RSS items
		doc.Find("item, entry").Each(func(i int, item *goquery.Selection) {

			// Try different link element patterns
			var articleURL string

			// Try link element first
			link := item.Find("link").First()

			if link.Length() > 0 {
				articleURL = strings.TrimSpace(link.Text())
				if articleURL == "" {
					articleURL, _ = link.Attr("href")
				}
			}

			// If no link found, try guid
			if articleURL == "" {
				guid := item.Find("guid").First()

				if guid.Length() > 0 {
					articleURL = strings.TrimSpace(guid.Text())
					if articleURL == "" {
						articleURL, _ = guid.Attr("href")
					}
				}
			}

			// If still no link found, try enclosure
			if articleURL == "" {
				enclosure := item.Find("enclosure").First()

				if enclosure.Length() > 0 {
					articleURL, _ = enclosure.Attr("url")
				}
			}

			// If still no link found, try listing all links and picking the first valid one
			if articleURL == "" {
				item.Find("link").EachWithBreak(func(i int, linkSel *goquery.Selection) bool {
					href, exists := linkSel.Attr("href")
					if exists && href != "" {
						articleURL = strings.TrimSpace(href)
						return false // break the loop
					}
					return true // continue the loop
				})
			}

			// Otherwise try to find any URL in the item text using regex
			if articleURL == "" {
				re := regexp.MustCompile(constants.RE_URL)
				matches := re.FindStringSubmatch(item.Text())

				// Take the first match which is likely a valid URL by our heuristics
				for _, match := range matches {
					if newspaper.IsLikelyArticleURL(match) {
						articleURL = match
						break
					}
				}

			}

			// Clean up the URL and validate
			articleURL = urls.PrepareURL(articleURL, s.URL)

			if articleURL != "" && newspaper.IsLikelyArticleURL(articleURL) {
				// Only include articles from the same domain as the source
				parsedArticleURL, err := urls.Parse(articleURL)
				if err != nil {
					return
				}
				if parsedArticleURL.Domain == s.ParsedURL.Domain && parsedArticleURL.String() != parsedFeedURL.String() {
					article := newspaper.Article{
						URL:       articleURL,
						SourceURL: feed.URL,
						Config:    s.Config,
					}
					articles = append(articles, article)
				}
			}
		})
	}

	return articles
}

// extractFeedURLs extracts feed URLs from categories
// Only includes feeds from the same domain as the source URL
func (s *DefaultSource) extractFeedURLs(categories []newspaper.Category) []string {
	feedURLs := []string{}
	sourceDomain := s.ParsedURL.Domain

	for _, cat := range categories {
		if cat.Doc == nil {
			continue
		}
		cat.Doc.Find("link[type='application/rss+xml'], link[type='application/atom+xml']").Each(func(i int, sel *goquery.Selection) {
			href, exists := sel.Attr("href")
			if exists {
				fullURL := urls.PrepareURL(href, cat.URL)
				if newspaper.IsValidCategoryURL(fullURL) {
					// Only include feed URLs from the same domain as the source
					categoryDomain, err := urls.Parse(fullURL)
					if err != nil {
						return
					}
					if categoryDomain.Domain == sourceDomain {
						feedURLs = append(feedURLs, fullURL)
					}
				}
			}
		})
	}
	return feedURLs
}

// categoriesToArticles returns articles from categories
// Only includes articles from the same domain as the source URL
func (s *DefaultSource) categoriesToArticles() []newspaper.Article {
	articles := []newspaper.Article{}
	sourceDomain := s.ParsedURL.Domain

	for _, cat := range s.Categories {
		if cat.Doc == nil && cat.HTML != "" {
			doc, err := parsers.FromString(cat.HTML)
			if err != nil {
				continue
			}
			cat.Doc = doc
		} else if cat.Doc == nil && cat.HTML == "" {
			continue
		}
		cat.Doc.Find("a").Each(func(i int, sel *goquery.Selection) {
			href, exists := sel.Attr("href")
			if exists && href != "" && href != "/" && href != "#" {
				articleURL := urls.PrepareURL(href, cat.URL)
				if newspaper.IsLikelyArticleURL(articleURL) && articleURL != s.URL && articleURL != cat.URL {
					// Only include articles from the same domain as the source
					parsedArticleURL, err := urls.Parse(articleURL)
					if err != nil {
						return
					}
					if parsedArticleURL.Domain == sourceDomain {
						// Skip navigation links - only include links that look like articles
						// Articles typically have IDs, dates, or specific patterns
						title := sel.Text()
						article := newspaper.Article{
							URL:       articleURL,
							SourceURL: cat.URL,
							Title:     title,
							Config:    s.Config,
						}
						articles = append(articles, article)
					}
				}
			}
		})
	}

	return articles
}

// GetArticles creates the list of Article objects
func (s *DefaultSource) GetArticles(limit int, onlyInPath bool) []newspaper.Article {
	categoryArticles := s.categoriesToArticles()
	feedArticles := s.feedsToArticles()

	allArticles := append(feedArticles, categoryArticles...)

	// Remove duplicates
	uniqueArticles := helpers.UniqueStructByKey(
		allArticles,
		func(a newspaper.Article) string {
			return a.URL
		},
		helpers.UniqueOptions{CaseSensitive: true, PreserveOrder: true},
	)

	if onlyInPath {
		currentDomain := s.ParsedURL.Domain
		currentPath := s.ParsedURL.Path
		pathChunks := strings.Split(strings.Trim(currentPath, "/"), "/")
		if len(pathChunks) > 0 && (strings.HasSuffix(pathChunks[len(pathChunks)-1], ".html") || strings.HasSuffix(pathChunks[len(pathChunks)-1], ".php")) {
			pathChunks = pathChunks[:len(pathChunks)-1]
		}
		currentPath = "/" + strings.Join(pathChunks, "/") + "/"

		filteredArticles := []newspaper.Article{}
		for _, article := range uniqueArticles {
			parsedArticleURL, err := urls.Parse(article.URL)
			if err != nil {
				continue
			}
			if currentDomain == parsedArticleURL.Domain && strings.HasPrefix(parsedArticleURL.Path, currentPath) {
				filteredArticles = append(filteredArticles, article)
			}
		}
		uniqueArticles = filteredArticles
	}

	if limit > 0 && len(uniqueArticles) > limit {
		s.Articles = uniqueArticles[:limit]
	} else {
		s.Articles = uniqueArticles
	}

	return s.Articles
}

// -----------------------------------------------------------------
// Utility methods
// -----------------------------------------------------------------

// extractDescription sets the description from meta tags
func (s *DefaultSource) extractDescription() {
	if s.Doc == nil {
		return
	}
	description, exists := s.Doc.Find("meta[name='description']").Attr("content")
	if exists {
		s.Description = description
	}
}

// Size returns the number of articles
func (s *DefaultSource) Size() int {
	return len(s.Articles)
}
