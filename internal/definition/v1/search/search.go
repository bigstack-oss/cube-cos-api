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

func WrapWilcard(keyword string) string {
	return "*" + strings.ToLower(keyword) + "*"
}
