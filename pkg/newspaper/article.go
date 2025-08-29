package newspaper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/cleaner"
	"github.com/tguidoux/newspaper4k-go/internal/nlp"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/constants"
	"golang.org/x/text/language"
)

// ArticleDownloadState represents the download state for the Article object.
type ArticleDownloadState int

const (
	NotStarted     ArticleDownloadState = 0
	FailedResponse ArticleDownloadState = 1
	Success        ArticleDownloadState = 2
)

// Article abstraction for
// This object fetches and holds information for a single article.
type Article struct {
	SourceURL            string               // URL to the main page of the news source
	URL                  string               // The article link (may differ from original URL)
	OriginalURL          string               // The original URL passed to the constructor
	Title                string               // Parsed title of the article
	ReadMoreLink         string               // XPath selector for the link to the full article
	TopImage             string               // Top image URL of the article
	MetaImg              string               // Image URL provided by metadata
	Images               []string             // List of all image URLs in the article
	Movies               []string             // List of video links in the article body
	Text                 string               // Parsed version of the article body
	TextCleaned          string               // Deprecated: same as Text
	Keywords             []string             // Inferred list of keywords for this article
	KeywordScores        map[string]float64   // Dictionary of keywords and their scores
	MetaKeywords         []string             // List of keywords provided by the meta data
	Tags                 map[string]string    // Extracted tag set from the article body
	Authors              []string             // Author list parsed from the article
	PublishDate          *time.Time           // Parsed publishing date from the article
	Summary              string               // Summarization of the article
	HTML                 string               // Raw HTML of the article page
	ArticleHTML          string               // Raw HTML of the article body
	IsParsed             bool                 // True if parse() has been called
	DownloadState        ArticleDownloadState // Download state
	DownloadExceptionMsg string               // Exception message if download() failed
	History              []string             // Redirection history from requests
	MetaDescription      string               // Description extracted from meta data
	MetaLang             string               // Language extracted from meta data
	MetaFavicon          string               // Website's favicon URL
	MetaSiteName         string               // Website's name
	MetaData             map[string]string    // Additional meta data from meta tags
	CanonicalLink        string               // Canonical URL for the article
	Categories           []*urls.URL          // Extracted category URLs from the source
	TopNode              *goquery.Selection   // Top node of the original DOM tree (HTML element)
	Doc                  *goquery.Document    // Full DOM of the downloaded HTML
	CleanDoc             *goquery.Document    // Cleaned version of the DOM tree
	Language             language.Tag         // Detected language of the article
	Config               *configuration.Configuration
}

// ParseRequest represents parameters for creating and parsing an Article.
type ParseRequest struct {
	URL           string
	Configuration *configuration.Configuration
	Extractors    []Extractor
	InputHTML     string
}

// Build builds a lone article from a URL. Calls Download(), Parse(), and NLP() in succession.
func (a *Article) Build(extractors []Extractor) {
	a.Download()
	a.Parse(extractors)
	a.NLP()
}

// Download downloads the link's HTML content.
func (a *Article) Download() *Article {

	inputHTML := a.Config.DownloadOptions.InputHTML

	if inputHTML == "" && a.HTML != "" {
		inputHTML = a.HTML
	}

	if inputHTML == "" {
		// Implement HTTP request logic
		resp, err := http.Get(a.URL)
		if err != nil {
			a.DownloadState = FailedResponse
			a.DownloadExceptionMsg = err.Error()
			return a
		}
		defer func() { _ = resp.Body.Close() }()

		// Read HTML
		htmlBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			a.DownloadState = FailedResponse
			a.DownloadExceptionMsg = err.Error()
			return a
		}
		htmlContent := string(htmlBytes)

		// Use goquery to parse
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			a.DownloadState = FailedResponse
			a.DownloadExceptionMsg = err.Error()
			return a
		}
		a.Doc = doc
		a.HTML = htmlContent
		a.DownloadState = Success
		a.HTML = inputHTML // Keep the original HTML
		a.DownloadState = Success
	} else {
		a.HTML = inputHTML
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(inputHTML))
		if err != nil {
			a.DownloadState = FailedResponse
			a.DownloadExceptionMsg = err.Error()
			return a
		}
		a.Doc = doc
		a.DownloadState = Success
	}

	return a
}

