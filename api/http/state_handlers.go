package httpapi

import (
	"encoding/json"
	"net/http"

	statedomain "github.com/infra-orchestration/controlplane/internal/state/domain"
)

func (h *handler) listStates(w http.ResponseWriter, r *http.Request) {
	states, err := h.deps.State.List(r.Context(), r.URL.Query().Get("environment_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, states)
}

func (h *handler) getState(w http.ResponseWriter, r *http.Request) {
	state, err := h.deps.State.Get(r.Context(), r.PathValue("resource_id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *handler) reportActual(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ActualState json.RawMessage             `json:"actual_state"`
		Status      statedomain.ExecutionStatus `json:"status"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	state, err := h.deps.State.ReportActual(r.Context(), r.PathValue("resource_id"), input.ActualState, input.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *handler) snapshotState(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.deps.State.Snapshot(r.Context(), r.PathValue("resource_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, snapshot)
}

func (h *handler) rollbackState(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SnapshotID string `json:"snapshot_id"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	state, err := h.deps.State.Rollback(r.Context(), r.PathValue("resource_id"), input.SnapshotID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}
