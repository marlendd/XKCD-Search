package core

import (
	"context"
)

type Searcher interface {
	Search(context.Context, string, int) (SearchResult, error)
}

type ISearcher interface {
	ISearch(context.Context, string, int) (SearchResult, error)
}

type DB interface {
	Search(context.Context, []string, int) ([]Comic, error)
	AllComics(context.Context) ([]StoredComic, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type Indexer interface {
	BuildIndex(context.Context) error
}