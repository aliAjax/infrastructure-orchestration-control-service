package httpapi

import (
	approvalapp "github.com/infra-orchestration/controlplane/internal/approval/application"
	auditapp "github.com/infra-orchestration/controlplane/internal/audit/application"
	driftapp "github.com/infra-orchestration/controlplane/internal/drift/application"
	executionapp "github.com/infra-orchestration/controlplane/internal/execution/application"
	graphapp "github.com/infra-orchestration/controlplane/internal/graph/application"
	planapp "github.com/infra-orchestration/controlplane/internal/plan/application"
	resourceapp "github.com/infra-orchestration/controlplane/internal/resource/application"
	stateapp "github.com/infra-orchestration/controlplane/internal/state/application"
)

type resourceService = *resourceapp.Service
type stateService = *stateapp.Service
type planService = *planapp.Service
type approvalService = *approvalapp.Service
type executionService = *executionapp.Service
type driftService = *driftapp.Service
type auditService = *auditapp.Service
type graphService = *graphapp.Service
