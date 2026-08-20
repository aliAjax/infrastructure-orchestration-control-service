# Infrastructure Orchestration Control Plane

This repository is a pure Go control plane for declaring infrastructure resources,
generating executable plans, enforcing approvals, executing changes through remote
runners, detecting drift, and keeping a full audit trail. It exposes REST, gRPC, and
a CLI. It does not include a browser frontend.

## Architecture

```text
cmd/
  server/    HTTP + gRPC control plane
  cli/       API client
  runner/    standalone remote runner example
api/
  http/      REST handlers and server
  grpc/      JSON-coded gRPC service
  middleware logging, recovery, rate limiting, metrics
internal/
  resource   project/environment/resource/variable declarations
  state      versioned desired/actual state, snapshots, rollback
  plan       diff generation and persisted plan history
  approval   approval requests and policy enforcement
  execution  horizontally scalable task execution and remote runners
  graph      dependency DAG validation and level execution
  lock       Redis distributed locks and explicit lock declarations
  drift      periodic drift detection and remediation plans
  audit      immutable audit event storage
```

Every domain uses the same layering:

```text
domain/          models and repository contracts
application/     use cases and orchestration
adapter/         DTO mapping
infrastructure/  PostgreSQL or Redis adapters
```

## Requirements

- Go 1.22+
- Docker with Docker Compose
- PostgreSQL 16
- Redis 7

## Quick Start

```bash
docker compose up -d
go mod tidy
go build ./...
go run ./cmd/server -config=configs/config.yaml
```

The service starts HTTP on `:8080`, gRPC on `:9090`, and an in-process execution
worker plus drift detector.

## Configuration

Configuration is read from `configs/config.yaml` by default. Every important value
can be overridden with an environment variable:

```bash
export INFRA_HTTP_ADDR=:8080
export INFRA_GRPC_ADDR=:9090
export INFRA_DATABASE_URL='postgres://infra:infra@localhost:5432/infra_controlplane?sslmode=disable'
export INFRA_REDIS_ADDR=localhost:6379
export INFRA_LOG_LEVEL=debug
```

## REST API

Health and observability:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl http://localhost:8080/metrics
```

Resource declarations:

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"platform","description":"shared infrastructure"}'

curl -X POST http://localhost:8080/api/v1/environments \
  -H 'Content-Type: application/json' \
  -d '{"project_id":"<project_id>","name":"production","kind":"production","production":true}'

curl -X POST http://localhost:8080/api/v1/resources \
  -H 'Content-Type: application/json' \
  -d '{"environment_id":"<env_id>","name":"database","type":"postgres","desired_state":{"engine":"16","storage_gb":100}}'
```

Plan, approval, and execution:

```bash
curl -X POST http://localhost:8080/api/v1/plans/generate \
  -H 'Content-Type: application/json' \
  -d '{"environment_id":"<env_id>","created_by":"alice"}'

curl -X POST http://localhost:8080/api/v1/approvals \
  -H 'Content-Type: application/json' \
  -d '{"plan_id":"<plan_id>","environment_id":"<env_id>","requested_by":"alice","reason":"production change"}'

curl -X POST http://localhost:8080/api/v1/approvals/<approval_id>/decision \
  -H 'Content-Type: application/json' \
  -d '{"approved_by":"manager","approved":true,"decision_note":"approved"}'

curl -X POST http://localhost:8080/api/v1/executions/start \
  -H 'Content-Type: application/json' \
  -d '{"plan_id":"<plan_id>","max_attempts":3}'
```

Drift:

```bash
curl -X POST http://localhost:8080/api/v1/drift/detect \
  -H 'Content-Type: application/json' \
  -d '{"environment_id":"<env_id>"}'
```

## gRPC

The gRPC endpoint uses the JSON codec to keep this repository self-contained without
generated protobuf files. The service name is:

```text
infra.controlplane.v1.ControlPlane
```

Methods include `Health`, `ListProjects`, and `GeneratePlan`. Use a gRPC client with
the `json` content subtype.

## CLI

```bash
go run ./cmd/cli -server http://localhost:8080 health
go run ./cmd/cli -server http://localhost:8080 projects list
go run ./cmd/cli -server http://localhost:8080 projects create platform
go run ./cmd/cli -server http://localhost:8080 resources upsert <env_id> cache redis '{"memory_mb":512}'
go run ./cmd/cli -server http://localhost:8080 plans generate <env_id> alice
```

## Declarations

Example YAML, JSON, and HCL declarations are in `examples/`. The resource
application package can parse all three formats and applies static validation such
as required fields, duplicate names, and valid JSON state.

## Runner

The standalone runner is a simple remote executor:

```bash
go run ./cmd/runner -addr :7070
```

The production control plane can be pointed at a remote runner by replacing the
`internal/execution/infrastructure.MockRunner` with
`internal/execution/infrastructure.HTTPRunner` in the composition root.

## Verification

```bash
docker compose up -d
go build ./...
go run ./cmd/server -config=configs/config.yaml
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/readyz
curl -s http://localhost:8080/metrics | head
```

Then use the REST or CLI examples to create a project, environment, resource, plan,
approval, and execution task.
