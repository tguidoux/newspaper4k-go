package newspaper4k

import (
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// CategoryExtractor extracts category URLs from a news source
type CategoryExtractor struct {
	config     *configuration.Configuration
	categories []string
}

// NewCategoryExtractor creates a new CategoryExtractor
func NewCategoryExtractor(config *configuration.Configuration) *CategoryExtractor {
	return &CategoryExtractor{
		config:     config,
		categories: []string{},
	}
}

// Parse extracts categories from the source URL and updates the article in-place
func (ce *CategoryExtractor) Parse(a *newspaper.Article) error {
	ce.categories = []string{}

	var doc *goquery.Document
	var err error

	// Use Doc field if available, otherwise parse HTML using parser
	if a.Doc != nil {
		doc = a.Doc
	} else {
		doc, err = parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
	}

	categories := ce.parse(a.SourceURL, doc)
	ce.categories = categories
	a.Categories = categories

	return nil
}

// parse extracts category URLs from the source
func (ce *CategoryExtractor) parse(sourceURL string, doc *goquery.Document) []string {
	domainTLD := ce.extractTLD(sourceURL)

	linksInDoc := ce.getLinksInDoc(doc)

	categoryCandidates := []map[string]any{}

	for _, pURL := range linksInDoc {
		ok, parsedURL := ce.isValidLink(pURL, domainTLD["domain"])
		if ok {
			if parsedURL["domain"] == "" {
				parsedURL["domain"] = urls.GetDomain(sourceURL)
				parsedURL["scheme"] = urls.GetScheme(sourceURL)
				parsedURL["tld"] = domainTLD
			}
			categoryCandidates = append(categoryCandidates, parsedURL)
		}
	}

	validCategories := ce.filterValidCategories(categoryCandidates, sourceURL, domainTLD)

	if len(validCategories) == 0 {
		otherLinksInDoc := ce.getOtherLinks(doc, domainTLD["domain"])
		for _, pURL := range otherLinksInDoc {
			ok, parsedURL := ce.isValidLink(pURL, domainTLD["domain"])
			if ok {
				path := strings.ToLower(parsedURL["path"].(string))
				pathChunks := ce.getPathChunks(path)
				subdomain := strings.ToLower(parsedURL["tld"].(map[string]string)["subdomain"])
				subdomainParts := strings.Split(subdomain, ".")

				conjunction := append(pathChunks, subdomainParts...)
				stopWords := ce.getStopWords()

				intersection := ce.intersection(conjunction, stopWords)
				if len(intersection) == 0 {
					validCategories = append(validCategories, ce.buildCategoryURL(parsedURL))
				}
			}
		}
	}

	validCategories = append(validCategories, "/") // add the root
	validCategories = ce.unique(validCategories)

	categoryURLs := []string{}
	for _, pURL := range validCategories {
		if pURL != "" {
			preparedURL := urls.PrepareURL(pURL, sourceURL)
			if preparedURL != "" {
				categoryURLs = append(categoryURLs, preparedURL)
			}
		}
	}

	sort.Strings(categoryURLs)
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

	return ce.unique(links)
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

	path := urls.GetPath(candidate)
	pathChunks := ce.getPathChunks(path)
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
func (ce *CategoryExtractor) filterValidCategories(candidates []map[string]any, sourceURL string, domainTLD map[string]string) []string {
	validCategories := []string{}
	stopWords := ce.getStopWords()

	for _, pURL := range candidates {
		path := strings.ToLower(pURL["path"].(string))
		pathChunks := ce.getPathChunks(path)
		subdomain := strings.ToLower(pURL["tld"].(map[string]string)["subdomain"])
		subdomainParts := strings.Split(subdomain, ".")

		conjunction := append(pathChunks, subdomainParts...)
		intersection := ce.intersection(conjunction, stopWords)

		if len(intersection) == 0 {
			categoryURL := ce.buildCategoryURL(pURL)
			validCategories = append(validCategories, categoryURL)
		}
	}

	return validCategories
}

// isValidLink checks if a URL is a valid category link
func (ce *CategoryExtractor) isValidLink(urlStr, filterTLD string) (bool, map[string]any) {
	parsedURL := map[string]any{
		"scheme": urls.GetScheme(urlStr),
		"domain": urls.GetDomain(urlStr),
		"path":   urls.GetPath(urlStr),
		"tld":    ce.extractTLD(urlStr),
	}

	// Remove any URL that starts with #
	if path, ok := parsedURL["path"].(string); ok && strings.HasPrefix(path, "#") {
		return false, parsedURL
	}

	// Remove URLs that are not http or https
	if scheme, ok := parsedURL["scheme"].(string); ok {
		if scheme != "" && scheme != "http" && scheme != "https" {
			return false, parsedURL
		}
	}

	pathChunks := ce.getPathChunks(parsedURL["path"].(string))

	// Remove index.html
	for i, chunk := range pathChunks {
		if chunk == "index.html" {
			pathChunks = append(pathChunks[:i], pathChunks[i+1:]...)
			break
		}
	}

	if parsedURL["domain"] != "" {
		tldData := parsedURL["tld"].(map[string]string)
		childSubdomainParts := strings.Split(tldData["subdomain"], ".")

		// Domain filtering logic
		if tldData["domain"] != filterTLD && !ce.contains(childSubdomainParts, filterTLD) {
			return false, parsedURL
		}

		if ce.contains([]string{"m", "i"}, tldData["subdomain"]) {
			return false, parsedURL
		}

		subd := ""
		if tldData["subdomain"] == "www" {
			subd = ""
		} else {
			subd = tldData["subdomain"]
		}

		if len(subd) > 0 && len(pathChunks) == 0 {
			return true, parsedURL // Allow http://category.domain.tld/
		}
	}

	// We want a path with just one subdir
	if len(pathChunks) > 2 || len(pathChunks) == 0 {
		return false, parsedURL
	}

	if ce.hasInvalidPrefixes(pathChunks) {
		return false, parsedURL
	}

	if len(pathChunks) == 2 && ce.contains(CATEGORY_URL_PREFIXES, pathChunks[0]) {
		return true, parsedURL
	}

	return len(pathChunks) == 1 && len(pathChunks[0]) > 1 && len(pathChunks[0]) < 20, parsedURL
}

// getPathChunks splits path into chunks
func (ce *CategoryExtractor) getPathChunks(path string) []string {
	parts := strings.Split(path, "/")
	chunks := []string{}
	for _, part := range parts {
		if part != "" {
			chunks = append(chunks, part)
		}
	}
	return chunks
}

// hasInvalidPrefixes checks for invalid path prefixes
func (ce *CategoryExtractor) hasInvalidPrefixes(pathChunks []string) bool {
	for _, chunk := range pathChunks {
		if strings.HasPrefix(chunk, "_") || strings.HasPrefix(chunk, "#") {
			return true
		}
	}
	return false
}

// buildCategoryURL builds a category URL from parsed URL data
func (ce *CategoryExtractor) buildCategoryURL(parsedURL map[string]any) string {
	scheme := parsedURL["scheme"].(string)
	if scheme == "" {
		scheme = "http"
	}

	domain := parsedURL["domain"].(string)
	path := parsedURL["path"].(string)

	path = strings.TrimSuffix(path, "/")

	return scheme + "://" + domain + path
}

// getStopWords returns the set of stopwords
func (ce *CategoryExtractor) getStopWords() []string {
	return URL_STOPWORDS
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

// unique returns unique strings from slice
func (ce *CategoryExtractor) unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
