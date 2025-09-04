# newspaper4k-go release notes

## 2025-09-04 - v1.5.1

### Fixed/Improvements

- Fix: deduplicate category URLs using helper function

## 2025-09-04 - v1.5.0

### Added/Refactoring/Deprecation

- Refactor: replace interface{} with any for improved type safety

### Fixed/Improvements

- Fix: enrich file IOCs with full URLs when available and remove duplicate file entries
- Fix: update SourceURL assignment to use ParsedURL for consistency
- Fix: validate category URLs to ensure host, TLD, and domain are present
- Fix: enhance keyword filtering to exclude unk and enforce minimum length

## 2025-09-01 - v1.4.0

### Added/Refactoring/Deprecation

- Feat: have bert tokenizer implemented locally

## 2025-09-01 - v1.3.0

### Added/Refactoring/Deprecation

- Feat(nlp): implement MultiSegmenter for lightweight multilingual tokenization

## 2025-09-01 - v1.2.0

### Added/Refactoring/Deprecation

- Refactor: replace hardcoded English stop words with text.StopwordsEN

## 2025-08-30 - v1.1.0

### Added/Refactoring/Deprecation

- Feat: enhance async source functionality with improved article building and category downloading
- Refactor: simplify GetArticles method and update DefaultSource to use default parameters
- Feat: implement DefaultSource struct and its methods for source handling
- Refactor: rename package from mymodule to newspaper4kgo and remove unused Clone function and tests

### Fixed/Improvements

- Fix: update User-Agent header to reflect correct module version

## 2025-08-29 - v1.0.0

### Removed

- Remove nlp package and associated GetStopWordsForLanguage function
- Remove main binary from repository and update .gitignore

### Added/Refactoring/Deprecation

- Feat: Enhance main functionality with article fetching and parsing, and add simple-source example
- Feat: Implement article fetching and parsing from URL in simple-article example
- Refactor: Simplify URL handling in NewArticle function
- Refactor: Remove unused DownloadArticles and ParseArticles methods from DefaultSource
- Feat: Enhance article JSON conversion with error handling and output display
- Refactor: Remove unused comment for demonstrateURL function in demo code
- Feat: Enhance IOCsExtractor to parse and normalize additional URLs from articles
- Refactor: Update article JSON methods for clarity; consolidate ToJSON functionality
- Feat: Enhance Article struct with additional fields and implement ToSimpleJSON method for simplified JSON representation
- Feat: Add IOCsExtractor to the default extractors for enhanced IOC extraction
- Refactor: Remove deprecated fields and streamline article initialization; enhance language detection in Article struct
- Refactor: Restructure scoreWeights to a struct for improved clarity and maintainability; enhance language detection logic in MetadataExtractor
- Refactor: Improve error handling and simplify client initialization in tests
- Refactor: Normalize key to lowercase in UniqueStructByKey for consistent case handling
- Refactor: Update main function to enhance article and source demonstration flow
- Refactor: Update Source interface and methods for improved clarity and functionality
- Refactor IsValidCategoryURL function to improve URL validation and streamline prefix checks
- Add IsLikelyArticleURL function to improve URL validation and enhance article detection
- Add constants for various extraction configurations and stopwords
- Refactor extractors to use constants for attributes, stopwords, and URL handling
- Add cleanURL function to sanitize URLs and implement tests for it
- Refactor article and list helpers: remove unused functions and clean up imports
- Add URL demonstration: implement URL parsing and manipulation examples in main function
- Refactor article and category handling in newspaper4k

- Introduced a new Category struct to represent category objects.
- Updated Article struct to use []*urls.URL for Categories.
- Added validation for category URLs in the new Category package.
- Refactored video and title extractors to utilize a helper function for document parsing.
- Consolidated common constants and variables into defines.go for better organization.
- Enhanced feed handling by creating a Feed struct and updating related methods.
- Improved error handling in the DefaultSource methods for better reliability.
- Streamlined the Build process in DefaultSource to handle articles and feeds more efficiently.
- Refactor error handling and improve resource management: use blank identifier for error results and replace deprecated string replacement methods
- Refactor CodeQL config and CI workflow: remove paths-ignore section and update test command to cover all packages
- Refactor module references and update versioning in Go files

### Fixed/Improvements

- Fix: Correct version number in release notes from v1.0.0 to 0.1.0
- Fix: Improve error handling when closing response body in Download method
- Fix: Remove Title field from Feed struct to simplify RSS feed representation

## 2025-08-28 - 0.1.0

### Added/Refactoring

- **Initial release** of newspaper4k-go - A Go implementation of newspaper4k
- Complete article extraction functionality from URLs and HTML strings
- Comprehensive metadata extraction including:
  - Article title, authors, and publication date
  - Full article text content
  - Images, videos, and media extraction
  - Keywords and automatic summary generation
  - Language detection and site metadata
  - Favicon, description, and other site information

### Features

- `NewArticleFromURL()` - Extract and parse articles directly from web URLs
- `NewArticleFromHTML()` - Parse articles from HTML string content
- Automatic content cleaning and article structure detection
- Multi-language support with automatic language detection
- Image and video extraction with metadata
- Keyword analysis and content summarization
- Robust error handling for network requests and malformed content

### Others

- Full project setup with Go module structure
- Comprehensive test suite
- Complete documentation and usage examples
- Ready for production use
