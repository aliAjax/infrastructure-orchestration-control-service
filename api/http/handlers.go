package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	approvaldomain "github.com/infra-orchestration/controlplane/internal/approval/domain"
	auditdomain "github.com/infra-orchestration/controlplane/internal/audit/domain"
	driftdomain "github.com/infra-orchestration/controlplane/internal/drift/domain"
	executiondomain "github.com/infra-orchestration/controlplane/internal/execution/domain"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
	"github.com/infra-orchestration/controlplane/internal/resource/domain"
	statedomain "github.com/infra-orchestration/controlplane/internal/state/domain"
)

type handler struct {
	deps   Dependencies
	logger *slog.Logger
}

func newHandler(deps Dependencies) *handler {
	return &handler{deps: deps, logger: deps.Logger}
}

func (h *handler) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /readyz", h.readyz)
	mux.HandleFunc("GET /metrics", h.metrics)

	mux.HandleFunc("POST /api/v1/projects", h.createProject)
	mux.HandleFunc("GET /api/v1/projects", h.listProjects)
	mux.HandleFunc("GET /api/v1/projects/{id}", h.getProject)
	mux.HandleFunc("PUT /api/v1/projects/{id}", h.updateProject)

	mux.HandleFunc("POST /api/v1/environments", h.createEnvironment)
	mux.HandleFunc("GET /api/v1/environments", h.listEnvironments)
	mux.HandleFunc("GET /api/v1/environments/{id}", h.getEnvironment)

	mux.HandleFunc("POST /api/v1/resources", h.upsertResource)
	mux.HandleFunc("GET /api/v1/resources", h.listResources)
	mux.HandleFunc("GET /api/v1/resources/{id}", h.getResource)

	mux.HandleFunc("POST /api/v1/variables", h.upsertVariable)
	mux.HandleFunc("GET /api/v1/variables", h.listVariables)

	mux.HandleFunc("GET /api/v1/states", h.listStates)
	mux.HandleFunc("GET /api/v1/states/{resource_id}", h.getState)
	mux.HandleFunc("POST /api/v1/states/{resource_id}/actual", h.reportActual)
	mux.HandleFunc("POST /api/v1/states/{resource_id}/snapshot", h.snapshotState)
	mux.HandleFunc("POST /api/v1/states/{resource_id}/rollback", h.rollbackState)

	mux.HandleFunc("POST /api/v1/plans/generate", h.generatePlan)
	mux.HandleFunc("GET /api/v1/plans", h.listPlans)
	mux.HandleFunc("GET /api/v1/plans/{id}", h.getPlan)
	mux.HandleFunc("GET /api/v1/plans/{id}/diff", h.getPlanDiff)

	mux.HandleFunc("POST /api/v1/approvals", h.createApproval)
	mux.HandleFunc("POST /api/v1/approvals/{id}/decision", h.decideApproval)
	mux.HandleFunc("GET /api/v1/approvals", h.listApprovals)

	mux.HandleFunc("POST /api/v1/executions/start", h.startExecution)
	mux.HandleFunc("GET /api/v1/executions", h.listExecutions)
	mux.HandleFunc("POST /api/v1/executions/{id}/retry", h.retryExecution)
	mux.HandleFunc("POST /api/v1/executions/{id}/cancel", h.cancelExecution)

	mux.HandleFunc("POST /api/v1/drift/detect", h.detectDrift)
	mux.HandleFunc("GET /api/v1/drift", h.listDrift)
	mux.HandleFunc("POST /api/v1/drift/remediate", h.remediateDrift)

	mux.HandleFunc("GET /api/v1/audit", h.listAudit)
	mux.HandleFunc("GET /api/v1/graph/{environment_id}", h.graph)
	return mux
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.routes().ServeHTTP(w, r)
}

