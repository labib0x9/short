package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/labib0x9/short/internal/app/url"
)

type Cleaner struct {
	srv url.Service
}

func NewCleaner(srv url.Service) Cleaner {
	return Cleaner{
		srv: srv,
	}
}

func (c *Cleaner) Run(ctx context.Context) error {
	timer := time.NewTicker(5 * time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down cleaner")
			return nil
		case <-timer.C:
			{
				slog.Info("Cleaner cleaning started")
				if err := c.srv.DeleteByExpireAt(ctx); err != nil {
					slog.Error("Cleaner Run()", "error", err)
				}
				slog.Info("Cleaner cleaning done")
			}
		}
	}
}
