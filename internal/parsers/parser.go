package parsers

import (
	"container/list"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/nlp"
	htmlpkg "golang.org/x/net/html"
)

// DropTags removes the tag(s), but not its children or text.
// The children and text are merged into the parent.
func DropTags(nodes ...*goquery.Selection) {
	for _, node := range nodes {
		if node != nil {
			node.Remove()
		}
	}
}

// GetUnicodeHTML handles encoding detection and returns proper UTF-8 HTML
func GetUnicodeHTML(htmlContent string) string {
	// Go natively handles UTF-8, so we just return the string as-is
	// In the original Python code, this used UnicodeDammit from bs4
	// but Go's string type is already UTF-8
	return htmlContent
}

// FromString parses HTML string into a goquery document
func FromString(htmlContent string) (*goquery.Document, error) {
	htmlContent = GetUnicodeHTML(htmlContent)

	// Remove XML declarations if present
	if strings.HasPrefix(htmlContent, "<?") {
		re := regexp.MustCompile(`<\?.*?\?>`)
		htmlContent = re.ReplaceAllString(htmlContent, "")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("FromString() returned an invalid string: %s...", htmlContent[:min(20, len(htmlContent))])
		return nil, err
	}

	return doc, nil
}

// NodeToString converts the tree under node to a string representation
func NodeToString(node *goquery.Selection) string {
	html, err := node.Html()
	if err != nil {
		return ""
	}
	return html
}

// GetTagsRegex gets list of elements of a certain tag with regex matching attributes
func GetTagsRegex(node *goquery.Selection, tag string, attribs map[string]string) []*goquery.Selection {
	if attribs == nil {
		return GetTags(node, tag, nil, "exact", false)
	}

	var results []*goquery.Selection

	for attr, pattern := range attribs {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		selector := fmt.Sprintf("%s[%s]", tag, attr)
		node.Find(selector).Each(func(i int, s *goquery.Selection) {
			attrVal, exists := s.Attr(attr)
			if exists && regex.MatchString(attrVal) {
				results = append(results, s)
			}
		})
	}

	return results
}

// GetTags gets list of elements of a certain tag with exact matching attributes
func GetTags(node *goquery.Selection, tag string, attribs map[string]string, attribsMatch string, ignoreDashes bool) []*goquery.Selection {
	if attribs == nil {
		selector := tag
		if tag == "" {
			selector = "*"
		}
		var results []*goquery.Selection
		node.Find(selector).Each(func(i int, s *goquery.Selection) {
			results = append(results, s)
		})
		return results
	}

	var results []*goquery.Selection

	for attr, value := range attribs {
		var selector string

		switch attribsMatch {
		case "exact":
			if ignoreDashes {
				selector = fmt.Sprintf("%s[%s='%s']", tag, attr, value)
			} else {
				selector = fmt.Sprintf("%s[%s='%s']", tag, attr, value)
			}
		case "substring":
			selector = fmt.Sprintf("%s[%s*='%s']", tag, attr, value)
		case "word":
			selector = fmt.Sprintf("%s[%s~='%s']", tag, attr, value)
		default:
			log.Printf("attribs_match must be one of 'exact', 'substring' or 'word'")
			return results
		}

		node.Find(selector).Each(func(i int, s *goquery.Selection) {
			results = append(results, s)
		})
	}

	return results
}

// GetElementsByAttribs gets list of elements with exact matching attributes
func GetElementsByAttribs(node *goquery.Selection, attribs map[string]string, attribsMatch string) []*goquery.Selection {
	return GetTags(node, "", attribs, attribsMatch, false)
}

