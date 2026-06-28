package confidence

import (
	"context"
	"errors"
	"testing"
	"time"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestDeterministicScorer_ScoreHypotheses_ScoresFullMemoryLeakEvidenceHigh(t *testing.T) {
	var _ investigation.ConfidenceScorer = DeterministicScorer{}

	scorer := NewDeterministicScorer()
	hypothesis := memoryLeakHypothesis(
		[]string{
			"evidence-deployment",
			"evidence-memory",
			"evidence-oom",
			"evidence-restarts",
			"evidence-5xx",
			"evidence-log",
		},
		nil,
	)
	evidence := []investigation.Evidence{
		evidenceWithSignal("evidence-deployment", investigation.EvidenceSignalRecentDeployment),
		evidenceWithSignal("evidence-memory", investigation.EvidenceSignalMemoryIncrease),
		evidenceWithSignal("evidence-oom", investigation.EvidenceSignalOOMKilled),
		evidenceWithSignal("evidence-restarts", investigation.EvidenceSignalRestartCountIncrease),
		evidenceWithSignal("evidence-5xx", investigation.EvidenceSignalHTTP5xxSpike),
		evidenceWithSignal("evidence-log", investigation.EvidenceSignalRelevantLog),
	}

	scored, err := scorer.ScoreHypotheses(context.Background(), []investigation.Hypothesis{hypothesis}, evidence)
	if err != nil {
		t.Fatalf("ScoreHypotheses returned error: %v", err)
	}

	if len(scored) != 1 {
		t.Fatalf("scored count = %d", len(scored))
	}
	if scored[0].Confidence != 1 {
		t.Fatalf("confidence = %v, want 1", scored[0].Confidence)
	}
	if len(scored[0].SupportingEvidence) != 6 {
		t.Fatalf("supporting evidence count = %d", len(scored[0].SupportingEvidence))
	}
	if len(scored[0].MissingEvidence) != 0 {
		t.Fatalf("missing evidence = %#v", scored[0].MissingEvidence)
	}
}

func TestDeterministicScorer_ScoreHypotheses_ScoresPartialEvidenceLower(t *testing.T) {
	scorer := NewDeterministicScorer()
	hypothesis := memoryLeakHypothesis(
		[]string{"evidence-deployment", "evidence-oom"},
		[]string{"memory_increase"},
	)
	evidence := []investigation.Evidence{
		evidenceWithSignal("evidence-deployment", investigation.EvidenceSignalRecentDeployment),
		evidenceWithSignal("evidence-oom", investigation.EvidenceSignalOOMKilled),
	}

	scored, err := scorer.ScoreHypotheses(context.Background(), []investigation.Hypothesis{hypothesis}, evidence)
	if err != nil {
		t.Fatalf("ScoreHypotheses returned error: %v", err)
	}

	if len(scored) != 1 {
		t.Fatalf("scored count = %d", len(scored))
	}
	if scored[0].Confidence != 0.55 {
		t.Fatalf("confidence = %v, want 0.55", scored[0].Confidence)
	}
	if len(scored[0].SupportingEvidence) != 2 {
		t.Fatalf("supporting evidence count = %d", len(scored[0].SupportingEvidence))
	}
	assertStringSliceEqual(t, scored[0].MissingEvidence, []string{"memory_increase"})
}

func TestDeterministicScorer_ScoreHypotheses_DerivesMissingCoreEvidenceWhenHypothesisOmitsIt(t *testing.T) {
	scorer := NewDeterministicScorer()
	hypothesis := memoryLeakHypothesis(
		[]string{"evidence-deployment", "evidence-oom"},
		nil,
	)
	evidence := []investigation.Evidence{
		evidenceWithSignal("evidence-deployment", investigation.EvidenceSignalRecentDeployment),
		evidenceWithSignal("evidence-oom", investigation.EvidenceSignalOOMKilled),
	}

	scored, err := scorer.ScoreHypotheses(context.Background(), []investigation.Hypothesis{hypothesis}, evidence)
	if err != nil {
		t.Fatalf("ScoreHypotheses returned error: %v", err)
	}

	assertStringSliceEqual(t, scored[0].MissingEvidence, []string{"memory_increase"})
}

func TestDeterministicScorer_ScoreHypotheses_ScoresUnknownRootCauseZero(t *testing.T) {
	scorer := NewDeterministicScorer()
	hypothesis := investigation.Hypothesis{
		ID:            "hypothesis-unknown",
		RootCauseCode: "unknown_root_cause",
		Summary:       "unknown",
	}

	scored, err := scorer.ScoreHypotheses(context.Background(), []investigation.Hypothesis{hypothesis}, nil)
	if err != nil {
		t.Fatalf("ScoreHypotheses returned error: %v", err)
	}

	if len(scored) != 1 {
		t.Fatalf("scored count = %d", len(scored))
	}
	if scored[0].Confidence != 0 {
		t.Fatalf("confidence = %v, want 0", scored[0].Confidence)
	}
	if len(scored[0].MissingEvidence) == 0 {
		t.Fatal("expected missing evidence for unknown root cause")
	}
}

func TestDeterministicScorer_ScoreHypotheses_ReturnsContextErrorWhenCanceled(t *testing.T) {
	scorer := NewDeterministicScorer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := scorer.ScoreHypotheses(ctx, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func memoryLeakHypothesis(supportingEvidenceIDs []string, missingEvidence []string) investigation.Hypothesis {
	return investigation.Hypothesis{
		ID:                    "hypothesis-memory-leak-after-deployment",
		RootCauseCode:         investigation.RootCauseMemoryLeakAfterDeployment,
		Summary:               "Memory leak after deployment is plausible.",
		SupportingEvidenceIDs: supportingEvidenceIDs,
		MissingEvidence:       missingEvidence,
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
