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
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// Source interface defines the methods for a news source
type Source interface {
	Build(inputHTML string, onlyHomepage bool, onlyInPath bool)
	Download()
	Parse()
	SetCategories()
	GetFeeds(limitFeeds int)
	SetDescription()
	DownloadCategories()
	ParseCategories()
	FeedsToArticles() []newspaper.Article
	CategoriesToArticles() []newspaper.Article
	GenerateArticles(limit int, onlyInPath bool)
	DownloadArticles() []newspaper.Article
	ParseArticles()
	Size() int
	CleanMemoCache()
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
	InputHTML     string
	OnlyHomepage  bool
	OnlyInPath    bool
	LimitArticles int
	LimitFeeds    int
}

// NewDefaultSource creates a new DefaultSource
func NewDefaultSource(request SourceRequest) (*DefaultSource, error) {

	url := request.URL
	config := request.Config

	if !urls.ValidateURL(url) {
		return nil, fmt.Errorf("input url is bad")
	}

	source := &DefaultSource{
		URL:          urls.PrepareURL(url, ""),
		Config:       &config,
		Categories:   []newspaper.Category{},
		Feeds:        []newspaper.Feed{},
		Articles:     []newspaper.Article{},
		IsParsed:     false,
		IsDownloaded: false,
	}

	ParsedURL, err := urls.Parse(source.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	source.ParsedURL = ParsedURL

	return source, nil
}

// Build encapsulates download and basic parsing
func (s *DefaultSource) BuildWithParams(params BuildParams) error {

	inputHTML := params.InputHTML
	onlyHomepage := params.OnlyHomepage
	onlyInPath := params.OnlyInPath
	limitArticles := params.LimitArticles
	if limitArticles <= 0 {
		limitArticles = 5000
	}
	limitFeeds := params.LimitFeeds
	if limitFeeds <= 0 {
		limitFeeds = 100
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
		err := s.SetCategories()
		if err != nil {
			return fmt.Errorf("failed to set categories: %v", err)
		}
		s.DownloadCategories()
	}
	s.ParseCategories()

	// Step 3: Download and parse feeds, generate articles
	// we skip feeds if onlyHomepage is true
	if !onlyHomepage {
		s.GetFeeds(limitFeeds)
	}

	s.GenerateArticles(limitArticles, onlyInPath)
	s.DownloadArticles()
	s.ParseArticles()

	return nil
}

// Build encapsulates download and basic parsing
func (s *DefaultSource) Build() error {
	return s.BuildWithParams(BuildParams{InputHTML: "", OnlyHomepage: false, OnlyInPath: false, LimitArticles: 5000})
}

// Download downloads the HTML of the source
func (s *DefaultSource) Download() error {
	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)
	resp, err := client.Get(s.URL)
	if err != nil {
		// Handle error - could log or set a flag
		return fmt.Errorf("failed to download URL: %v", err)
	}
	defer resp.Body.Close()

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
	s.SetDescription()
	s.IsParsed = true

	return nil
}

// SetCategories sets the categories for the source
// Only includes categories from the same domain as the source URL
func (s *DefaultSource) SetCategories() error {
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
				categoryDomain, err := urls.Parse(fullURL)
				if err != nil {
					return
				}
				if categoryDomain.Domain == sourceDomain {
					categoryURLs = append(categoryURLs, fullURL)
				}
			}
		}
	})

	// Remove duplicates
	uniqueURLs := helpers.UniqueStringsSimple(categoryURLs)

	// Limit categories for demo purposes
	if len(uniqueURLs) > 5 {
		uniqueURLs = uniqueURLs[:5]
	}

	s.Categories = make([]newspaper.Category, len(uniqueURLs))
	for i, u := range uniqueURLs {
		s.Categories[i] = newspaper.Category{URL: u}
	}

	return nil
}

