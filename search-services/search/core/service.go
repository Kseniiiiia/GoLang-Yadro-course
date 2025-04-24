package core

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
)

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
	index Index
	mu    sync.RWMutex
}

func NewService(log *slog.Logger, db DB, words Words) (*Service, error) {
	return &Service{
		log:   log,
		db:    db,
		words: words,
		index: make(Index),
	}, nil
}

func (s *Service) Search(ctx context.Context, phrase string, limit int) (SearchResult, error) {
	words, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return SearchResult{}, fmt.Errorf("normalization failed: %w", err)
	}

	allComics, err := s.db.SearchComics(ctx, words, limit)
	if err != nil {
		return SearchResult{}, fmt.Errorf("db search failed: %w", err)
	}

	resultComics := allComics
	if limit > 0 && len(allComics) > limit {
		resultComics = allComics[:limit]
	}

	return SearchResult{
		Comics: resultComics,
		Total:  len(allComics),
	}, nil
}

func (s *Service) IndexSearch(ctx context.Context, phrase string, limit int) (SearchResult, error) {
	words, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return SearchResult{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	idSet := make(map[int]struct{})
	wordToIds := make(map[string]map[int]struct{})

	for _, word := range words {
		if ids, exists := s.index[word]; exists {
			wordToIds[word] = make(map[int]struct{})
			for _, id := range ids {
				wordToIds[word][id] = struct{}{}
				idSet[id] = struct{}{}
			}
		}
	}

	ids := make([]int, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	comics, err := s.db.GetComicsByIDs(ctx, ids)
	if err != nil {
		return SearchResult{}, err
	}

	sort.Slice(comics, func(i, j int) bool {
		iUnique, iTotal := countMatches(comics[i].ID, comics[i].Words, words, wordToIds)
		jUnique, jTotal := countMatches(comics[j].ID, comics[j].Words, words, wordToIds)

		// Сначала комиксы с большим количеством уникальных совпадений
		if iUnique != jUnique {
			return iUnique > jUnique
		}
		// Затем по общему количеству совпадений
		return iTotal > jTotal
	})

	if limit > 0 && len(comics) > limit {
		comics = comics[:limit]
	}

	return SearchResult{Comics: comics, Total: len(comics)}, nil
}

func countMatches(id int, comicWords, searchWords []string, wordToIds map[string]map[int]struct{}) (int, int) {
	uniqueMatches := 0
	totalMatches := 0

	comicWordCount := make(map[string]int)
	for _, word := range comicWords {
		comicWordCount[word]++
	}

	for _, word := range searchWords {
		if ids, exists := wordToIds[word]; exists {
			if _, hasWord := ids[id]; hasWord {
				uniqueMatches++
				totalMatches += comicWordCount[word]
			}
		}
	}

	return uniqueMatches, totalMatches
}

func (s *Service) GetIndex(ctx context.Context) Index {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index
}

func (s *Service) BuildIndex(ctx context.Context) error {
	comics, err := s.db.AllComics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get comics: %w", err)
	}

	newIndex := make(Index)
	for _, comic := range comics {
		for _, word := range comic.Words {
			newIndex[word] = append(newIndex[word], comic.ID)
		}
	}

	s.mu.RLock()
	s.index = newIndex
	s.mu.RUnlock()

	s.log.Info("Index rebuilt",
		"total_comics", len(comics),
		"unique_words", len(newIndex))

	return nil
}

func (s *Service) Stats(ctx context.Context) (DBStats, error) {
	return s.db.Stats(ctx)
}

func (s *Service) Ping(ctx context.Context) error {
	return nil
}
