package azure

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

// Adaptor — Azure Quantum platform
type Adaptor struct {
	APIKey  string
	BaseURL string
	client  *http.Client
}

func (a *Adaptor) ProviderName() string { return "azure_quantum" }

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
	provider := "ionq"
	target := req.Backend
	if target == "" {
		target = "ionq_simulator"
	}

	payload := map[string]any{
		"name":       "quantumclaw-" + fmt.Sprint(time.Now().Unix()),
		"providerId": provider,
		"target":     target,
		"input": map[string]string{
			"format": "qir.v1",
			"data":   req.Circuit.QASM,
		},
		"shots": req.Shots,
	}
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "POST", a.BaseURL+"/jobs", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("azure: create: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("azure: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var job struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&job)

	return &quantum.QuantumTaskResult{
		TaskID:  job.ID,
		Status:  mapAzureStatus(job.Status),
		Backend: target,
	}, nil
}

func (a *Adaptor) QueryTask(ctx context.Context, taskID string) (*quantum.QuantumTaskResult, error) {
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "GET", a.BaseURL+"/jobs/"+taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("azure: create: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("azure: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("azure: HTTP %d", resp.StatusCode)
	}

	var job struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Target string `json:"target"`
	}
	json.NewDecoder(resp.Body).Decode(&job)

	return &quantum.QuantumTaskResult{
		TaskID:  taskID,
		Status:  mapAzureStatus(job.Status),
		Backend: job.Target,
	}, nil
}

func (a *Adaptor) CancelTask(ctx context.Context, taskID string) error {
	body := bytes.NewReader([]byte(`{"status":"cancelling"}`))
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "PUT", a.BaseURL+"/jobs/"+taskID, body)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return fmt.Errorf("azure: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("azure: cancel HTTP %d", resp.StatusCode)
	}
	return nil
}

func (a *Adaptor) ListBackends(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "GET", a.BaseURL+"/providers", nil)
	if err != nil {
		return nil, fmt.Errorf("azure: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return []string{"azure_quantinium", "azure_honeywell", "azure_ionq"}, nil
	}
	defer resp.Body.Close()

	var result struct {
		Providers []struct {
			ID      string   `json:"id"`
			Targets []string `json:"targets"`
		} `json:"providers"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	backends := make([]string, 0)
	for _, p := range result.Providers {
		for _, t := range p.Targets {
			backends = append(backends, p.ID+":"+t)
		}
	}
	if len(backends) == 0 {
		return []string{"azure_quantinium", "azure_honeywell"}, nil
	}
	return backends, nil
}

func mapAzureStatus(s string) string {
	switch s {
	case "Queued", "Submitted", "Waiting":
		return "queued"
	case "Running":
		return "running"
	case "Succeeded", "Completed":
		return "completed"
	case "Failed":
		return "failed"
	case "Cancelled":
		return "cancelled"
	default:
		return s
	}
}

