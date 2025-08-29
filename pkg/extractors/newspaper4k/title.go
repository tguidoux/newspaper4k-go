package newspaper4k

import (
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/languages"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
	"golang.org/x/text/language"
)

// TitleExtractor extracts the title from an article
type TitleExtractor struct {
	config *configuration.Configuration
	title  string
}

// NewTitleExtractor creates a new TitleExtractor
func NewTitleExtractor(config *configuration.Configuration) *TitleExtractor {
	return &TitleExtractor{
		config: config,
		title:  "",
	}
}

// Parse extracts the article title and updates the article in-place
func (te *TitleExtractor) Parse(a *newspaper.Article) error {
	te.title = ""

	if a.Doc == nil {
		doc, err := parsers.FromString(a.HTML)
		if err != nil {
			return err
		}
		a.Doc = doc
	}

	titleElement := a.Doc.Find("title").First()
	if titleElement.Length() == 0 {
		return nil
	}

	titleText := strings.TrimSpace(titleElement.Text())
	usedDelimiter := false

	// title from h1
	titleTextH1 := te.getTitleFromH1(a.Doc)

	// title from og:title and similar meta tags
	titleTextFB := te.getTitleFromMeta(a.Doc)

	// create filtered versions for comparison
	langTag := a.GetLanguage()

	if langTag == language.Und {
		// if language is undetermined, try to get it from config or article meta
		if te.config != nil && te.config.Language() != "" {
			langTag = languages.GetTagFromISO639_1(te.config.Language())
		} else if a.MetaLang != "" {
			langTag = languages.GetTagFromISO639_1(a.MetaLang)
		}

		// assume English for title comparison
		if langTag == language.Und {
			langTag = language.English
		}

	}

	regexChars := languages.LanguageRegex(langTag)
	filterRegex := regexp.MustCompile("[^" + regexChars + "]")
	filterTitleText := filterRegex.ReplaceAllString(strings.ToLower(titleText), "")
	filterTitleTextH1 := filterRegex.ReplaceAllString(strings.ToLower(titleTextH1), "")
	filterTitleTextFB := filterRegex.ReplaceAllString(strings.ToLower(titleTextFB), "")

	// check for better alternatives
	if titleTextH1 == titleText {
		usedDelimiter = true
	} else if filterTitleTextH1 != "" && filterTitleTextH1 == filterTitleTextFB {
		titleText = titleTextH1
		usedDelimiter = true
	} else if filterTitleTextH1 != "" && strings.Contains(filterTitleText, filterTitleTextH1) &&
		filterTitleTextFB != "" && strings.Contains(filterTitleText, filterTitleTextFB) &&
		len(titleTextH1) > len(titleTextFB) {
		titleText = titleTextH1
		usedDelimiter = true
	} else if filterTitleTextFB != "" && filterTitleTextFB != filterTitleText &&
		strings.HasPrefix(filterTitleText, filterTitleTextFB) {
		titleText = titleTextFB
		usedDelimiter = true
	}

	if !usedDelimiter {
		delimiters := []string{"|", "-", "_", "/", " » "}
		for _, delimiter := range delimiters {
			if strings.Contains(titleText, delimiter) {
				titleText = te.splitTitle(titleText, delimiter, titleTextH1)
				break
			}
		}
	}

	title := strings.ReplaceAll(titleText, constants.MOTLEY_REPLACEMENT[0], constants.MOTLEY_REPLACEMENT[1])

	// prefer h1 if very similar
	filterTitle := filterRegex.ReplaceAllString(strings.ToLower(title), "")
	if filterTitleTextH1 == filterTitle {
		title = titleTextH1
	}

	te.title = strings.TrimSpace(title)
	a.Title = te.title

	return nil
}

// getTitleFromH1 extracts the longest title from h1 elements
func (te *TitleExtractor) getTitleFromH1(doc *goquery.Document) string {
	var titles []string
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			titles = append(titles, text)
		}
	})

	if len(titles) == 0 {
		return ""
	}

	// sort by length descending
	sort.Slice(titles, func(i, j int) bool {
		return len(titles[i]) > len(titles[j])
	})

	longest := titles[0]

	// discard if too short (fewer than 2 words)
	words := strings.Fields(longest)
	if len(words) <= 2 {
		return ""
	}

	// clean double spaces
	return strings.Join(words, " ")
}

// getTitleFromMeta extracts title from meta tags
func (te *TitleExtractor) getTitleFromMeta(doc *goquery.Document) string {
	for _, metaName := range constants.TITLE_META_INFO {
		// Use parser's GetMetatags method
		metaElements := parsers.GetMetatags(doc.Selection, metaName)
		for _, metaElement := range metaElements {
			content := parsers.GetAttribute(metaElement, "content", nil, "")
			if contentStr, ok := content.(string); ok && contentStr != "" {
				return strings.TrimSpace(contentStr)
			}
		}
	}
	return ""
}

// splitTitle splits the title using the best delimiter
func (te *TitleExtractor) splitTitle(title, delimiter, hint string) string {
	pieces := strings.Split(title, delimiter)

	var filterRegex *regexp.Regexp
	if hint != "" {
		filterRegex = regexp.MustCompile(`[^a-zA-Z0-9\ ]`)
		hint = filterRegex.ReplaceAllString(strings.ToLower(hint), "")
	}

	largestIndex := 0
	largestLength := 0

	for i, piece := range pieces {
		current := strings.TrimSpace(piece)

		if hint != "" && strings.Contains(filterRegex.ReplaceAllString(strings.ToLower(current), ""), hint) {
			largestIndex = i
			break
		}

		if len(current) > largestLength {
			largestLength = len(current)
			largestIndex = i
		}
	}

	result := pieces[largestIndex]
	return strings.ReplaceAll(result, constants.TITLE_REPLACEMENTS[0], constants.TITLE_REPLACEMENTS[1])
}
