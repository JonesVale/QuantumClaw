package rigetti

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quantumclaw/quantumclaw/relay/quantum"
)

// Adaptor — Rigetti quantum platform
type Adaptor struct {
	APIKey  string
	BaseURL string
	client  *http.Client
}

func (a *Adaptor) ProviderName() string { return "rigetti" }

func (a *Adaptor) getClient() *http.Client {
	if a.client == nil {
		a.client = &http.Client{Timeout: 60 * time.Second}
	}
	return a.client
}

func safeCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (a *Adaptor) RunTask(ctx context.Context, req *quantum.QuantumTaskRequest) (*quantum.QuantumTaskResult, error) {
	payload := map[string]any{
		"program":     req.Circuit.QASM,
		"backend":     req.Backend,
		"shot_count":  req.Shots,
	}
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "POST", a.BaseURL+"/v1/jobs", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("rigetti: create: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rigetti: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rigetti: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var job struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&job)

	status := "queued"
	if job.Status != "" {
		status = mapRigettiStatus(job.Status)
	}

	return &quantum.QuantumTaskResult{
		TaskID:  job.ID,
		Status:  status,
		Backend: req.Backend,
	}, nil
}

func (a *Adaptor) QueryTask(ctx context.Context, taskID string) (*quantum.QuantumTaskResult, error) {
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "GET", a.BaseURL+"/v1/jobs/"+taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("rigetti: create: %w", err)
	}
	httpReq.Header.Set("X-API-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rigetti: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rigetti: HTTP %d", resp.StatusCode)
	}

	var job struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&job)

	return &quantum.QuantumTaskResult{
		TaskID: taskID,
		Status: mapRigettiStatus(job.Status),
	}, nil
}

func (a *Adaptor) CancelTask(ctx context.Context, taskID string) error {
	body := bytes.NewReader([]byte(`{"status":"cancelled"}`))
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "PATCH", a.BaseURL+"/v1/jobs/"+taskID, body)
	if err != nil {
		return fmt.Errorf("rigetti: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return fmt.Errorf("rigetti: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rigetti: cancel HTTP %d", resp.StatusCode)
	}
	return nil
}

func (a *Adaptor) ListBackends(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "GET", a.BaseURL+"/v1/backends", nil)
	if err != nil {
		return nil, fmt.Errorf("rigetti: %w", err)
	}
	httpReq.Header.Set("X-API-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return []string{"rigetti_aspem", "rigetti_ankaa", "rigetti_qvm"}, nil
	}
	defer resp.Body.Close()

	var backends []string
	json.NewDecoder(resp.Body).Decode(&backends)
	if len(backends) == 0 {
		return []string{"rigetti_aspem", "rigetti_ankaa"}, nil
	}
	return backends, nil
}

func mapRigettiStatus(s string) string {
	switch s {
	case "queued":
		return "queued"
	case "running":
		return "running"
	case "completed", "finished":
		return "completed"
	case "failed":
		return "failed"
	case "cancelled":
		return "cancelled"
	default:
		return s
	}
}


