package urls

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/publicsuffix"
)

const (
	// Date regex patterns for detecting dates in URLs
	dateRegexPattern      = `([\./\-_\s]?(19|20)\d{2})[\./\-_\s]?(([0-3]?[0-9][\./\-_\s])|(\w{3,5}[\./\-_\s]))([0-3]?[0-9]([\./\-\+\?]|$))`
	strictDateRegexPrefix = `(?<=\W)`
)

var (
	// Allowed file types for articles
	allowedTypes = []string{
		"html", "htm", "md", "rst", "aspx", "jsp", "rhtml", "cgi",
		"xhtml", "jhtml", "asp", "shtml",
	}

	// Good path keywords that indicate article content
	goodPaths = []string{
		"story", "article", "feature", "featured", "slides",
		"slideshow", "gallery", "news", "video", "media", "v",
		"radio", "press",
	}

	// Bad path chunks that indicate non-article content
	badChunks = []string{
		"careers", "contact", "about", "faq", "terms", "privacy",
		"advert", "preferences", "feedback", "info", "browse",
		"howto", "account", "subscribe", "donate", "shop", "admin",
		"auth_user", "emploi", "annonces", "blog", "courrierdeslecteurs",
		"page_newsletters", "adserver", "clicannonces", "services",
		"contribution", "boutique", "espaceclient",
	}

	// Bad domains to filter out
	badDomains = []string{
		"amazon", "doubleclick", "twitter", "facebook", "google",
		"youtube", "instagram", "pinterest",
	}

	// Compiled regex patterns
	dateRegex = regexp.MustCompile(dateRegexPattern)
)

type URL struct {
	Domain    string
	Subdomain string
	TLD       string
	Port      string
	ICANN     bool
	*url.URL
}

func (u *URL) Brand() string {
	return u.Domain
}

func (u *URL) Copy() *URL {
	if u == nil {
		return nil
	}
	copiedURL := *u.URL
	return &URL{
		Domain:    u.Domain,
		Subdomain: u.Subdomain,
		TLD:       u.TLD,
		Port:      u.Port,
		ICANN:     u.ICANN,
		URL:       &copiedURL,
	}
}

func (u *URL) Prepare(source *URL) *URL {
	if u == nil {
		return nil
	}
	prepared := u.Copy()
	preparedStr := PrepareURL(prepared.String(), source.String())
	parsed, err := Parse(preparedStr)
	if err != nil {
		return prepared
	}
	return parsed
}

// RedirectBack handles URL redirects from services like Pinterest
// that redirect to their site with the real news URL as a GET parameter
func RedirectBack(urlStr, sourceDomain string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	query := parsedURL.RawQuery

	// Parse query parameters
	queryValues, err := url.ParseQuery(query)
	if err != nil {
		return urlStr
	}

	// Check for 'url' parameter
	if urls := queryValues["url"]; len(urls) > 0 {
		return urls[0]
	}

	return urlStr
}

// PrepareURL cleans a URL, removes arguments, handles redirects,
// and merges relative URLs with absolute ones
func PrepareURL(urlStr string, sourceURL string) string {
	var properURL string

	if sourceURL != "" {
		sourceParsed, err := url.Parse(sourceURL)
		if err != nil {
			return ""
		}
		sourceDomain := sourceParsed.Host

		properURL = joinURL(sourceURL, urlStr)
		properURL = RedirectBack(properURL, sourceDomain)
	} else {
		properURL = urlStr
	}

	return properURL
}

// joinURL joins a base URL with a relative URL
func joinURL(baseURL, relativeURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return relativeURL
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}

	return base.ResolveReference(rel).String()
}

func ValidateURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// ValidURL checks if a URL is a valid news article URL
func ValidURL(urlStr string, test bool) bool {
	// If testing, preprocess the URL
	if test {
		urlStr = PrepareURL(urlStr, "")
	}

	// Check minimum length (11 chars is shortest valid URL, e.g., http://x.co)
	if urlStr == "" || len(urlStr) < 11 {
		return false
	}

	// Check for mailto or missing http/https
	if strings.Contains(urlStr, "mailto:") ||
		(!strings.Contains(urlStr, "http://") && !strings.Contains(urlStr, "https://")) {
		return false
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	path := parsedURL.Path

	// Input URL is not in valid form
	if !strings.HasPrefix(path, "/") {
		return false
	}

	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")

	// Split path into chunks
	pathChunks := strings.Split(path, "/")
	var filteredChunks []string
	for _, chunk := range pathChunks {
		if chunk != "" {
			filteredChunks = append(filteredChunks, chunk)
		}
	}

	// Extract file type
	if len(filteredChunks) > 0 {
		fileType := URLToFileType(urlStr)
		if fileType != "" && !slices.Contains(allowedTypes, fileType) {
			return false
		}

		lastChunkParts := strings.Split(filteredChunks[len(filteredChunks)-1], ".")
		if len(lastChunkParts) > 1 {
			filteredChunks[len(filteredChunks)-1] = lastChunkParts[len(lastChunkParts)-2]
		}
	}

	// Remove "index" chunks
	for i := len(filteredChunks) - 1; i >= 0; i-- {
		if filteredChunks[i] == "index" {
			filteredChunks = append(filteredChunks[:i], filteredChunks[i+1:]...)
		}
	}

	// Extract TLD data (simplified version)
	tldData, err := Parse(urlStr)
	if err != nil {
		return false
	}
	subdomain := tldData.Subdomain
	tld := strings.ToLower(tldData.Domain)

	urlSlug := ""
	if len(filteredChunks) > 0 {
		urlSlug = filteredChunks[len(filteredChunks)-1]
	}

	// Check bad domains
	if slices.Contains(badDomains, tld) {
		return false
	}

	dashCount := strings.Count(urlSlug, "-")
	underscoreCount := strings.Count(urlSlug, "_")

	// Check for news slug pattern
	if urlSlug != "" && (dashCount > 4 || underscoreCount > 4) {
		var parts []string

		if dashCount >= underscoreCount {
			parts = strings.Split(urlSlug, "-")
		} else {
			parts = strings.Split(urlSlug, "_")
		}

		// Check if TLD is not in the slug parts
		tldInParts := false
		for _, part := range parts {
			if strings.ToLower(part) == tld {
				tldInParts = true
				break
			}
		}

		if !tldInParts {
			return true
		}
	}

	// Must have at least 2 path chunks
	if len(filteredChunks) <= 1 {
		return false
	}

	// Check for bad chunks in path or subdomain
	for _, badChunk := range badChunks {
		if slices.Contains(filteredChunks, badChunk) || badChunk == subdomain {
			return false
		}
	}

	// Check for date pattern
	if dateRegex.MatchString(urlStr) {
		return true
	}

	// Check for numeric ID patterns
	if len(filteredChunks) >= 2 && len(filteredChunks) <= 3 {
		lastChunk := filteredChunks[len(filteredChunks)-1]
		if matched, _ := regexp.MatchString(`\d{3,}$`, lastChunk); matched {
			return true
		}

		if len(filteredChunks) == 3 {
			middleChunk := filteredChunks[1]
			if matched, _ := regexp.MatchString(`\d{3,}$`, middleChunk); matched {
				return true
			}
		}
	}

	// Check for good paths
	for _, goodPath := range goodPaths {
		for _, chunk := range filteredChunks {
			if strings.EqualFold(chunk, goodPath) {
				return true
			}
		}
	}

	// If URL has an allowed file type and passes basic checks, consider it valid
	if len(filteredChunks) >= 2 {
		fileType := URLToFileType(urlStr)
		if fileType != "" && slices.Contains(allowedTypes, fileType) {
			return true
		}
	}

	return false
}

// URLToFileType extracts the file type from a URL
func URLToFileType(absURL string) string {
	parsedURL, err := url.Parse(absURL)
	if err != nil {
		return ""
	}

	path := strings.TrimSuffix(parsedURL.Path, "/")

	pathChunks := strings.Split(path, "/")
	if len(pathChunks) == 0 {
		return ""
	}

	lastChunk := pathChunks[len(pathChunks)-1]
	parts := strings.Split(lastChunk, ".")

	if len(parts) < 2 {
		return ""
	}

	fileType := parts[len(parts)-1]

	// Assume file extension is maximum 5 characters long
	if len(fileType) <= 5 || slices.Contains(allowedTypes, strings.ToLower(fileType)) {
		return strings.ToLower(fileType)
	}

	return ""
}

// IsAbsURL checks if a URL is absolute
func IsAbsURL(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// URLJoinIfValid joins a base URL and a possibly relative URL safely
func URLJoinIfValid(baseURL, relativeURL string) string {
	result := joinURL(baseURL, relativeURL)
	return result
}

// GetPathChunks splits path into chunks
func GetPathChunks(path string) []string {
	parts := strings.Split(path, "/")
	chunks := []string{}
	for _, part := range parts {
		if part != "" {
			chunks = append(chunks, part)
		}
	}
	return chunks
}

// Parse mirrors net/url.Parse except instead it returns
// a URL, which contains extra fields.
func Parse(s string) (*URL, error) {
	url, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	if url.Host == "" {
		return &URL{URL: url}, nil
	}
	dom, port := domainPort(url.Host)
	//etld+1
	etld1, err := publicsuffix.EffectiveTLDPlusOne(dom)
	suffix, icann := publicsuffix.PublicSuffix(strings.ToLower(dom))
	// HACK: attempt to support valid domains which are not registered with ICAN
	if err != nil && !icann && suffix == dom {
		etld1 = dom
		err = nil
	}
	if err != nil {
		return nil, err
	}
	//convert to domain name, and tld
	i := strings.Index(etld1, ".")
	if i < 0 {
		return nil, fmt.Errorf("tld: failed parsing %q", s)
	}
	domName := etld1[0:i]
	tld := etld1[i+1:]
	//and subdomain
	sub := ""
	if rest := strings.TrimSuffix(dom, "."+etld1); rest != dom {
		sub = rest
	}
	return &URL{
		Subdomain: sub,
		Domain:    domName,
		TLD:       tld,
		Port:      port,
		ICANN:     icann,
		URL:       url,
	}, nil
}

func domainPort(host string) (string, string) {
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i], host[i+1:]
		} else if host[i] < '0' || host[i] > '9' {
			return host, ""
		}
	}
	//will only land here if the string is all digits,
	//net/url should prevent that from happening
	return host, ""
}
