package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/infra-orchestration/controlplane/api/middleware"
	"github.com/infra-orchestration/controlplane/internal/platform"
)

type Dependencies struct {
	Logger          *slog.Logger
	HTTPAddr        string
	ShutdownTimeout time.Duration
	Metrics         *middleware.Metrics
	Resource        resourceService
	State           stateService
	Plan            planService
	Approval        approvalService
	Execution       executionService
	Drift           driftService
	Audit           auditService
	Graph           graphService
	ReadyCheck      func(context.Context) error
}

type Server struct {
	cfg     platform.ServerConfig
	logger  *slog.Logger
	handler http.Handler
	metrics *middleware.Metrics
	server  *http.Server
}

func NewServer(deps Dependencies) *Server {
	metrics := middleware.NewMetrics()
	deps.Metrics = metrics
	h := newHandler(deps)
	var chain http.Handler = h
	chain = middleware.RateLimit(200, time.Minute, chain)
	chain = metrics.Track(chain)
	chain = middleware.Recovery(deps.Logger, chain)
	chain = middleware.Logging(deps.Logger, chain)
	return &Server{
		cfg:     platform.ServerConfig{HTTPAddr: deps.HTTPAddr, ShutdownTimeout: deps.ShutdownTimeout},
		logger:  deps.Logger,
		handler: chain,
		metrics: metrics,
	}
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           s.handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	s.logger.Info("starting http server", "addr", s.cfg.HTTPAddr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
