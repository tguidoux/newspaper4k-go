package helpers

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/tguidoux/newspaper4k-go/internal/parsers"
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

// GetDocFromArticle returns a goquery document from an article, parsing HTML if necessary
func GetDocFromArticle(a *newspaper.Article) (*goquery.Document, error) {
	if a.Doc != nil {
		return a.Doc, nil
	}
	return parsers.FromString(a.HTML)
}
