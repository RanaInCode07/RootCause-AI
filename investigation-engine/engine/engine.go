package engine

import (
	"context"
	"fmt"

	"rootcause.ai/investigation-engine/engine/classifier"
	"rootcause.ai/investigation-engine/engine/collector"
	"rootcause.ai/investigation-engine/engine/confidence"
	"rootcause.ai/investigation-engine/engine/hypothesis"
	"rootcause.ai/investigation-engine/engine/loader"
	"rootcause.ai/investigation-engine/engine/report"
	"rootcause.ai/investigation-engine/pkg/investigation"
)

type Config struct {
	IncidentLoader      investigation.IncidentLoader
	Classifier          investigation.Classifier
	EvidenceCollector   investigation.EvidenceCollector
	HypothesisGenerator investigation.HypothesisGenerator
	ConfidenceScorer    investigation.ConfidenceScorer
	ReportGenerator     investigation.ReportGenerator
}

type Engine struct {
	incidentLoader      investigation.IncidentLoader
	classifier          investigation.Classifier
	evidenceCollector   investigation.EvidenceCollector
	hypothesisGenerator investigation.HypothesisGenerator
	confidenceScorer    investigation.ConfidenceScorer
	reportGenerator     investigation.ReportGenerator
}

type Result struct {
	Incident         investigation.Incident
	Classification   investigation.Classification
	Evidence         []investigation.Evidence
	Hypotheses       []investigation.Hypothesis
	ScoredHypotheses []investigation.ScoredHypothesis
	Report           investigation.Report
}

func NewDefault() Engine {
	return Engine{
		incidentLoader:      loader.NewJSONFileLoader(),
		classifier:          classifier.NewDeterministicClassifier(),
		evidenceCollector:   collector.NewDeterministicCollector(),
		hypothesisGenerator: hypothesis.NewDeterministicGenerator(),
		confidenceScorer:    confidence.NewDeterministicScorer(),
		reportGenerator:     report.NewDeterministicGenerator(),
	}
}

func New(config Config) (Engine, error) {
	if config.IncidentLoader == nil {
		return Engine{}, fmt.Errorf("incident loader is required")
	}
	if config.Classifier == nil {
		return Engine{}, fmt.Errorf("classifier is required")
	}
	if config.EvidenceCollector == nil {
		return Engine{}, fmt.Errorf("evidence collector is required")
	}
	if config.HypothesisGenerator == nil {
		return Engine{}, fmt.Errorf("hypothesis generator is required")
	}
	if config.ConfidenceScorer == nil {
		return Engine{}, fmt.Errorf("confidence scorer is required")
	}
	if config.ReportGenerator == nil {
		return Engine{}, fmt.Errorf("report generator is required")
	}

	return Engine{
		incidentLoader:      config.IncidentLoader,
		classifier:          config.Classifier,
		evidenceCollector:   config.EvidenceCollector,
		hypothesisGenerator: config.HypothesisGenerator,
		confidenceScorer:    config.ConfidenceScorer,
		reportGenerator:     config.ReportGenerator,
	}, nil
}

func (engine Engine) InvestigateFile(ctx context.Context, path string) (Result, error) {
	incident, err := engine.incidentLoader.LoadIncident(ctx, path)
	if err != nil {
		return Result{}, fmt.Errorf("load incident: %w", err)
	}

	classification, err := engine.classifier.Classify(ctx, incident)
	if err != nil {
		return Result{}, fmt.Errorf("classify incident: %w", err)
	}

	evidence, err := engine.evidenceCollector.CollectEvidence(ctx, incident, classification)
	if err != nil {
		return Result{}, fmt.Errorf("collect evidence: %w", err)
	}

	hypotheses, err := engine.hypothesisGenerator.GenerateHypotheses(ctx, incident, classification, evidence)
	if err != nil {
		return Result{}, fmt.Errorf("generate hypotheses: %w", err)
	}

	scoredHypotheses, err := engine.confidenceScorer.ScoreHypotheses(ctx, hypotheses, evidence)
	if err != nil {
		return Result{}, fmt.Errorf("score hypotheses: %w", err)
	}

	report, err := engine.reportGenerator.GenerateReport(ctx, incident, classification, evidence, scoredHypotheses)
	if err != nil {
		return Result{}, fmt.Errorf("generate report: %w", err)
	}

	return Result{
		Incident:         incident,
		Classification:   classification,
		Evidence:         evidence,
		Hypotheses:       hypotheses,
		ScoredHypotheses: scoredHypotheses,
		Report:           report,
	}, nil
}
