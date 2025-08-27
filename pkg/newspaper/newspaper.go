package newspaper

type Extractor interface {
	Parse(a *Article) error
}
