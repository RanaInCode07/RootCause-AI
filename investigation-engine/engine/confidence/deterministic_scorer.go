package confidence

import (
	"context"
	"math"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

var memoryLeakWeights = map[investigation.EvidenceSignal]float64{
	investigation.EvidenceSignalRecentDeployment:     0.25,
	investigation.EvidenceSignalMemoryIncrease:       0.35,
	investigation.EvidenceSignalOOMKilled:            0.30,
	investigation.EvidenceSignalRestartCountIncrease: 0.04,
	investigation.EvidenceSignalHTTP5xxSpike:         0.03,
	investigation.EvidenceSignalRelevantLog:          0.03,
}

var memoryLeakCoreSignals = []investigation.EvidenceSignal{
	investigation.EvidenceSignalRecentDeployment,
	investigation.EvidenceSignalMemoryIncrease,
	investigation.EvidenceSignalOOMKilled,
}

type DeterministicScorer struct{}

func NewDeterministicScorer() DeterministicScorer {
	return DeterministicScorer{}
}

func (DeterministicScorer) ScoreHypotheses(ctx context.Context, hypotheses []investigation.Hypothesis, evidence []investigation.Evidence) ([]investigation.ScoredHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	byID := evidenceByID(evidence)
	scored := make([]investigation.ScoredHypothesis, 0, len(hypotheses))
	for _, hypothesis := range hypotheses {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		scored = append(scored, scoreHypothesis(hypothesis, byID))
	}
	return scored, nil
}

func scoreHypothesis(hypothesis investigation.Hypothesis, evidenceByID map[string]investigation.Evidence) investigation.ScoredHypothesis {
	supportingEvidence := supportingEvidenceForHypothesis(hypothesis, evidenceByID)
	switch hypothesis.RootCauseCode {
	case investigation.RootCauseMemoryLeakAfterDeployment:
		return investigation.ScoredHypothesis{
			Hypothesis:         hypothesis,
			Confidence:         scoreMemoryLeakAfterDeployment(supportingEvidence),
			SupportingEvidence: supportingEvidence,
			MissingEvidence:    missingMemoryLeakEvidence(hypothesis, supportingEvidence),
		}
	default:
		return investigation.ScoredHypothesis{
			Hypothesis:         hypothesis,
			Confidence:         0,
			SupportingEvidence: supportingEvidence,
			MissingEvidence:    []string{"unsupported_root_cause_rule"},
		}
	}
}

func supportingEvidenceForHypothesis(hypothesis investigation.Hypothesis, evidenceByID map[string]investigation.Evidence) []investigation.Evidence {
	supportingEvidence := make([]investigation.Evidence, 0, len(hypothesis.SupportingEvidenceIDs))
	for _, id := range hypothesis.SupportingEvidenceIDs {
		item, ok := evidenceByID[id]
		if ok {
			supportingEvidence = append(supportingEvidence, item)
		}
	}
	return supportingEvidence
}

func scoreMemoryLeakAfterDeployment(evidence []investigation.Evidence) float64 {
	seen := make(map[investigation.EvidenceSignal]bool, len(evidence))
	score := 0.0
	for _, item := range evidence {
		if seen[item.Signal] {
			continue
		}
		seen[item.Signal] = true
		score += memoryLeakWeights[item.Signal]
	}
	return roundConfidence(clamp(score, 0, 1))
}

func missingMemoryLeakEvidence(hypothesis investigation.Hypothesis, evidence []investigation.Evidence) []string {
	missing := make([]string, 0, len(memoryLeakCoreSignals))
	presentSignals := make(map[investigation.EvidenceSignal]bool, len(evidence))
	for _, item := range evidence {
		presentSignals[item.Signal] = true
	}

	alreadyMissing := make(map[string]bool, len(hypothesis.MissingEvidence))
	for _, item := range hypothesis.MissingEvidence {
		if item == "" || alreadyMissing[item] {
			continue
		}
		alreadyMissing[item] = true
		missing = append(missing, item)
	}

	for _, signal := range memoryLeakCoreSignals {
		if presentSignals[signal] {
			continue
		}
		value := string(signal)
		if alreadyMissing[value] {
			continue
		}
		missing = append(missing, value)
	}

	return missing
}

func evidenceByID(evidence []investigation.Evidence) map[string]investigation.Evidence {
	byID := make(map[string]investigation.Evidence, len(evidence))
	for _, item := range evidence {
		if item.ID == "" {
			continue
		}
		byID[item.ID] = item
	}
	return byID
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func roundConfidence(value float64) float64 {
	return math.Round(value*100) / 100
}
