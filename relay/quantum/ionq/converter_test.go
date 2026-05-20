package ionq

import (
	"encoding/json"
	"testing"

	"github.com/quantumclaw/quantumclaw/relay/quantum"
)

func TestConvertToIonQRequest_WithGates(t *testing.T) {
	req := &quantum.QuantumTaskRequest{
		Backend: "ionq_harmony",
		Shots:   1000,
		Circuit: quantum.Circuit{
			Qubits: 2,
			Gates: []quantum.Gate{
				{Name: "h", Targets: []int{0}},
				{Name: "cx", Control: []int{0}, Targets: []int{1}},
			},
		},
	}

	ionQReq, err := ConvertToIonQRequest(req)
	if err != nil {
		t.Fatalf("ConvertToIonQRequest failed: %v", err)
	}

	if ionQReq.Target != "ionq_harmony" {
		t.Errorf("Target = %q, want ionq_harmony", ionQReq.Target)
	}
	if ionQReq.Shots != 1000 {
		t.Errorf("Shots = %d, want 1000", ionQReq.Shots)
	}
	if ionQReq.Circuit == nil {
		t.Fatal("Circuit is nil")
	}

	// Verify JSON serialization
	data, err := json.Marshal(ionQReq)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if decoded["target"] != "ionq_harmony" {
		t.Errorf("JSON target = %v, want ionq_harmony", decoded["target"])
	}
}

func TestConvertToIonQRequest_WithQASM(t *testing.T) {
	req := &quantum.QuantumTaskRequest{
		Shots: 1000,
		Circuit: quantum.Circuit{
			Qubits: 2,
			QASM:   "OPENQASM 2.0; qreg q[2]; h q[0]; cx q[0],q[1];",
		},
	}

	ionQReq, err := ConvertToIonQRequest(req)
	if err != nil {
		t.Fatalf("ConvertToIonQRequest failed: %v", err)
	}

	if ionQReq.Circuit == nil {
		t.Fatal("Circuit is nil for QASM input")
	}
}

func TestConvertToIonQRequest_EmptyCircuit(t *testing.T) {
	req := &quantum.QuantumTaskRequest{
		Shots:   1000,
		Circuit: quantum.Circuit{Qubits: 0}, // no gates, no qasm
	}

	_, err := ConvertToIonQRequest(req)
	if err == nil {
		t.Error("Expected error for empty circuit, got nil")
	}
}

func TestConvertToIonQRequest_AutoBackend(t *testing.T) {
	req := &quantum.QuantumTaskRequest{
		Backend: "auto",
		Shots:   500,
		Circuit: quantum.Circuit{
			Qubits: 1,
			Gates:  []quantum.Gate{{Name: "h", Targets: []int{0}}},
		},
	}

	ionQReq, err := ConvertToIonQRequest(req)
	if err != nil {
		t.Fatalf("ConvertToIonQRequest failed: %v", err)
	}

	if ionQReq.Target == "auto" {
		t.Error("Target should not be 'auto', should default to qpu.harmony")
	}
	if ionQReq.Shots != 500 {
		t.Errorf("Shots = %d, want 500", ionQReq.Shots)
	}
}

func TestConvertToIonQRequest_ZeroShots(t *testing.T) {
	req := &quantum.QuantumTaskRequest{
		Shots: 0,
		Circuit: quantum.Circuit{
			Qubits: 2,
			Gates:  []quantum.Gate{{Name: "h", Targets: []int{0}}},
		},
	}

	ionQReq, err := ConvertToIonQRequest(req)
	if err != nil {
		t.Fatalf("ConvertToIonQRequest failed: %v", err)
	}

	if ionQReq.Shots != 1000 {
		t.Errorf("Shots = %d, want 1000 (default)", ionQReq.Shots)
	}
}

func TestParseIonQResponse_Completed(t *testing.T) {
	jsonData := `{
		"id": "job-test-123",
		"status": "completed",
		"name": "ionq_harmony",
		"result": {
			"histogram": {"00": 0.498, "11": 0.502}
		},
		"cost": {
			"exec_time_ms": 1234,
			"cost": 0.005
		}
	}`

	result, err := ParseIonQResponse([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseIonQResponse failed: %v", err)
	}

	if result.TaskID != "job-test-123" {
		t.Errorf("TaskID = %q, want job-test-123", result.TaskID)
	}
	if result.Status != "completed" {
		t.Errorf("Status = %q, want completed", result.Status)
	}
	if result.Backend != "ionq_harmony" {
		t.Errorf("Backend = %q, want ionq_harmony", result.Backend)
	}
	if result.Results == nil {
		t.Fatal("Results is nil")
	}
	if result.Results.Shots == 0 {
		t.Error("Shots should not be 0")
	}
	if result.CostQuota <= 0 {
		t.Errorf("CostQuota = %d, want > 0", result.CostQuota)
	}
}

func TestParseIonQResponse_Failed(t *testing.T) {
	jsonData := `{
		"id": "job-fail-456",
		"status": "failed",
		"failure": {
			"code": "CIRCUIT_ERROR",
			"error": "Invalid gate sequence"
		}
	}`

	result, err := ParseIonQResponse([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseIonQResponse failed: %v", err)
	}

	if result.Status != "failed" {
		t.Errorf("Status = %q, want failed", result.Status)
	}
	if result.Error != "Invalid gate sequence" {
		t.Errorf("Error = %q, want Invalid gate sequence", result.Error)
	}
	if result.Results != nil {
		t.Error("Results should be nil for failed task")
	}
}

func TestParseIonQResponse_InvalidJSON(t *testing.T) {
	_, err := ParseIonQResponse([]byte(`{invalid json`))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
