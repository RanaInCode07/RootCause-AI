package hypothesis

import (
	"context"
	"errors"
	"testing"
	"time"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestDeterministicGenerator_GenerateHypotheses_GeneratesMemoryLeakAfterDeployment(t *testing.T) {
	var _ investigation.HypothesisGenerator = DeterministicGenerator{}

	generator := NewDeterministicGenerator()
	evidence := []investigation.Evidence{
		evidenceWithSignal("evidence-deployment", investigation.EvidenceSignalRecentDeployment),
		evidenceWithSignal("evidence-memory", investigation.EvidenceSignalMemoryIncrease),
		evidenceWithSignal("evidence-oom", investigation.EvidenceSignalOOMKilled),
		evidenceWithSignal("evidence-restarts", investigation.EvidenceSignalRestartCountIncrease),
		evidenceWithSignal("evidence-5xx", investigation.EvidenceSignalHTTP5xxSpike),
		evidenceWithSignal("evidence-log", investigation.EvidenceSignalRelevantLog),
	}

	hypotheses, err := generator.GenerateHypotheses(context.Background(), investigation.Incident{}, memoryClassification(), evidence)
	if err != nil {
		t.Fatalf("GenerateHypotheses returned error: %v", err)
	}

	if len(hypotheses) != 1 {
		t.Fatalf("hypothesis count = %d", len(hypotheses))
	}
	got := hypotheses[0]
	if got.ID != "hypothesis-memory-leak-after-deployment" {
		t.Fatalf("hypothesis ID = %q", got.ID)
	}
	if got.RootCauseCode != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("root cause = %q", got.RootCauseCode)
	}
	if len(got.MissingEvidence) != 0 {
		t.Fatalf("missing evidence = %#v", got.MissingEvidence)
	}

	wantSupport := []string{
		"evidence-deployment",
		"evidence-memory",
		"evidence-oom",
		"evidence-restarts",
		"evidence-5xx",
		"evidence-log",
	}
	assertStringSliceEqual(t, got.SupportingEvidenceIDs, wantSupport)
}

func TestDeterministicGenerator_GenerateHypotheses_ListsMissingCoreEvidence(t *testing.T) {
	generator := NewDeterministicGenerator()
	evidence := []investigation.Evidence{
		evidenceWithSignal("evidence-deployment", investigation.EvidenceSignalRecentDeployment),
		evidenceWithSignal("evidence-oom", investigation.EvidenceSignalOOMKilled),
	}

	hypotheses, err := generator.GenerateHypotheses(context.Background(), investigation.Incident{}, memoryClassification(), evidence)
	if err != nil {
		t.Fatalf("GenerateHypotheses returned error: %v", err)
	}

	if len(hypotheses) != 1 {
		t.Fatalf("hypothesis count = %d", len(hypotheses))
	}
	assertStringSliceEqual(t, hypotheses[0].SupportingEvidenceIDs, []string{"evidence-deployment", "evidence-oom"})
	assertStringSliceEqual(t, hypotheses[0].MissingEvidence, []string{"memory_increase"})
}

func TestDeterministicGenerator_GenerateHypotheses_DoesNotGenerateWithInsufficientEvidence(t *testing.T) {
	generator := NewDeterministicGenerator()
	evidence := []investigation.Evidence{
		evidenceWithSignal("evidence-deployment", investigation.EvidenceSignalRecentDeployment),
	}

	hypotheses, err := generator.GenerateHypotheses(context.Background(), investigation.Incident{}, memoryClassification(), evidence)
	if err != nil {
		t.Fatalf("GenerateHypotheses returned error: %v", err)
	}

	if len(hypotheses) != 0 {
		t.Fatalf("hypotheses = %#v", hypotheses)
	}
}

func TestDeterministicGenerator_GenerateHypotheses_IgnoresNonMemoryClassification(t *testing.T) {
	generator := NewDeterministicGenerator()
	evidence := []investigation.Evidence{
		evidenceWithSignal("evidence-deployment", investigation.EvidenceSignalRecentDeployment),
		evidenceWithSignal("evidence-memory", investigation.EvidenceSignalMemoryIncrease),
		evidenceWithSignal("evidence-oom", investigation.EvidenceSignalOOMKilled),
	}
	classification := memoryClassification()
	classification.AlertType = "latency"

	hypotheses, err := generator.GenerateHypotheses(context.Background(), investigation.Incident{}, classification, evidence)
	if err != nil {
		t.Fatalf("GenerateHypotheses returned error: %v", err)
	}

	if len(hypotheses) != 0 {
		t.Fatalf("hypotheses = %#v", hypotheses)
	}
}

func TestDeterministicGenerator_GenerateHypotheses_ReturnsContextErrorWhenCanceled(t *testing.T) {
	generator := NewDeterministicGenerator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := generator.GenerateHypotheses(ctx, investigation.Incident{}, memoryClassification(), nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func memoryClassification() investigation.Classification {
	return investigation.Classification{
		AlertType: investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		Service:   "checkout-api",
		Priority:  investigation.PriorityP1,
	}
}

func evidenceWithSignal(id string, signal investigation.EvidenceSignal) investigation.Evidence {
	return investigation.Evidence{
		ID:         id,
		Signal:     signal,
		ObservedAt: time.Date(2026, 6, 28, 9, 45, 0, 0, time.UTC),
		Summary:    string(signal),
	}
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	}
}
