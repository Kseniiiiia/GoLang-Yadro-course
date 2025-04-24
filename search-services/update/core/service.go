package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	mu          sync.Mutex
	updates     bool
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
	}, nil
}

func (s *Service) Update(ctx context.Context) (err error) {
	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last id: %w", err)
	}

	existIDs, err := s.db.IDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get existing IDs: %w", err)
	}

	existIDsMap := make(map[int]struct{}, len(existIDs))
	for _, id := range existIDs {
		existIDsMap[id] = struct{}{}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, s.concurrency)
	var errFinal error
	var once sync.Once

	for id := 1; id <= lastID; id++ {
		if _, exists := existIDsMap[id]; exists {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()

			info, err := s.xkcd.Get(ctx, id)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					return
				}
				once.Do(func() {
					errFinal = fmt.Errorf("failed to get comics %d: %w", id, err)
				})
				return
			}

			words, err := s.words.Norm(ctx, info.Title+" "+info.Description)
			if err != nil {
				once.Do(func() {
					errFinal = fmt.Errorf("failed to normalize words for comics %d: %w", id, err)
				})
				return
			}

			comics := Comics{
				ID:    info.NUM,
				URL:   info.URL,
				Words: words,
			}

			if err := s.db.Add(ctx, comics); err != nil {
				once.Do(func() {
					errFinal = fmt.Errorf("failed to add comics %d to db: %w", id, err)
				})
				return
			}
		}(id)
	}

	wg.Wait()

	return errFinal
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		return ServiceStats{}, fmt.Errorf("failed to get db stats: %w", err)
	}

	comicsTotal, err := s.Count(ctx)
	if err != nil {
		return ServiceStats{}, fmt.Errorf("failed to count comics: %w", err)
	}

	return ServiceStats{
		DBStats:     dbStats,
		ComicsTotal: comicsTotal,
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.updates {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	err := s.db.Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}
	return nil
}

func (s *Service) Count(ctx context.Context) (int, error) {
	lastID, err := s.xkcd.LastID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get last ID: %w", err)
	}

	missingIDs := s.xkcd.MissingIds(ctx)
	return lastID - len(missingIDs), nil
}
