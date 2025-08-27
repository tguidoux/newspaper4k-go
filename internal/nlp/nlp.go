package nlp

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/go-ego/gse"
	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	"github.com/tguidoux/newspaper4k-go/internal/resources/text"
)

// StopWords represents a collection of stop words and a tokenizer for a language
type StopWords struct {
	StopWords map[string]bool
	Tokenizer *tokenizer.Tokenizer
	Gse       *gse.Segmenter
}

// NewStopWords creates a new StopWords instance for the given language
func NewStopWords(language string) (*StopWords, error) {
	stopWords := make(map[string]bool)

	// Load stop words from the text package based on language
	stopWordsSlice := GetStopWordsForLanguage(language)

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

	// Initialize tokenizer - using BERT tokenizer as example
	tk := pretrained.BertBaseUncased()

	// Initialize GSE for Chinese text
	var gseSeg *gse.Segmenter
	if language == "zh" {
		gseSeg = new(gse.Segmenter)
		_ = gseSeg.LoadDict()
	}

	return &StopWords{
		StopWords: stopWords,
		Tokenizer: tk,
		Gse:       gseSeg,
	}, nil
}

// GetStopWordsForLanguage returns the appropriate stopwords slice for the given language
func GetStopWordsForLanguage(language string) []string {
	switch strings.ToLower(language) {
	case "af":
		return text.StopwordsAF
	case "ar":
		return text.StopwordsAR
	case "as":
		return text.StopwordsAS
	case "be":
		return text.StopwordsBE
	case "bg":
		return text.StopwordsBG
	case "bn":
		return text.StopwordsBN
	case "br":
		return text.StopwordsBR
	case "ca":
		return text.StopwordsCA
	case "cs":
		return text.StopwordsCS
	case "cy":
		return text.StopwordsCY
	case "da":
		return text.StopwordsDA
	case "de":
		return text.StopwordsDE
	case "el":
		return text.StopwordsEL
	case "en":
		return text.StopwordsEN
	case "eo":
		return text.StopwordsEO
	case "es":
		return text.StopwordsES
	case "et":
		return text.StopwordsET
	case "eu":
		return text.StopwordsEU
	case "fa":
		return text.StopwordsFA
	case "fi":
		return text.StopwordsFI
	case "fr":
		return text.StopwordsFR
	case "ga":
		return text.StopwordsGA
	case "gl":
		return text.StopwordsGL
	case "gu":
		return text.StopwordsGU
	case "ha":
		return text.StopwordsHA
	case "he":
		return text.StopwordsHE
	case "hi":
		return text.StopwordsHI
	case "hr":
		return text.StopwordsHR
	case "hu":
		return text.StopwordsHU
	case "hy":
		return text.StopwordsHY
	case "id":
		return text.StopwordsID
	case "is":
		return text.StopwordsIS
	case "it":
		return text.StopwordsIT
	case "ja":
		return text.StopwordsJA
	case "ka":
		return text.StopwordsKA
	case "kn":
		return text.StopwordsKN
	case "ko":
		return text.StopwordsKO
	case "ku":
		return text.StopwordsKU
	case "lb":
		return text.StopwordsLB
	case "lt":
		return text.StopwordsLT
	case "lv":
		return text.StopwordsLV
	case "mk":
		return text.StopwordsMK
	case "mn":
		return text.StopwordsMN
	case "mr":
		return text.StopwordsMR
	case "ms":
		return text.StopwordsMS
	case "my":
		return text.StopwordsMY
	case "nb":
		return text.StopwordsNB
	case "ne":
		return text.StopwordsNE
	case "nl":
		return text.StopwordsNL
	case "no":
		return text.StopwordsNO
	case "pa":
		return text.StopwordsPA
	case "pl":
		return text.StopwordsPL
	case "ps":
		return text.StopwordsPS
	case "pt":
		return text.StopwordsPT
	case "rn":
		return text.StopwordsRN
	case "ro":
		return text.StopwordsRO
	case "ru":
		return text.StopwordsRU
	case "rw":
		return text.StopwordsRW
	case "si":
		return text.StopwordsSI
	case "sk":
		return text.StopwordsSK
	case "sl":
		return text.StopwordsSL
	case "so":
		return text.StopwordsSO
	case "sr":
		return text.StopwordsSR
	case "st":
		return text.StopwordsST
	case "sv":
		return text.StopwordsSV
	case "sw":
		return text.StopwordsSW
	case "ta":
		return text.StopwordsTA
	case "te":
		return text.StopwordsTE
	case "th":
		return text.StopwordsTH
	case "tl":
		return text.StopwordsTL
	case "tr":
		return text.StopwordsTR
	case "tt":
		return text.StopwordsTT
	case "uk":
		return text.StopwordsUK
	case "ur":
		return text.StopwordsUR
	case "uz":
		return text.StopwordsUZ
	case "vi":
		return text.StopwordsVI
	case "yo":
		return text.StopwordsYO
	case "zh":
		return text.StopwordsZH
	case "zu":
		return text.StopwordsZU
	default:
		return []string{} // Return empty slice for unknown languages
	}
}

