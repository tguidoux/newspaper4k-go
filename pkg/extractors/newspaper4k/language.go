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
	// If configuration forces a language, use it
	if le.config != nil && le.config.Language() != "" {
		a.MetaLang = le.config.Language()
		a.Language = languages.GetTagFromISO639_1(a.MetaLang)
		return nil
	}

	// If meta language already present, leave it as-is
	if a.MetaLang != "" {
		a.Language = languages.GetTagFromISO639_1(a.MetaLang)
		return nil
	}

	// Build text to detect language from: prefer article text + title, fall back to whole HTML
	var text string

	if a.Text != "" || a.Title != "" {
		text = strings.TrimSpace(a.Title + " " + a.Text)
	} else {
		// try to parse the HTML to extract textual content
		var doc, err = parsers.FromString(a.HTML)
		if err == nil && doc != nil {
			text = strings.TrimSpace(parsers.GetText(doc.Selection))
		} else if err != nil {
			// If parsing failed, as a last resort use raw HTML
			text = strings.TrimSpace(a.HTML)
			_ = err
		}
	}

	if text == "" {
		return nil
	}

	info := languages.FromString(text)
	lang := info.LanguageCode()
	if lang == "" || lang == "und" {
		return nil
	}

	a.MetaLang = lang
	a.Language = languages.GetTagFromISO639_1(lang)
	return nil
}
