package classifier

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestDeterministicClassifier_Classify_ReturnsAlertClassification(t *testing.T) {
	var _ investigation.Classifier = DeterministicClassifier{}

	classifier := NewDeterministicClassifier()
	incident := incidentWithAlert(investigation.Alert{
		ID:        "alert-1",
		Name:      "Checkout memory high",
		Type:      investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		Service:   "checkout-api",
		StartedAt: validTime(),
	})

	classification, err := classifier.Classify(context.Background(), incident)
	if err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	if classification.AlertType != investigation.AlertTypeMemory {
		t.Fatalf("alert type = %q", classification.AlertType)
	}
	if classification.Severity != investigation.SeverityCritical {
		t.Fatalf("severity = %q", classification.Severity)
	}
	if classification.Service != "checkout-api" {
		t.Fatalf("service = %q", classification.Service)
	}
	if classification.Priority != investigation.PriorityP1 {
		t.Fatalf("priority = %q", classification.Priority)
	}
}

func TestDeterministicClassifier_Classify_UsesExplicitPriorityLabel(t *testing.T) {
	classifier := NewDeterministicClassifier()
	incident := incidentWithAlert(investigation.Alert{
		ID:        "alert-1",
		Name:      "Checkout memory high",
		Type:      investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		Service:   "checkout-api",
		StartedAt: validTime(),
		Labels: map[string]string{
			"priority": "P0",
		},
	})

	classification, err := classifier.Classify(context.Background(), incident)
	if err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	if classification.Priority != investigation.PriorityP0 {
		t.Fatalf("priority = %q", classification.Priority)
	}
}

func TestDeterministicClassifier_Classify_MapsWarningToP2(t *testing.T) {
	classifier := NewDeterministicClassifier()
	incident := incidentWithAlert(investigation.Alert{
		ID:        "alert-1",
		Name:      "Checkout memory warning",
		Type:      investigation.AlertTypeMemory,
		Severity:  investigation.SeverityWarning,
		Service:   "checkout-api",
		StartedAt: validTime(),
	})

	classification, err := classifier.Classify(context.Background(), incident)
	if err != nil {
		t.Fatalf("Classify returned error: %v", err)
	}

	if classification.Priority != investigation.PriorityP2 {
		t.Fatalf("priority = %q", classification.Priority)
	}
}

func TestDeterministicClassifier_Classify_ReturnsValidationErrorForInvalidAlert(t *testing.T) {
	classifier := NewDeterministicClassifier()
	incident := incidentWithAlert(investigation.Alert{
		ID:        "alert-1",
		Name:      "Checkout memory high",
		Type:      investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		StartedAt: validTime(),
	})

	_, err := classifier.Classify(context.Background(), incident)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "validate alert") {
		t.Fatalf("error = %q, want validation context", err.Error())
	}
}

func TestDeterministicClassifier_Classify_ReturnsContextErrorWhenCanceled(t *testing.T) {
	classifier := NewDeterministicClassifier()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := classifier.Classify(ctx, incidentWithAlert(investigation.Alert{
		ID:        "alert-1",
		Name:      "Checkout memory high",
		Type:      investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		Service:   "checkout-api",
		StartedAt: validTime(),
	}))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func incidentWithAlert(alert investigation.Alert) investigation.Incident {
	return investigation.Incident{
		ID:    "inc-1",
		Alert: alert,
	}
}

func validTime() time.Time {
	return time.Date(2026, 6, 28, 9, 45, 0, 0, time.UTC)
}
