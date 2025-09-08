package source

import (
	"fmt"
	"io"
	"math/rand/v2"
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
	return s.BuildWithParams(DefaultBuildParams())
}

// Build encapsulates download and basic parsing
func (s *DefaultSource) BuildWithParams(params BuildParams) error {

	// Step 1: Download and parse homepage
	// if InputHTML is provided, use it instead of downloading
	if params.InputHTML != "" {
		s.HTML = params.InputHTML
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
	if params.OnlyHomepage {
		s.Categories = []newspaper.Category{{URL: s.URL, HTML: s.HTML, Doc: s.Doc}}
	} else {
		err := s.SearchCategories()
		if err != nil {
			return fmt.Errorf("failed to set categories: %v", err)
		}
		s.DownloadCategories()
	}
	s.BuildCategories()

	if len(s.Categories) > params.LimitCategories {
		s.Categories = s.Categories[:params.LimitCategories]
	}

	// Step 3: Download and parse feed
	// we skip feeds if onlyHomepage is true
	if !params.OnlyHomepage {
		s.GetFeedsWithParams(params)
	}

	return nil
}

// Download downloads the HTML of the source
func (s *DefaultSource) Download() error {

	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)
	resp, err := client.Get(s.URL)
	if err != nil {
		// Handle error - could log or set a flag
		return fmt.Errorf("failed to download: %v", err)
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

	var validCategories []newspaper.Category

	for _, cat := range s.Categories {
		err := s.downloadCategory(&cat)
		if err == nil {
			validCategories = append(validCategories, cat)
		}
	}

	s.Categories = validCategories
}

// downloadCategories downloads HTML for all categories
func (s *DefaultSource) downloadCategory(category *newspaper.Category) error {

	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)
	resp, err := client.Get(category.URL)
	if err != nil || resp.StatusCode >= 400 {
		return fmt.Errorf("failed to get category")
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to get category body")
	}
	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to close category body")
	}
	category.HTML = string(htmlBytes)
	return nil
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

func (s *DefaultSource) getCommonFeeds() []string {
	commonFeedSuffixes := constants.COMMON_FEED_SUFFIXES
	commonFeedURLs := []string{}

	for _, suffix := range commonFeedSuffixes {
		commonFeedURLs = append(commonFeedURLs, s.URL+suffix)
	}

	// medium.com special-case
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

	if s.Config.MaxFeeds > 0 && len(commonFeedURLs) > s.Config.MaxFeeds {
		commonFeedURLs = commonFeedURLs[:s.Config.MaxFeeds]
	}

	return commonFeedURLs
}

func (s *DefaultSource) checkFeed(feedURL string) (string, bool, error) {
	client := helpers.CreateHTTPClient(s.Config.RequestsParams.Timeout)

	resp, err := client.Get(feedURL)
	if err != nil || resp.StatusCode >= 300 {
		return "", false, fmt.Errorf("invalid status code while fetching rss")
	}

	rssBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, fmt.Errorf("failed to read rss")
	}
	err = resp.Body.Close()
	if err != nil {
		return "", false, fmt.Errorf("failed to close rss body")
	}
	rss := string(rssBytes)

	return rss, true, nil
}

func (s *DefaultSource) GetFeedsWithParams(params BuildParams) {

	commonFeedURLs := s.getCommonFeeds()

	// Download and check feeds
	validFeeds := []newspaper.Feed{}

	for _, feedURL := range commonFeedURLs {
		url := urls.PrepareURL(feedURL, feedURL)
		rss, valid, err := s.checkFeed(url)
		if valid && err != nil {
			feed := newspaper.Feed{URL: url, RSS: rss}
			validFeeds = append(validFeeds, feed)
		}
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

func (s *DefaultSource) GetFeeds() {
	s.GetFeedsWithParams(DefaultBuildParams())
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
						SourceURL: s.ParsedURL.String(),
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
							SourceURL: s.ParsedURL.String(),
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

func (s *DefaultSource) GetArticlesWithParams(params BuildParams) []newspaper.Article {
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

	if params.OnlySameDomain {

		filteredArticles := []newspaper.Article{}
		for _, article := range uniqueArticles {
			parsedArticleURL, err := urls.Parse(article.URL)
			if err != nil {
				continue
			}
			if s.ParsedURL.Domain == parsedArticleURL.Domain {
				if params.AllowSubDomain || (!params.AllowSubDomain && s.ParsedURL.Subdomain == parsedArticleURL.Subdomain) {
					filteredArticles = append(filteredArticles, article)
				}
			}
		}
		uniqueArticles = filteredArticles
	}

	if params.Shuffle {
		rand.Shuffle(len(uniqueArticles), func(i, j int) {
			uniqueArticles[i], uniqueArticles[j] = uniqueArticles[j], uniqueArticles[i]
		})
	}

	if params.LimitArticles > 0 && len(uniqueArticles) > params.LimitArticles {
		s.Articles = uniqueArticles[:params.LimitArticles]
	} else {
		s.Articles = uniqueArticles
	}

	return s.Articles
}

// GetArticles creates the list of Article objects
func (s *DefaultSource) GetArticles() []newspaper.Article {
	return s.GetArticlesWithParams(DefaultBuildParams())
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
