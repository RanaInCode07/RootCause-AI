package evaluator

import (
	"context"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

type DeterministicEvaluator struct{}

func NewDeterministicEvaluator() DeterministicEvaluator {
	return DeterministicEvaluator{}
}

func (DeterministicEvaluator) Evaluate(ctx context.Context, report investigation.Report, groundTruth investigation.GroundTruth) (investigation.EvaluationResult, error) {
	if err := ctx.Err(); err != nil {
		return investigation.EvaluationResult{}, err
	}

	status := investigation.EvaluationStatusFail
	if report.RootCause.Code == groundTruth.RootCauseCode {
		status = investigation.EvaluationStatusPass
	}

	matched, missing := compareEvidence(report, groundTruth)
	return investigation.EvaluationResult{
		Status:                 status,
		ExpectedRootCauseCode:  groundTruth.RootCauseCode,
		PredictedRootCauseCode: report.RootCause.Code,
		Confidence:             report.RootCause.Confidence,
		MatchedEvidenceIDs:     matched,
		MissingEvidenceIDs:     missing,
	}, nil
}

func compareEvidence(report investigation.Report, groundTruth investigation.GroundTruth) ([]string, []string) {
	citedIDs := citedEvidenceIDs(report)
	matched := make([]string, 0, len(groundTruth.EvidenceIDs))
	missing := make([]string, 0, len(groundTruth.EvidenceIDs))

	for _, expectedID := range groundTruth.EvidenceIDs {
		if citedIDs[expectedID] {
			matched = append(matched, expectedID)
			continue
		}
		missing = append(missing, expectedID)
	}
	return matched, missing
}

func citedEvidenceIDs(report investigation.Report) map[string]bool {
	cited := make(map[string]bool, len(report.RootCause.SupportingEvidenceIDs)*2)
	supporting := make(map[string]bool, len(report.RootCause.SupportingEvidenceIDs))
	for _, id := range report.RootCause.SupportingEvidenceIDs {
		if id == "" {
			continue
		}
		cited[id] = true
		supporting[id] = true
	}

	for _, item := range report.Evidence {
		if !supporting[item.ID] {
			continue
		}
		if item.Source.ID != "" {
			cited[item.Source.ID] = true
		}
	}
	return cited
}