// isLikelyArticleURL checks if a URL is likely to be an article rather than a navigation link
func (s *DefaultSource) isLikelyArticleURL(urlStr string) bool {
	// Skip obvious navigation/category URLs
	if strings.Contains(urlStr, "/newest") ||
		strings.Contains(urlStr, "/past") ||
		strings.Contains(urlStr, "/comments") ||
		strings.Contains(urlStr, "/ask") ||
		strings.Contains(urlStr, "/show") ||
		strings.Contains(urlStr, "/jobs") ||
		strings.Contains(urlStr, "/submit") ||
		strings.Contains(urlStr, "/news") ||
		strings.Contains(urlStr, "/front") ||
		strings.Contains(urlStr, "/newcomments") ||
		strings.Contains(urlStr, "/login") ||
		strings.Contains(urlStr, "/logout") ||
		strings.Contains(urlStr, "/user") {
		return false
	}

	// For Hacker News, articles have /item?id= pattern
	if strings.Contains(urlStr, "/item?id=") {
		return true
	}

	// For other sites, check for common article patterns
	if strings.Contains(urlStr, "/article") ||
		strings.Contains(urlStr, "/story") ||
		strings.Contains(urlStr, "/post") ||
		strings.Contains(urlStr, "/news/") ||
		strings.Contains(urlStr, "/blog/") ||
		strings.Contains(urlStr, "/p/") ||
		strings.Contains(urlStr, "/entry") ||
		strings.Contains(urlStr, "/content") {
		return true
	}

	// Check if URL has a date-like pattern (YYYY/MM/DD or similar)
	datePattern := regexp.MustCompile(`/(\d{4})/(\d{1,2})/(\d{1,2})/`)
	if datePattern.MatchString(urlStr) {
		return true
	}

	// If it has query parameters, it might be an article
	ParsedURL, err := url.Parse(urlStr)
	if err == nil && ParsedURL.RawQuery != "" {
		return true
	}

	// Default: if it's not obviously a category/navigation URL, consider it an article
	return false
}
func (s *DefaultSource) GetFeeds(limitFeeds int) {
	commonFeedSuffixes := newspaper.COMMON_FEED_SUFFIXES
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
		resp.Body.Close()
		if err != nil {
			continue
		}
		rss := string(rssBytes)

		feed := newspaper.Feed{URL: feedURL, RSS: rss}
		validFeeds = append(validFeeds, feed)
	}

	// Add main page as a category for feed extraction
	mainCategory := newspaper.Category{URL: s.URL, HTML: s.HTML, Doc: s.Doc}
	allCategories := append(s.Categories, mainCategory)

	// Extract feed URLs from categories
	feedURLs := s.extractFeedURLs(allCategories)
	for _, feedURL := range feedURLs {

		// Skip if already added
		found := false
		for _, feed := range validFeeds {
			if feed.URL == feedURL {
				found = true
				break
			}
		}
		if !found {
			validFeeds = append(validFeeds, newspaper.Feed{URL: feedURL})
		}
	}

	s.Feeds = validFeeds
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

// SetDescription sets the description from meta tags
func (s *DefaultSource) SetDescription() {
	if s.Doc == nil {
		return
	}
	description, exists := s.Doc.Find("meta[name='description']").Attr("content")
	if exists {
		s.Description = description
	}
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
		resp.Body.Close()
		if err != nil {
			continue
		}
		s.Categories[i].HTML = string(htmlBytes)
	}
}

// ParseCategories parses the HTML into goquery documents
func (s *DefaultSource) ParseCategories() {
	for i, cat := range s.Categories {
		if cat.HTML != "" {
			doc, err := parsers.FromString(cat.HTML)
			if err == nil {
				s.Categories[i].Doc = doc
			}
		}
	}
}

