package languages

import (
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// convertUnicodeEscapes converts Unicode escape sequences like \u0400 to actual Unicode characters
func convertUnicodeEscapes(s string) string {
	// Regular expression to match \u followed by exactly 4 hexadecimal digits
	re := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)

	// Replace all matches with the corresponding Unicode character
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract the hex part (skip the \u)
		hexStr := match[2:]
		// Convert hex to int
		codePoint, err := strconv.ParseInt(hexStr, 16, 32)
		if err != nil {
			// If parsing fails, return the original match
			return match
		}
		// Convert to Unicode character
		return string(rune(codePoint))
	})

	return result
}

// LanguageTuple represents a language with its ISO code and name
type LanguageTuple struct {
	Code string
	Name string
}

// LanguagesTuples contains all supported languages with their ISO codes and names
var LanguagesTuples = []LanguageTuple{
	{"aa", "Afar"},
	{"ab", "Abkhazian"},
	{"ae", "Avestan"},
	{"af", "Afrikaans"},
	{"ak", "Akan"},
	{"am", "Amharic"},
	{"an", "Aragonese"},
	{"ar", "Arabic"},
	{"as", "Assamese"},
	{"av", "Avaric"},
	{"ay", "Aymara"},
	{"az", "Azerbaijani"},
	{"ba", "Bashkir"},
	{"be", "Belarusian"},
	{"bg", "Bulgarian"},
	{"bh", "Bihari languages"},
	{"bi", "Bislama"},
	{"bm", "Bambara"},
	{"bn", "Bengali"},
	{"bo", "Tibetan"},
	{"br", "Breton"},
	{"bs", "Bosnian"},
	{"ca", "Catalan; Valencian"},
	{"ce", "Chechen"},
	{"ch", "Chamorro"},
	{"co", "Corsican"},
	{"cr", "Cree"},
	{"cs", "Czech"},
	{"cv", "Chuvash"},
	{"cy", "Welsh"},
	{"da", "Danish"},
	{"de", "German"},
	{"dv", "Divehi; Dhivehi; Maldivian"},
	{"dz", "Dzongkha"},
	{"ee", "Ewe"},
	{"el", "Greek, Modern (1453-)"},
	{"en", "English"},
	{"eo", "Esperanto"},
	{"es", "Spanish; Castilian"},
	{"et", "Estonian"},
	{"eu", "Basque"},
	{"fa", "Persian"},
	{"ff", "Fulah"},
	{"fi", "Finnish"},
	{"fj", "Fijian"},
	{"fo", "Faroese"},
	{"fr", "French"},
	{"fy", "Western Frisian"},
	{"ga", "Irish"},
	{"gd", "Gaelic; Scottish Gaelic"},
	{"gl", "Galician"},
	{"gn", "Guarani"},
	{"gu", "Gujarati"},
	{"gv", "Manx"},
	{"ha", "Hausa"},
	{"he", "Hebrew"},
	{"hi", "Hindi"},
	{"ho", "Hiri Motu"},
	{"hr", "Croatian"},
	{"ht", "Haitian; Haitian Creole"},
	{"hu", "Hungarian"},
	{"hy", "Armenian"},
	{"hz", "Herero"},
	{"id", "Indonesian"},
	{"ig", "Igbo"},
	{"ii", "Sichuan Yi; Nuosu"},
	{"ik", "Inupiaq"},
	{"io", "Ido"},
	{"is", "Icelandic"},
	{"it", "Italian"},
	{"iu", "Inuktitut"},
	{"ja", "Japanese"},
	{"jv", "Javanese"},
	{"ka", "Georgian"},
	{"kg", "Kongo"},
	{"ki", "Kikuyu; Gikuyu"},
	{"kj", "Kuanyama; Kwanyama"},
	{"kk", "Kazakh"},
	{"kl", "Kalaallisut; Greenlandic"},
	{"km", "Central Khmer"},
	{"kn", "Kannada"},
	{"ko", "Korean"},
	{"kr", "Kanuri"},
	{"ks", "Kashmiri"},
	{"ku", "Kurdish"},
	{"kv", "Komi"},
	{"kw", "Cornish"},
	{"ky", "Kirghiz; Kyrgyz"},
	{"la", "Latin"},
	{"lb", "Luxembourgish; Letzeburgesch"},
	{"lg", "Ganda"},
	{"li", "Limburgan; Limburger; Limburgish"},
	{"ln", "Lingala"},
	{"lo", "Lao"},
	{"lt", "Lithuanian"},
	{"lu", "Luba-Katanga"},
	{"lv", "Latvian"},
	{"mg", "Malagasy"},
	{"mh", "Marshallese"},
	{"mi", "Maori"},
	{"mk", "Macedonian"},
	{"ml", "Malayalam"},
	{"mn", "Mongolian"},
	{"mr", "Marathi"},
	{"ms", "Malay"},
	{"mt", "Maltese"},
	{"my", "Burmese"},
	{"na", "Nauru"},
	{"nb", "Bokmål, Norwegian; Norwegian Bokmål"},
	{"nd", "Ndebele, North; North Ndebele"},
	{"ne", "Nepali"},
	{"ng", "Ndonga"},
	{"nl", "Dutch; Flemish"},
	{"nn", "Norwegian Nynorsk; Nynorsk, Norwegian"},
	{"no", "Norwegian"},
	{"nr", "Ndebele, South; South Ndebele"},
	{"nv", "Navajo; Navaho"},
	{"oc", "Occitan (post 1500)"},
	{"oj", "Ojibwa"},
	{"om", "Oromo"},
	{"or", "Oriya"},
	{"os", "Ossetian; Ossetic"},
	{"pa", "Panjabi; Punjabi"},
	{"pi", "Pali"},
	{"pl", "Polish"},
	{"ps", "Pushto; Pashto"},
	{"pt", "Portuguese"},
	{"qu", "Quechua"},
	{"rm", "Romansh"},
	{"rn", "Rundi"},
	{"ro", "Romanian; Moldavian; Moldovan"},
	{"ru", "Russian"},
	{"rw", "Kinyarwanda"},
	{"sa", "Sanskrit"},
	{"sc", "Sardinian"},
	{"sd", "Sindhi"},
	{"se", "Northern Sami"},
	{"sg", "Sango"},
	{"si", "Sinhala; Sinhalese"},
	{"sk", "Slovak"},
	{"sl", "Slovenian"},
	{"sm", "Samoan"},
	{"sn", "Shona"},
	{"so", "Somali"},
	{"sq", "Albanian"},
	{"sr", "Serbian"},
	{"ss", "Swati"},
	{"st", "Sotho, Southern"},
	{"su", "Sundanese"},
	{"sv", "Swedish"},
	{"sw", "Swahili"},
	{"ta", "Tamil"},
	{"te", "Telugu"},
	{"tg", "Tajik"},
	{"th", "Thai"},
	{"ti", "Tigrinya"},
	{"tk", "Turkmen"},
	{"tl", "Tagalog"},
	{"tn", "Tswana"},
	{"to", "Tonga (Tonga Islands)"},
	{"tr", "Turkish"},
	{"ts", "Tsonga"},
	{"tt", "Tatar"},
	{"tw", "Twi"},
	{"ty", "Tahitian"},
	{"ug", "Uighur; Uyghur"},
	{"uk", "Ukrainian"},
	{"ur", "Urdu"},
	{"uz", "Uzbek"},
	{"ve", "Venda"},
	{"vi", "Vietnamese"},
	{"vo", "Volapük"},
	{"wa", "Walloon"},
	{"wo", "Wolof"},
	{"xh", "Xhosa"},
	{"yi", "Yiddish"},
	{"yo", "Yoruba"},
	{"za", "Zhuang; Chuang"},
	{"zh", "Chinese"},
	{"zu", "Zulu"},
}