// Parse parses the previously downloaded article.
func (a *Article) Parse(extractors []Extractor) *Article {
	if err := a.ThrowIfNotDownloadedVerbose(); err != nil {
		// Handle error, perhaps log or return
		return a
	}

	// Run extractors
	for _, ext := range extractors {
		_ = ext.Parse(a)
	}

	// Clean the top node if it exists
	if a.TopNode != nil {
		documentCleaner := cleaner.NewDocumentCleaner()
		a.TopNode = documentCleaner.Clean(a.TopNode)
		// Update article HTML and text from cleaned node
		a.ArticleHTML = parsers.OuterHTML(a.TopNode)
		a.Text = parsers.GetText(a.TopNode)
	}

	a.IsParsed = true
	return a
}

// NLP performs keyword extraction and summarization.
func (a *Article) NLP() {
	if err := a.ThrowIfNotParsedVerbose(); err != nil {
		// Handle error
		return
	}

	// Get language for stop words
	language := a.GetLanguage().String()
	if language == "" {
		language = "en" // Default to English
	}

	// Create StopWords instance
	stopwords, err := nlp.NewStopWords(language)
	if err != nil {
		// Fallback to basic method if StopWords creation fails
		a.extractKeywordsBasic()
		a.generateSummaryBasic()
		return
	}

	// Extract keywords using NLP package
	a.extractKeywordsWithNLP(stopwords)

	// Generate summary using NLP package
	a.generateSummaryWithNLP(stopwords)
}

// extractKeywordsWithNLP extracts keywords using the NLP package
func (a *Article) extractKeywordsWithNLP(stopwords *nlp.StopWords) {
	text := a.Text
	if text == "" {
		return
	}

	// Use NLP package to extract keywords
	maxKeywords := a.Config.MaxKeywords
	if maxKeywords <= 0 {
		maxKeywords = 10
	}

	keywordScores := nlp.Keywords(text, stopwords, maxKeywords)

	// Filter keywords to remove special characters and ensure minimum length
	keywordScores = a.filterKeywords(keywordScores)

	// Convert to the format expected by Article
	a.KeywordScores = make(map[string]float64)
	a.Keywords = make([]string, 0, len(keywordScores))

	// Sort keywords by score
	type wordScore struct {
		word  string
		score float64
	}

	var wordScores []wordScore
	for word, score := range keywordScores {
		wordScores = append(wordScores, wordScore{word: word, score: score})
	}

	sort.Slice(wordScores, func(i, j int) bool {
		return wordScores[i].score > wordScores[j].score
	})

	for _, ws := range wordScores {
		a.Keywords = append(a.Keywords, ws.word)
		a.KeywordScores[ws.word] = ws.score
	}

	// Combine with title keywords
	a.combineTitleKeywords()
}

