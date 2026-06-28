package report

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestDeterministicGenerator_GenerateReport_BuildsReportFromTopHypothesis(t *testing.T) {
	var _ investigation.ReportGenerator = DeterministicGenerator{}

	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	generator := NewDeterministicGenerator(WithClock(func() time.Time { return now }))

	incident := validIncident()
	classification := validClassification()
	evidence := validEvidence()
	hypotheses := []investigation.ScoredHypothesis{
		scoredHypothesis("hypothesis-low", "dependency_timeout", 0.41, nil, []string{"dependency_error_rate"}),
		scoredHypothesis(
			"hypothesis-memory-leak-after-deployment",
			investigation.RootCauseMemoryLeakAfterDeployment,
			1,
			[]string{"evidence-deployment", "evidence-memory", "evidence-oom"},
			nil,
		),
	}

	report, err := generator.GenerateReport(context.Background(), incident, classification, evidence, hypotheses)
	if err != nil {
		t.Fatalf("GenerateReport returned error: %v", err)
	}
	if err := report.Validate(); err != nil {
		t.Fatalf("report validation failed: %v", err)
	}

	if report.IncidentID != incident.ID {
		t.Fatalf("incident id = %q", report.IncidentID)
	}
	if !report.GeneratedAt.Equal(now) {
		t.Fatalf("generated_at = %s", report.GeneratedAt)
	}
	if report.RootCause.Code != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("root cause = %q", report.RootCause.Code)
	}
	if report.RootCause.Confidence != 1 {
		t.Fatalf("confidence = %v", report.RootCause.Confidence)
	}
	assertStringSliceEqual(t, report.RootCause.SupportingEvidenceIDs, []string{"evidence-deployment", "evidence-memory", "evidence-oom"})
	if len(report.RootCause.MissingEvidence) != 0 {
		t.Fatalf("missing evidence = %#v", report.RootCause.MissingEvidence)
	}
	if len(report.Evidence) != len(evidence) {
		t.Fatalf("evidence count = %d", len(report.Evidence))
	}
	if len(report.Alternatives) != 1 {
		t.Fatalf("alternative count = %d", len(report.Alternatives))
	}
	if report.Alternatives[0].Hypothesis.ID != "hypothesis-low" {
		t.Fatalf("alternative ID = %q", report.Alternatives[0].Hypothesis.ID)
	}
	if !strings.Contains(strings.ToLower(report.Recommendation), "rollback") {
		t.Fatalf("recommendation = %q, want rollback guidance", report.Recommendation)
	}
	if !strings.Contains(report.Summary, "checkout-api") {
		t.Fatalf("summary = %q, want service name", report.Summary)
	}
}

func TestDeterministicGenerator_GenerateReport_BuildsChronologicalTimeline(t *testing.T) {
	generator := NewDeterministicGenerator(WithClock(func() time.Time {
		return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	}))

	report, err := generator.GenerateReport(
		context.Background(),
		validIncident(),
		validClassification(),
		validEvidence(),
		[]investigation.ScoredHypothesis{
			scoredHypothesis(
				"hypothesis-memory-leak-after-deployment",
				investigation.RootCauseMemoryLeakAfterDeployment,
				1,
				[]string{"evidence-deployment", "evidence-memory", "evidence-oom"},
				nil,
			),
		},
	)
	if err != nil {
		t.Fatalf("GenerateReport returned error: %v", err)
	}

	if len(report.Timeline) < 5 {
		t.Fatalf("timeline count = %d", len(report.Timeline))
	}
	for i := 1; i < len(report.Timeline); i++ {
		if report.Timeline[i].Timestamp.Before(report.Timeline[i-1].Timestamp) {
			t.Fatalf("timeline is not sorted: %#v", report.Timeline)
		}
	}
	if report.Timeline[0].Source != "deployment" {
		t.Fatalf("first timeline source = %q", report.Timeline[0].Source)
	}
	if !timelineHasSource(report.Timeline, "evidence") {
		t.Fatalf("timeline missing evidence event: %#v", report.Timeline)
	}
}

func TestDeterministicGenerator_GenerateReport_UsesLowConfidenceRecommendation(t *testing.T) {
	generator := NewDeterministicGenerator()
	report, err := generator.GenerateReport(
		context.Background(),
		validIncident(),
		validClassification(),
		validEvidence(),
		[]investigation.ScoredHypothesis{
			scoredHypothesis(
				"hypothesis-memory-leak-after-deployment",
				investigation.RootCauseMemoryLeakAfterDeployment,
				0.55,
				[]string{"evidence-deployment", "evidence-oom"},
				[]string{"memory_increase"},
			),
		},
	)
	if err != nil {
		t.Fatalf("GenerateReport returned error: %v", err)
	}

	if !strings.Contains(strings.ToLower(report.Recommendation), "collect") {
		t.Fatalf("recommendation = %q, want collect guidance", report.Recommendation)
	}
	assertStringSliceEqual(t, report.RootCause.MissingEvidence, []string{"memory_increase"})
}

