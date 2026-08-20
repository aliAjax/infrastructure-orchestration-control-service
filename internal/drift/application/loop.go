package application

import (
	"context"
	"log/slog"
	"time"
)

type Loop struct {
	service  *Service
	logger   *slog.Logger
	interval time.Duration
}

func NewLoop(service *Service, logger *slog.Logger, interval time.Duration) *Loop {
	return &Loop{service: service, logger: logger, interval: interval}
}

func (l *Loop) Run(ctx context.Context) error {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			l.runOnce(ctx)
		}
	}
}

func (l *Loop) runOnce(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		l.logger.Warn("drift loop cancelled", "error", err)
		return
	}
	if l.service == nil || l.service.resourceRepo == nil {
		l.logger.Warn("drift loop has no resource repository")
		return
	}
	environments, err := l.service.resourceRepo.ListResources(ctx, "")
	if err != nil {
		l.logger.Error("drift loop list resources failed", "error", err)
		return
	}
	seen := make(map[string]struct{})
	for _, resource := range environments {
		envID := resource.EnvironmentID
		if _, ok := seen[envID]; ok {
			continue
		}
		seen[envID] = struct{}{}
		if _, err := l.service.Detect(ctx, envID); err != nil {
			l.logger.Error("drift detection failed", "environment_id", envID, "error", err)
		}
	}
}
