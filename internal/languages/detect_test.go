package languages

import (
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func TestInfoMethods(t *testing.T) {
	info := Info{
		lang:        "en",
		probability: 0.8,
		langTag:     language.English,
	}

	t.Run("Tag", func(t *testing.T) {
		if info.Tag() != language.English {
			t.Errorf("Tag() = %v, want %v", info.Tag(), language.English)
		}
	})

	t.Run("LanguageCode", func(t *testing.T) {
		if info.LanguageCode() != "en" {
			t.Errorf("LanguageCode() = %q, want %q", info.LanguageCode(), "en")
		}
	})

	t.Run("Confidence", func(t *testing.T) {
		if info.Confidence() != 0.8 {
			t.Errorf("Confidence() = %v, want %v", info.Confidence(), 0.8)
		}
	})

	t.Run("LanguageName", func(t *testing.T) {
		name := info.LanguageName()
		if name != "English" {
			t.Errorf("LanguageName() = %q, want %q", name, "English")
		}
	})

	t.Run("SelfName", func(t *testing.T) {
		name := info.SelfName()
		if name != "English" {
			t.Errorf("SelfName() = %q, want %q", name, "English")
		}
	})
}

func TestInfoLanguageCode_LongCode(t *testing.T) {
	info := Info{
		lang:        "zh-Hans",
		probability: 0.9,
		langTag:     language.SimplifiedChinese,
	}

	code := info.LanguageCode()
	if code != "zh" {
		t.Errorf("LanguageCode() = %q, want %q", code, "zh")
	}
}

func TestFromReader(t *testing.T) {
	text := "Hello world, this is a test in English."
	reader := strings.NewReader(text)

	info, err := FromReader(reader)
	if err != nil {
		t.Fatalf("FromReader() returned error: %v", err)
	}

	if info.LanguageCode() != "en" {
		t.Errorf("FromReader() detected language %q, want %q", info.LanguageCode(), "en")
	}

	if info.Confidence() <= 0 || info.Confidence() > 1 {
		t.Errorf("FromReader() confidence = %v, want in range (0,1]", info.Confidence())
	}
}

func TestFromString_English(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog. This is a test of English language detection."
	info := FromString(text)

	if info.LanguageCode() != "en" {
		t.Errorf("FromString() detected language %q, want %q", info.LanguageCode(), "en")
	}

	if info.Confidence() <= 0 || info.Confidence() > 1 {
		t.Errorf("FromString() confidence = %v, want in range (0,1]", info.Confidence())
	}
}

func TestFromString_French(t *testing.T) {
	text := "Le renard brun rapide saute par-dessus le chien paresseux. Ceci est un test de détection de langue française."
	info := FromString(text)

	if info.LanguageCode() != "fr" {
		t.Errorf("FromString() detected language %q, want %q", info.LanguageCode(), "fr")
	}
}

func TestFromString_German(t *testing.T) {
	text := "Der schnelle braune Fuchs springt über den faulen Hund. Dies ist ein Test der deutschen Spracherkennung."
	info := FromString(text)

	if info.LanguageCode() != "de" {
		t.Errorf("FromString() detected language %q, want %q", info.LanguageCode(), "de")
	}
}

func TestFromString_Arabic(t *testing.T) {
	text := "مرحبا بالعالم، هذا اختبار للكشف عن اللغة العربية."
	info := FromString(text)

	// Check that confidence is in valid range
	if info.Confidence() <= 0 || info.Confidence() > 1 {
		t.Errorf("FromString() confidence = %v, want in range (0,1]", info.Confidence())
	}

	// The function should not panic and should return some result
	if info.LanguageCode() == "" {
		t.Error("FromString() returned empty language code")
	}

	// Log the actual result for debugging
	t.Logf("Arabic text detected as: %s (confidence: %v)", info.LanguageCode(), info.Confidence())
}

func TestFromString_Chinese(t *testing.T) {
	text := "你好世界，这是一个中文语言检测测试。"
	info := FromString(text)

	// Check that confidence is in valid range
	if info.Confidence() <= 0 || info.Confidence() > 1 {
		t.Errorf("FromString() confidence = %v, want in range (0,1]", info.Confidence())
	}

	// The function should not panic and should return some result
	if info.LanguageCode() == "" {
		t.Error("FromString() returned empty language code")
	}

	// Log the actual result for debugging
	t.Logf("Chinese text detected as: %s (confidence: %v)", info.LanguageCode(), info.Confidence())
}

func TestFromString_MixedScripts(t *testing.T) {
	text := "Hello 世界 مرحبا"
	info := FromString(text)

	// Check that confidence is in valid range
	if info.Confidence() <= 0 || info.Confidence() > 1 {
		t.Errorf("FromString() confidence = %v, want in range (0,1]", info.Confidence())
	}

	// The function should not panic and should return some result
	if info.LanguageCode() == "" {
		t.Error("FromString() returned empty language code")
	}

	// Log the actual result for debugging
	t.Logf("Mixed script text detected as: %s (confidence: %v)", info.LanguageCode(), info.Confidence())
}

func TestFromString_Empty(t *testing.T) {
	info := FromString("")

	if info.LanguageCode() != "und" {
		t.Errorf("FromString() on empty string detected language %q, want %q", info.LanguageCode(), "und")
	}
}

func TestFromString_ShortText(t *testing.T) {
	text := "Hi"
	info := FromString(text)

	// Short text might not be confidently detected, but should not panic
	if info.Confidence() < 0 || info.Confidence() > 1 {
		t.Errorf("FromString() confidence = %v, want in range [0,1]", info.Confidence())
	}
}

func TestFromString_Spanish(t *testing.T) {
	text := "El zorro marrón rápido salta sobre el perro perezoso. Esta es una prueba de detección de idioma español."
	info := FromString(text)

	if info.LanguageCode() != "es" {
		t.Errorf("FromString() detected language %q, want %q", info.LanguageCode(), "es")
	}
}

func TestFromString_Italian(t *testing.T) {
	text := "La volpe marrone veloce salta sopra il cane pigro. Questo è un test di rilevamento della lingua italiana."
	info := FromString(text)

	if info.LanguageCode() != "it" {
		t.Errorf("FromString() detected language %q, want %q", info.LanguageCode(), "it")
	}
}