// combineTitleKeywords combines keywords from title and text
func (a *Article) combineTitleKeywords() {
	if a.Title == "" || len(a.KeywordScores) == 0 {
		return
	}

	// Extract keywords from title
	titleWords := strings.Fields(strings.ToLower(a.Title))
	titleKeywordSet := make(map[string]bool)
	stopWords := a.getStopWords()

	for _, word := range titleWords {
		word = strings.TrimSpace(word)
		cleaned := a.cleanKeyword(word)
		if cleaned != "" && !a.isStopWord(cleaned, stopWords) {
			titleKeywordSet[cleaned] = true
		}
	}

	// Boost scores for keywords that appear in title
	for keyword, score := range a.KeywordScores {
		if titleKeywordSet[strings.ToLower(keyword)] {
			a.KeywordScores[keyword] = score * 1.5 // Boost by 50%
		}
	}

	// Add title keywords that aren't already in the list
	for word := range titleKeywordSet {
		if _, exists := a.KeywordScores[word]; !exists {
			a.Keywords = append([]string{word}, a.Keywords...) // Add to front
			a.KeywordScores[word] = 0.1                        // Give a base score
		}
	}

	// Re-sort keywords by score
	type wordScore struct {
		word  string
		score float64
	}

	var wordScores []wordScore
	for word, score := range a.KeywordScores {
		wordScores = append(wordScores, wordScore{word: word, score: score})
	}

	sort.Slice(wordScores, func(i, j int) bool {
		return wordScores[i].score > wordScores[j].score
	})

	// Update keywords list
	maxKeywords := a.Config.MaxKeywords
	if maxKeywords <= 0 {
		maxKeywords = 10
	}

	a.Keywords = make([]string, 0, maxKeywords)
	for i, ws := range wordScores {
		if i >= maxKeywords {
			break
		}
		a.Keywords = append(a.Keywords, ws.word)
	}
}

// generateSummaryWithNLP generates summary using the NLP package
func (a *Article) generateSummaryWithNLP(stopwords *nlp.StopWords) {
	title := a.Title
	text := a.Text
	if text == "" {
		return
	}

	// Use NLP package to generate summary
	maxSentences := a.Config.MaxSummarySent
	if maxSentences <= 0 {
		maxSentences = 5
	}

	summarySentences := nlp.Summarize(title, text, stopwords, maxSentences)
	a.Summary = strings.Join(summarySentences, " ")
}

// extractKeywordsBasic is a fallback keyword extraction without gse
func (a *Article) extractKeywordsBasic() {
	text := a.Text
	if text == "" {
		return
	}

	// Simple word splitting
	words := strings.Fields(text)
	wordFreq := make(map[string]int)
	stopWords := a.getStopWords()

	for _, word := range words {
		word = strings.TrimSpace(strings.ToLower(word))
		cleaned := a.cleanKeyword(word)
		if cleaned != "" && len(cleaned) >= 3 && !a.isStopWord(cleaned, stopWords) {
			wordFreq[cleaned]++
		}
	}

	// Convert to slice and sort by frequency
	type wordScore struct {
		word  string
		score float64
	}

	var wordScores []wordScore
	totalWords := 0
	for _, freq := range wordFreq {
		totalWords += freq
	}

	for word, freq := range wordFreq {
		tf := float64(freq) / float64(totalWords)
		wordScores = append(wordScores, wordScore{word: word, score: tf})
	}

	sort.Slice(wordScores, func(i, j int) bool {
		return wordScores[i].score > wordScores[j].score
	})

	maxKeywords := a.Config.MaxKeywords
	if maxKeywords <= 0 {
		maxKeywords = 10
	}

	a.Keywords = make([]string, 0, maxKeywords)
	a.KeywordScores = make(map[string]float64)

	for i, ws := range wordScores {
		if i >= maxKeywords {
			break
		}
		a.Keywords = append(a.Keywords, ws.word)
		a.KeywordScores[ws.word] = ws.score
	}
}

