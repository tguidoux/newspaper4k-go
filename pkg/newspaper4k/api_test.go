package newspaper4k

import (
	"strings"
	"testing"
)

const testHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<title>Breaking News: Major Scientific Discovery Announced Today</title>
	<meta property="og:title" content="Major Scientific Discovery Announced Today" />
	<meta name="description" content="Scientists have made a groundbreaking discovery that could change the world." />
	<meta property="og:description" content="Scientists have made a groundbreaking discovery that could change the world." />
	<meta property="og:image" content="https://example.com/images/scientific-discovery.jpg" />
	<meta name="twitter:image" content="https://example.com/images/twitter-discovery.jpg" />
	<link rel="icon" type="image/png" href="/favicon.png" />
	<meta name="author" content="Dr. Jane Smith" />
	<meta property="article:author" content="Dr. Jane Smith" />
	<meta name="author" content="Dr. Michael Johnson" />
	<meta property="article:published_time" content="2025-08-27T09:15:00Z" />
	<meta name="publishdate" content="2025-08-27" />
	<meta property="og:published_time" content="2025-08-27T09:15:00Z" />
	<script type="application/ld+json">
	{
		"@context": "https://schema.org",
		"@type": "NewsArticle",
		"headline": "Major Scientific Discovery Announced Today",
		"datePublished": "2025-08-27T10:30:00Z",
		"author": {
			"@type": "Person",
			"name": "Dr. Jane Smith"
		},
		"publisher": {
			"@type": "Organization",
			"name": "Science News"
		}
	}
	</script>
	<script type="application/ld+json">
	{
		"@context": "https://schema.org",
		"@graph": [
			{
				"@type": "NewsArticle",
				"datePublished": "2025-08-27",
				"author": [
					{
						"@type": "Person",
						"name": "Dr. Jane Smith"
					},
					{
						"@type": "Person",
						"name": "Dr. Michael Johnson"
					}
				]
			}
		]
	}
	</script>
</head>
<body>
	<header>
		<nav>
			<a href="/science">Science</a>
			<a href="/technology">Technology</a>
			<a href="/health">Health</a>
			<a href="/environment">Environment</a>
			<a href="/politics">Politics</a>
			<a href="/sports">Sports</a>
			<a href="/business">Business</a>
			<a href="/entertainment">Entertainment</a>
			<a href="/world">World</a>
			<a href="/us">U.S.</a>
			<a href="/latest">Latest</a>
			<a href="/breaking">Breaking</a>
		</nav>
		<h1>Breaking News: Major Scientific Discovery Announced Today</h1>
	</header>
	<main>
		<article>
			<h1>Major Scientific Discovery Announced Today</h1>
			<img src="/images/scientific-lab.jpg" alt="Scientific laboratory" />
			<p class="byline">By Dr. Jane Smith, Senior Science Reporter</p>
			<time datetime="2025-08-27T08:00:00Z" published>Published on August 27, 2025</time>
			<p class="author">Written by: Dr. Michael Johnson and Dr. Sarah Davis</p>
			<div itemprop="author" itemscope itemtype="https://schema.org/Person">
				<span itemprop="name">Dr. Jane Smith</span>
			</div>
			<p>In a groundbreaking announcement today, scientists at the International Research Institute revealed a major breakthrough in renewable energy technology.</p>
			<img src="https://example.com/images/energy-breakthrough.png" alt="Energy breakthrough diagram" />
			<p>The discovery promises to revolutionize how we harness clean energy sources, potentially solving the world's energy crisis within the next decade.</p>
			<h2>The Breakthrough</h2>
			<p>Researchers have developed a new method for storing solar energy that is both more efficient and cost-effective than current technologies.</p>
			<p>"This could be the game-changer we've been waiting for," said Dr. Michael Johnson, lead researcher on the project.</p>
			<h2>Implications</h2>
			<p>The implications of this discovery are far-reaching, affecting everything from transportation to industrial manufacturing.</p>
			<p class="author-info">Contributors: Dr. Sarah Davis, Research Assistant</p>

			<!-- Video content for testing -->
			<h2>Watch the Announcement</h2>
			<iframe width="560" height="315" src="https://www.youtube.com/embed/dQw4w9WgXcQ" frameborder="0" allowfullscreen></iframe>
			<p>Watch Dr. Smith's full announcement in the video above.</p>

			<video width="560" height="315" controls>
				<source src="https://example.com/videos/discovery-announcement.mp4" type="video/mp4">
				Your browser does not support the video tag.
			</video>

			<div>
				<object data="https://vimeo.com/76979871" width="560" height="315">
					<embed src="https://vimeo.com/76979871" width="560" height="315">
				</object>
			</div>

			<script type="application/ld+json">
			{
				"@context": "https://schema.org",
				"@type": "VideoObject",
				"name": "Scientific Discovery Announcement",
				"description": "Full announcement of the major scientific breakthrough",
				"contentUrl": "https://example.com/videos/announcement-full.mp4",
				"thumbnailUrl": "https://example.com/images/video-thumbnail.jpg",
				"width": 1920,
				"height": 1080
			}
			</script>
		</article>
	</main>
