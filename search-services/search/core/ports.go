package core

import "context"

type Searcher interface {
	Search(ctx context.Context, phrase string, limit int) (SearchResult, error)
	IndexSearch(ctx context.Context, phrase string, limit int) (SearchResult, error)
}

type Indexer interface {
	BuildIndex(ctx context.Context) error
	GetIndex(ctx context.Context) Index
}

type DB interface {
	SearchComics(ctx context.Context, words []string, limit int) ([]Comics, error)
	AllComics(ctx context.Context) ([]Comics, error)
	Stats(ctx context.Context) (DBStats, error)
	GetComicsByIDs(ctx context.Context, ids []int) ([]Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}