// generateSummaryBasic generates a basic summary from the article text
func (a *Article) generateSummaryBasic() {
	text := a.Text
	if text == "" {
		return
	}

	sentences := strings.Split(text, ".")
	if len(sentences) == 0 {
		a.Summary = text
		return
	}

	// Score sentences based on keyword frequency
	type sentenceScore struct {
		sentence string
		score    float64
	}

	var sentenceScores []sentenceScore
	keywordSet := make(map[string]bool)
	for _, keyword := range a.Keywords {
		keywordSet[strings.ToLower(keyword)] = true
	}

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Calculate score based on keyword presence
		words := strings.Fields(strings.ToLower(sentence))
		score := 0.0
		for _, word := range words {
			if keywordSet[word] {
				score += 1.0
			}
		}

		// Boost score for sentence length (prefer substantial sentences)
		if len(words) > 5 && len(words) < 30 {
			score += 0.5
		}

		sentenceScores = append(sentenceScores, sentenceScore{
			sentence: sentence,
			score:    score,
		})
	}

	// Sort by score (descending)
	sort.Slice(sentenceScores, func(i, j int) bool {
		return sentenceScores[i].score > sentenceScores[j].score
	})

	// Take top sentences
	maxSentences := a.Config.MaxSummarySent
	if maxSentences <= 0 {
		maxSentences = 3
	}

	var summarySentences []string
	for i, ss := range sentenceScores {
		if i >= maxSentences {
			break
		}
		summarySentences = append(summarySentences, ss.sentence)
	}

	a.Summary = strings.Join(summarySentences, ". ")
	if a.Summary != "" && !strings.HasSuffix(a.Summary, ".") {
		a.Summary += "."
	}
}

// getStopWords returns a language-specific set of stop words
func (a *Article) getStopWords() map[string]bool {
	stopWords := make(map[string]bool)

	// Try to load stop words from the text package based on article's language
	language := a.MetaLang
	if language == "" {
		language = "en" // Default to English
	}

	stopWordsSlice := nlp.GetStopWordsForLanguage(language)

	for _, word := range stopWordsSlice {
		if word != "" {
			stopWords[strings.ToLower(word)] = true
		}
	}

	// Fallback to hardcoded English stop words if no stopwords found
	if len(stopWords) == 0 {
		englishStopWords := []string{
			"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of",
			"with", "by", "is", "are", "was", "were", "be", "been", "being", "have", "has", "had",
			"do", "does", "did", "will", "would", "could", "should", "may", "might", "must", "can", "shall",
			"this", "that", "these", "those", "i", "you", "he", "she", "it", "we", "they", "me",
			"him", "her", "us", "them", "my", "your", "his", "its", "our", "their",
		}
		for _, word := range englishStopWords {
			stopWords[word] = true
		}
	}

	return stopWords
}

// isStopWord checks if a word is a stop word
func (a *Article) isStopWord(word string, stopWords map[string]bool) bool {
	return stopWords[strings.ToLower(word)]
}

// GetTitle returns the title of the article.
func (a *Article) GetTitle() string {
	return a.Title
}

// SetTitle sets the title of the article.
func (a *Article) SetTitle(value string) {
	if value != "" {
		if len(value) > a.Config.MaxTitle {
			a.Title = value[:a.Config.MaxTitle]
		} else {
			a.Title = value
		}
	} else {
		a.Title = ""
	}
}

// GetText returns the text content of the article.
func (a *Article) GetText() string {
	return a.Text
}

// SetText sets the text of the article.
func (a *Article) SetText(value string) {
	if value != "" {
		if len(value) > a.Config.MaxText {
			a.Text = value[:a.Config.MaxText]
		} else {
			a.Text = value
		}
	} else {
		a.Text = ""
	}
}

// GetHTML returns the HTML content of the article.
func (a *Article) GetHTML() string {
	return a.HTML
}

// SetHTML sets the HTML content of the article.
func (a *Article) SetHTML(value string) {
	a.DownloadState = Success
	if value != "" {
		a.HTML = value
	} else {
		a.HTML = ""
	}
}

// GetSummary returns the summary of the article.
func (a *Article) GetSummary() string {
	return a.Summary
}

// SetSummary sets the summary of the article.
func (a *Article) SetSummary(value string) {
	if value != "" {
		if len(value) > a.Config.MaxSummary {
			a.Summary = value[:a.Config.MaxSummary]
		} else {
			a.Summary = value
		}
	} else {
		a.Summary = ""
	}
}

