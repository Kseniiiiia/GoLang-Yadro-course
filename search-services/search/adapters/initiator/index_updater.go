package initiator

import (
	"context"
	"log/slog"
	"time"

	"yadro.com/course/search/core"
)

type Initiator struct {
	log     *slog.Logger
	service core.Indexer
	ttl     time.Duration
}

func NewInit(log *slog.Logger, service core.Indexer, ttl time.Duration) *Initiator {
	return &Initiator{log: log, service: service, ttl: ttl}
}

func (i *Initiator) Start(ctx context.Context) {
	i.buildIndex(ctx)

	ticker := time.NewTicker(i.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.buildIndex(ctx)
		case <-ctx.Done():
			i.log.Info("Initiator stopped")
			return
		}
	}
}

func (i *Initiator) buildIndex(ctx context.Context) {
	i.log.Info("Building index...")
	if err := i.service.BuildIndex(ctx); err != nil {
		i.log.Error("Index build failed", "error", err)
	}
}