// FeedsToArticles returns articles from RSS feeds
func (s *DefaultSource) FeedsToArticles() []newspaper.Article {
	articles := []newspaper.Article{}

	for _, feed := range s.Feeds {
		if feed.RSS == "" {
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

			// Clean up the URL and validate
			articleURL = strings.TrimSpace(articleURL)
			if articleURL != "" && newspaper.IsValidCategoryURL(articleURL) {
				// Only include articles from the same domain as the source
				parsedArticleURL, err := urls.Parse(articleURL)
				if err != nil {
					return
				}
				if parsedArticleURL.Domain == s.ParsedURL.Domain {
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

// CategoriesToArticles returns articles from categories
// Only includes articles from the same domain as the source URL
func (s *DefaultSource) CategoriesToArticles() []newspaper.Article {
	articles := []newspaper.Article{}
	sourceDomain := s.ParsedURL.Domain

	for _, cat := range s.Categories {
		if cat.Doc == nil {
			continue
		}
		cat.Doc.Find("a").Each(func(i int, sel *goquery.Selection) {
			href, exists := sel.Attr("href")
			if exists && href != "" && href != "/" && href != "#" {
				articleURL := urls.PrepareURL(href, cat.URL)
				if newspaper.IsValidCategoryURL(articleURL) && articleURL != s.URL && articleURL != cat.URL {
					// Only include articles from the same domain as the source
					parsedArticleURL, err := urls.Parse(articleURL)
					if err != nil {
						return
					}
					fmt.Println("Found article URL:", parsedArticleURL.String())
					if parsedArticleURL.Domain == sourceDomain {
						// Skip navigation links - only include links that look like articles
						// Articles typically have IDs, dates, or specific patterns
						if s.isLikelyArticleURL(articleURL) {
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
			}
		})
	}

	return articles
}

// GenerateArticles creates the list of Article objects
func (s *DefaultSource) GenerateArticles(limit int, onlyInPath bool) {
	categoryArticles := s.CategoriesToArticles()
	feedArticles := s.FeedsToArticles()

	allArticles := append(feedArticles, categoryArticles...)

	// Remove duplicates
	uniqueMap := make(map[string]newspaper.Article)
	for _, article := range allArticles {
		uniqueMap[article.URL] = article
	}
	uniqueArticles := make([]newspaper.Article, 0, len(uniqueMap))
	for _, article := range uniqueMap {
		uniqueArticles = append(uniqueArticles, article)
	}

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

}

// DownloadArticles downloads all articles
func (s *DefaultSource) DownloadArticles() []newspaper.Article {
	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)
	for i, article := range s.Articles {
		// Simple download
		resp, err := client.Get(article.URL)
		if err != nil || resp.StatusCode >= 400 {
			continue
		}

		htmlBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		s.Articles[i].HTML = string(htmlBytes)
		s.Articles[i].DownloadState = newspaper.Success
	}
	return s.Articles
}

// ParseArticles parses all articles
func (s *DefaultSource) ParseArticles() {
	for i, article := range s.Articles {
		if article.HTML != "" {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(article.HTML))
			if err == nil {
				s.Articles[i].Doc = doc
				// Basic parsing
				title := doc.Find("title").First().Text()
				if title != "" {
					s.Articles[i].Title = title
				}
				// Add more parsing logic as needed
			}
		}
	}
	s.IsParsed = true
}

// Size returns the number of articles
func (s *DefaultSource) Size() int {
	return len(s.Articles)
}

// CleanMemoCache clears the memoization cache
func (s *DefaultSource) CleanMemoCache() {
	// TODO: implement if caching is added
}

// FeedURLs returns a list of feed URLs
func (s *DefaultSource) FeedURLs() []string {
	urls := []string{}
	for _, feed := range s.Feeds {
		urls = append(urls, feed.URL)
	}
	return urls
}

// CategoryURLs returns a list of category URLs
func (s *DefaultSource) CategoryURLs() []string {
	urls := []string{}
	for _, cat := range s.Categories {
		urls = append(urls, cat.URL)
	}
	return urls
}

// ArticleURLs returns a list of article URLs
func (s *DefaultSource) ArticleURLs() []string {
	urls := []string{}
	for _, article := range s.Articles {
		urls = append(urls, article.URL)
	}
	return urls
}

// PrintSummary prints a summary of the source
func (s *DefaultSource) PrintSummary() {
	fmt.Println(s.String())
}

// String returns a string representation of the source
func (s *DefaultSource) String() string {
	res := fmt.Sprintf("Source (\n\turl=%s\n\tbrand=%s\n\tdomain=%s\n\tlen(articles)=%d\n\tdescription=%s\n)",
		s.URL, s.ParsedURL.Brand(), s.ParsedURL.Domain, len(s.Articles), s.Description[:helpers.Min(50, len(s.Description))])

	res += "\n 10 sample Articles: \n"
	for i, article := range s.Articles[:helpers.Min(10, len(s.Articles))] {
		res += fmt.Sprintf("%d: %s\n", i+1, article.URL)
	}

	res += "\ncategory_urls: \n" + strings.Join(s.CategoryURLs(), "\n")
	res += "\nfeed_urls:\n" + strings.Join(s.FeedURLs(), "\n")

	return res
}
