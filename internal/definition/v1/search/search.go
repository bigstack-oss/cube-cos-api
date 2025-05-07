package search

import (
	"strings"

	"github.com/blevesearch/bleve/v2"
)

const (
	MaxSearchResults = 100000
)

func New() (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	return bleve.NewMemOnly(mapping)
}

func NormalizedKeyword(keyword string) string {
	keyword = strings.ToLower(keyword)
	return strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		":", "",
		".", "",
		",", "",
		"@", "",
		"!", "",
		"#", "",
	).Replace(keyword)
}

func WrapWilcard(keyword string) string {
	return "*" + strings.ToLower(keyword) + "*"
}

func WildcardQuery(keyword string) *bleve.SearchRequest {
	return bleve.NewSearchRequestOptions(
		bleve.NewWildcardQuery(WrapWilcard(keyword)),
		MaxSearchResults,
		0,
		false,
	)
}
