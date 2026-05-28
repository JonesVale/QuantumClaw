package braket

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

// Adaptor — AWS Braket quantum platform
type Adaptor struct {
	APIKey  string
	BaseURL string
	Region  string
	client  *http.Client
}

func (a *Adaptor) ProviderName() string { return "aws_braket" }

func (a *Adaptor) getClient() *http.Client {
	if a.client == nil {
		a.client = &http.Client{Timeout: 60 * time.Second}
	}
	return a.client
}

func (a *Adaptor) apiURL(path string) string {
	region := a.Region
	if region == "" {
		region = "us-west-1"
	}
	base := a.BaseURL
	if base == "" {
		base = "https://braket." + region + ".amazonaws.com"
	}
	return base + path
}

func safeCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (a *Adaptor) RunTask(ctx context.Context, req *quantum.QuantumTaskRequest) (*quantum.QuantumTaskResult, error) {
	payload := map[string]any{
		"deviceArn":      req.Backend,
		"shots":          req.Shots,
		"outputS3Bucket": "quantumclaw-output",
		"outputS3Key":    "braket/" + fmt.Sprint(time.Now().Unix()),
		"action": map[string]any{
			"braketSchemaHeader": map[string]string{"schemaVersion": "1.0"},
			"program":           req.Circuit.QASM,
		},
	}
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "POST", a.apiURL("/quantumTasks"), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("aws_braket: create: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("aws_braket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("aws_braket: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var task struct {
		TaskArn  string `json:"quantumTaskArn"`
		Status   string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&task)

	status := "queued"
	if task.Status != "" {
		status = mapBraketStatus(task.Status)
	}

	return &quantum.QuantumTaskResult{
		TaskID:  task.TaskArn,
		Status:  status,
		Backend: req.Backend,
	}, nil
}

func (a *Adaptor) QueryTask(ctx context.Context, taskID string) (*quantum.QuantumTaskResult, error) {
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "GET", a.apiURL("/quantumTasks/"+taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("aws_braket: create: %w", err)
	}
	httpReq.Header.Set("X-Api-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("aws_braket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aws_braket: HTTP %d", resp.StatusCode)
	}

	var task struct {
		Status     string `json:"status"`
		DeviceArn  string `json:"deviceArn"`
	}
	json.NewDecoder(resp.Body).Decode(&task)

	return &quantum.QuantumTaskResult{
		TaskID:  taskID,
		Status:  mapBraketStatus(task.Status),
		Backend: task.DeviceArn,
	}, nil
}

func (a *Adaptor) CancelTask(ctx context.Context, taskID string) error {
	body := bytes.NewReader([]byte(`{}`))
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "PUT", a.apiURL("/quantumTasks/"+taskID+"/cancel"), body)
	if err != nil {
		return fmt.Errorf("aws_braket: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return fmt.Errorf("aws_braket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("aws_braket: cancel HTTP %d", resp.StatusCode)
	}
	return nil
}

func (a *Adaptor) ListBackends(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(safeCtx(ctx), "GET", a.apiURL("/devices"), nil)
	if err != nil {
		return nil, fmt.Errorf("aws_braket: %w", err)
	}
	httpReq.Header.Set("X-Api-Key", a.APIKey)

	resp, err := a.getClient().Do(httpReq)
	if err != nil {
		return []string{"braket_sv1", "braket_dm1", "braket_rigetti", "braket_ionq"}, nil
	}
	defer resp.Body.Close()

	var devices []struct {
		DeviceName string `json:"deviceName"`
	}
	json.NewDecoder(resp.Body).Decode(&devices)
	if len(devices) == 0 {
		return []string{"braket_sv1", "braket_dm1"}, nil
	}
	names := make([]string, 0, len(devices))
	for _, d := range devices {
		names = append(names, d.DeviceName)
	}
	return names, nil
}

func mapBraketStatus(s string) string {
	switch s {
	case "CREATED", "QUEUED":
		return "queued"
	case "RUNNING":
		return "running"
	case "COMPLETED":
		return "completed"
	case "FAILED":
		return "failed"
	case "CANCELLED":
		return "cancelled"
	default:
		return s
	}
}

