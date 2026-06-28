package investigation

import "context"

type IncidentLoader interface {
	LoadIncident(ctx context.Context, path string) (Incident, error)
}

type Classifier interface {
	Classify(ctx context.Context, incident Incident) (Classification, error)
}

type EvidenceCollector interface {
	CollectEvidence(ctx context.Context, incident Incident, classification Classification) ([]Evidence, error)
}

type HypothesisGenerator interface {
	GenerateHypotheses(ctx context.Context, incident Incident, classification Classification, evidence []Evidence) ([]Hypothesis, error)
}

type ConfidenceScorer interface {
	ScoreHypotheses(ctx context.Context, hypotheses []Hypothesis, evidence []Evidence) ([]ScoredHypothesis, error)
}

type ReportGenerator interface {
	GenerateReport(ctx context.Context, incident Incident, classification Classification, evidence []Evidence, hypotheses []ScoredHypothesis) (Report, error)
}

type Evaluator interface {
	Evaluate(ctx context.Context, report Report, groundTruth GroundTruth) (EvaluationResult, error)
}

type ReasoningEngine interface {
	Analyze(ctx context.Context, input ReasoningInput) (ReasoningResult, error)
}

type Summarizer interface {
	Summarize(ctx context.Context, input SummaryInput) (SummaryResult, error)
}

type Planner interface {
	Plan(ctx context.Context, input PlanningInput) (PlanningResult, error)
}