// Tokenize tokenizes the given text
func (sw *StopWords) Tokenize(text string) []string {
	if sw.Gse != nil {
		// Use GSE for Chinese
		return sw.Gse.Cut(text, true)
	}

	// Use sugarme/tokenizer for other languages
	encoded, err := sw.Tokenizer.EncodeSingle(text, true)
	if err != nil {
		return strings.Fields(text) // fallback to simple split
	}

	return encoded.GetTokens()
}

// Keywords gets the top keywords and their frequency scores
func Keywords(text string, stopwords *StopWords, maxKeywords int) map[string]float64 {
	tokenisedText := stopwords.Tokenize(text)
	if len(text) == 0 {
		return map[string]float64{}
	}
	// Number of words before removing blacklist words
	numWords := len(tokenisedText)
	if numWords == 0 {
		numWords = 1
	}

	// Filter out stop words
	filteredTokens := []string{}
	for _, token := range tokenisedText {
		if !stopwords.StopWords[strings.ToLower(token)] {
			filteredTokens = append(filteredTokens, token)
		}
	}

	freq := make(map[string]int)
	for _, token := range filteredTokens {
		freq[token]++
	}

	// Get most common
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range freq {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	keywordsDict := make(map[string]float64)
	limit := len(sorted)
	if maxKeywords > 0 && maxKeywords < limit {
		limit = maxKeywords
	}
	for i := 0; i < limit; i++ {
		k := sorted[i].Key
		v := sorted[i].Value
		keywordsDict[k] = float64(v)*1.5/float64(numWords) + 1
	}

	return keywordsDict
}

// Constants for NLP settings
const (
	MeanSentenceLen       = 20.0
	SummarizeKeywordCount = 10
)

// SplitSentences splits a large string into sentences
func SplitSentences(text string) []string {
	// Simple sentence splitter using regex
	re := regexp.MustCompile(`[.!?]+`)
	sentences := re.Split(text, -1)
	var cleaned []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 10 {
			cleaned = append(cleaned, s)
		}
	}
	return cleaned
}

// Summarize summarizes an article into the most relevant sentences
func Summarize(title, text string, stopwords *StopWords, maxSents int) []string {
	if len(text) == 0 || len(title) == 0 || maxSents <= 0 {
		return []string{}
	}

	summaries := []string{}
	sentences := SplitSentences(text)
	keys := Keywords(text, stopwords, SummarizeKeywordCount)
	titleWords := stopwords.Tokenize(title)

	// Score sentences
	ranks := ScoredSentences(sentences, titleWords, keys, stopwords)

	// Filter out the first maxSents relevant sentences
	if len(ranks) > maxSents {
		ranks = ranks[:maxSents]
	}
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Index < ranks[j].Index // Sort by sentence order in the text
	})
	for _, rank := range ranks {
		summaries = append(summaries, rank.Sentence)
	}
	return summaries
}

// SentenceRank represents a sentence with its score and index
type SentenceRank struct {
	Index    int
	Sentence string
	Score    float64
}

