package hypothesis

import (
	"context"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

var memoryLeakSupportSignals = []investigation.EvidenceSignal{
	investigation.EvidenceSignalRecentDeployment,
	investigation.EvidenceSignalMemoryIncrease,
	investigation.EvidenceSignalOOMKilled,
	investigation.EvidenceSignalRestartCountIncrease,
	investigation.EvidenceSignalHTTP5xxSpike,
	investigation.EvidenceSignalRelevantLog,
}

var memoryLeakCoreSignals = []investigation.EvidenceSignal{
	investigation.EvidenceSignalRecentDeployment,
	investigation.EvidenceSignalMemoryIncrease,
	investigation.EvidenceSignalOOMKilled,
}

type DeterministicGenerator struct{}

func NewDeterministicGenerator() DeterministicGenerator {
	return DeterministicGenerator{}
}

func (DeterministicGenerator) GenerateHypotheses(ctx context.Context, incident investigation.Incident, classification investigation.Classification, evidence []investigation.Evidence) ([]investigation.Hypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if classification.AlertType != investigation.AlertTypeMemory {
		return nil, nil
	}

	hypothesis, ok := memoryLeakAfterDeploymentHypothesis(evidence)
	if !ok {
		return nil, nil
	}
	return []investigation.Hypothesis{hypothesis}, nil
}

func memoryLeakAfterDeploymentHypothesis(evidence []investigation.Evidence) (investigation.Hypothesis, bool) {
	bySignal := evidenceBySignal(evidence)
	corePresent := 0
	missing := make([]string, 0, len(memoryLeakCoreSignals))
	for _, signal := range memoryLeakCoreSignals {
		if len(bySignal[signal]) > 0 {
			corePresent++
			continue
		}
		missing = append(missing, string(signal))
	}

	if corePresent < 2 {
		return investigation.Hypothesis{}, false
	}

	return investigation.Hypothesis{
		ID:                    "hypothesis-memory-leak-after-deployment",
		RootCauseCode:         investigation.RootCauseMemoryLeakAfterDeployment,
		Summary:               "Memory leak after deployment is plausible because deployment, memory, and OOM evidence point to a post-release memory regression.",
		SupportingEvidenceIDs: supportingEvidenceIDs(bySignal, memoryLeakSupportSignals),
		MissingEvidence:       missing,
	}, true
}

func evidenceBySignal(evidence []investigation.Evidence) map[investigation.EvidenceSignal][]investigation.Evidence {
	bySignal := make(map[investigation.EvidenceSignal][]investigation.Evidence, len(evidence))
	for _, item := range evidence {
		bySignal[item.Signal] = append(bySignal[item.Signal], item)
	}
	return bySignal
}

func supportingEvidenceIDs(bySignal map[investigation.EvidenceSignal][]investigation.Evidence, signals []investigation.EvidenceSignal) []string {
	ids := make([]string, 0)
	for _, signal := range signals {
		for _, item := range bySignal[signal] {
			if item.ID != "" {
				ids = append(ids, item.ID)
			}
		}
	}
	return ids
}
