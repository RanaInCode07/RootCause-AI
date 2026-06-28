package evaluator

import (
	"context"
	"errors"
	"testing"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestDeterministicEvaluator_Evaluate_ReturnsPassForMatchingRootCause(t *testing.T) {
	var _ investigation.Evaluator = DeterministicEvaluator{}

	evaluator := NewDeterministicEvaluator()

	result, err := evaluator.Evaluate(context.Background(), reportWithRootCause(investigation.RootCauseMemoryLeakAfterDeployment), groundTruth())
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if result.Status != investigation.EvaluationStatusPass {
		t.Fatalf("status = %q", result.Status)
	}
	if result.ExpectedRootCauseCode != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("expected root cause = %q", result.ExpectedRootCauseCode)
	}
	if result.PredictedRootCauseCode != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("predicted root cause = %q", result.PredictedRootCauseCode)
	}
	if result.Confidence != 0.92 {
		t.Fatalf("confidence = %v", result.Confidence)
	}
	assertStringSliceEqual(t, result.MatchedEvidenceIDs, []string{"deploy-1", "metric-memory", "kube-oom"})
	assertStringSliceEqual(t, result.MissingEvidenceIDs, []string{"log-oom"})
}

func TestDeterministicEvaluator_Evaluate_ReturnsFailForDifferentRootCause(t *testing.T) {
	evaluator := NewDeterministicEvaluator()

	result, err := evaluator.Evaluate(context.Background(), reportWithRootCause("dependency_timeout"), groundTruth())
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if result.Status != investigation.EvaluationStatusFail {
		t.Fatalf("status = %q", result.Status)
	}
	if result.ExpectedRootCauseCode != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("expected root cause = %q", result.ExpectedRootCauseCode)
	}
	if result.PredictedRootCauseCode != "dependency_timeout" {
		t.Fatalf("predicted root cause = %q", result.PredictedRootCauseCode)
	}
	if result.Confidence != 0.92 {
		t.Fatalf("confidence = %v", result.Confidence)
	}
}

func TestDeterministicEvaluator_Evaluate_DoesNotMatchUncitedEvidence(t *testing.T) {
	evaluator := NewDeterministicEvaluator()
	report := reportWithRootCause(investigation.RootCauseMemoryLeakAfterDeployment)
	report.RootCause.SupportingEvidenceIDs = []string{"evidence-deployment"}

	result, err := evaluator.Evaluate(context.Background(), report, groundTruth())
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	assertStringSliceEqual(t, result.MatchedEvidenceIDs, []string{"deploy-1"})
	assertStringSliceEqual(t, result.MissingEvidenceIDs, []string{"metric-memory", "kube-oom", "log-oom"})
}

func TestDeterministicEvaluator_Evaluate_MatchesDirectSupportingEvidenceIDs(t *testing.T) {
	evaluator := NewDeterministicEvaluator()
	report := investigation.Report{
		RootCause: investigation.RootCausePrediction{
			Code:                  investigation.RootCauseMemoryLeakAfterDeployment,
			Confidence:            0.7,
			SupportingEvidenceIDs: []string{"deploy-1", "metric-memory"},
		},
	}

	result, err := evaluator.Evaluate(context.Background(), report, groundTruth())
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	assertStringSliceEqual(t, result.MatchedEvidenceIDs, []string{"deploy-1", "metric-memory"})
	assertStringSliceEqual(t, result.MissingEvidenceIDs, []string{"kube-oom", "log-oom"})
}

func TestDeterministicEvaluator_Evaluate_ReturnsContextErrorWhenCanceled(t *testing.T) {
	evaluator := NewDeterministicEvaluator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := evaluator.Evaluate(ctx, reportWithRootCause(investigation.RootCauseMemoryLeakAfterDeployment), groundTruth())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func reportWithRootCause(rootCause investigation.RootCauseCode) investigation.Report {
	return investigation.Report{
		RootCause: investigation.RootCausePrediction{
			Code:                  rootCause,
			Confidence:            0.92,
			SupportingEvidenceIDs: []string{"evidence-deployment", "evidence-memory", "evidence-oom"},
		},
		Evidence: []investigation.Evidence{
			{ID: "evidence-deployment", Source: investigation.EvidenceSource{Kind: "deployment", ID: "deploy-1"}},
			{ID: "evidence-memory", Source: investigation.EvidenceSource{Kind: "metric", ID: "metric-memory"}},
			{ID: "evidence-oom", Source: investigation.EvidenceSource{Kind: "kubernetes_event", ID: "kube-oom"}},
			{ID: "evidence-log", Source: investigation.EvidenceSource{Kind: "log", ID: "log-oom"}},
		},
	}
}

func groundTruth() investigation.GroundTruth {
	return investigation.GroundTruth{
		RootCauseCode: investigation.RootCauseMemoryLeakAfterDeployment,
		EvidenceIDs:   []string{"deploy-1", "metric-memory", "kube-oom", "log-oom"},
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
