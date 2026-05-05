package core

import (
	"context"
	"log/slog"
	"sort"
	"sync"
)

//go:generate mockgen -source=ports.go -destination=../mocks/mocks.go -package=mocks

type Index struct {
	idx map[string][]int
	comics map[int]Comic
	mu  sync.RWMutex
}

type comicScore struct {
	ID int
	Score int
}

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
	index Index
}

func NewService(log *slog.Logger, db DB, words Words) *Service {
	return &Service{
		log:   log,
		db:    db,
		words: words,
		index: Index{},
	}
}

func (s *Service) Search(ctx context.Context, phrase string, limit int) (SearchResult, error) {
	norm, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return SearchResult{}, err
	}

	comics, err := s.db.Search(ctx, norm, limit)
	if err != nil {
		return SearchResult{}, err
	}

	return SearchResult{Comics: comics}, nil
}

func (s *Service) ISearch(ctx context.Context, phrase string, limit int) (SearchResult, error) {
	norm, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return SearchResult{}, err
	}
	scores := make(map[int] int, len(norm))

	s.index.mu.RLock()
	defer s.index.mu.RUnlock()

	for _, word := range norm {
		ids := s.index.idx[word]
		for _, id := range ids {
			scores[id]++
		}
	}

	var comics []comicScore

	for id, score := range scores {
		comics = append(comics, comicScore{
			ID: id,
			Score: score,
		})
	}

	sort.Slice(comics, func(i, j int) bool {
		if comics[i].Score != comics[j].Score {
			return comics[i].Score > comics[j].Score
		}
		return comics[i].ID < comics[j].ID
	})

	n := min(limit, len(comics))
	result := make([]Comic, n)

	for i, comic := range comics[:n] {
		result[i] = s.index.comics[comic.ID]
	}

	return SearchResult{Comics: result}, nil
}

func (s *Service) BuildIndex(ctx context.Context) error {
	comics, err := s.db.AllComics(ctx)
	if err != nil {
		s.log.Error("failed to get all comics", "error", err)
		return err
	}
	idx := make(map[string][]int, len(comics))
	comicsMap := make(map[int]Comic, len(comics))

	for _, comic := range comics {
		for _, word := range comic.Words {
			idx[word] = append(idx[word], comic.ID)
		}
		comicsMap[comic.ID] = Comic{
			ID: comic.ID,
			URL: comic.URL,
		}
	}

	s.index.mu.Lock()
	s.index.idx = idx
	s.index.comics = comicsMap
	s.index.mu.Unlock()

	return nil
}