package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	lockapplication "github.com/infra-orchestration/controlplane/internal/lock/application"
)

type Worker struct {
	service  *Service
	lock     *lockapplication.Service
	logger   *slog.Logger
	owner    string
	interval time.Duration
}

func NewWorker(service *Service, lock *lockapplication.Service, logger *slog.Logger, owner string, interval time.Duration) *Worker {
	if owner == "" {
		owner = fmt.Sprintf("worker-%d", time.Now().UnixNano())
	}
	return &Worker{service: service, lock: lock, logger: logger, owner: owner, interval: interval}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.processNext(ctx); err != nil {
				w.logger.Error("worker iteration failed", "error", err)
			}
		}
	}
}

func (w *Worker) processNext(ctx context.Context) error {
	return w.lock.WithLock(ctx, "execution-worker", w.owner, 10*time.Second, func(ctx context.Context) error {
		task, err := w.service.ProcessOne(ctx, w.owner)
		if err != nil {
			return err
		}
		if task == nil {
			return nil
		}
		w.logger.Info("task processed", "task_id", task.ID, "resource_id", task.ResourceID, "status", task.Status)
		return nil
	})
}
