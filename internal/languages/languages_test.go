package languages

import (
	"testing"

	"golang.org/x/text/language"
)

func TestGetTagFromISO639_1(t *testing.T) {
	tests := []struct {
		name     string
		isoCode  string
		expected language.Tag
	}{
		{
			name:     "valid ISO code 'en'",
			isoCode:  "en",
			expected: language.English,
		},
		{
			name:     "valid ISO code 'de'",
			isoCode:  "de",
			expected: language.German,
		},
		{
			name:     "invalid ISO code",
			isoCode:  "xyz",
			expected: language.Und,
		},
		{
			name:     "empty string",
			isoCode:  "",
			expected: language.Und,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTagFromISO639_1(tt.isoCode)
			if result != tt.expected {
				t.Errorf("GetTagFromISO639_1(%s) = %v, want %v", tt.isoCode, result, tt.expected)
			}
		})
	}
}

func TestValidLanguages(t *testing.T) {
	result := ValidLanguages()
	if len(result) == 0 {
		t.Error("ValidLanguages() returned empty slice")
	}
	if result[0].Code != "aa" || result[0].Name != "Afar" {
		t.Errorf("ValidLanguages()[0] = %v, want {Code: 'aa', Name: 'Afar'}", result[0])
	}
	// Check that it returns the same as LanguagesTuples
	if len(result) != len(LanguagesTuples) {
		t.Errorf("ValidLanguages() length = %d, want %d", len(result), len(LanguagesTuples))
	}
}

func TestLanguageRegex(t *testing.T) {
	tests := []struct {
		name     string
		tag      language.Tag
		expected string
	}{
		{
			name:     "English language",
			tag:      language.English,
			expected: `a-zA-Z0-9\s`,
		},
		{
			name:     "Arabic language",
			tag:      language.Arabic,
			expected: `\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe`,
		},
		{
			name:     "Chinese language",
			tag:      language.Chinese,
			expected: `\u3100-\u312f\u31a0-\u31bf\u3200-\u32ff\u3300-\u33ff\u3400-\u4db5\u4e00-\u9fff\ua000-\ua48f\ua490-\ua4cf\uf900-\ufaff\ufe30-\ufe4f`,
		},
		{
			name:     "Unknown language",
			tag:      language.Und,
			expected: `a-zA-Z0-9\s`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LanguageRegex(tt.tag)
			if result != tt.expected {
				t.Errorf("LanguageRegex(%v) = %q, want %q", tt.tag, result, tt.expected)
			}
		})
	}
}