// ThrowIfNotDownloadedVerbose checks if the article has been downloaded.
func (a *Article) ThrowIfNotDownloadedVerbose() error {
	switch a.DownloadState {
	case NotStarted:
		return fmt.Errorf("you must download() an article first")
	case FailedResponse:
		return fmt.Errorf("article download() failed with %s on URL %s", a.DownloadExceptionMsg, a.URL)
	}
	return nil
}

// ThrowIfNotParsedVerbose checks if the article has been parsed.
func (a *Article) ThrowIfNotParsedVerbose() error {
	if !a.IsParsed {
		return fmt.Errorf("you must parse() an article first")
	}
	return nil
}

// IsValidURL checks if the URL is valid.
func (a *Article) IsValidURL() bool {
	// Implement URL validation
	return true // TODO: Add actual validation logic from urls.IsValidNewsArticleURL(...)
}

// IsValidBody checks if the article body is valid.
func (a *Article) IsValidBody() bool {
	if !a.IsParsed {
		return false
	}
	wordCount := len(strings.Fields(a.Text))
	return wordCount >= a.Config.MinWordCount
}

// GetTopKeywords returns the top keywords with their scores
func (a *Article) GetTopKeywords() map[string]float64 {
	return a.KeywordScores
}

// GetTopKeywordsList returns the top keywords as a list
func (a *Article) GetTopKeywordsList() []string {
	return a.Keywords
}

func (a *Article) GetLanguage() language.Tag {
	return a.Language
}

func (a *Article) SetLanguage(lang language.Tag) {
	a.Language = lang
}

// GetCleanDoc returns the cleaned version of the document
func (a *Article) GetCleanDoc() *goquery.Document {
	if a.CleanDoc == nil && a.Doc != nil {
		documentCleaner := cleaner.NewDocumentCleaner()
		// Clone the document for cleaning
		docHTML := parsers.OuterHTML(a.Doc.Find("html").First())
		var err error
		a.CleanDoc, err = goquery.NewDocumentFromReader(strings.NewReader(docHTML))
		if err == nil {
			// Convert document to selection for cleaning
			rootSelection := a.CleanDoc.Find("html")
			if rootSelection.Length() == 0 {
				rootSelection = a.CleanDoc.Find("body")
			}
			if rootSelection.Length() == 0 {
				rootSelection = a.CleanDoc.Selection
			}
			cleanSelection := documentCleaner.Clean(rootSelection)
			// Create a new document from the cleaned selection
			cleanHTML := parsers.OuterHTML(cleanSelection)
			a.CleanDoc, _ = goquery.NewDocumentFromReader(strings.NewReader(cleanHTML))
		}
	}
	return a.CleanDoc
}

// ToJSON creates a JSON string from the article data
func (a *Article) ToJSON() (string, error) {
	if err := a.ThrowIfNotParsedVerbose(); err != nil {
		return "", err
	}

	// Create a map with the most important article data
	articleData := map[string]interface{}{
		"title":            a.Title,
		"text":             a.Text,
		"authors":          a.Authors,
		"publish_date":     a.PublishDate,
		"summary":          a.Summary,
		"keywords":         a.Keywords,
		"keyword_scores":   a.KeywordScores,
		"source_url":       a.SourceURL,
		"url":              a.URL,
		"top_image":        a.TopImage,
		"images":           a.Images,
		"movies":           a.Movies,
		"meta_description": a.MetaDescription,
		"meta_lang":        a.MetaLang,
		"is_parsed":        a.IsParsed,
	}

	// Convert to JSON (simple implementation)
	jsonStr := "{"
	first := true
	for key, value := range articleData {
		if !first {
			jsonStr += ","
		}
		jsonStr += fmt.Sprintf("\"%s\":", key)
		switch v := value.(type) {
		case string:
			jsonStr += fmt.Sprintf("\"%s\"", v)
		case []string:
			jsonStr += "["
			for i, item := range v {
				if i > 0 {
					jsonStr += ","
				}
				jsonStr += fmt.Sprintf("\"%s\"", item)
			}
			jsonStr += "]"
		case map[string]float64:
			jsonStr += "{"
			firstInner := true
			for k, val := range v {
				if !firstInner {
					jsonStr += ","
				}
				jsonStr += fmt.Sprintf("\"%s\":%f", k, val)
				firstInner = false
			}
			jsonStr += "}"
		case *time.Time:
			if v != nil {
				jsonStr += fmt.Sprintf("\"%s\"", v.Format(time.RFC3339))
			} else {
				jsonStr += "null"
			}
		case bool:
			if v {
				jsonStr += "true"
			} else {
				jsonStr += "false"
			}
		default:
			jsonStr += "null"
		}
		first = false
	}
	jsonStr += "}"

	return jsonStr, nil
}

