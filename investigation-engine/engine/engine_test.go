package engine

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"rootcause.ai/investigation-engine/evaluator"
	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestDefaultEngine_InvestigateFile_RunsOOMFixtureEndToEnd(t *testing.T) {
	engine := NewDefault()
	path := filepath.Join("..", "dataset", "incidents", "oom_after_deployment.json")

	result, err := engine.InvestigateFile(context.Background(), path)
	if err != nil {
		t.Fatalf("InvestigateFile returned error: %v", err)
	}

	if result.Incident.ID != "inc-oom-checkout-2026-06-28" {
		t.Fatalf("incident ID = %q", result.Incident.ID)
	}
	if result.Classification.Service != "checkout-api" {
		t.Fatalf("service = %q", result.Classification.Service)
	}
	if len(result.Evidence) < 6 {
		t.Fatalf("evidence count = %d", len(result.Evidence))
	}
	if len(result.Hypotheses) != 1 {
		t.Fatalf("hypothesis count = %d", len(result.Hypotheses))
	}
	if len(result.ScoredHypotheses) != 1 {
		t.Fatalf("scored hypothesis count = %d", len(result.ScoredHypotheses))
	}
	if result.Report.RootCause.Code != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("root cause = %q", result.Report.RootCause.Code)
	}
	if result.Report.RootCause.Confidence != 1 {
		t.Fatalf("confidence = %v", result.Report.RootCause.Confidence)
	}

	evaluation, err := evaluator.NewDeterministicEvaluator().Evaluate(context.Background(), result.Report, result.Incident.GroundTruth)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if evaluation.Status != investigation.EvaluationStatusPass {
		t.Fatalf("evaluation status = %q", evaluation.Status)
	}
	if len(evaluation.MatchedEvidenceIDs) != 5 {
		t.Fatalf("matched evidence = %#v", evaluation.MatchedEvidenceIDs)
	}
	if len(evaluation.MissingEvidenceIDs) != 0 {
		t.Fatalf("missing evidence = %#v", evaluation.MissingEvidenceIDs)
	}
}

func TestNew_ReturnsErrorForMissingDependency(t *testing.T) {
	_, err := New(Config{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "incident loader") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestEngine_InvestigateFile_WrapsStageError(t *testing.T) {
	wantErr := errors.New("loader failed")
	engine, err := New(Config{
		IncidentLoader:      fakeLoader{err: wantErr},
		Classifier:          fakeClassifier{},
		EvidenceCollector:   fakeCollector{},
		HypothesisGenerator: fakeHypothesisGenerator{},
		ConfidenceScorer:    fakeConfidenceScorer{},
		ReportGenerator:     fakeReportGenerator{},
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = engine.InvestigateFile(context.Background(), "incident.json")
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "load incident") {
		t.Fatalf("error = %q, want stage context", err.Error())
	}
}

func TestEngine_InvestigateFile_ReturnsContextErrorWhenCanceled(t *testing.T) {
	engine := NewDefault()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := engine.InvestigateFile(ctx, filepath.Join("..", "dataset", "incidents", "oom_after_deployment.json"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

type fakeLoader struct {
	err error
}

func (loader fakeLoader) LoadIncident(ctx context.Context, path string) (investigation.Incident, error) {
	return investigation.Incident{}, loader.err
}

type fakeClassifier struct{}

func (fakeClassifier) Classify(ctx context.Context, incident investigation.Incident) (investigation.Classification, error) {
	return investigation.Classification{}, nil
}

type fakeCollector struct{}

func (fakeCollector) CollectEvidence(ctx context.Context, incident investigation.Incident, classification investigation.Classification) ([]investigation.Evidence, error) {
	return nil, nil
}

type fakeHypothesisGenerator struct{}

func (fakeHypothesisGenerator) GenerateHypotheses(ctx context.Context, incident investigation.Incident, classification investigation.Classification, evidence []investigation.Evidence) ([]investigation.Hypothesis, error) {
	return nil, nil
}

type fakeConfidenceScorer struct{}

func (fakeConfidenceScorer) ScoreHypotheses(ctx context.Context, hypotheses []investigation.Hypothesis, evidence []investigation.Evidence) ([]investigation.ScoredHypothesis, error) {
	return nil, nil
}

type fakeReportGenerator struct{}

func (fakeReportGenerator) GenerateReport(ctx context.Context, incident investigation.Incident, classification investigation.Classification, evidence []investigation.Evidence, hypotheses []investigation.ScoredHypothesis) (investigation.Report, error) {
	return investigation.Report{}, nil
}
