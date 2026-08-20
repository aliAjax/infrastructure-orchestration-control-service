package main

import (
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/infra-orchestration/controlplane/internal/execution/domain"
)

func main() {
	addr := flag.String("addr", ":7070", "runner HTTP listen address")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /execute", func(w http.ResponseWriter, r *http.Request) {
		var request domain.RunnerRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}
		time.Sleep(10 * time.Millisecond)
		output, _ := json.Marshal(map[string]any{"resource_id": request.ResourceID, "type": request.Type, "result": "remote-ok", "ts": time.Now().UTC()})
		writeJSON(w, http.StatusOK, domain.RunnerResponse{Output: output})
	})
	logger.Info("remote runner started", "addr", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		logger.Error("runner stopped", "error", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
