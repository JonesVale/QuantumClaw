package ibmq

import (
	"encoding/json"
	"testing"

	"github.com/quantumclaw/quantumclaw/relay/quantum"
)

func TestCircuitToQASM_WithGates(t *testing.T) {
	circuit := &quantum.Circuit{
		Qubits: 2,
		Gates: []quantum.Gate{
			{Name: "h", Targets: []int{0}},
			{Name: "cx", Control: []int{0}, Targets: []int{1}},
			{Name: "rx", Targets: []int{0}, Params: []float64{1.5708}},
			{Name: "measure", Targets: []int{0}},
		},
	}

	qasm, err := CircuitToQASM(circuit)
	if err != nil {
		t.Fatalf("CircuitToQASM failed: %v", err)
	}

	if qasm == "" {
		t.Fatal("QASM string is empty")
	}

	// Check basic QASM structure
	expectedParts := []string{
		"OPENQASM 2.0",
		"qreg q[2]",
		"creg c[2]",
		"h q[0]",
		"cx q[0], q[1]",
		"measure q[0] -> c[0]",
		"measure q[1] -> c[1]",
	}
	for _, part := range expectedParts {
		if !contains(qasm, part) {
			t.Errorf("QASM missing expected content: %s", part)
		}
	}
}

func TestCircuitToQASM_WithQASMInput(t *testing.T) {
	inputQASM := "OPENQASM 2.0; qreg q[2]; h q[0];"
	circuit := &quantum.Circuit{
		Qubits: 2,
		QASM:   inputQASM,
		// Gates should be ignored when QASM is present
		Gates: []quantum.Gate{{Name: "x", Targets: []int{0}}},
	}

	qasm, err := CircuitToQASM(circuit)
	if err != nil {
		t.Fatalf("CircuitToQASM failed: %v", err)
	}

	if qasm != inputQASM {
		t.Errorf("QASM = %q, want input unchanged: %q", qasm, inputQASM)
	}
}

func TestCircuitToQASM_NilCircuit(t *testing.T) {
	_, err := CircuitToQASM(nil)
	if err == nil {
		t.Error("Expected error for nil circuit, got nil")
	}
}

func TestCircuitToQASM_EmptyCircuit(t *testing.T) {
	circuit := &quantum.Circuit{Qubits: 3} // no gates, no qasm
	_, err := CircuitToQASM(circuit)
	if err == nil {
		t.Error("Expected error for empty circuit, got nil")
	}
}

func TestParseIBMQResponse_Completed(t *testing.T) {
	jsonData := `{
		"id": "ibm-job-789",
		"status": "COMPLETED",
		"result": {"00": 498, "11": 502},
		"execution_time_ms": 5678
	}`

	result, err := ParseIBMQResponse([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseIBMQResponse failed: %v", err)
	}

	if result.TaskID != "ibm-job-789" {
		t.Errorf("TaskID = %q, want ibm-job-789", result.TaskID)
	}
	if result.Status != "completed" {
		t.Errorf("Status = %q, want completed", result.Status)
	}
	if result.ExecTimeMs != 5678 {
		t.Errorf("ExecTimeMs = %d, want 5678", result.ExecTimeMs)
	}
	if result.Results.Shots != 1000 {
		t.Errorf("Results.Shots = %d, want 1000", result.Results.Shots)
	}
	if result.Results.Counts["00"] != 498 {
		t.Errorf("Counts[00] = %d, want 498", result.Results.Counts["00"])
	}
}

func TestParseIBMQResponse_Failed(t *testing.T) {
	jsonData := `{
		"id": "ibm-fail-abc",
		"status": "FAILED",
		"error": "Qubit not available"
	}`

	result, err := ParseIBMQResponse([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseIBMQResponse failed: %v", err)
	}

	if result.Status != "failed" {
		t.Errorf("Status = %q, want failed", result.Status)
	}
	if result.Error != "Qubit not available" {
		t.Errorf("Error = %q, want Qubit not available", result.Error)
	}
}

func TestParseIBMQResponse_JSONRoundTrip(t *testing.T) {
	// Verify that the result can be serialized back to JSON (no unmarshalable types)
	result := &quantum.QuantumTaskResult{
		TaskID:   "roundtrip-test",
		Status:   "completed",
		Provider: "ibmq",
		Results: &quantum.TaskResults{
			Counts: map[string]int{"000": 125, "001": 375},
			Shots:  500,
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded quantum.QuantumTaskResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.TaskID != result.TaskID {
		t.Errorf("TaskID = %q, want %q", decoded.TaskID, result.TaskID)
	}
	if decoded.Results.Shots != 500 {
		t.Errorf("Shots = %d, want 500", decoded.Results.Shots)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
