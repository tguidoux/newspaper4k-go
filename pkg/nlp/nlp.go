package nlp

import "github.com/tguidoux/newspaper4k-go/internal/nlp"

func GetStopWordsForLanguage(code string) []string {
	return nlp.GetStopWordsForLanguage(code)
}
