package newspaper4k

import (
	"strings"

	"github.com/tguidoux/newspaper4k-go/internal/languages"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// LanguageExtractor detects the article language and sets Article.MetaLang
// It respects an explicitly configured language (Configuration.Language) and
// will not override an existing Article.MetaLang.
type LanguageExtractor struct {
	config *configuration.Configuration
}

// NewLanguageExtractor creates a new LanguageExtractor
func NewLanguageExtractor(config *configuration.Configuration) *LanguageExtractor {
	return &LanguageExtractor{config: config}
}

// Parse detects the language and updates the article's MetaLang when appropriate
func (le *LanguageExtractor) Parse(a *newspaper.Article) error {
	// 1) honor configuration override
	if le.config != nil && le.config.Language() != "" {
		a.MetaLang = le.config.Language()
		a.Language = languages.GetTagFromISO639_1(a.MetaLang)
		return nil
	}

	// 2) if meta language already present, populate language tag and exit
	if a.MetaLang != "" {
		a.Language = languages.GetTagFromISO639_1(a.MetaLang)
		return nil
	}

	// 3) build a short piece of text to detect language from
	text := buildDetectionText(a)
	if text == "" {
		return nil
	}

	// 4) detect language and set fields when detection yields a usable code
	info := languages.FromString(text)
	lang := info.LanguageCode()
	if lang == "" || lang == "und" {
		return nil
	}

	a.MetaLang = lang
	a.Language = languages.GetTagFromISO639_1(lang)
	return nil
}

// buildDetectionText chooses the best available text to run language detection on.
// Priority: Title + Text (if available) -> parsed visible text from HTML -> raw HTML fallback.
func buildDetectionText(a *newspaper.Article) string {
	if a == nil {
		return ""
	}

	if a.Title != "" || a.Text != "" {
		return strings.TrimSpace(a.Title + " " + a.Text)
	}

	// Try to parse HTML and extract visible text
	if a.HTML != "" {
		if doc, err := parsers.FromString(a.HTML); err == nil && doc != nil {
			if txt := strings.TrimSpace(parsers.GetText(doc.Selection)); txt != "" {
				return txt
			}
		}
		// last resort: use raw HTML
		return strings.TrimSpace(a.HTML)
	}

	return ""
}
