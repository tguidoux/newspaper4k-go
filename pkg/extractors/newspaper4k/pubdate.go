package newspaper4k

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/araddon/dateparse"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// DateMatch represents a date with a score
type DateMatch struct {
	date  time.Time
	score int
}

// PubdateExtractor extracts publication dates from articles.
type PubdateExtractor struct {
	config  *configuration.Configuration
	pubdate *time.Time
}

// NewPubdateExtractor creates a new PubdateExtractor.
func NewPubdateExtractor(config *configuration.Configuration) *PubdateExtractor {
	return &PubdateExtractor{
		config: config,
	}
}

// Parse extracts the publication date and updates the article in-place
func (p *PubdateExtractor) Parse(a *newspaper.Article) error {
	p.pubdate = nil

	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
		a.Doc = doc
	}

	// Call the existing parsing logic
	pubdate := p.parseWithDoc(a.URL, a.Doc)
	a.PublishDate = pubdate
	return nil
}

// parseWithDoc extracts the publication date using multiple strategies.
func (p *PubdateExtractor) parseWithDoc(articleURL string, doc *goquery.Document) *time.Time {
	// Helper function to parse date string
	parseDateStr := func(dateStr string) *time.Time {
		if dateStr == "" {
			return nil
		}
		t, err := dateparse.ParseAny(dateStr)
		if err != nil {
			return nil
		}
		return &t
	}

	var dateMatches []DateMatch

	// Strategy 1: Pubdate from URL
	strictDateRegex := regexp.MustCompile(`\d{4}[/-]\d{1,2}[/-]\d{1,2}`)
	if match := strictDateRegex.FindString(articleURL); match != "" {
		if dt := parseDateStr(match); dt != nil {
			dateMatches = append(dateMatches, DateMatch{date: *dt, score: 10})
		}
	}

	// Strategy 2: Pubdate from JSON-LD or structured data using parser
	jsonObjects := parsers.GetLdJsonObject(doc.Selection)
	for _, jsonData := range jsonObjects {
		dateMatches = p.extractDateFromJSON(jsonData, dateMatches)
	}

	// Strategy 3: Pubdate from <time> tags
	doc.Find("time").Each(func(i int, s *goquery.Selection) {
		if datetime, exists := s.Attr("datetime"); exists {
			if dt := parseDateStr(datetime); dt != nil {
				score := 5
				text := strings.ToLower(s.Text())
				if strings.Contains(text, "published") || strings.Contains(text, "on:") {
					score = 8
				}
				dateMatches = append(dateMatches, DateMatch{date: *dt, score: score})
			}
		}
	})

	// Strategy 4: Pubdate from meta tags using parser
	for _, metaInfo := range constants.PUBLISH_DATE_META_INFO {
		metaElements := parsers.GetMetatags(doc.Selection, metaInfo)
		for _, metaElement := range metaElements {
			content := parsers.GetAttribute(metaElement, "content", nil, "")
			if contentStr, ok := content.(string); ok && contentStr != "" {
				if dt := parseDateStr(contentStr); dt != nil {
					score := 6
					// Check if it's a meta tag (it should be since we're getting from GetMetatags)
					score += 1
					daysDiff := int(time.Since(*dt).Hours() / 24)
					if daysDiff < 0 {
						score -= 2
					} else if daysDiff > 25*365 {
						score -= 1
					}
					dateMatches = append(dateMatches, DateMatch{date: *dt, score: score})
				}
			}
		}
	}

	// Sort by score descending
	sort.Slice(dateMatches, func(i, j int) bool {
		return dateMatches[i].score > dateMatches[j].score
	})

	if len(dateMatches) > 0 {
		p.pubdate = &dateMatches[0].date
		return p.pubdate
	}
	return nil
}

// extractDateFromJSON extracts dates from JSON-LD data
func (p *PubdateExtractor) extractDateFromJSON(data interface{}, dateMatches []DateMatch) []DateMatch {
	switch v := data.(type) {
	case map[string]interface{}:
		if graph, ok := v["@graph"]; ok {
			if graphSlice, ok := graph.([]interface{}); ok {
				for _, item := range graphSlice {
					if itemMap, ok := item.(map[string]interface{}); ok {
						dateMatches = p.extractDateFromMap(itemMap, dateMatches, 10)
					}
				}
			}
		} else {
			dateMatches = p.extractDateFromMap(v, dateMatches, 9)
		}
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				dateMatches = p.extractDateFromMap(itemMap, dateMatches, 9)
			}
		}
	}
	return dateMatches
}

// extractDateFromMap extracts dates from a map
func (p *PubdateExtractor) extractDateFromMap(data map[string]interface{}, dateMatches []DateMatch, score int) []DateMatch {
	for _, key := range []string{"datePublished", "dateCreated"} {
		if dateStr, ok := data[key]; ok {
			if str, ok := dateStr.(string); ok {
				if dt := p.parseDateStr(str); dt != nil {
					dateMatches = append(dateMatches, DateMatch{date: *dt, score: score})
				}
			}
		}
	}
	return dateMatches
}

// parseDateStr parses a date string
func (p *PubdateExtractor) parseDateStr(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	t, err := dateparse.ParseAny(dateStr)
	if err != nil {
		return nil
	}
	return &t
}
