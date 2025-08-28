# newspaper4k-go

[![Go Report Card](https://goreportcard.com/badge/github.com/tguidoux/newspaper4k-go)](https://goreportcard.com/report/github.com/tguidoux/newspaper4k-go)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/tguidoux/newspaper4k-go)](https://pkg.go.dev/github.com/tguidoux/newspaper4k-go)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Actions Status](https://github.com/tguidoux/newspaper4k-go/workflows/ci/badge.svg)](https://github.com/tguidoux/newspaper4k-go/actions)
[![codecov](https://codecov.io/gh/tguidoux/newspaper4k-go/branch/main/graph/badge.svg)](https://codecov.io/gh/tguidoux/newspaper4k-go)
[![Release](https://img.shields.io/github/release/tguidoux/newspaper4k-go.svg?style=flat-square)](RELEASE-NOTES.md)

A Go library for extracting and parsing articles from web pages. This is a Go implementation of [newspaper4k](https://github.com/AndyTheFactory/newspaper4k), inspired by the [newspaper](https://github.com/codelucas/newspaper) Python library. It provides functionality to scrape article content, extract metadata such as title, authors, publish date, and more from HTML pages or URLs.

**Note:** This library does not support all features of newspaper4k yet, but it can successfully parse articles! Pull requests are welcome to add more features.

## Installation

To use this module in your Go project, you can use the `go get` command:

```bash
$ go get -u github.com/tguidoux/newspaper4k-go
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/tguidoux/newspaper4k-go/pkg/newspaper4k"
)

func main() {
	// Extract article from URL
	article, err := newspaper4k.NewArticleFromURL("https://example.com/article")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Title: %s\n", article.Title)
	fmt.Printf("Authors: %v\n", article.Authors)
	fmt.Printf("Text: %s\n", article.Text)
	fmt.Printf("Publish Date: %v\n", article.PublishDate)

	// Or extract from HTML string
	html := `<html><body><h1>Test Article</h1><p>This is the content.</p></body></html>`
	article2, err := newspaper4k.NewArticleFromHTML(html)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Title: %s\n", article2.Title)
}
```

## Configuration

This library does not require any configuration. Simply import and use the functions as shown in the Usage section.

## Contributing

We welcome contributions! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Acknowledgments

- This is a Go port of [newspaper4k](https://github.com/AndyTheFactory/newspaper4k) by [AndyTheFactory](https://github.com/AndyTheFactory)
- Thanks to [codelucas](https://github.com/codelucas) for the original [newspaper](https://github.com/codelucas/newspaper) Python library
- Thanks to all contributors and the Go community

## Contact

Théo Guidoux - [GitHub](https://github.com/tguidoux)

## Release History

See the [RELEASE-NOTES.md](RELEASE-NOTES.md) file for details.

## Support

If you encounter any issues or have questions, feel free to [create an issue](https://github.com/tguidoux/newspaper4k-go/issues).
