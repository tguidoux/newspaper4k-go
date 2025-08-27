package newspaper4k

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// MetadataExtractor extracts metadata from articles
type MetadataExtractor struct {
	config *configuration.Configuration
}

// NewMetadataExtractor creates a new MetadataExtractor
func NewMetadataExtractor(config *configuration.Configuration) *MetadataExtractor {
	return &MetadataExtractor{
		config: config,
	}
}

// Parse extracts metadata from the article and updates the article in-place
func (me *MetadataExtractor) Parse(a *newspaper.Article) error {
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

	// Extract metadata
	a.MetaLang = me.getMetaLanguage(doc)
	a.CanonicalLink = me.getCanonicalLink(a.URL, doc)
	a.MetaSiteName = me.getMetaField(doc, "og:site_name")
	a.MetaDescription = me.getMetaField(doc, []string{"description", "og:description"})
	a.MetaKeywords = me.getMetaKeywords(doc)
	a.MetaData = me.getMetadata(doc)

	return nil
}

// getMetaLanguage extracts the language from meta tags
func (me *MetadataExtractor) getMetaLanguage(doc *goquery.Document) string {
	getIfValid := func(s string) string {
		if s == "" || len(s) < 2 {
			return ""
		}
		s = s[:2]
		matched, _ := regexp.MatchString(RE_LANG, s)
		if matched {
			return strings.ToLower(s)
		}
		return ""
	}

	// Check lang attribute on html element
	if lang, exists := doc.Find("html").Attr("lang"); exists {
		if valid := getIfValid(lang); valid != "" {
			return valid
		}
	}

	// Check meta language tags
	for _, elem := range META_LANGUAGE_TAGS {
		tag := elem["tag"]
		attr := elem["attr"]
		value := elem["value"]

		selection := doc.Find(tag + "[" + attr + "='" + value + "']")
		if selection.Length() > 0 {
			if content, exists := selection.First().Attr("content"); exists {
				if valid := getIfValid(content); valid != "" {
					return valid
				}
			}
		}
	}

	return ""
}

// getCanonicalLink extracts the canonical URL
func (me *MetadataExtractor) getCanonicalLink(articleURL string, doc *goquery.Document) string {
	candidates := []string{}

	// Get canonical links using parser's GetTags method
	canonicalElements := parsers.GetTags(doc.Selection, "link", map[string]string{"rel": "canonical"}, "exact", false)
	for _, element := range canonicalElements {
		href := parsers.GetAttribute(element, "href", nil, "")
		if hrefStr, ok := href.(string); ok && hrefStr != "" {
			candidates = append(candidates, hrefStr)
		}
	}

	// Get og:url
	if ogURL := me.getMetaField(doc, "og:url"); ogURL != "" {
		candidates = append(candidates, ogURL)
	}

	// Filter and clean candidates
	validCandidates := []string{}
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c != "" {
			validCandidates = append(validCandidates, c)
		}
	}

	if len(validCandidates) == 0 {
		return ""
	}

	metaURL := validCandidates[0]
	parsedMetaURL, err := url.Parse(metaURL)
	if err != nil {
		return metaURL
	}

	if parsedMetaURL.Host == "" {
		// Might not have a hostname
		parsedArticleURL, err := url.Parse(articleURL)
		if err != nil {
			return metaURL
		}

		stripHostnameInMetaPath := regexp.MustCompile(`.*` + regexp.QuoteMeta(parsedArticleURL.Host) + `(?=/)/(.*)`)
		matches := stripHostnameInMetaPath.FindStringSubmatch(parsedMetaURL.Path)
		if len(matches) > 1 {
			truePath := matches[1]
			metaURL = (&url.URL{
				Scheme: parsedArticleURL.Scheme,
				Host:   parsedArticleURL.Host,
				Path:   truePath,
			}).String()
		} else {
			metaURL = (&url.URL{
				Scheme:   parsedArticleURL.Scheme,
				Host:     parsedArticleURL.Host,
				Path:     parsedMetaURL.Path,
				RawQuery: parsedMetaURL.RawQuery,
				Fragment: parsedMetaURL.Fragment,
			}).String()
		}
	}

	return metaURL
}

// getMetadata extracts all metadata from meta tags
func (me *MetadataExtractor) getMetadata(doc *goquery.Document) map[string]string {
	data := make(map[string]string)

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		var key, value string

		if property, exists := s.Attr("property"); exists {
			key = property
			if content, exists := s.Attr("content"); exists {
				value = content
			}
		} else if name, exists := s.Attr("name"); exists {
			key = name
			if content, exists := s.Attr("content"); exists {
				value = content
			}
		} else if itemprop, exists := s.Attr("itemprop"); exists {
			key = itemprop
			if content, exists := s.Attr("content"); exists {
				value = content
			}
		}

		if key != "" && value != "" {
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)

			// Handle numeric values
			if regexp.MustCompile(`^\d+$`).MatchString(value) {
				// Keep as string for now
			}

			data[key] = value
		}
	})

	return data
}

// getMetaField extracts a specific meta field
func (me *MetadataExtractor) getMetaField(doc *goquery.Document, fields interface{}) string {
	var fieldList []string

	switch v := fields.(type) {
	case string:
		fieldList = []string{v}
	case []string:
		fieldList = v
	default:
		return ""
	}

	for _, f := range fieldList {
		// Use parser's GetMetatags method
		metaElements := parsers.GetMetatags(doc.Selection, f)
		for _, metaElement := range metaElements {
			content := parsers.GetAttribute(metaElement, "content", nil, "")
			if contentStr, ok := content.(string); ok && contentStr != "" {
				return strings.TrimSpace(contentStr)
			}
		}
	}
	return ""
}

// getMetaKeywords extracts keywords from meta tags
func (me *MetadataExtractor) getMetaKeywords(doc *goquery.Document) []string {
	keywordsStr := me.getMetaField(doc, "keywords")
	if keywordsStr == "" {
		return []string{}
	}

	keywords := strings.Split(keywordsStr, ",")
	result := make([]string, 0, len(keywords))
	for _, k := range keywords {
		trimmed := strings.TrimSpace(k)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
