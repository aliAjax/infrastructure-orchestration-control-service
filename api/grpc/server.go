package grpcapi

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	approvalapp "github.com/infra-orchestration/controlplane/internal/approval/application"
	auditapp "github.com/infra-orchestration/controlplane/internal/audit/application"
	driftapp "github.com/infra-orchestration/controlplane/internal/drift/application"
	executionapp "github.com/infra-orchestration/controlplane/internal/execution/application"
	graphapp "github.com/infra-orchestration/controlplane/internal/graph/application"
	planapp "github.com/infra-orchestration/controlplane/internal/plan/application"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
	resourceapp "github.com/infra-orchestration/controlplane/internal/resource/application"
	stateapp "github.com/infra-orchestration/controlplane/internal/state/application"
)

type Dependencies struct {
	Logger    *slog.Logger
	Resource  *resourceapp.Service
	State     *stateapp.Service
	Plan      *planapp.Service
	Approval  *approvalapp.Service
	Execution *executionapp.Service
	Drift     *driftapp.Service
	Audit     *auditapp.Service
	Graph     *graphapp.Service
}

type HealthRequest struct{}
type HealthResponse struct {
	Status string `json:"status"`
}

type ListProjectsRequest struct{}
type ListProjectsResponse struct {
	Projects []any `json:"projects"`
}

type GeneratePlanRequest struct {
	EnvironmentID string `json:"environment_id"`
	CreatedBy     string `json:"created_by"`
}

type GeneratePlanResponse struct {
	Plan any `json:"plan"`
}

type Server struct {
	cfg        string
	logger     *slog.Logger
	deps       Dependencies
	grpcServer *grpc.Server
}

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

func NewServer(addr string, deps Dependencies) *Server {
	server := grpc.NewServer(grpc.ForceServerCodec(jsonCodec{}))
	s := &Server{cfg: addr, logger: deps.Logger, deps: deps, grpcServer: server}
	server.RegisterService(&ServiceDesc, s)
	return s
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.cfg)
	if err != nil {
		return fmt.Errorf("listen grpc %s: %w", s.cfg, err)
	}
	s.logger.Info("starting grpc server", "addr", s.cfg)
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

func (s *Server) Stop() {
	s.grpcServer.Stop()
}

func (s *Server) Health(ctx context.Context, req *HealthRequest) (*HealthResponse, error) {
	return &HealthResponse{Status: "ok"}, nil
}

func (s *Server) ListProjects(ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error) {
	projects, err := s.deps.Resource.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, len(projects))
	for _, project := range projects {
		out = append(out, project)
	}
	return &ListProjectsResponse{Projects: out}, nil
}

func (s *Server) GeneratePlan(ctx context.Context, req *GeneratePlanRequest) (*GeneratePlanResponse, error) {
	plan, err := s.deps.Plan.Generate(ctx, plandomain.GenerateInput{EnvironmentID: req.EnvironmentID, CreatedBy: req.CreatedBy})
	if err != nil {
		return nil, err
	}
	return &GeneratePlanResponse{Plan: plan}, nil
}

var ServiceDesc = grpc.ServiceDesc{
	ServiceName: "infra.controlplane.v1.ControlPlane",
	HandlerType: (*ControlPlaneServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Health",
			Handler: unaryHandler(func(s *Server, ctx context.Context, req *HealthRequest) (*HealthResponse, error) {
				return s.Health(ctx, req)
			}),
		},
		{
			MethodName: "ListProjects",
			Handler: unaryHandler(func(s *Server, ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error) {
				return s.ListProjects(ctx, req)
			}),
		},
		{
			MethodName: "GeneratePlan",
			Handler: unaryHandler(func(s *Server, ctx context.Context, req *GeneratePlanRequest) (*GeneratePlanResponse, error) {
				return s.GeneratePlan(ctx, req)
			}),
		},
	},
}

type ControlPlaneServer interface {
	Health(context.Context, *HealthRequest) (*HealthResponse, error)
	ListProjects(context.Context, *ListProjectsRequest) (*ListProjectsResponse, error)
	GeneratePlan(context.Context, *GeneratePlanRequest) (*GeneratePlanResponse, error)
}

func unaryHandler[Req, Resp any](fn func(*Server, context.Context, *Req) (*Resp, error)) func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	return func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		req := new(Req)
		if err := dec(req); err != nil {
			return nil, err
		}
		s, ok := srv.(*Server)
		if !ok {
			return nil, fmt.Errorf("invalid server type")
		}
		if interceptor == nil {
			return fn(s, ctx, req)
		}
		info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/infra.controlplane.v1.ControlPlane/Unknown"}
		handler := func(ctx context.Context, req any) (any, error) {
			return fn(s, ctx, req.(*Req))
		}
		return interceptor(ctx, req, info, handler)
	}
}