func TestDeterministicGenerator_GenerateReport_ProducesJSONSerializableReport(t *testing.T) {
	generator := NewDeterministicGenerator()
	report, err := generator.GenerateReport(
		context.Background(),
		validIncident(),
		validClassification(),
		validEvidence(),
		[]investigation.ScoredHypothesis{
			scoredHypothesis(
				"hypothesis-memory-leak-after-deployment",
				investigation.RootCauseMemoryLeakAfterDeployment,
				1,
				[]string{"evidence-deployment", "evidence-memory", "evidence-oom"},
				nil,
			),
		},
	)
	if err != nil {
		t.Fatalf("GenerateReport returned error: %v", err)
	}

	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if !strings.Contains(string(raw), `"root_cause"`) {
		t.Fatalf("marshaled report missing root_cause: %s", raw)
	}
}

func TestDeterministicGenerator_GenerateReport_ReturnsErrorWithoutHypotheses(t *testing.T) {
	generator := NewDeterministicGenerator()

	_, err := generator.GenerateReport(context.Background(), validIncident(), validClassification(), validEvidence(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no scored hypotheses") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestDeterministicGenerator_GenerateReport_ReturnsContextErrorWhenCanceled(t *testing.T) {
	generator := NewDeterministicGenerator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := generator.GenerateReport(ctx, validIncident(), validClassification(), validEvidence(), nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func validIncident() investigation.Incident {
	return investigation.Incident{
		ID: "inc-oom-checkout-2026-06-28",
		Metadata: investigation.IncidentMetadata{
			Title:      "Checkout API memory exhaustion after deployment",
			DetectedAt: time.Date(2026, 6, 28, 9, 45, 0, 0, time.UTC),
		},
		Alert: investigation.Alert{
			ID:        "alert-checkout-memory-001",
			Name:      "Checkout API pod restarts and high memory",
			Type:      investigation.AlertTypeMemory,
			Severity:  investigation.SeverityCritical,
			Service:   "checkout-api",
			StartedAt: time.Date(2026, 6, 28, 9, 42, 0, 0, time.UTC),
		},
		Deployment: investigation.Deployment{
			ID:         "deploy-checkout-api-20260628-0915",
			Service:    "checkout-api",
			Version:    "2026.06.28.1",
			DeployedAt: time.Date(2026, 6, 28, 9, 15, 0, 0, time.UTC),
		},
	}
}

func validClassification() investigation.Classification {
	return investigation.Classification{
		AlertType: investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		Service:   "checkout-api",
		Priority:  investigation.PriorityP1,
	}
}

func validEvidence() []investigation.Evidence {
	return []investigation.Evidence{
		evidence("evidence-memory", investigation.EvidenceSignalMemoryIncrease, time.Date(2026, 6, 28, 9, 44, 0, 0, time.UTC)),
		evidence("evidence-deployment", investigation.EvidenceSignalRecentDeployment, time.Date(2026, 6, 28, 9, 15, 0, 0, time.UTC)),
		evidence("evidence-oom", investigation.EvidenceSignalOOMKilled, time.Date(2026, 6, 28, 9, 41, 12, 0, time.UTC)),
	}
}

func evidence(id string, signal investigation.EvidenceSignal, observedAt time.Time) investigation.Evidence {
	return investigation.Evidence{
		ID:         id,
		Signal:     signal,
		ObservedAt: observedAt,
		Summary:    string(signal),
	}
}

func scoredHypothesis(id string, rootCause investigation.RootCauseCode, confidence float64, supportingEvidenceIDs []string, missingEvidence []string) investigation.ScoredHypothesis {
	hypothesis := investigation.Hypothesis{
		ID:                    id,
		RootCauseCode:         rootCause,
		Summary:               "Memory leak after deployment is plausible.",
		SupportingEvidenceIDs: supportingEvidenceIDs,
		MissingEvidence:       missingEvidence,
	}
	return investigation.ScoredHypothesis{
		Hypothesis:      hypothesis,
		Confidence:      confidence,
		MissingEvidence: missingEvidence,
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

func timelineHasSource(timeline []investigation.TimelineEvent, source string) bool {
	for _, item := range timeline {
		if item.Source == source {
			return true
		}
	}
	return false
}