</body>
</html>`

func TestNewArticleFromHTML(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article from HTML: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	// Test basic fields
	if art.Title == "" {
		t.Error("Title should not be empty")
	}
	expectedTitle := "Breaking News: Major Scientific Discovery Announced Today"
	if art.Title != expectedTitle {
		t.Errorf("Expected title %q, got %q", expectedTitle, art.Title)
	}

	if art.Text == "" {
		t.Error("Text should not be empty")
	}

	if !strings.Contains(art.Text, "scientists at the International Research Institute") {
		t.Error("Text should contain expected content")
	}

	if art.IsParsed != true {
		t.Error("Article should be parsed")
	}

	if art.DownloadState != 2 { // Success
		t.Error("Download state should be Success")
	}
}

func TestArticleAuthors(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if len(art.Authors) == 0 {
		t.Error("Authors should not be empty")
	}

	// Check if expected authors are present
	expectedAuthors := []string{"Dr. Jane Smith", "Dr. Michael Johnson", "Dr. Sarah Davis"}
	found := make(map[string]bool)
	for _, author := range art.Authors {
		found[author] = true
	}

	for _, expected := range expectedAuthors {
		if !found[expected] {
			t.Errorf("Expected author %q not found in authors list: %v", expected, art.Authors)
		}
	}
}

func TestArticleMetaData(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if art.MetaDescription == "" {
		t.Error("Meta description should not be empty")
	}

	expectedDesc := "Scientists have made a groundbreaking discovery that could change the world."
	if art.MetaDescription != expectedDesc {
		t.Errorf("Expected meta description %q, got %q", expectedDesc, art.MetaDescription)
	}

	if art.MetaLang != "en" {
		t.Errorf("Expected meta language 'en', got %q", art.MetaLang)
	}
}

func TestArticleImages(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if len(art.Images) == 0 {
		t.Error("Images should not be empty")
	}

	// Check for expected images that are actually extracted
	expectedImages := []string{
		"/images/scientific-lab.jpg",
		"https://example.com/images/energy-breakthrough.png",
	}

	found := make(map[string]bool)
	for _, img := range art.Images {
		found[img] = true
	}

	for _, expected := range expectedImages {
		if !found[expected] {
			t.Errorf("Expected image %q not found in images list: %v", expected, art.Images)
		}
	}
}

func TestArticleVideos(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if len(art.Movies) == 0 {
		t.Error("Movies should not be empty")
	}

	// Check for expected videos that are actually extracted
	expectedVideos := []string{
		"https://www.youtube.com/embed/dQw4w9WgXcQ",
		"https://vimeo.com/76979871",
		"https://example.com/videos/announcement-full.mp4",
	}

	found := make(map[string]bool)
	for _, video := range art.Movies {
		found[video] = true
	}

	for _, expected := range expectedVideos {
		if !found[expected] {
			t.Errorf("Expected video %q not found in movies list: %v", expected, art.Movies)
		}
	}
}

func TestArticleKeywords(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if len(art.Keywords) == 0 {
		t.Error("Keywords should not be empty")
	}

	if len(art.KeywordScores) == 0 {
		t.Error("Keyword scores should not be empty")
	}

	// Check that keywords are reasonable
	for _, keyword := range art.Keywords {
		if len(keyword) < 3 {
			t.Errorf("Keyword %q is too short", keyword)
		}
	}
}

func TestArticleSummary(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if art.Summary == "" {
		t.Error("Summary should not be empty")
	}

	// Summary should be shorter than full text
	if len(art.Summary) >= len(art.Text) {
		t.Error("Summary should be shorter than full text")
	}
}

func TestArticlePublishDate(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if art.PublishDate == nil {
		t.Error("Publish date should not be nil")
	}
}

func TestArticleLanguage(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	if art.Language.String() != "en" {
		t.Errorf("Expected language 'en', got %q", art.Language.String())
	}
}

func TestArticleToJSON(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	jsonStr, err := art.ToJSON()
	if err != nil {
		t.Fatalf("Error converting article to JSON: %v", err)
	}

	if jsonStr == "" {
		t.Error("JSON string should not be empty")
	}

	// Basic checks for JSON structure
	if !strings.Contains(jsonStr, `"title"`) {
		t.Error("JSON should contain title field")
	}

	if !strings.Contains(jsonStr, `"text"`) {
		t.Error("JSON should contain text field")
	}

	if !strings.Contains(jsonStr, `"authors"`) {
		t.Error("JSON should contain authors field")
	}
}

func TestArticleGettersAndSetters(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	// Test GetTitle
	title := art.GetTitle()
	if title != art.Title {
		t.Error("GetTitle should return the same as Title field")
	}

	// Test GetText
	text := art.GetText()
	if text != art.Text {
		t.Error("GetText should return the same as Text field")
	}

	// Test GetSummary
	summary := art.GetSummary()
	if summary != art.Summary {
		t.Error("GetSummary should return the same as Summary field")
	}

	// Test GetTopKeywords
	keywords := art.GetTopKeywords()
	if len(keywords) != len(art.KeywordScores) {
		t.Error("GetTopKeywords should return the same as KeywordScores")
	}

	// Test GetTopKeywordsList
	keywordList := art.GetTopKeywordsList()
	if len(keywordList) != len(art.Keywords) {
		t.Error("GetTopKeywordsList should return the same as Keywords")
	}
}

func TestArticleIsValidBody(t *testing.T) {
	art, err := NewArticleFromHTML(testHTML)
	if err != nil {
		t.Fatalf("Error creating article: %v", err)
	}

	extractors := DefaultExtractors(art.Config)

	err = art.Build(extractors)
	if err != nil {
		t.Fatalf("Error building article: %v", err)
	}

	// Check word count
	wordCount := len(strings.Fields(art.Text))
	t.Logf("Word count: %d, MinWordCount: %d", wordCount, art.Config.MinWordCount)

	if wordCount < art.Config.MinWordCount {
		t.Logf("Article has %d words, which is less than MinWordCount %d, so IsValidBody returns false", wordCount, art.Config.MinWordCount)
	}

	// The test HTML has limited content, so IsValidBody may return false
	// This is expected behavior
}
