package languages

import (
	"bytes"
	"io"
	"math"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

const undetermined = "und"
const rescale = 0.5
const expOverflow = 7.09e+02

// Info is the language detection result (small API compatible with the inspiration)
type Info struct {
	lang        string
	probability float64
	langTag     language.Tag
}

// Tag returns the language.Tag of the detected language
func (info Info) Tag() language.Tag { return info.langTag }

// LanguageCode returns the ISO 639-1 code for the detected language
func (info Info) LanguageCode() string {
	if len(info.lang) < 4 {
		return info.lang
	}
	return info.lang[:2]
}

// Confidence returns a measure of reliability for the language classification in [0,1]
func (info Info) Confidence() float64 { return info.probability }

// LanguageName returns the English name of the detected language
func (info Info) LanguageName() string { return display.English.Tags().Name(info.langTag) }

// SelfName returns the name of the language in its own language
func (info Info) SelfName() string { return display.Self.Name(info.langTag) }

// FromReader detects the language from an io.Reader
func FromReader(r io.Reader) (Info, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return Info{}, err
	}
	return FromString(buf.String()), nil
}

// FromString detects the language from the given string.
// It uses Unicode-range heuristics for non-Latin scripts and small stopword
// matching for common Latin languages. The algorithm is lightweight and
// inspired by the getlang approach (counting evidence per language + softmax).
func FromString(text string) Info {
	matches := make(map[string]int)

	// Initialize with undetermined to avoid empty map
	matches[undetermined] = 1

	// 1) Use Unicode regexes defined in languages.go to count script evidence
	for code, pattern := range LanguagesUnicodeRegex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		hits := re.FindAllString(text, -1)
		if len(hits) > 0 {
			matches[code] += len(hits)
		}
	}

	// 2) Check a few common Latin-based languages using stopword heuristics
	latinStopwords := map[string][]string{
		"en": {" the ", " and ", " of ", " to ", " in ", " is ", " that "},
		"fr": {" le ", " la ", " et ", " les ", " des ", " un ", " une ", " que "},
		"de": {" der ", " die ", " das ", " und ", " ist ", " ein ", " eine "},
		"es": {" el ", " la ", " y ", " que ", " en ", " los ", " las ", " un ", " una "},
		"it": {" il ", " la ", " e ", " che ", " di ", " un "},
		"pt": {" o ", " a ", " e ", " que ", " de ", " do ", " da "},
		"nl": {" de ", " en ", " van ", " het ", " een "},
		"pl": {" i ", " w ", " z ", " że ", " się ", " nie "},
	}

	// lower-case text with surrounding spaces to simplify word boundary checks
	t := " " + strings.ToLower(text) + " "
	for code, words := range latinStopwords {
		var count int
		for _, w := range words {
			count += strings.Count(t, w)
		}
		if count > 0 {
			matches[code] += count * 3 // give stopwords a bit more weight
		}
	}

	// 3) Compute softmax over integer counts to get probabilities
	sm := softMax(matches)
	maxk := maxKey(matches)
	tag := language.MustParse(maxk)
	return Info{lang: maxk, probability: sm[maxk], langTag: tag}
}

// softMax converts evidence counts into probabilities similar to getlang
func softMax(mapping map[string]int) map[string]float64 {
	soft := make(map[string]float64)
	var denom float64
	overflowed := false
	for _, v := range mapping {
		denom += math.Exp(float64(v) * rescale)
		if v > expOverflow {
			overflowed = true
		}
	}
	for k := range mapping {
		if !overflowed {
			soft[k] = math.Exp(rescale*float64(mapping[k])) / denom
		} else {
			soft[k] = 1.0
		}
	}
	return soft
}

func maxKey(mapping map[string]int) string {
	var max int
	var key string
	// maintain deterministic order by sorting keys when values equal
	keys := make([]string, 0, len(mapping))
	for k := range mapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := mapping[k]
		if v > max || key == "" {
			max = v
			key = k
		}
	}
	return key
}