// GetMetatags gets list of meta tags with name, property or itemprop equal to value
func GetMetatags(node *goquery.Selection, value string) []*goquery.Selection {
	if value == "" {
		var results []*goquery.Selection
		node.Find("meta").Each(func(i int, s *goquery.Selection) {
			results = append(results, s)
		})
		return results
	}

	var results []*goquery.Selection

	selectors := []string{
		fmt.Sprintf("meta[name='%s']", value),
		fmt.Sprintf("meta[property='%s']", value),
		fmt.Sprintf("meta[itemprop='%s']", value),
	}

	for _, selector := range selectors {
		node.Find(selector).Each(func(i int, s *goquery.Selection) {
			results = append(results, s)
		})
	}

	return results
}

// GetElementsByTagslist gets list of elements with tag in tagList
func GetElementsByTagslist(node *goquery.Selection, tagList []string) []*goquery.Selection {
	var results []*goquery.Selection

	for _, tag := range tagList {
		node.Find(tag).Each(func(i int, s *goquery.Selection) {
			results = append(results, s)
		})
	}

	return results
}

// CreateElement creates a new HTML element
func CreateElement(tag, text, tail string) *goquery.Selection {
	html := fmt.Sprintf("<%s>%s</%s>", tag, htmlpkg.EscapeString(text), tag)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	return doc.Find(tag).First()
}

// Remove removes the node(s) from the tree
func Remove(nodes []*goquery.Selection, keepTags []string) {
	for _, node := range nodes {
		if node == nil {
			continue
		}

		// Keep specified tags before removing
		for _, tag := range keepTags {
			node.Find(tag).Each(func(i int, s *goquery.Selection) {
				// Move kept elements to parent
				if parent := node.Parent(); parent.Length() > 0 {
					parent.AppendSelection(s)
				}
			})
		}

		node.Remove()
	}
}

// GetText extracts text from node, cleaning up unwanted elements
func GetText(node *goquery.Selection) string {
	// Clone the node to avoid modifying the original
	cloned := node.Clone()

	// Remove unwanted elements
	cloned.Find("script, style, select, option, textarea").Remove()

	// Get all text content
	text := cloned.Text()

	// Unescape HTML entities
	text = html.UnescapeString(text)

	// Remove any remaining HTML tags that may have been included as literal
	tagRe := regexp.MustCompile(`<[^>]+>`)
	text = tagRe.ReplaceAllString(text, " ")

	// Trim whitespace
	return InnerTrim(text)
}

// InnerTrim trims whitespace and collapses multiple spaces
func InnerTrim(text string) string {
	// Replace multiple whitespace with single space
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// GetAttribute gets the unicode attribute of the node
func GetAttribute(node *goquery.Selection, attr string, type_ any, default_ any) any {
	attrVal, exists := node.Attr(attr)
	if !exists {
		return default_
	}

	attrVal = htmlpkg.UnescapeString(attrVal)

	if type_ != nil {
		switch type_.(type) {
		case int:
			if val, err := strconv.Atoi(attrVal); err == nil {
				return val
			}
		case float64:
			if val, err := strconv.ParseFloat(attrVal, 64); err == nil {
				return val
			}
		case bool:
			if val, err := strconv.ParseBool(attrVal); err == nil {
				return val
			}
		}
		return default_
	}

	return attrVal
}

// SetAttribute sets an attribute on the node
func SetAttribute(node *goquery.Selection, attr string, value any) {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case int:
		strValue = strconv.Itoa(v)
	case float64:
		strValue = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		strValue = strconv.FormatBool(v)
	default:
		strValue = fmt.Sprintf("%v", v)
	}

	node.SetAttr(attr, strValue)
}

// OuterHTML gets the outer HTML of the node
func OuterHTML(node *goquery.Selection) string {
	if node.Length() == 0 {
		return ""
	}

	// Get the first node
	htmlNode := node.Get(0)
	if htmlNode == nil {
		return ""
	}

	// Render the node to HTML string
	var buf strings.Builder
	if err := htmlpkg.Render(&buf, htmlNode); err != nil {
		return ""
	}

	return buf.String()
} // GetLdJsonObject gets the JSON-LD object from the node
func GetLdJsonObject(node *goquery.Selection) []map[string]any {
	var results []map[string]any

	node.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		jsonStr := s.Text()
		var jsonData any
		if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
			return
		}

		switch v := jsonData.(type) {
		case []any:
			for _, item := range v {
				if obj, ok := item.(map[string]any); ok {
					results = append(results, obj)
				}
			}
		case map[string]any:
			results = append(results, v)
		}
	})

	return results
}

