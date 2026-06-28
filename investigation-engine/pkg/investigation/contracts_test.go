package investigation

import (
	"context"
	"testing"
)

func TestCoreInterfaces_AreImplementableByFutureAdapters(t *testing.T) {
	var _ IncidentLoader = fakeIncidentLoader{}
	var _ Classifier = fakeClassifier{}
	var _ EvidenceCollector = fakeEvidenceCollector{}
	var _ HypothesisGenerator = fakeHypothesisGenerator{}
	var _ ConfidenceScorer = fakeConfidenceScorer{}
	var _ ReportGenerator = fakeReportGenerator{}
	var _ Evaluator = fakeEvaluator{}
	var _ ReasoningEngine = fakeReasoningEngine{}
	var _ Summarizer = fakeSummarizer{}
	var _ Planner = fakePlanner{}
}

type fakeIncidentLoader struct{}

func (fakeIncidentLoader) LoadIncident(ctx context.Context, path string) (Incident, error) {
	return Incident{}, ctx.Err()
}

type fakeClassifier struct{}

func (fakeClassifier) Classify(ctx context.Context, incident Incident) (Classification, error) {
	return Classification{}, ctx.Err()
}

type fakeEvidenceCollector struct{}

func (fakeEvidenceCollector) CollectEvidence(ctx context.Context, incident Incident, classification Classification) ([]Evidence, error) {
	return nil, ctx.Err()
}

type fakeHypothesisGenerator struct{}

func (fakeHypothesisGenerator) GenerateHypotheses(ctx context.Context, incident Incident, classification Classification, evidence []Evidence) ([]Hypothesis, error) {
	return nil, ctx.Err()
}

type fakeConfidenceScorer struct{}

func (fakeConfidenceScorer) ScoreHypotheses(ctx context.Context, hypotheses []Hypothesis, evidence []Evidence) ([]ScoredHypothesis, error) {
	return nil, ctx.Err()
}

type fakeReportGenerator struct{}

func (fakeReportGenerator) GenerateReport(ctx context.Context, incident Incident, classification Classification, evidence []Evidence, hypotheses []ScoredHypothesis) (Report, error) {
	return Report{}, ctx.Err()
}

type fakeEvaluator struct{}

func (fakeEvaluator) Evaluate(ctx context.Context, report Report, groundTruth GroundTruth) (EvaluationResult, error) {
	return EvaluationResult{}, ctx.Err()
}

type fakeReasoningEngine struct{}

func (fakeReasoningEngine) Analyze(ctx context.Context, input ReasoningInput) (ReasoningResult, error) {
	return ReasoningResult{}, ctx.Err()
}

type fakeSummarizer struct{}

func (fakeSummarizer) Summarize(ctx context.Context, input SummaryInput) (SummaryResult, error) {
	return SummaryResult{}, ctx.Err()
}

type fakePlanner struct{}

func (fakePlanner) Plan(ctx context.Context, input PlanningInput) (PlanningResult, error) {
	return PlanningResult{}, ctx.Err()
}
