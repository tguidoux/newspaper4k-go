package helpers

import (
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// UniqueArticlesByURL removes duplicate articles by URL
func UniqueArticlesByURL(articles []newspaper.Article) []newspaper.Article {
	seen := make(map[string]bool)
	result := make([]newspaper.Article, 0, len(articles))

	for _, article := range articles {
		if !seen[article.URL] {
			seen[article.URL] = true
			result = append(result, article)
		}
	}

	return result
}