// GetNodeDepth gets the depth of the node (how deep its children are)
func GetNodeDepth(node *goquery.Selection) int {
	queue := list.New()
	queue.PushBack(node)
	maxDepth := 1

	for queue.Len() > 0 {
		element := queue.Remove(queue.Front()).(*goquery.Selection)
		currentDepth := 1

		element.Children().Each(func(i int, child *goquery.Selection) {
			queue.PushBack(child)
			currentDepth++
		})

		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}
	}

	return maxDepth
}

// GetLevel gets the level of the node in the tree
func GetLevel(node *goquery.Selection) int {
	level := 0
	current := node

	for current.Length() > 0 {
		current = current.Parent()
		level++
	}

	return level - 1 // Subtract 1 because we start from the node itself
}

// GetNodesAtLevel gets the nodes at a certain level in the tree
func GetNodesAtLevel(root *goquery.Selection, level int) []*goquery.Selection {
	var results []*goquery.Selection

	if level == 0 {
		results = append(results, root)
		return results
	}

	queue := list.New()
	queue.PushBack(&levelNode{node: root, level: 0})

	for queue.Len() > 0 {
		element := queue.Remove(queue.Front()).(*levelNode)

		if element.level == level {
			results = append(results, element.node)
		} else if element.level < level {
			element.node.Children().Each(func(i int, child *goquery.Selection) {
				queue.PushBack(&levelNode{node: child, level: element.level + 1})
			})
		}
	}

	return results
}

type levelNode struct {
	node  *goquery.Selection
	level int
}

// IsHighlinkDensity checks the density of links within a node
func IsHighlinkDensity(node *goquery.Selection, language string) bool {
	links := GetElementsByTagslist(node, []string{"a", "button"})
	if len(links) == 0 {
		return false
	}

	nodeText := GetText(node)
	wordCount := getWordCount(nodeText, language)

	if wordCount == 0 {
		return len(links) > 0
	}

	linkWordCounts := 0
	for _, link := range links {
		linkText := GetText(link)
		linkWordCounts += getWordCount(linkText, language)
	}

	proportion := float64(linkWordCounts*100) / float64(wordCount)

	// Function that starts from 65-70% for small values and drops to 40-35% for larger values
	limitFunction := func(x float64) float64 {
		return 87 - 70/(1.3+math.Exp(1-x/200))
	}

	if proportion > limitFunction(float64(wordCount)) {
		return true
	}

	if wordCount < 50 && len(links) > 2 && proportion > 50 {
		return true
	}

	return false
}

// GetNodeGravityScore gets the gravity score from the node
func GetNodeGravityScore(node *goquery.Selection) float64 {
	scoreStr, exists := node.Attr("gravityScore")
	if !exists {
		return 0.0
	}

	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return 0.0
	}

	return score
}

// Helper functions

func getWordCount(text string, language string) int {
	var words []string

	if language != "" {
		stopWords, err := nlp.NewStopWords(language)
		if err == nil && stopWords != nil {
			// Use tokenizer if available
			if stopWords.Tokenizer != nil {
				// This would need proper tokenization implementation
				words = strings.FieldsFunc(text, func(r rune) bool {
					return !unicode.IsLetter(r) && !unicode.IsNumber(r)
				})
			} else {
				words = strings.Fields(text)
			}
		} else {
			words = strings.Fields(text)
		}
	} else {
		words = strings.FieldsFunc(text, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
	}

	return len(words)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
