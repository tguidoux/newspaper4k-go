package newspaper4k

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/helpers"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// AuthorsExtractor extracts author information from articles
type AuthorsExtractor struct {
	config  *configuration.Configuration
	authors []string
}

// NewAuthorsExtractor creates a new AuthorsExtractor
func NewAuthorsExtractor(config *configuration.Configuration) *AuthorsExtractor {
	return &AuthorsExtractor{
		config:  config,
		authors: []string{},
	}
}

// Parse extracts authors from the article and updates the article in-place
func (ae *AuthorsExtractor) Parse(a *newspaper.Article) error {
	ae.authors = []string{}

	doc, err := helpers.GetDocFromArticle(a)
	if err != nil {
		return err
	}

	authors := ae.extractAuthors(doc)

	// Clean up authors of stopwords
	authors = ae.cleanAuthors(authors)

	// Remove duplicates while preserving order
	authors = helpers.UniqueStrings(authors, helpers.UniqueOptions{
		CaseSensitive: false,
		PreserveOrder: true,
	})

	ae.authors = authors
	a.Authors = authors

	return nil
}

// extractAuthors extracts authors from various sources
func (ae *AuthorsExtractor) extractAuthors(doc *goquery.Document) []string {
	authors := []string{}

	// Try 1: Search JSON-LD structured data for authors
	authors = append(authors, ae.extractFromJSONLD(doc)...)

	// Try 2: Search popular author tags for authors
	authors = append(authors, ae.extractFromAuthorTags(doc)...)

	return authors
}

// extractFromJSONLD extracts authors from JSON-LD structured data
func (ae *AuthorsExtractor) extractFromJSONLD(doc *goquery.Document) []string {
	authors := []string{}

	// Use parser's GetLdJsonObject method
	jsonObjects := parsers.GetLdJsonObject(doc.Selection)

	for _, jsonData := range jsonObjects {
		// Handle @graph structure
		if graph, exists := jsonData["@graph"]; exists {
			if graphArray, ok := graph.([]interface{}); ok {
				for _, item := range graphArray {
					if itemMap, ok := item.(map[string]interface{}); ok {
						// Check for Person type
						if itemType, exists := itemMap["@type"]; exists && itemType == "Person" {
							if name, exists := itemMap["name"]; exists {
								if nameStr, ok := name.(string); ok {
									authors = append(authors, nameStr)
								}
							}
						}
						// Check for author field
						if author, exists := itemMap["author"]; exists {
							authors = append(authors, ae.extractAuthorNames(author)...)
						}
					}
				}
			}
		} else {
			// Check for author field in root
			if author, exists := jsonData["author"]; exists {
				authors = append(authors, ae.extractAuthorNames(author)...)
			}
		}
	}

	return authors
}

// extractAuthorNames extracts author names from various JSON-LD structures
func (ae *AuthorsExtractor) extractAuthorNames(author interface{}) []string {
	names := []string{}

	switch v := author.(type) {
	case string:
		names = append(names, v)
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if name, exists := itemMap["name"]; exists {
					if nameStr, ok := name.(string); ok {
						names = append(names, nameStr)
					}
				}
			} else if itemStr, ok := item.(string); ok {
				names = append(names, itemStr)
			}
		}
	case map[string]interface{}:
		if name, exists := v["name"]; exists {
			if nameStr, ok := name.(string); ok {
				names = append(names, nameStr)
			} else if nameArray, ok := name.([]interface{}); ok {
				for _, n := range nameArray {
					if nStr, ok := n.(string); ok {
						names = append(names, nStr)
					}
				}
			}
		}
	}

	return names
}

// extractFromAuthorTags extracts authors from HTML elements with author-related attributes
func (ae *AuthorsExtractor) extractFromAuthorTags(doc *goquery.Document) []string {
	authors := []string{}

	// Search for elements with author-related attributes and values
	for _, attr := range newspaper.AUTHOR_ATTRS {
		for _, val := range newspaper.AUTHOR_VALS {
			// Use parser's GetTags method for attribute-based searching
			attribs := map[string]string{attr: val}
			elements := parsers.GetTags(doc.Selection, "", attribs, "exact", false)

			for _, element := range elements {
				content := parsers.GetText(element)
				if content != "" {
					authors = append(authors, ae.parseByline(content)...)
				}
			}

			// Also search for meta tags with these attributes
			metaElements := parsers.GetMetatags(doc.Selection, val)
			for _, metaElement := range metaElements {
				content := parsers.GetAttribute(metaElement, "content", nil, "")
				if contentStr, ok := content.(string); ok && contentStr != "" {
					authors = append(authors, ae.parseByline(contentStr)...)
				}
			}
		}
	}

	return authors
}

// parseByline parses author names from a line of text
func (ae *AuthorsExtractor) parseByline(searchStr string) []string {
	// Remove HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]+>`)
	searchStr = htmlRegex.ReplaceAllString(searchStr, "")

	// Clean whitespace
	searchStr = regexp.MustCompile(`[\n\t\r\xa0]`).ReplaceAllString(searchStr, " ")

	// Remove "By:" or "From:" prefixes
	byRegex := regexp.MustCompile(`(?i)\b(by|from)[:\s](.*)`)
	if matches := byRegex.FindStringSubmatch(searchStr); len(matches) > 2 {
		searchStr = matches[2]
	}

	searchStr = strings.TrimSpace(searchStr)

	// Split by common separators
	nameTokens := regexp.MustCompile(`[·,|]|\sand\s|\set\s|\sund\s|\/`).Split(searchStr, -1)

	// Clean and filter tokens
	validTokens := []string{}
	for _, token := range nameTokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		// Skip if contains digits
		if regexp.MustCompile(`\d`).MatchString(token) {
			continue
		}

		// Skip if not 2-5 words or 1-5 words with reasonable length
		words := regexp.MustCompile(`\w+`).FindAllString(token, -1)
		if len(words) < 2 || len(words) > 5 {
			continue
		}

		validTokens = append(validTokens, token)
	}

	return validTokens
}

// cleanAuthors removes stopwords from author names
func (ae *AuthorsExtractor) cleanAuthors(authors []string) []string {
	// Create regex pattern for stopwords
	stopwordsPattern := strings.Join(newspaper.AUTHOR_STOP_WORDS, "|")
	stopwordsRegex := regexp.MustCompile(`(?i)\b(` + stopwordsPattern + `)\b`)

	cleaned := []string{}
	for _, author := range authors {
		// Remove stopwords
		cleanedAuthor := stopwordsRegex.ReplaceAllString(author, "")
		// Clean up extra punctuation and whitespace
		cleanedAuthor = regexp.MustCompile(`^[^\w]+|[^\w]+$`).ReplaceAllString(cleanedAuthor, "")
		cleanedAuthor = strings.TrimSpace(cleanedAuthor)

		if cleanedAuthor != "" {
			cleaned = append(cleaned, cleanedAuthor)
		}
	}

	return cleaned
}
