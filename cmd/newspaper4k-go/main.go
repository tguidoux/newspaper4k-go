package main

import (
	"fmt"

	"github.com/tguidoux/newspaper4k-go/internal/helpers"
	"github.com/tguidoux/newspaper4k-go/internal/urls"
	"github.com/tguidoux/newspaper4k-go/pkg/configuration"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper4k"
	"github.com/tguidoux/newspaper4k-go/pkg/source"
)

func main() {
	fmt.Println("Newspaper4k-Go Article Parser Demo")
	fmt.Println("==================================")

	// demonstrateURL()

	// Demonstrate Article functionality
	// demonstrateArticleUsage()

	// Demonstrate Source functionality
	demonstrateSourceUsage()
}

func demonstrateURL() {
	fmt.Println("\n\n=== URL PACKAGE DEMO ===")
	fmt.Println("Demonstrating URL parsing and manipulation")
	fmt.Println("=========================================")

	// Example URLs to parse
	urlsToParse := []string{
		"https://edition.cnn.com/2023/10/01/world/sample-article/index.html?utm_source=newsletter",
		"http://subdomain.example.co.uk/path/to/page?query=123#section",
		"https://www.bbc.com/news/world-12345678",
		"https://www.bbc.com/news/world-12345678#fasef",
		"https://example.com",
		"https://sub.sub2.example.com",
		"ftp://invalid-protocol.com/resource",
		"not-a-valid-url",
	}

	for _, urlStr := range urlsToParse {
		fmt.Printf("\nParsing URL: %s\n", urlStr)
		parsedURL, err := urls.Parse(urlStr)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		fmt.Printf("  Scheme: %s\n", parsedURL.Scheme)
		fmt.Printf("  Domain: %s\n", parsedURL.Domain)
		fmt.Printf("  Subdomain: %s\n", parsedURL.Subdomain)
		fmt.Printf("  TLD: %s\n", parsedURL.TLD)
		fmt.Printf("  Port: %s\n", parsedURL.Port)
		fmt.Printf("  Path: %s\n", parsedURL.Path)
		fmt.Printf("  RawQuery: %s\n", parsedURL.RawQuery)
		fmt.Printf("  Fragment: %s\n", parsedURL.Fragment)
		fmt.Printf("  String: %s\n", parsedURL.String())

		// Demonstrate URL preparation
		sourceURL, err := urls.Parse("https://edition.cnn.com")
		if err != nil {
			fmt.Printf("  Source URL Parse Error: %v\n", err)
			continue
		}
		err = parsedURL.Prepare(sourceURL)
		if err != nil {
			fmt.Printf("  Prepare Error: %v\n", err)
			continue
		}
		fmt.Printf("  Prepared URL: %s\n", parsedURL.String())
		fmt.Printf("  File Type: %s\n", parsedURL.FileType)
		fmt.Printf("  Is Valid: %t\n", parsedURL.IsValidNewsArticleURL())
	}
}

