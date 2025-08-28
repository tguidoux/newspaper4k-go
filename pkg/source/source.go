package source

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// Category represents a category object from a news source
type Category struct {
	URL  string
	HTML string
	Doc  *goquery.Document
}

// Feed represents an RSS feed from a news source
type Feed struct {
	URL   string
	RSS   string
	Title string
}

// Source interface defines the methods for a news source
type Source interface {
	Build(inputHTML string, onlyHomepage bool, onlyInPath bool)
	Download()
	Parse()
	SetCategories()
	SetFeeds()
	SetDescription()
	DownloadCategories()
	DownloadFeeds()
	ParseCategories()
	ParseFeeds()
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
	Config       *configuration.Configuration
	Domain       string
	Scheme       string
	Categories   []Category
	Feeds        []Feed
	Articles     []newspaper.Article
	HTML         string
	Doc          *goquery.Document
	LogoURL      string
	Favicon      string
	Brand        string
	Description  string
	ReadMoreLink string
	IsParsed     bool
	IsDownloaded bool
}

type SourceRequest struct {
	URL          string
	ReadMoreLink string
	Config       *configuration.Configuration
}

// NewDefaultSource creates a new DefaultSource
func NewDefaultSource(request SourceRequest) (*DefaultSource, error) {

	url := request.URL
	readMoreLink := request.ReadMoreLink
	config := request.Config

	if url == "" || !strings.Contains(url, "://") || !strings.HasPrefix(url, "http") {
		return nil, fmt.Errorf("input url is bad")
	}

	if config == nil {
		config = configuration.NewConfiguration()
	}

	source := &DefaultSource{
		URL:          urls.PrepareURL(url, ""),
		Config:       config,
		ReadMoreLink: readMoreLink,
		Categories:   []Category{},
		Feeds:        []Feed{},
		Articles:     []newspaper.Article{},
		IsParsed:     false,
		IsDownloaded: false,
	}

	source.Domain = urls.GetDomain(source.URL)
	source.Scheme = urls.GetScheme(source.URL)
	source.Brand = extractBrand(source.URL)

	return source, nil
}

// extractBrand extracts the domain root from the URL
func extractBrand(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	parts := strings.Split(parsed.Host, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return parsed.Host
}

// Build encapsulates download and basic parsing
func (s *DefaultSource) Build(inputHTML string, onlyHomepage bool, onlyInPath bool) {
	if inputHTML != "" {
		s.HTML = inputHTML
	} else {
		s.Download()
	}
	s.Parse()

	if onlyHomepage {
		s.Categories = []Category{{URL: s.URL, HTML: s.HTML, Doc: s.Doc}}
	} else {
		s.SetCategories()
		s.DownloadCategories()
	}
	s.ParseCategories()

	if !onlyHomepage {
		s.SetFeeds()
		s.DownloadFeeds()
		// s.ParseFeeds() // TODO: implement if needed
	}

	s.GenerateArticles(5000, onlyInPath)
}

// Download downloads the HTML of the source
func (s *DefaultSource) Download() {
	client := &http.Client{
		Timeout: time.Duration(s.Config.RequestsParams.Timeout) * time.Second,
	}
	resp, err := client.Get(s.URL)
	if err != nil {
		// Handle error - could log or set a flag
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// Handle HTTP error
		return
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		return
	}
	s.HTML = string(htmlBytes)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s.HTML))
	if err != nil {
		// Handle error
		return
	}
	s.Doc = doc
	s.IsDownloaded = true
}

// Parse sets the goquery document and sets description
func (s *DefaultSource) Parse() {
	if s.Doc == nil {
		doc, err := parsers.FromString(s.HTML)
		if err != nil {
			// Handle error
			return
		}
		s.Doc = doc
	}
	s.SetDescription()
	s.IsParsed = true
}

// SetCategories sets the categories for the source
// Only includes categories from the same domain as the source URL
func (s *DefaultSource) SetCategories() {
	// Simple implementation: extract categories from links on the homepage
	if s.Doc == nil {
		return
	}

	categoryURLs := []string{}
	sourceDomain := s.Domain

	s.Doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if exists && href != "" && href != "/" && href != "#" {
			fullURL := urls.PrepareURL(href, s.URL)
			if s.isValidCategoryURL(fullURL) && fullURL != s.URL {
				// Only include URLs from the same domain as the source
				categoryDomain := urls.GetDomain(fullURL)
				if categoryDomain == sourceDomain {
					categoryURLs = append(categoryURLs, fullURL)
				}
			}
		}
	})

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueURLs := []string{}
	for _, u := range categoryURLs {
		if !seen[u] {
			seen[u] = true
			uniqueURLs = append(uniqueURLs, u)
		}
	}

	// Limit categories for demo purposes
	if len(uniqueURLs) > 5 {
		uniqueURLs = uniqueURLs[:5]
	}

	s.Categories = make([]Category, len(uniqueURLs))
	for i, u := range uniqueURLs {
		s.Categories[i] = Category{URL: u}
	}
}

