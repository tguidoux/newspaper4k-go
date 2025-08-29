package newspaper4k

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/languages"
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
	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
		a.Doc = doc
	}
	// Extract metadata
	a.MetaLang = me.getMetaLanguage(a.Doc)
	a.CanonicalLink = me.getCanonicalLink(a.URL, a.Doc)
	a.MetaSiteName = me.getMetaField(a.Doc, "og:site_name")
	a.MetaDescription = me.getMetaField(a.Doc, "description", "og:description")
	a.MetaKeywords = me.getMetaKeywords(a.Doc)
	a.MetaData = me.getMetadata(a.Doc)

	return nil
}

// getMetaLanguage extracts the language from meta tags
func (me *MetadataExtractor) getMetaLanguage(doc *goquery.Document) string {
	// 1) prefer the `lang` attribute on <html>
	if raw, exists := doc.Find("html").Attr("lang"); exists {
		lang := strings.ToLower(strings.TrimSpace(raw))
		if languages.IsValidLanguageCode(lang) {
			return lang
		}
	}

	// 2) fallback to configured META_LANGUAGE_TAGS (e.g. <meta property/name=item> tags)
	for _, entry := range constants.META_LANGUAGE_TAGS {
		tag := entry["tag"]
		attr := entry["attr"]
		value := entry["value"]

		sel := doc.Find(tag + "[" + attr + "='" + value + "']").First()
		if content := getAttrContent(sel, "content"); content != "" {
			lang := strings.ToLower(strings.TrimSpace(content))
			if languages.IsValidLanguageCode(lang) {
				return lang
			}
		}
	}

	return ""
}

// getCanonicalLink extracts the canonical URL
func (me *MetadataExtractor) getCanonicalLink(articleURL string, doc *goquery.Document) string {
	// Collect candidate URLs: <link rel="canonical"> and og:url
	var candidates []string

	linkElems := parsers.GetTags(doc.Selection, "link", map[string]string{"rel": "canonical"}, "exact", false)
	for _, el := range linkElems {
		attr := parsers.GetAttribute(el, "href", nil, "")
		if hrefStr, ok := attr.(string); ok && strings.TrimSpace(hrefStr) != "" {
			candidates = append(candidates, strings.TrimSpace(hrefStr))
			break // prefer first canonical link
		}
	}

	if og := me.getMetaField(doc, "og:url"); og != "" {
		candidates = append(candidates, og)
	}

	if len(candidates) == 0 {
		return ""
	}

	raw := candidates[0]

	// Parse using net/url. If relative, resolve against articleURL.
	parsedMeta, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if parsedMeta.IsAbs() {
		return parsedMeta.String()
	}

	parsedArticle, err := urls.Parse(articleURL)
	if err != nil {
		return parsedMeta.String()
	}

	resolved := parsedArticle.ResolveReference(parsedMeta)
	return resolved.String()
}

// getMetadata extracts all metadata from meta tags
func (me *MetadataExtractor) getMetadata(doc *goquery.Document) map[string]string {
	out := make(map[string]string)

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		// Prefer common attributes in this order: property, name, itemprop
		var key string
		if v := getAttrContent(s, "property"); v != "" {
			key = v
		} else if v := getAttrContent(s, "name"); v != "" {
			key = v
		} else if v := getAttrContent(s, "itemprop"); v != "" {
			key = v
		}

		if key == "" {
			return
		}

		if content := getAttrContent(s, "content"); content != "" {
			out[strings.TrimSpace(key)] = strings.TrimSpace(content)
		}
	})

	return out
}

// getMetaField extracts a specific meta field
func (me *MetadataExtractor) getMetaField(doc *goquery.Document, fields ...string) string {
	for _, f := range fields {
		// Use parser helper to find meta tags (handles property/name/itemprop variations)
		metas := parsers.GetMetatags(doc.Selection, f)
		for _, m := range metas {
			content := parsers.GetAttribute(m, "content", nil, "")
			if contentStr, ok := content.(string); ok && strings.TrimSpace(contentStr) != "" {
				return strings.TrimSpace(contentStr)
			}
		}
	}
	return ""
}

// getMetaKeywords extracts keywords from meta tags
func (me *MetadataExtractor) getMetaKeywords(doc *goquery.Document) []string {
	ks := me.getMetaField(doc, "keywords")
	if ks == "" {
		return nil
	}

	parts := strings.Split(ks, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// getAttrContent returns the value of attr on sel or an empty string if missing.
func getAttrContent(sel *goquery.Selection, attr string) string {
	if sel == nil || sel.Length() == 0 {
		return ""
	}
	if v, ok := sel.Attr(attr); ok {
		return v
	}
	return ""
}
