package quantum

import (
	"testing"
)

func TestQuantumTaskRequestDefaults(t *testing.T) {
	req := &QuantumTaskRequest{
		Shots: 1000,
		Circuit: Circuit{
			Qubits: 2,
			Gates: []Gate{
				{Name: "h", Targets: []int{0}},
				{Name: "cx", Control: []int{0}, Targets: []int{1}},
				{Name: "measure", Targets: []int{0}},
			},
		},
	}

	if req.Shots != 1000 {
		t.Errorf("Shots = %d, want 1000", req.Shots)
	}
	if req.Circuit.Qubits != 2 {
		t.Errorf("Qubits = %d, want 2", req.Circuit.Qubits)
	}
	if len(req.Circuit.Gates) != 3 {
		t.Errorf("Gates count = %d, want 3", len(req.Circuit.Gates))
	}
}

func TestQuantumTaskResultDefaults(t *testing.T) {
	result := &QuantumTaskResult{
		TaskID:   "qt-test-123",
		Status:   "completed",
		Provider: "ionq",
		Backend:  "ionq_harmony",
		Results: &TaskResults{
			Counts: map[string]int{"00": 498, "11": 502},
			Shots:  1000,
		},
		CostQuota: 5000,
	}

	if result.TaskID != "qt-test-123" {
		t.Errorf("TaskID = %q, want qt-test-123", result.TaskID)
	}
	if result.Results.Shots != 1000 {
		t.Errorf("Shots = %d, want 1000", result.Results.Shots)
	}
	if result.Results.Counts["00"] != 498 {
		t.Errorf("Counts[00] = %d, want 498", result.Results.Counts["00"])
	}
}

func TestGateStructure(t *testing.T) {
	gate := Gate{
		Name:    "rx",
		Targets: []int{0},
		Params:  []float64{3.14159},
	}

	if gate.Name != "rx" {
		t.Errorf("Gate.Name = %q, want rx", gate.Name)
	}
	if len(gate.Params) != 1 || gate.Params[0] != 3.14159 {
		t.Errorf("Gate.Params = %v, want [3.14159]", gate.Params)
	}
}

func TestCircuitQASM(t *testing.T) {
	// Verify QASM takes priority over gates
	circuit := Circuit{
		Qubits: 2,
		QASM:   "OPENQASM 2.0;\ninclude \"qelib1.inc\";\nqreg q[2];\ncreg c[2];\nh q[0];\ncx q[0],q[1];\nmeasure q[0] -> c[0];\nmeasure q[1] -> c[1];\n",
		Gates: []Gate{
			{Name: "h", Targets: []int{0}},
		},
	}

	if circuit.QASM == "" {
		t.Error("QASM should not be empty")
	}
	if circuit.Qubits != 2 {
		t.Errorf("Qubits = %d, want 2", circuit.Qubits)
	}
}

func TestEmptyResults(t *testing.T) {
	result := &QuantumTaskResult{
		TaskID: "qt-empty",
		Status: "failed",
		Error:  "circuit error",
	}

	if result.Error != "circuit error" {
		t.Errorf("Error = %q, want circuit error", result.Error)
	}
	if result.Results != nil {
		t.Error("Results should be nil for failed task")
	}
}