// isValidCategoryURL performs basic validation for category URLs
func (s *DefaultSource) isValidCategoryURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	// Must be a valid URL with http/https
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Basic checks
	if parsedURL.Host == "" {
		return false
	}

	path := parsedURL.Path

	// Allow root path and simple paths
	if path == "" || path == "/" {
		return false // Skip homepage
	}

	// Skip URLs with query parameters (likely not categories)
	if parsedURL.RawQuery != "" {
		return false
	}

	// Skip URLs with fragments
	if parsedURL.Fragment != "" {
		return false
	}

	// Skip very long paths (likely articles, not categories)
	if len(path) > 50 {
		return false
	}

	// Skip paths with file extensions (likely files, not categories)
	if strings.Contains(path, ".") {
		parts := strings.Split(path, ".")
		if len(parts) > 1 {
			ext := strings.ToLower(parts[len(parts)-1])
			if ext != "" && len(ext) <= 4 { // Common file extensions
				return false
			}
		}
	}

	return true
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
	parsedURL, err := url.Parse(urlStr)
	if err == nil && parsedURL.RawQuery != "" {
		return true
	}

	// Default: if it's not obviously a category/navigation URL, consider it an article
	return false
}
func (s *DefaultSource) SetFeeds() {
	commonFeedSuffixes := []string{"/feed", "/feeds", "/rss"}
	commonFeedURLs := []string{}

	for _, suffix := range commonFeedSuffixes {
		commonFeedURLs = append(commonFeedURLs, s.URL+suffix)
	}

	// Check for medium.com specific feeds
	if strings.Contains(s.URL, "medium.com") {
		parsed, _ := url.Parse(s.URL)
		if strings.HasPrefix(parsed.Path, "/@") {
			newPath := "/feed/" + strings.Split(parsed.Path, "/")[1]
			newURL := s.Scheme + "://" + s.Domain + newPath
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
	validFeeds := []Feed{}
	client := &http.Client{
		Timeout: time.Duration(s.Config.RequestsParams.Timeout) * time.Second,
	}
	// Limit feed URLs to check
	if len(commonFeedURLs) > 3 {
		commonFeedURLs = commonFeedURLs[:3]
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

		feed := Feed{URL: feedURL, RSS: rss}
		validFeeds = append(validFeeds, feed)
	}

	// Add main page as a category for feed extraction
	mainCategory := Category{URL: s.URL, HTML: s.HTML, Doc: s.Doc}
	allCategories := append(s.Categories, mainCategory)

	// Extract feed URLs from categories
	feedURLs := s.extractFeedURLs(allCategories)
	for _, feedURL := range feedURLs {
		if !containsFeed(validFeeds, feedURL) {
			validFeeds = append(validFeeds, Feed{URL: feedURL})
		}
	}

	s.Feeds = validFeeds
}

// extractFeedURLs extracts feed URLs from categories
// Only includes feeds from the same domain as the source URL
func (s *DefaultSource) extractFeedURLs(categories []Category) []string {
	feedURLs := []string{}
	sourceDomain := s.Domain

	for _, cat := range categories {
		if cat.Doc == nil {
			continue
		}
		cat.Doc.Find("link[type='application/rss+xml'], link[type='application/atom+xml']").Each(func(i int, sel *goquery.Selection) {
			href, exists := sel.Attr("href")
			if exists {
				fullURL := urls.PrepareURL(href, cat.URL)
				if s.isValidCategoryURL(fullURL) {
					// Only include feed URLs from the same domain as the source
					feedDomain := urls.GetDomain(fullURL)
					if feedDomain == sourceDomain {
						feedURLs = append(feedURLs, fullURL)
					}
				}
			}
		})
	}
	return feedURLs
}

// containsFeed checks if a feed URL is already in the list
func containsFeed(feeds []Feed, url string) bool {
	for _, f := range feeds {
		if f.URL == url {
			return true
		}
	}
	return false
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
	client := &http.Client{
		Timeout: time.Duration(s.Config.RequestsParams.Timeout) * time.Second,
	}
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

// DownloadFeeds downloads RSS for all feeds
func (s *DefaultSource) DownloadFeeds() {
	client := &http.Client{
		Timeout: time.Duration(s.Config.RequestsParams.Timeout) * time.Second,
	}
	for i, feed := range s.Feeds {
		if feed.RSS != "" {
			continue // already downloaded
		}
		resp, err := client.Get(feed.URL)
		if err != nil || resp.StatusCode >= 400 {
			continue
		}

		rssBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		s.Feeds[i].RSS = string(rssBytes)
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

// ParseFeeds adds titles to feeds
func (s *DefaultSource) ParseFeeds() {
	for i, feed := range s.Feeds {
		if feed.RSS == "" {
			continue
		}
		doc, err := parsers.FromString(feed.RSS)
		if err != nil {
			continue
		}
		title := doc.Find("title").First().Text()
		if title == "" {
			title = s.Brand
		}
		s.Feeds[i].Title = title
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
			if articleURL != "" && s.isValidCategoryURL(articleURL) {
				// Only include articles from the same domain as the source
				articleDomain := urls.GetDomain(articleURL)
				if articleDomain == s.Domain {
					article := newspaper.Article{
						URL:          articleURL,
						SourceURL:    feed.URL,
						ReadMoreLink: s.ReadMoreLink,
						Config:       s.Config,
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
	sourceDomain := s.Domain

	for _, cat := range s.Categories {
		if cat.Doc == nil {
			continue
		}
		cat.Doc.Find("a").Each(func(i int, sel *goquery.Selection) {
			href, exists := sel.Attr("href")
			if exists && href != "" && href != "/" && href != "#" {
				articleURL := urls.PrepareURL(href, cat.URL)
				if s.isValidCategoryURL(articleURL) && articleURL != s.URL && articleURL != cat.URL {
					// Only include articles from the same domain as the source
					articleDomain := urls.GetDomain(articleURL)
					if articleDomain == sourceDomain {
						// Skip navigation links - only include links that look like articles
						// Articles typically have IDs, dates, or specific patterns
						if s.isLikelyArticleURL(articleURL) {
							title := sel.Text()
							article := newspaper.Article{
								URL:          articleURL,
								SourceURL:    cat.URL,
								Title:        title,
								ReadMoreLink: s.ReadMoreLink,
								Config:       s.Config,
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
	seen := make(map[string]bool)
	uniqueArticles := []newspaper.Article{}
	for _, article := range allArticles {
		if !seen[article.URL] {
			seen[article.URL] = true
			uniqueArticles = append(uniqueArticles, article)
		}
	}

	if onlyInPath {
		currentDomain := urls.GetDomain(s.URL)
		currentPath := urls.GetPath(s.URL)
		pathChunks := strings.Split(strings.Trim(currentPath, "/"), "/")
		if len(pathChunks) > 0 && (strings.HasSuffix(pathChunks[len(pathChunks)-1], ".html") || strings.HasSuffix(pathChunks[len(pathChunks)-1], ".php")) {
			pathChunks = pathChunks[:len(pathChunks)-1]
		}
		currentPath = "/" + strings.Join(pathChunks, "/") + "/"

		filteredArticles := []newspaper.Article{}
		for _, article := range uniqueArticles {
			if currentDomain == urls.GetDomain(article.URL) && strings.HasPrefix(urls.GetPath(article.URL), currentPath) {
				filteredArticles = append(filteredArticles, article)
			}
		}
		uniqueArticles = filteredArticles
	}

	if len(uniqueArticles) > limit {
		s.Articles = uniqueArticles[:limit]
	} else {
		s.Articles = uniqueArticles
	}

	// For demo purposes, limit to 10 articles max
	if len(s.Articles) > 10 {
		s.Articles = s.Articles[:10]
	}
}

// DownloadArticles downloads all articles
func (s *DefaultSource) DownloadArticles() []newspaper.Article {
	client := &http.Client{
		Timeout: time.Duration(s.Config.RequestsParams.Timeout) * time.Second,
	}
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
		s.URL, s.Brand, s.Domain, len(s.Articles), s.Description[:min(50, len(s.Description))])

	res += "\n 10 sample Articles: \n"
	for i, article := range s.Articles[:min(10, len(s.Articles))] {
		res += fmt.Sprintf("%d: %s\n", i+1, article.URL)
	}

	res += "\ncategory_urls: \n" + strings.Join(s.CategoryURLs(), "\n")
	res += "\nfeed_urls:\n" + strings.Join(s.FeedURLs(), "\n")

	return res
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
