package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/infra-orchestration/controlplane/internal/execution/domain"
)

type HTTPRunner struct {
	client *http.Client
	url    string
}

func NewHTTPRunner(url string) *HTTPRunner {
	return &HTTPRunner{url: url, client: &http.Client{Timeout: 30 * time.Second}}
}

func (r *HTTPRunner) Execute(ctx context.Context, request domain.RunnerRequest) (domain.RunnerResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return domain.RunnerResponse{}, fmt.Errorf("marshal runner request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(payload))
	if err != nil {
		return domain.RunnerResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return domain.RunnerResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.RunnerResponse{}, fmt.Errorf("remote runner returned %s", resp.Status)
	}
	var out domain.RunnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return domain.RunnerResponse{}, err
	}
	return out, nil
}
