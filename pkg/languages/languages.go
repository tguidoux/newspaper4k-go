package languages

import (
	"strings"

	"github.com/tguidoux/newspaper4k-go/internal/languages"
)

// IsValidLanguageCode checks if the given language code is valid
func IsValidLanguageCode(code string) bool {
	_, exists := languages.LanguagesDict[code]
	return exists
}

// GetLanguageCodeFromName attempts to find a language code from a language name (case-insensitive)
func GetLanguageCodeFromName(name string) string {
	name = strings.ToLower(name)
	for code, langName := range languages.LanguagesDict {
		if strings.ToLower(langName) == name {
			return code
		}
	}
	return ""
}

// GetAvailableLanguages returns a list of available language codes
// Note: This is a simplified version since we don't have the stopwords directory structure
func GetAvailableLanguages() []string {
	var codes []string
	for code := range languages.LanguagesDict {
		codes = append(codes, code)
	}
	return codes
}

// GetLanguageNameFromCode returns the language name from the ISO 639-1 code
func GetLanguageNameFromCode(code string) string {
	if lang, exists := languages.LanguagesDict[code]; exists {
		return lang
	}
	return ""
}
