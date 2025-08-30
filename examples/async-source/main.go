package main

import (
	"fmt"

	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/source"
)

func main() {

	config := configuration.NewConfiguration()

	// Example 1: Create a source
	fmt.Println("1. Creating source")
	src, err := source.NewAsyncSource(source.SourceRequest{URL: "https://www.lemonde.fr/", Config: *config})
	if err != nil {
		fmt.Printf("Error creating source: %v\n", err)
		return
	}

	fmt.Printf("   Source created successfully!")
	fmt.Printf("   Domain: %s\n", src.ParsedURL.Domain)
	fmt.Printf("   Scheme: %s\n", src.ParsedURL.Scheme)
	fmt.Printf("   Subdomain: %s\n", src.ParsedURL.Subdomain)
	fmt.Printf("   TLD: %s\n", src.ParsedURL.TLD)

	// Build the source
	fmt.Println("2. Building source...")
	if err := src.Build(); err != nil {
		fmt.Printf("Error building source: %v\n", err)
		return
	}

	fmt.Printf("   Download status: %t\n", src.IsDownloaded)
	fmt.Printf("   Parse status: %t\n", src.IsParsed)
	fmt.Printf("   Categories found: %d\n", len(src.Categories))
	fmt.Printf("   Feeds found: %d\n", len(src.Feeds))
	fmt.Printf("   Description: %s\n", src.Description)
	fmt.Printf("   Articles generated: %d\n", src.Size())

	// Show some categories
	if len(src.Categories) > 0 {
		fmt.Println("3. Sample categories:")
		for i, cat := range src.Categories {
			fmt.Printf("   %d. %s\n", i+1, cat.URL)
		}
	}

	// Show some feeds
	if len(src.Feeds) > 0 {
		fmt.Println("3. Sample feeds:")
		for i, feed := range src.Feeds {
			fmt.Printf("   %d. %s\n", i+1, feed.URL)
		}
	}

	fmt.Println("4. Getting articles:")
	articles := src.GetArticles(10000, false)
	fmt.Printf("   Articles retrieved (not built): %d\n", len(articles))
}
