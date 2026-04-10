package initiator

import (
	"context"
	"log/slog"
	"time"

	"yadro.com/course/search/core"
)

type Initiator struct {
	indexer core.Indexer
	ttl     time.Duration
	log     *slog.Logger
}

func New(indexer core.Indexer, ttl time.Duration, log *slog.Logger) *Initiator {
	return &Initiator{indexer: indexer, ttl: ttl, log: log}
}

func (i *Initiator) Run(ctx context.Context) {
	if err := i.indexer.BuildIndex(ctx); err != nil {
		i.log.Error("failed to build index", "error", err)
	}

	ticker := time.NewTicker(i.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := i.indexer.BuildIndex(ctx); err != nil {
				i.log.Error("failed to build index", "error", err)
			}
		case <-ctx.Done():
			return
		}

	}
}