// TitleScore calculates the title score for a sentence
func TitleScore(titleTokens, sentenceTokens []string, stopwords *StopWords) float64 {
	filteredTitle := []string{}
	for _, token := range titleTokens {
		if !stopwords.StopWords[strings.ToLower(token)] {
			filteredTitle = append(filteredTitle, token)
		}
	}
	if len(filteredTitle) == 0 {
		return 0.0
	}

	intersection := 0
	for _, st := range sentenceTokens {
		for _, tt := range filteredTitle {
			if strings.EqualFold(st, tt) && !stopwords.StopWords[strings.ToLower(st)] {
				intersection++
				break
			}
		}
	}
	return float64(intersection) / float64(len(filteredTitle))
}

// ScoredSentences scores sentences based on different features
func ScoredSentences(sentences, titleWords []string, keywords map[string]float64, stopwords *StopWords) []SentenceRank {
	sentenceCount := len(sentences)
	ranks := make([]SentenceRank, 0, sentenceCount)

	for i, s := range sentences {
		sentenceTokens := stopwords.Tokenize(s)
		titleFeatures := TitleScore(titleWords, sentenceTokens, stopwords)
		sentLen := LengthScore(len(sentenceTokens))
		sentPos := SentencePositionScore(i+1, sentenceCount)
		sbsFeature := SBS(sentenceTokens, keywords)
		dbsFeature := DBS(sentenceTokens, keywords)
		frequency := (sbsFeature + dbsFeature) / 2.0 * 10.0
		// Weighted average of scores from four categories
		totalScore := (titleFeatures*1.5 + frequency*2.0 + sentLen*1.0 + sentPos*1.0) / 4.0
		ranks = append(ranks, SentenceRank{
			Index:    i,
			Sentence: s,
			Score:    totalScore,
		})
	}

	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Score > ranks[j].Score
	})
	return ranks
}

// LengthScore calculates the length score
func LengthScore(sentenceLen int) float64 {
	return 1 - math.Abs(MeanSentenceLen-float64(sentenceLen))/MeanSentenceLen
}

// SentencePositionScore calculates the sentence position score
func SentencePositionScore(i, size int) float64 {
	normalized := float64(i) / float64(size)

	ranges := []struct {
		threshold float64
		value     float64
	}{
		{1.0, 0},
		{0.9, 0.15},
		{0.8, 0.04},
		{0.7, 0.04},
		{0.6, 0.06},
		{0.5, 0.04},
		{0.4, 0.05},
		{0.3, 0.08},
		{0.2, 0.14},
		{0.1, 0.23},
		{0, 0.17},
	}

	for _, r := range ranges {
		if normalized > r.threshold {
			return r.value
		}
	}
	return 0
}

// SBS calculates Sentence-Based Score
func SBS(words []string, keywords map[string]float64) float64 {
	if len(words) == 0 || len(keywords) == 0 {
		return 0.0
	}

	score := 0.0
	for _, word := range words {
		if val, exists := keywords[word]; exists {
			score += val
		}
	}
	score /= float64(len(words))
	score /= 10.0
	return score
}

// DBS calculates Document-Based Score
func DBS(words []string, keywords map[string]float64) float64 {
	if len(words) == 0 || len(keywords) == 0 {
		return 0.0
	}

	summ := 0.0
	wordsInKeys := []struct {
		index int
		score float64
		word  string
	}{}
	for i, word := range words {
		if score, exists := keywords[word]; exists {
			wordsInKeys = append(wordsInKeys, struct {
				index int
				score float64
				word  string
			}{i, score, word})
		}
	}
	if len(wordsInKeys) == 0 {
		return 0
	}

	for j := 0; j < len(wordsInKeys)-1; j++ {
		first := wordsInKeys[j]
		second := wordsInKeys[j+1]
		dif := second.index - first.index
		summ += first.score * second.score / float64(dif*dif)
	}

	intersection := make(map[string]bool)
	for _, wik := range wordsInKeys {
		intersection[wik.word] = true
	}
	k := len(intersection) + 1
	return 1 / (float64(k) * (float64(k) + 1.0)) * summ
}
