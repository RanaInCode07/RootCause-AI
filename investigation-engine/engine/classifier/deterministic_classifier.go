package classifier

import (
	"context"
	"fmt"
	"strings"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

type DeterministicClassifier struct{}

func NewDeterministicClassifier() DeterministicClassifier {
	return DeterministicClassifier{}
}

func (DeterministicClassifier) Classify(ctx context.Context, incident investigation.Incident) (investigation.Classification, error) {
	if err := ctx.Err(); err != nil {
		return investigation.Classification{}, err
	}

	if err := incident.Alert.Validate(); err != nil {
		return investigation.Classification{}, fmt.Errorf("validate alert: %w", err)
	}

	return investigation.Classification{
		AlertType: incident.Alert.Type,
		Severity:  incident.Alert.Severity,
		Service:   incident.Alert.Service,
		Priority:  priorityForAlert(incident.Alert),
	}, nil
}

func priorityForAlert(alert investigation.Alert) investigation.Priority {
	if priority, ok := explicitPriority(alert.Labels["priority"]); ok {
		return priority
	}

	switch alert.Severity {
	case investigation.SeverityCritical:
		return investigation.PriorityP1
	case investigation.SeverityWarning:
		return investigation.PriorityP2
	default:
		return investigation.PriorityP2
	}
}

func explicitPriority(value string) (investigation.Priority, bool) {
	switch investigation.Priority(strings.ToUpper(strings.TrimSpace(value))) {
	case investigation.PriorityP0:
		return investigation.PriorityP0, true
	case investigation.PriorityP1:
		return investigation.PriorityP1, true
	case investigation.PriorityP2:
		return investigation.PriorityP2, true
	default:
		return "", false
	}
}