//	func demonstrateArticleUsage() {
//		// Sample HTML content for testing
//		testHTML := `
//		 <!DOCTYPE html>
//		 <html lang="en">
//		 <head>
//		 	<title>Breaking News: Major Scientific Discovery Announced Today</title>
//		 	<meta property="og:title" content="Major Scientific Discovery Announced Today" />
//		 	<meta name="description" content="Scientists have made a groundbreaking discovery that could change the world." />
//		 	<meta property="og:description" content="Scientists have made a groundbreaking discovery that could change the world." />
//		 	<meta property="og:image" content="https://example.com/images/scientific-discovery.jpg" />
//		 	<meta name="twitter:image" content="https://example.com/images/twitter-discovery.jpg" />
//		 	<link rel="icon" type="image/png" href="/favicon.png" />
//		 	<meta name="author" content="Dr. Jane Smith" />
//		 	<meta property="article:author" content="Dr. Jane Smith" />
//		 	<meta name="author" content="Dr. Michael Johnson" />
//		 	<meta property="article:published_time" content="2025-08-27T09:15:00Z" />
//		 	<meta name="publishdate" content="2025-08-27" />
//		 	<meta property="og:published_time" content="2025-08-27T09:15:00Z" />
//		 	<script type="application/ld+json">
//		 	{
//		 		"@context": "https://schema.org",
//		 		"@type": "NewsArticle",
//		 		"headline": "Major Scientific Discovery Announced Today",
//		 		"datePublished": "2025-08-27T10:30:00Z",
//		 		"author": {
//		 			"@type": "Person",
//		 			"name": "Dr. Jane Smith"
//		 		},
//		 		"publisher": {
//		 			"@type": "Organization",
//		 			"name": "Science News"
//		 		}
//		 	}
//		 	</script>
//		 	<script type="application/ld+json">
//		 	{
//		 		"@context": "https://schema.org",
//		 		"@graph": [
//		 			{
//		 				"@type": "NewsArticle",
//		 				"datePublished": "2025-08-27",
//		 				"author": [
//		 					{
//		 						"@type": "Person",
//		 						"name": "Dr. Jane Smith"
//		 					},
//		 					{
//		 						"@type": "Person",
//		 						"name": "Dr. Michael Johnson"
//		 					}
//		 				]
//		 			}
//		 		]
//		 	}
//		 	</script>
//		 </head>
//		 <body>
//		 	<header>
//		 		<nav>
//		 			<a href="/science">Science</a>
//		 			<a href="/technology">Technology</a>
//		 			<a href="/health">Health</a>
//		 			<a href="/environment">Environment</a>
//		 			<a href="/politics">Politics</a>
//		 			<a href="/sports">Sports</a>
//		 			<a href="/business">Business</a>
//		 			<a href="/entertainment">Entertainment</a>
//		 			<a href="/world">World</a>
//		 			<a href="/us">U.S.</a>
//		 			<a href="/latest">Latest</a>
//		 			<a href="/breaking">Breaking</a>
//		 		</nav>
//		 		<h1>Breaking News: Major Scientific Discovery Announced Today</h1>
//		 	</header>
//		 	<main>
//		 		<article>
//		 			<h1>Major Scientific Discovery Announced Today</h1>
//		 			<img src="/images/scientific-lab.jpg" alt="Scientific laboratory" />
//		 			<p class="byline">By Dr. Jane Smith, Senior Science Reporter</p>
//		 			<time datetime="2025-08-27T08:00:00Z" published>Published on August 27, 2025</time>
//		 			<p class="author">Written by: Dr. Michael Johnson and Dr. Sarah Davis</p>
//		 			<div itemprop="author" itemscope itemtype="https://schema.org/Person">
//		 				<span itemprop="name">Dr. Jane Smith</span>
//		 			</div>
//		 			<p>In a groundbreaking announcement today, scientists at the International Research Institute revealed a major breakthrough in renewable energy technology.</p>
//		 			<img src="https://example.com/images/energy-breakthrough.png" alt="Energy breakthrough diagram" />
//		 			<p>The discovery promises to revolutionize how we harness clean energy sources, potentially solving the world's energy crisis within the next decade.</p>
//		 			<h2>The Breakthrough</h2>
//		 			<p>Researchers have developed a new method for storing solar energy that is both more efficient and cost-effective than current technologies.</p>
//		 			<p>"This could be the game-changer we've been waiting for," said Dr. Michael Johnson, lead researcher on the project.</p>
//		 			<h2>Implications</h2>
//		 			<p>The implications of this discovery are far-reaching, affecting everything from transportation to industrial manufacturing.</p>
//		 			<p class="author-info">Contributors: Dr. Sarah Davis, Research Assistant</p>
//
//		 			<!-- Video content for testing -->
//		 			<h2>Watch the Announcement</h2>
//		 			<iframe width="560" height="315" src="https://www.youtube.com/embed/dQw4w9WgXcQ" frameborder="0" allowfullscreen></iframe>
//		 			<p>Watch Dr. Smith's full announcement in the video above.</p>
//
//		 			<video width="560" height="315" controls>
//		 				<source src="https://example.com/videos/discovery-announcement.mp4" type="video/mp4">
//		 				Your browser does not support the video tag.
//		 			</video>
//
//		 			<div>
//		 				<object data="https://vimeo.com/76979871" width="560" height="315">
//		 					<embed src="https://vimeo.com/76979871" width="560" height="315">
//		 				</object>
//		 			</div>
//
//		 			<script type="application/ld+json">
//		 			{
//		 				"@context": "https://schema.org",
//		 				"@type": "VideoObject",
//		 				"name": "Scientific Discovery Announcement",
//		 				"description": "Full announcement of the major scientific breakthrough",
//		 				"contentUrl": "https://example.com/videos/announcement-full.mp4",
//		 				"thumbnailUrl": "https://example.com/images/video-thumbnail.jpg",
//		 				"width": 1920,
//		 				"height": 1080
//		 			}
//		 			</script>
//		 		</article>
//		 	</main>
//		 </body>
//		 </html>`
//		// art, _ := newspaper.NewArticleFromHTML(testHTML)
//		art, err := newspaper4k.NewArticleFromURL("https://www.welovetennis.fr/us-open/daniil-medvedev-quand-jaurai-35-ans-je-boycotterai-peut-etre-les-matches-de-11-heures-je-ferai-forfait")
//		if err != nil {
//			fmt.Printf("Error fetching article: %v\n", err)
//			return
//		}
//
//		// Display results
//		helpers.PrintArticleResults(art)
//
//		art2, err := newspaper4k.NewArticleFromHTML(testHTML)
//		if err != nil {
//			fmt.Printf("Error fetching article: %v\n", err)
//			return
//		}
//		helpers.PrintArticleResults(art2)
//
//		fmt.Println("\nDemo completed successfully!")
//	}
func demonstrateSourceUsage() {
	fmt.Println("\n\n=== SOURCE PACKAGE DEMO ===")
	fmt.Println("Demonstrating news source crawling and article discovery")
	fmt.Println("====================================================")

	// Create a configuration
	config := configuration.NewConfiguration()

	// Example HTML content for testing
	//	mockHTML := `
	//	<!DOCTYPE html>
	//	<html>
	//	<head>
	//		<title>Hacker News</title>
	//		<meta name="description" content="Hacker News is a social news website focusing on computer science and entrepreneurship">
	//	</head>
	//	<body>
	//		<header>
	//			<nav>
	//				<a href="/newest">New</a>
	//				<a href="/past">Past</a>
	//				<a href="/comments">Comments</a>
	//				<a href="/ask">Ask</a>
	//				<a href="/show">Show</a>
	//				<a href="/jobs">Jobs</a>
	//				<a href="/submit">Submit</a>
	//			</nav>
	//		</header>
	//		<main>
	//			<table>
	//				<tr>
	//					<td><a href="/item?id=1">First Story Title</a></td>
	//					<td><a href="/item?id=1">(42 comments)</a></td>
	//				</tr>
	//				<tr>
	//					<td><a href="/item?id=2">Second Story Title</a></td>
	//					<td><a href="/item?id=2">(15 comments)</a></td>
	//				</tr>
	//				<tr>
	//					<td><a href="/item?id=3">Third Story Title</a></td>
	//					<td><a href="/item?id=3">(8 comments)</a></td>
	//				</tr>
	//			</table>
	//		</main>
	//	</body>
	//	</html>`

	// Example 1: Create a source with mock HTML
	fmt.Printf("\n1. Creating source\n")
	src, err := source.NewDefaultSource(source.SourceRequest{URL: "https://www.lemonde.fr/", Config: *config})
	if err != nil {
		fmt.Printf("Error creating source: %v\n", err)
		return
	}

	fmt.Printf("   Source created successfully!\n")
	fmt.Printf("   Domain: %s\n", src.ParsedURL.Domain)
	fmt.Printf("   Scheme: %s\n", src.ParsedURL.Scheme)
	fmt.Printf("   Subdomain: %s\n", src.ParsedURL.Subdomain)
	fmt.Printf("   TLD: %s\n", src.ParsedURL.TLD)

	// Build the source with mock HTML
	fmt.Printf("\n2. Building source...\n")
	src.Build() // inputHTML=mockHTML, onlyHomepage=false, onlyInPath=false

	fmt.Printf("   Download status: %t\n", src.IsDownloaded)
	fmt.Printf("   Parse status: %t\n", src.IsParsed)
	fmt.Printf("   Categories found: %d\n", len(src.Categories))
	fmt.Printf("   Feeds found: %d\n", len(src.Feeds))
	fmt.Printf("   Description: %s\n", src.Description)
	fmt.Printf("   Articles generated: %d\n", src.Size())

	// Show some categories
	if len(src.Categories) > 0 {
		fmt.Printf("\n3. Sample categories:\n")
		for i, cat := range src.Categories {
			fmt.Printf("   %d. %s\n", i+1, cat.URL)
		}
	}

	// Show some feeds
	if len(src.Feeds) > 0 {
		fmt.Printf("\n3. Sample feeds:\n")
		for i, feed := range src.Feeds {
			fmt.Printf("   %d. %s\n", i+1, feed.URL)
		}
	}

	// Show some articles
	if src.Size() > 0 {
		fmt.Printf("\n4. Sample articles:\n")
		articles := src.Articles
		for i, a := range articles {
			fmt.Printf("   %d. %s\n", i+1, a.URL)
		}

		// Show article details without downloading
		fmt.Printf("\n5. Article details (from mock data):\n")
		for i, article := range src.Articles[:helpers.Min(3, len(src.Articles))] {
			fmt.Printf("   %d. Title: %s\n", i+1, article.Title)
			fmt.Printf("      URL: %s\n", article.URL)
			fmt.Printf("      Source URL: %s\n", article.SourceURL)
		}
	}

	fmt.Println("Source built successfully!")

	fmt.Println("Getting articles")
	articles := src.GetArticles(10000, false)
	fmt.Printf("   Articles retrieved: %d\n", len(articles))
	extractors := newspaper4k.DefaultExtractors(config)
	for i, article := range articles[:helpers.Min(3, len(articles))] {
		article.Build(extractors)
		if err != nil {
			fmt.Printf("Error building article: %v\n", err)
			continue
		}

		parsedURL, err := urls.Parse(article.URL)
		if err != nil {
			fmt.Printf("Error parsing article URL: %v\n", err)
			continue
		}
		fmt.Printf("   %d. Title: %s\n", i+1, article.Title)
		fmt.Printf("      URL: %s\n", parsedURL.String())
		fmt.Printf("      Source URL: %s\n", article.SourceURL)
		fmt.Printf("      Authors: %v\n", article.Authors)
		fmt.Printf("      Publish Date: %v\n", article.PublishDate)
		fmt.Printf("      Top Image: %s\n", article.TopImage)
		fmt.Printf("      Text (first 100 chars): %.100s\n", article.Text)

	}

	// Example 2: Using only homepage parsing
	//	fmt.Printf("\n\n6. Example with only homepage parsing:\n")
	//	src2, err := source.NewDefaultSource(source.SourceRequest{URL: "https://news.ycombinator.com", Config: *config})
	//	if err != nil {
	//		fmt.Printf("Error creating source: %v\n", err)
	//		return
	//	}
	//
	//	src2.Build(mockHTML, true, false) // onlyHomepage=true
	//	fmt.Printf("   Homepage-only parsing - Articles found: %d\n", src2.Size())

	fmt.Printf("\nSource demo completed!\n")
}
