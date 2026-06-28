package investigation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIncidentFixture_ValidateOOMAfterDeployment(t *testing.T) {
	path := filepath.Join("..", "..", "dataset", "incidents", "oom_after_deployment.json")

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var incident Incident
	if err := json.Unmarshal(raw, &incident); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if err := incident.Validate(); err != nil {
		t.Fatalf("validate fixture: %v", err)
	}

	if incident.ID != "inc-oom-checkout-2026-06-28" {
		t.Fatalf("incident ID = %q", incident.ID)
	}
	if incident.Alert.Service != "checkout-api" {
		t.Fatalf("service = %q", incident.Alert.Service)
	}
	if incident.IncidentWindow.Start.IsZero() || incident.IncidentWindow.End.IsZero() {
		t.Fatal("expected incident window")
	}
	if incident.Deployment == nil {
		t.Fatal("expected deployment")
	}
	if incident.GroundTruth.RootCauseCode != RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("ground truth root cause = %q", incident.GroundTruth.RootCauseCode)
	}
	if len(incident.Metrics) == 0 {
		t.Fatal("expected metrics")
	}
	if len(incident.KubernetesEvents) == 0 {
		t.Fatal("expected kubernetes events")
	}
	if len(incident.Logs) == 0 {
		t.Fatal("expected logs")
	}
}

func TestIncidentJSONSchema_IsPresentAndValidJSON(t *testing.T) {
	path := filepath.Join("..", "..", "dataset", "schema", "incident.schema.json")

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	if schema["title"] != "AIE Incident" {
		t.Fatalf("schema title = %v", schema["title"])
	}
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("schema required field is missing or invalid")
	}
	for _, field := range []string{"id", "metadata", "incident_window", "alert", "deployment", "metrics", "kubernetes_events", "logs", "ground_truth"} {
		if !containsAnyString(required, field) {
			t.Fatalf("schema required fields missing %q", field)
		}
	}
}

func TestIncident_ValidateRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Incident)
		wantErr string
	}{
		{
			name: "missing incident id",
			mutate: func(incident *Incident) {
				incident.ID = ""
			},
			wantErr: "incident id is required",
		},
		{
			name: "missing incident window",
			mutate: func(incident *Incident) {
				incident.IncidentWindow = IncidentWindow{}
			},
			wantErr: "incident window: start is required",
		},
		{
			name: "missing alert service",
			mutate: func(incident *Incident) {
				incident.Alert.Service = ""
			},
			wantErr: "alert: service is required",
		},
		{
			name: "missing ground truth",
			mutate: func(incident *Incident) {
				incident.GroundTruth = GroundTruth{}
			},
			wantErr: "ground truth: root_cause_code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incident := validIncident()
			tt.mutate(&incident)

			err := incident.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestIncident_ValidateAllowsNullableDeploymentAndEmptyEvidenceArrays(t *testing.T) {
	incident := validIncident()
	incident.Deployment = nil
	incident.Metrics = []MetricSeries{}
	incident.KubernetesEvents = []KubernetesEvent{}
	incident.Logs = []LogEntry{}

	if err := incident.Validate(); err != nil {
		t.Fatalf("validate incident: %v", err)
	}
}

func TestMetricSeries_ValidateRejectsSeriesWithoutPoints(t *testing.T) {
	series := MetricSeries{
		ID:   "metric-memory",
		Name: "container_memory_usage_mb",
		Unit: "MB",
	}

	if err := series.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestReport_ValidateRequiresRootCausePrediction(t *testing.T) {
	report := Report{
		IncidentID: "inc-1",
		Summary:    "summary",
	}

	if err := report.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func validIncident() Incident {
	return Incident{
		ID:               "inc-1",
		IncidentWindow:   validIncidentWindow(),
		Alert:            validAlert(),
		Deployment:       validDeployment(),
		Metrics:          []MetricSeries{validMetricSeries()},
		KubernetesEvents: []KubernetesEvent{validKubernetesEvent()},
		Logs:             []LogEntry{validLogEntry()},
		GroundTruth:      validGroundTruth(),
	}
}

func validAlert() Alert {
	return Alert{
		ID:        "alert-1",
		Name:      "High memory",
		Type:      AlertTypeMemory,
		Severity:  SeverityCritical,
		Service:   "checkout-api",
		StartedAt: validTime(),
	}
}

func validIncidentWindow() IncidentWindow {
	return IncidentWindow{
		Start: validTime().Add(-10 * time.Minute),
		End:   validTime().Add(10 * time.Minute),
	}
}

func validDeployment() *Deployment {
	return &Deployment{
		ID:         "deploy-1",
		Service:    "checkout-api",
		Version:    "2026.06.28.1",
		DeployedAt: validTime(),
	}
}

func validMetricSeries() MetricSeries {
	return MetricSeries{
		ID:   "metric-memory",
		Name: "container_memory_usage_mb",
		Unit: "MB",
		Points: []MetricPoint{
			{
				Timestamp: validTime(),
				Value:     2048,
			},
		},
	}
}

func validKubernetesEvent() KubernetesEvent {
	return KubernetesEvent{
		ID:        "kube-event-1",
		Timestamp: validTime(),
		Type:      "Warning",
		Reason:    "OOMKilled",
		Pod:       "checkout-api-abc",
		Container: "checkout-api",
		Message:   "container exceeded memory limit",
	}
}

func validLogEntry() LogEntry {
	return LogEntry{
		ID:        "log-1",
		Timestamp: validTime(),
		Level:     "error",
		Service:   "checkout-api",
		Message:   "out of memory",
	}
}

func validGroundTruth() GroundTruth {
	return GroundTruth{
		RootCauseCode: RootCauseMemoryLeakAfterDeployment,
		Summary:       "A deployment introduced a memory leak.",
	}
}

func validTime() time.Time {
	return time.Date(2026, 6, 28, 9, 45, 0, 0, time.UTC)
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if got, ok := value.(string); ok && got == want {
			return true
		}
	}
	return false
}