// LanguagesDict maps ISO codes to language names
var LanguagesDict = map[string]string{}

// LanguagesTags maps ISO codes to language.Tag for convenient usage
var LanguagesTags = map[string]language.Tag{}

// LanguagesUnicodeRegex contains Unicode regex patterns for various languages
var LanguagesUnicodeRegex = map[string]string{
	"aa": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"ab": `[\u0400-\u04ff]`,
	"ae": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"am": `[\u1200-\u137f]`,
	"ar": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"as": `[\u0980-\u09ff]`,
	"av": `[\u0400-\u04ff\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"az": `[\u0600-\u06ff]`,
	"ba": `[\u0400-\u04ff]`,
	"be": `[\u0400-\u04ff]`,
	"bg": `[\u0400-\u04ff]`,
	"bh": `[\u0900-\u097f]`,
	"bn": `[\u0980-\u09ff]`,
	"bo": `[\u0f00-\u0fff]`,
	"bs": `[\u0400-\u04ff]`,
	"ce": `[\u0400-\u04ff]`,
	"cr": `[\u1400-\u167f]`,
	"cv": `[\u0400-\u04ff]`,
	"cy": `[\u1680-\u169f]`,
	"dv": `[\u0f00-\u0fff]`,
	"dz": `[\u0780-\u07bf\u0f00-\u0fff]`,
	"el": `[\u0370-\u03ff\u1f00-\u1fff]`,
	"fa": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"ff": `[\u0600-\u06ff]`,
	"ga": `[\u1680-\u169f]`,
	"gu": `[\u0a80-\u0aff]`,
	"ha": `[\u0600-\u06ff]`,
	"he": `[\u0590-\u05ff]`,
	"hi": `[\u0900-\u097f]`,
	"hy": `[\u0530-\u058f]`,
	"ii": `[\u3100-\u312f\u31a0-\u31bf\u3200-\u32ff\u3300-\u33ff\u3400-\u4db5\u4e00-\u9fff\ua000-\ua48f\ua490-\ua4cf\uf900-\ufaff\ufe30-\ufe4f]`,
	"ja": `[\u3040-\u309f\u30a0-\u30ff\u3190-\u319f\u3200-\u32ff\u3300-\u33ff\u3400-\u4db5\u4e00-\u9fff\uf900-\ufaff\ufe30-\ufe4f]`,
	"jv": `[\u0600-\u06ff]`,
	"ka": `[\u10a0-\u10ff]`,
	"kk": `[\u0400-\u04ff]`,
	"km": `[\u1780-\u17ff]`,
	"kn": `[\u0c80-\u0cff]`,
	"ko": `[\u1100-\u11ff\u3130-\u318f\u3200-\u32ff\u3300-\u33ff\u3400-\u4db5\u4e00-\u9fff\uac00-\ud7a3\uf900-\ufaff\ufe30-\ufe4f]`,
	"kr": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"ks": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"ku": `[\u0600-\u06ff]`,
	"kv": `[\u0400-\u04ff]`,
	"ky": `[\u0400-\u04ff\u0600-\u06ff]`,
	"lg": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"lo": `[\u0e80-\u0eff]`,
	"mk": `[\u0400-\u04ff]`,
	"ml": `[\u0d00-\u0d7f]`,
	"mn": `[\u1800-\u18af\u0400-\u04ff]`,
	"mr": `[\u0900-\u097f]`,
	"my": `[\u1000-\u109f]`,
	"ne": `[\u0900-\u097f]`,
	"oj": `[\u1400-\u167f]`,
	"or": `[\u0b00-\u0b7f]`,
	"os": `[\u0400-\u04ff]`,
	"pa": `[\u0a00-\u0a7f]`,
	"pi": `[\u1000-\u109f\u1780-\u17ff]`,
	"ps": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"ru": `[\u0400-\u04ff]`,
	"sa": `[\u0900-\u097f\u0980-\u09ff\u0c80-\u0cff\u0d00-\u0d7f\u0f00-\u0fff\u1000-\u109f]`,
	"sd": `[\ufb50-\ufdff\ufe70-\ufefe]`,
	"si": `[\u0d80-\u0dff]`,
	"sr": `[\u0400-\u04ff]`,
	"sw": `[\u0600-\u06ff]`,
	"ta": `[\u0b80-\u0bff]`,
	"te": `[\u0c00-\u0c7f]`,
	"tg": `[\u0400-\u04ff]`,
	"th": `[\u0e00-\u0e7f]`,
	"ti": `[\u1200-\u137f]`,
	"tk": `[\u0400-\u04ff\u0600-\u06ff]`,
	"tt": `[\u0400-\u04ff\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"ug": `[\u0400-\u04ff\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"uk": `[\u0400-\u04ff]`,
	"ur": `[\u0600-\u06ff\ufb50-\ufdff\ufe70-\ufefe]`,
	"uz": `[\u0400-\u04ff\u0600-\u06ff]`,
	"wo": `[\u0600-\u06ff]`,
	"yi": `[\u0590-\u05ff]`,
	"za": `[\u0400-\u04ff]`,
	"zh": `[\u3100-\u312f\u31a0-\u31bf\u3200-\u32ff\u3300-\u33ff\u3400-\u4db5\u4e00-\u9fff\ua000-\ua48f\ua490-\ua4cf\uf900-\ufaff\ufe30-\ufe4f]`,
}

func init() {
	// Initialize the LanguagesDict from LanguagesTuples
	for _, lang := range LanguagesTuples {
		LanguagesDict[lang.Code] = lang.Name
	}
	// Initialize LanguagesTags map
	for _, lang := range LanguagesTuples {
		LanguagesTags[lang.Code] = language.Make(lang.Code)
	}
}

// GetTagFromISO639_1 returns a language.Tag for the given ISO 639-1 code.
// If the code is unknown, it returns language.Und (undetermined).
func GetTagFromISO639_1(iso639_1 string) language.Tag {
	if tag, exists := LanguagesTags[iso639_1]; exists {
		return tag
	}
	return language.Und
}

// ValidLanguages returns the list of available languages as tuples (iso-code, language)
func ValidLanguages() []LanguageTuple {
	return LanguagesTuples
}

// LanguageRegex returns the regex pattern for characters used by the given language
func LanguageRegex(tag language.Tag) string {
	// Return a pattern for characters that should NOT be filtered out
	// This is used in [^...] to match characters to keep
	// Use canonical tag string (may include script) and try direct lookup
	canonical := tag.String()
	if pat, ok := LanguagesUnicodeRegex[canonical]; ok {
		pat = strings.TrimSpace(pat)
		if strings.HasPrefix(pat, "[") && strings.HasSuffix(pat, "]") {
			return pat[1 : len(pat)-1]
		}
		return pat
	}

	// Try base language (e.g., "en" from "en-US")
	base := canonical
	if i := strings.Index(canonical, "-"); i > 0 {
		base = canonical[:i]
	}
	if len(base) > 2 {
		base = base[:2]
	}
	if pat, ok := LanguagesUnicodeRegex[base]; ok {
		pat = strings.TrimSpace(pat)
		if strings.HasPrefix(pat, "[") && strings.HasSuffix(pat, "]") {
			return pat[1 : len(pat)-1]
		}
		return pat
	}

	// Default to Latin-range characters
	return `a-zA-Z0-9\s`
}

// LanguageRegexPattern returns a regex pattern with Unicode escapes converted to actual characters
// This is meant for use in regexp compilation
func LanguageRegexPattern(tag language.Tag) string {
	// Get the raw pattern from LanguageRegex
	pattern := LanguageRegex(tag)
	// Convert Unicode escapes to actual Unicode characters
	return convertUnicodeEscapes(pattern)
}