func (h *handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.deps.ReadyCheck != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.deps.ReadyCheck(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *handler) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	if h.deps.Metrics == nil {
		http.Error(w, "metrics unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Write(h.deps.Metrics.Prometheus())
}

func (h *handler) createProject(w http.ResponseWriter, r *http.Request) {
	var input domain.ProjectInput
	if !readJSON(w, r, &input) {
		return
	}
	project, err := h.deps.Resource.CreateProject(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.recordAudit(r, auditdomain.EventPlanCreated, "project", project.ID, project)
	writeJSON(w, http.StatusCreated, project)
}

func (h *handler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.deps.Resource.ListProjects(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *handler) getProject(w http.ResponseWriter, r *http.Request) {
	project, err := h.deps.Resource.GetProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *handler) updateProject(w http.ResponseWriter, r *http.Request) {
	var input domain.ProjectInput
	if !readJSON(w, r, &input) {
		return
	}
	project, err := h.deps.Resource.UpdateProject(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *handler) createEnvironment(w http.ResponseWriter, r *http.Request) {
	var input domain.EnvironmentInput
	if !readJSON(w, r, &input) {
		return
	}
	env, err := h.deps.Resource.CreateEnvironment(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, env)
}

func (h *handler) listEnvironments(w http.ResponseWriter, r *http.Request) {
	envs, err := h.deps.Resource.ListEnvironments(r.Context(), r.URL.Query().Get("project_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, envs)
}

func (h *handler) getEnvironment(w http.ResponseWriter, r *http.Request) {
	env, err := h.deps.Resource.GetEnvironment(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, env)
}

func (h *handler) upsertResource(w http.ResponseWriter, r *http.Request) {
	var input domain.ResourceInput
	if !readJSON(w, r, &input) {
		return
	}
	res, err := h.deps.Resource.UpsertResource(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	_, _ = h.deps.State.ApplyDesired(r.Context(), res.ID, res.DesiredState)
	writeJSON(w, http.StatusOK, res)
}

func (h *handler) listResources(w http.ResponseWriter, r *http.Request) {
	res, err := h.deps.Resource.ListResources(r.Context(), r.URL.Query().Get("environment_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *handler) getResource(w http.ResponseWriter, r *http.Request) {
	res, err := h.deps.Resource.GetResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *handler) upsertVariable(w http.ResponseWriter, r *http.Request) {
	var input domain.VariableInput
	if !readJSON(w, r, &input) {
		return
	}
	variable, err := h.deps.Resource.UpsertVariable(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, variable)
}

func (h *handler) listVariables(w http.ResponseWriter, r *http.Request) {
	variables, err := h.deps.Resource.ListVariables(r.Context(), r.URL.Query().Get("project_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, variables)
}

func (h *handler) generatePlan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EnvironmentID string `json:"environment_id"`
		CreatedBy     string `json:"created_by"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	if input.CreatedBy == "" {
		input.CreatedBy = r.Header.Get("X-Actor")
	}
	plan, err := h.deps.Plan.Generate(r.Context(), plandomain.GenerateInput{EnvironmentID: input.EnvironmentID, CreatedBy: input.CreatedBy})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.ensureApprovalForPlan(r, plan)
	writeJSON(w, http.StatusCreated, plan)
}

func (h *handler) ensureApprovalForPlan(r *http.Request, plan plandomain.Plan) {
	items, err := h.deps.Plan.DiffItems(r.Context(), plan.ID)
	if err != nil || len(items) == 0 {
		return
	}
	env, err := h.deps.Resource.GetEnvironment(r.Context(), plan.EnvironmentID)
	if err != nil {
		return
	}
	requires := env.Production
	for _, item := range items {
		if item.Operation == plandomain.OperationDelete {
			requires = true
		}
	}
	resources, _ := h.deps.Resource.ListResources(r.Context(), plan.EnvironmentID)
	resourceMap := make(map[string]domain.Resource, len(resources))
	for _, res := range resources {
		resourceMap[res.ID] = res
	}
	for _, item := range items {
		if res, ok := resourceMap[item.ResourceID]; ok && (res.Sensitive || res.ApprovalRequired) {
			requires = true
		}
	}
	if !requires {
		return
	}
	_, _ = h.deps.Approval.Create(r.Context(), approvaldomain.CreateInput{
		PlanID:        plan.ID,
		EnvironmentID: plan.EnvironmentID,
		RequestedBy:   actorFromRequest(r),
		Reason:        "plan requires approval by policy",
	})
}

func (h *handler) listPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.deps.Plan.List(r.Context(), r.URL.Query().Get("environment_id"), queryInt(r, "limit", 50))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

func (h *handler) getPlan(w http.ResponseWriter, r *http.Request) {
	plan, err := h.deps.Plan.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (h *handler) getPlanDiff(w http.ResponseWriter, r *http.Request) {
	items, err := h.deps.Plan.DiffItems(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *handler) createApproval(w http.ResponseWriter, r *http.Request) {
	var input approvaldomain.CreateInput
	if !readJSON(w, r, &input) {
		return
	}
	request, err := h.deps.Approval.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, request)
}

func (h *handler) decideApproval(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ApprovedBy   string `json:"approved_by"`
		Approved     bool   `json:"approved"`
		DecisionNote string `json:"decision_note"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	request, err := h.deps.Approval.Decide(r.Context(), approvaldomain.DecisionInput{
		ID:           r.PathValue("id"),
		ApprovedBy:   input.ApprovedBy,
		Approved:     input.Approved,
		DecisionNote: input.DecisionNote,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Approved {
		_, _ = h.deps.Plan.UpdateStatus(r.Context(), request.PlanID, plandomain.StatusApproved)
	} else {
		_, _ = h.deps.Plan.UpdateStatus(r.Context(), request.PlanID, plandomain.StatusRejected)
	}
	writeJSON(w, http.StatusOK, request)
}

func (h *handler) listApprovals(w http.ResponseWriter, r *http.Request) {
	requests, err := h.deps.Approval.ForPlan(r.Context(), r.URL.Query().Get("plan_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, requests)
}

func (h *handler) startExecution(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PlanID      string `json:"plan_id"`
		MaxAttempts int    `json:"max_attempts"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	if err := h.deps.Approval.EnsureApproved(r.Context(), input.PlanID); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	tasks, err := h.deps.Execution.StartPlan(r.Context(), input.PlanID, input.MaxAttempts)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, tasks)
}

func (h *handler) listExecutions(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.deps.Execution.List(r.Context(), r.URL.Query().Get("plan_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *handler) retryExecution(w http.ResponseWriter, r *http.Request) {
	task, err := h.deps.Execution.Retry(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *handler) cancelExecution(w http.ResponseWriter, r *http.Request) {
	task, err := h.deps.Execution.Cancel(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *handler) detectDrift(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EnvironmentID string `json:"environment_id"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	records, err := h.deps.Drift.Detect(r.Context(), input.EnvironmentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (h *handler) listDrift(w http.ResponseWriter, r *http.Request) {
	records, err := h.deps.Drift.List(r.Context(), r.URL.Query().Get("environment_id"), r.URL.Query().Get("unresolved") == "true")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (h *handler) remediateDrift(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EnvironmentID string `json:"environment_id"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	plan, err := h.deps.Drift.Remediate(r.Context(), input.EnvironmentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, plan)
}

func (h *handler) listAudit(w http.ResponseWriter, r *http.Request) {
	events, err := h.deps.Audit.List(r.Context(), r.URL.Query().Get("entity_type"), r.URL.Query().Get("entity_id"), queryInt(r, "limit", 50))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h *handler) graph(w http.ResponseWriter, r *http.Request) {
	envID := r.PathValue("environment_id")
	resources, err := h.deps.Resource.ListResources(r.Context(), envID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	levels, err := h.deps.Graph.Levels(r.Context(), resources)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"environment_id": envID, "levels": levels})
}

func (h *handler) recordAudit(r *http.Request, eventType auditdomain.EventType, entityType, entityID string, content any) {
	raw, _ := json.Marshal(content)
	_, _ = h.deps.Audit.Record(r.Context(), auditdomain.CreateInput{
		Type:       eventType,
		Actor:      actorFromRequest(r),
		EntityType: entityType,
		EntityID:   entityID,
		Content:    raw,
		Result:     "success",
	})
}

func actorFromRequest(r *http.Request) string {
	if actor := r.Header.Get("X-Actor"); actor != "" {
		return actor
	}
	return "anonymous"
}

func readJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

var _ = fmt.Sprintf
var _ = statedomain.StatusPending
var _ = executiondomain.StatusPending
var _ = driftdomain.DriftRecord{}
