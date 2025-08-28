package helpers

import (
	"fmt"
	"strings"

	"github.com/tguidoux/newspaper4k-go/pkg/newspaper"
)

// UniqueOptions configures the behavior of UniqueStrings
type UniqueOptions struct {
	CaseSensitive bool
	PreserveOrder bool
}

// UniqueStrings removes duplicates from a slice of strings with configurable options
func UniqueStrings(slice []string, opts UniqueOptions) []string {
	if len(slice) == 0 {
		return slice
	}

	if opts.PreserveOrder {
		seen := make(map[string]bool)
		result := make([]string, 0, len(slice))

		for _, item := range slice {
			key := item
			if !opts.CaseSensitive {
				key = strings.ToLower(item)
			}

			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		return result
	} else {
		// Use map for O(1) lookup, order not preserved
		seen := make(map[string]bool)
		result := make([]string, 0, len(slice))

		for _, item := range slice {
			key := item
			if !opts.CaseSensitive {
				key = strings.ToLower(item)
			}

			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		return result
	}
}

// UniqueStringsSimple removes duplicates from a slice of strings (case-sensitive, order preserved)
func UniqueStringsSimple(slice []string) []string {
	return UniqueStrings(slice, UniqueOptions{
		CaseSensitive: true,
		PreserveOrder: true,
	})
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UniqueStringsByKey removes duplicates from a slice using a key function
func UniqueStringsByKey(items []string, keyFunc func(string) string, opts UniqueOptions) []string {
	if len(items) == 0 {
		return items
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(items))

	for _, item := range items {
		key := keyFunc(item)
		if !opts.CaseSensitive && opts.PreserveOrder {
			key = strings.ToLower(key)
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result
}

// PrintArticleResults prints the results of an article parsing to stdout
func PrintArticleResults(art *newspaper.Article) {
	fmt.Println("\n=== PARSED ARTICLE RESULTS ===")
	fmt.Printf("Title: %s\n", art.Title)
	fmt.Printf("Source URL: %s\n", art.SourceURL)
	fmt.Printf("Is Parsed: %t\n", art.IsParsed)
	fmt.Printf("Authors: %v\n", art.Authors)
	fmt.Printf("Meta Description: %s\n", art.MetaDescription)
	fmt.Printf("Meta Language: %s\n", art.MetaLang)
	fmt.Printf("Meta Site Name: %s\n", art.MetaSiteName)
	fmt.Printf("Meta Keywords: %v\n", art.MetaKeywords)
	fmt.Printf("Canonical Link: %s\n", art.CanonicalLink)
	fmt.Printf("Categories: %v\n", art.Categories)
	fmt.Printf("Top Image: %s\n", art.TopImage)
	fmt.Printf("Meta Image: %s\n", art.MetaImg)
	fmt.Printf("Images: %v\n", art.Images)
	fmt.Printf("Favicon: %s\n", art.MetaFavicon)
	fmt.Printf("Movies: %v\n", art.Movies)
	fmt.Printf("Pub Date: %v\n", art.PublishDate)
	fmt.Printf("Language: %v\n", art.Language)
	fmt.Printf("Text: %v\n", art.Text)
	fmt.Printf("Keywords: %v\n", art.GetTopKeywordsList())
	fmt.Printf("Summary: %s\n", art.GetSummary())
}
