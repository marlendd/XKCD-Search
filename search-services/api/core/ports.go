package core

import "context"

//go:generate mockgen -source=ports.go -destination=../mocks/mocks.go -package=mocks

type Normalizer interface {
	Norm(context.Context, string) ([]string, error)
}

type Pinger interface {
	Ping(context.Context) error
}

type Updater interface {
	Update(context.Context) error
	Stats(context.Context) (UpdateStats, error)
	Status(context.Context) (UpdateStatus, error)
	Drop(context.Context) error
}

type Searcher interface {
	Search(context.Context, string, int) (SearchResult, error)
}

type ISearcher interface {
	ISearch(context.Context, string, int) (SearchResult, error)
}