// cleanKeyword filters keywords to ensure they are simple words with no special characters and minimum 3 characters
func (a *Article) cleanKeyword(keyword string) string {
	// Remove special characters and keep only letters
	cleaned := ""
	for _, r := range keyword {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			cleaned += string(r)
		}
	}

	// Convert to lowercase
	cleaned = strings.ToLower(cleaned)

	// Check minimum length
	if len(cleaned) < 3 {
		return ""
	}

	return cleaned
}

// filterKeywords applies cleaning to a map of keyword scores
func (a *Article) filterKeywords(keywordScores map[string]float64) map[string]float64 {
	filtered := make(map[string]float64)
	stopWords := a.getStopWords()

	for keyword, score := range keywordScores {
		cleaned := a.cleanKeyword(keyword)
		if cleaned != "" && !a.isStopWord(cleaned, stopWords) {
			// If multiple keywords map to the same cleaned version, keep the highest score
			if existingScore, exists := filtered[cleaned]; !exists || score > existingScore {
				filtered[cleaned] = score
			}
		}
	}

	return filtered
}

func (a *Article) String() string {
	if err := a.ThrowIfNotParsedVerbose(); err != nil {
		return fmt.Sprintf("Article not parsed: %v", err)
	}

	return fmt.Sprintf("Article Title: %s\nURL: %s\nAuthors: %v\nPublish Date: %v\nTop Image: %s\nMeta Description: %s\nKeywords: %v\nSummary: %s\nText: %s",
		a.Title,
		a.URL,
		a.Authors,
		a.PublishDate,
		a.TopImage,
		a.MetaDescription,
		a.Keywords,
		a.Summary,
		a.Text)
}

// IsLikelyArticleURL checks if a URL is likely to be an article rather than a navigation link
func IsLikelyArticleURL(urlStr string) bool {
	// Skip obvious navigation/category

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// is contains any stopwords from URL_STOPWORDS
	for _, stopword := range constants.COMMON_NOT_ARTICLE_URL_STOPWORDS {
		if strings.Contains(parsedURL.Path, stopword) || strings.HasSuffix(parsedURL.Path, "/"+stopword) {
			return false
		}
	}

	// For Hacker News, articles have /item?id= pattern
	if strings.Contains(urlStr, "/item?id=") {
		return true
	}

	// For other sites, check for common article patterns, URL_GOODWORDS
	for _, goodword := range constants.COMMON_ARTICLE_URL_GOODWORDS {
		if strings.Contains(parsedURL.Path, goodword) || strings.HasSuffix(parsedURL.Path, "/"+goodword) {
			return true
		}
	}

	// Check if URL has a date-like pattern (YYYY/MM/DD or similar)
	datePattern := regexp.MustCompile(`/(\d{4})/(\d{1,2})/(\d{1,2})/`)
	if datePattern.MatchString(urlStr) {
		return true
	}

	// If it has query parameters, it might be an article
	if parsedURL.RawQuery != "" {
		return true
	}

	// Default: if it's not obviously a category/navigation URL, consider it an article
	return false
}
