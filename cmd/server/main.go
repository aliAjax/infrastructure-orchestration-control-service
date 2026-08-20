package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	grpcapi "github.com/infra-orchestration/controlplane/api/grpc"
	httpapi "github.com/infra-orchestration/controlplane/api/http"
	approvalapp "github.com/infra-orchestration/controlplane/internal/approval/application"
	approvalinfra "github.com/infra-orchestration/controlplane/internal/approval/infrastructure"
	auditapp "github.com/infra-orchestration/controlplane/internal/audit/application"
	auditinfra "github.com/infra-orchestration/controlplane/internal/audit/infrastructure"
	driftapp "github.com/infra-orchestration/controlplane/internal/drift/application"
	driftinfra "github.com/infra-orchestration/controlplane/internal/drift/infrastructure"
	executionapp "github.com/infra-orchestration/controlplane/internal/execution/application"
	executioninfra "github.com/infra-orchestration/controlplane/internal/execution/infrastructure"
	graphapp "github.com/infra-orchestration/controlplane/internal/graph/application"
	lockapp "github.com/infra-orchestration/controlplane/internal/lock/application"
	lockinfra "github.com/infra-orchestration/controlplane/internal/lock/infrastructure"
	planapp "github.com/infra-orchestration/controlplane/internal/plan/application"
	planinfra "github.com/infra-orchestration/controlplane/internal/plan/infrastructure"
	"github.com/infra-orchestration/controlplane/internal/platform"
	resourceapp "github.com/infra-orchestration/controlplane/internal/resource/application"
	resourceinfra "github.com/infra-orchestration/controlplane/internal/resource/infrastructure"
	stateapp "github.com/infra-orchestration/controlplane/internal/state/application"
	stateinfra "github.com/infra-orchestration/controlplane/internal/state/infrastructure"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	migrateOnly := flag.Bool("migrate-only", false, "run migrations and exit")
	disableWorker := flag.Bool("disable-worker", false, "disable the in-process execution worker")
	flag.Parse()

	cfg, err := platform.LoadConfig(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	logger := platform.NewLogger(cfg.Logging.Level)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := platform.OpenPostgres(ctx, cfg.Database)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := platform.RunMigrations(ctx, db, "migrations"); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}
	if *migrateOnly {
		logger.Info("migrations complete")
		return
	}

	redisClient, err := platform.OpenRedis(ctx, cfg.Redis)
	if err != nil {
		logger.Error("open redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	resourceRepo := resourceinfra.NewPostgresRepository(db)
	stateRepo := stateinfra.NewPostgresRepository(db)
	planRepo := planinfra.NewPostgresRepository(db)
	approvalRepo := approvalinfra.NewPostgresRepository(db)
	auditRepo := auditinfra.NewPostgresRepository(db)
	executionRepo := executioninfra.NewPostgresRepository(db)
	driftRepo := driftinfra.NewPostgresRepository(db)

	graphService := graphapp.NewService()
	resourceService := resourceapp.NewService(resourceRepo)
	stateService := stateapp.NewService(stateRepo)
	planService := planapp.NewService(planRepo, resourceRepo, stateRepo, graphService)
	approvalService := approvalapp.NewService(approvalRepo)
	auditService := auditapp.NewService(auditRepo)
	mockRunner := executioninfra.NewMockRunner()
	executionService := executionapp.NewService(executionRepo, mockRunner, planService, stateService)
	driftService := driftapp.NewService(driftRepo, resourceRepo, stateRepo, planService)

	lockManager := lockinfra.NewRedisManager(redisClient)
	lockService := lockapp.NewService(lockManager)
	worker := executionapp.NewWorker(executionService, lockService, logger, "server-worker", cfg.Executor.PollInterval)
	driftLoop := driftapp.NewLoop(driftService, logger, cfg.Drift.Interval)

	httpDeps := httpapi.Dependencies{
		Logger:          logger,
		HTTPAddr:        cfg.Server.HTTPAddr,
		ShutdownTimeout: cfg.Server.ShutdownTimeout,
		Resource:        resourceService,
		State:           stateService,
		Plan:            planService,
		Approval:        approvalService,
		Execution:       executionService,
		Drift:           driftService,
		Audit:           auditService,
		Graph:           graphService,
		ReadyCheck: func(ctx context.Context) error {
			if err := db.Ping(ctx); err != nil {
				return err
			}
			return redisClient.Ping(ctx).Err()
		},
	}
	httpServer := httpapi.NewServer(httpDeps)
	grpcDeps := grpcapi.Dependencies{
		Logger:    logger,
		Resource:  resourceService,
		State:     stateService,
		Plan:      planService,
		Approval:  approvalService,
		Execution: executionService,
		Drift:     driftService,
		Audit:     auditService,
		Graph:     graphService,
	}
	grpcServer := grpcapi.NewServer(cfg.Server.GRPCAddr, grpcDeps)

	errCh := make(chan error, 4)
	go func() { errCh <- httpServer.Start() }()
	go func() { errCh <- grpcServer.Start() }()
	if !*disableWorker {
		go func() { errCh <- worker.Run(ctx) }()
	}
	go func() { errCh <- driftLoop.Run(ctx) }()

	logger.Info("control plane started", "http", cfg.Server.HTTPAddr, "grpc", cfg.Server.GRPCAddr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal", "signal", sig.String())
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
		}
	}
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()
	logger.Info("shutdown complete")
}
