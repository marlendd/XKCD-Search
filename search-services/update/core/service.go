package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	running     atomic.Bool
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
	if !s.running.CompareAndSwap(false, true) {
		return ErrAlreadyExists
	}
	defer s.running.Store(false)

	defer func() {
		if err != nil {
			s.log.Error("update failed", "error", err)
		}
	}()

	g, ctx := errgroup.WithContext(ctx)

	toDownload, err := s.toDownloadIDs(ctx)
	if err != nil {
		return err
	}

	idChan := make(chan int)

	g.Go(func() error {
		defer close(idChan)
		for _, id := range toDownload {
			select {
			case idChan <- id:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	for range s.concurrency {
		g.Go(func() error {
			for id := range idChan {
				if err := s.fetchAndStore(ctx, id); err != nil {
					return err
				}
			}
			return nil
		})
	}

	err = g.Wait()
	return err
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("failed to get stats", "error", err)
		return ServiceStats{}, err
	}

	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get last id", "error", err)
		return ServiceStats{}, err
	}

	return ServiceStats{
		DBStats:     dbStats,
		ComicsTotal: comicsTotal,
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if s.running.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	if s.running.Load() {
		return ErrAlreadyExists
	}
	return s.db.Drop(ctx)
}

func (s *Service) toDownloadIDs(ctx context.Context) ([]int, error) {
	ids, err := s.db.IDs(ctx)
	if err != nil {
		s.log.Error("failed to get ids from db", "error", err)
		return nil, err
	}

	lastId, err := s.xkcd.LastID(ctx)
	if err != nil {
		return nil, err
	}

	existing := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		existing[id] = struct{}{}
	}

	toDownload := make([]int, 0, lastId-len(ids))
	for id := 1; id <= lastId; id++ {
		if _, ok := existing[id]; !ok {
			toDownload = append(toDownload, id)
		}
	}

	return toDownload, nil
}

func (s *Service) fetchAndStore(ctx context.Context, id int) error {
	info, err := s.xkcd.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = s.db.Add(ctx, Comics{ID: id, URL: "", Words: []string{}})
			return nil
		}
		return err
	}

	words, err := s.words.Norm(ctx, info.Description+" "+info.Title+" "+info.SafeTitle+" "+info.Transcript)
	if err != nil {
		if errors.Is(err, ErrBadArguments) {
			s.log.Warn("phrase too long, skipping words", "id", id)
			words = []string{}
		} else {
			return err
		}
	}

	return s.db.Add(ctx, Comics{
		ID:    id,
		URL:   info.URL,
		Words: words,
	})
}
