package newspaper4k

import (
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// CategoryExtractor extracts category URLs from a news source
type CategoryExtractor struct {
	config     *configuration.Configuration
	categories []*urls.URL
}

// NewCategoryExtractor creates a new CategoryExtractor
func NewCategoryExtractor(config *configuration.Configuration) *CategoryExtractor {
	return &CategoryExtractor{
		config:     config,
		categories: []*urls.URL{},
	}
}

// Parse extracts categories from the source URL and updates the article in-place
func (ce *CategoryExtractor) Parse(a *newspaper.Article) error {
	ce.categories = []*urls.URL{}

	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
		a.Doc = doc
	}

	categories := ce.parse(a.SourceURL, a.Doc)
	ce.categories = categories
	a.Categories = categories

	return nil
}

// parse extracts category URLs from the source
func (ce *CategoryExtractor) parse(sourceURL string, doc *goquery.Document) []*urls.URL {
	parsedSourceURL, err := urls.Parse(sourceURL)
	if err != nil {
		return []*urls.URL{}
	}

	linksInDoc := ce.getLinksInDoc(doc)

	categoryCandidates := []*urls.URL{}

	for _, pURL := range linksInDoc {
		ok := newspaper.IsValidCategoryURL(pURL)

		if ok {
			parsedURL, err := urls.Parse(pURL)

			if err != nil {
				continue
			}

			if parsedURL.Domain == "" || parsedURL.TLD == "" {
				continue
			}

			categoryCandidates = append(categoryCandidates, parsedURL)
		}
	}

	validCategories := ce.filterValidCategories(categoryCandidates, parsedSourceURL)

	if len(validCategories) == 0 {
		otherLinksInDoc := ce.getOtherLinks(doc, parsedSourceURL.Domain)
		for _, pURL := range otherLinksInDoc {
			ok := newspaper.IsValidCategoryURL(pURL)
			if ok {

				parsedURL, err := urls.Parse(pURL)
				if err != nil {
					continue
				}

				pathChunks := parsedURL.GetPathChunks()
				subdomain := strings.ToLower(parsedURL.Subdomain)
				subdomainParts := strings.Split(subdomain, ".")

				conjunction := append(pathChunks, subdomainParts...)
				stopWords := constants.URL_STOPWORDS

				intersection := ce.intersection(conjunction, stopWords)
				if len(intersection) == 0 {
					validCategories = append(validCategories, parsedURL)
				}
			}
		}
	}

	// RootCategory URL
	rootCategory := parsedSourceURL.Copy()
	rootCategory.Path = "/"
	validCategories = append(validCategories, rootCategory) // add the root

	categoryURLs := []*urls.URL{}
	for _, pURL := range validCategories {
		if pURL.String() != "" {
			err := pURL.Prepare(parsedSourceURL)
			if err != nil {
				continue
			}
			categoryURLs = append(categoryURLs, pURL)
		}
	}

	sort.SliceStable(categoryURLs, func(i, j int) bool {
		return categoryURLs[i].String() < categoryURLs[j].String()
	})
	categoryURLs = slices.Compact(categoryURLs)
	return categoryURLs
}

// extractTLD extracts TLD information from a URL
func (ce *CategoryExtractor) extractTLD(urlStr string) map[string]string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return map[string]string{"domain": "", "subdomain": ""}
	}

	host := parsedURL.Host
	parts := strings.Split(host, ".")

	result := map[string]string{
		"domain":    "",
		"subdomain": "",
	}

	if len(parts) >= 2 {
		result["domain"] = parts[len(parts)-2]
		if len(parts) >= 3 {
			result["subdomain"] = strings.Join(parts[:len(parts)-2], ".")
		}
	}

	return result
}

// getLinksInDoc gets all href links from anchor tags
func (ce *CategoryExtractor) getLinksInDoc(doc *goquery.Document) []string {
	links := []string{}

	// Use parser's GetElementsByTagslist method
	linkElements := parsers.GetElementsByTagslist(doc.Selection, []string{"a"})
	for _, element := range linkElements {
		href := parsers.GetAttribute(element, "href", nil, "")
		if hrefStr, ok := href.(string); ok && hrefStr != "" {
			links = append(links, hrefStr)
		}
	}

	return slices.Compact(links)
}

// getOtherLinks gets links from non-anchor sources (javascript, json, etc.)
func (ce *CategoryExtractor) getOtherLinks(doc *goquery.Document, filterTLD string) []string {
	html := ce.nodeToString(doc)
	candidates := ce.findURLsInHTML(html)

	candidates = ce.cleanURLs(candidates)

	filtered := []string{}
	for _, candidate := range candidates {
		if ce.filterOtherLink(candidate, filterTLD) {
			filtered = append(filtered, candidate)
		}
	}

	return filtered
}

// nodeToString converts a goquery document to string
func (ce *CategoryExtractor) nodeToString(doc *goquery.Document) string {
	html, _ := doc.Html()
	return html
}

// findURLsInHTML finds URLs in HTML using regex
func (ce *CategoryExtractor) findURLsInHTML(html string) []string {
	re := regexp.MustCompile(`"(https?://[^"]*)"`)

	matches := re.FindAllStringSubmatch(html, -1)
	urls := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, match[1])
		}
	}

	return urls
}

// cleanURLs cleans escaped URLs
func (ce *CategoryExtractor) cleanURLs(urls []string) []string {
	cleaned := []string{}
	for _, u := range urls {
		u = strings.ReplaceAll(u, `\/`, "/")
		u = strings.ReplaceAll(u, `/\/`, "/")
		cleaned = append(cleaned, u)
	}
	return cleaned
}

// filterOtherLink filters other links based on criteria
func (ce *CategoryExtractor) filterOtherLink(candidate, filterTLD string) bool {
	if filterTLD != "" {
		tldData := ce.extractTLD(candidate)
		if tldData["domain"] != filterTLD {
			return false
		}
	}

	if ce.isAssetURL(candidate) {
		return false
	}

	parsedCandidate, err := urls.Parse(candidate)
	if err != nil {
		return false
	}

	pathChunks := parsedCandidate.GetPathChunks()
	if len(pathChunks) > 2 || len(pathChunks) == 0 {
		return false
	}

	return true
}

// isAssetURL checks if URL is an asset (css, js, etc.)
func (ce *CategoryExtractor) isAssetURL(urlStr string) bool {
	assetExtensions := []string{".css", ".js", ".json", ".xml", ".rss", ".jpg", ".jpeg", ".png"}
	for _, ext := range assetExtensions {
		if strings.Contains(strings.ToLower(urlStr), ext) {
			return true
		}
	}
	return false
}

// filterValidCategories filters category candidates
func (ce *CategoryExtractor) filterValidCategories(candidates []*urls.URL, sourceURL *urls.URL) []*urls.URL {
	validCategories := []*urls.URL{}
	stopWords := constants.URL_STOPWORDS

	for _, pURL := range candidates {
		pathChunks := pURL.GetPathChunks()
		subdomain := strings.ToLower(pURL.Subdomain)
		subdomainParts := strings.Split(subdomain, ".")

		conjunction := append(pathChunks, subdomainParts...)
		intersection := ce.intersection(conjunction, stopWords)

		if len(intersection) == 0 {
			validCategories = append(validCategories, pURL)
		}
	}

	return validCategories
}

// intersection returns intersection of two string slices
func (ce *CategoryExtractor) intersection(a, b []string) []string {
	result := []string{}
	for _, item := range a {
		if ce.contains(b, item) {
			result = append(result, item)
		}
	}
	return result
}

// contains checks if slice contains string
func (ce *CategoryExtractor) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
