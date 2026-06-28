package collector

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestDeterministicCollector_CollectEvidence_ExtractsOOMFixtureSignals(t *testing.T) {
	var _ investigation.EvidenceCollector = DeterministicCollector{}

	collector := NewDeterministicCollector()
	incident := loadFixture(t)
	classification := fixtureClassification()

	evidence, err := collector.CollectEvidence(context.Background(), incident, classification)
	if err != nil {
		t.Fatalf("CollectEvidence returned error: %v", err)
	}

	wantSignals := []investigation.EvidenceSignal{
		investigation.EvidenceSignalRecentDeployment,
		investigation.EvidenceSignalMemoryIncrease,
		investigation.EvidenceSignalOOMKilled,
		investigation.EvidenceSignalRestartCountIncrease,
		investigation.EvidenceSignalHTTP5xxSpike,
		investigation.EvidenceSignalRelevantLog,
	}
	for _, signal := range wantSignals {
		if !hasSignal(evidence, signal) {
			t.Fatalf("expected evidence signal %q in %#v", signal, evidence)
		}
	}
	for _, item := range evidence {
		if strings.Contains(item.Summary, "%!") {
			t.Fatalf("evidence summary contains formatter artifact: %q", item.Summary)
		}
	}

	recentDeployment := evidenceBySignal(t, evidence, investigation.EvidenceSignalRecentDeployment)
	if recentDeployment.Type != investigation.EvidenceTypeDeployment {
		t.Fatalf("recent deployment evidence type = %q", recentDeployment.Type)
	}
	if recentDeployment.Source.ID != "deploy-checkout-api-20260628-0915" {
		t.Fatalf("deployment evidence source = %q", recentDeployment.Source.ID)
	}
	if recentDeployment.Attributes["minutes_before_alert"] != "27" {
		t.Fatalf("minutes_before_alert = %q", recentDeployment.Attributes["minutes_before_alert"])
	}

	memoryIncrease := evidenceBySignal(t, evidence, investigation.EvidenceSignalMemoryIncrease)
	if memoryIncrease.Source.ID != "metric-checkout-memory" {
		t.Fatalf("memory evidence source = %q", memoryIncrease.Source.ID)
	}
	if memoryIncrease.Attributes["start_value"] != "418" {
		t.Fatalf("memory start = %q", memoryIncrease.Attributes["start_value"])
	}
	if memoryIncrease.Attributes["end_value"] != "2048" {
		t.Fatalf("memory end = %q", memoryIncrease.Attributes["end_value"])
	}

	oomKilled := evidenceBySignal(t, evidence, investigation.EvidenceSignalOOMKilled)
	if oomKilled.Source.ID != "kube-event-checkout-oom-1" {
		t.Fatalf("oom evidence source = %q", oomKilled.Source.ID)
	}
	if oomKilled.Attributes["restart_count"] != "1" {
		t.Fatalf("oom restart_count = %q", oomKilled.Attributes["restart_count"])
	}

	restartIncrease := evidenceBySignal(t, evidence, investigation.EvidenceSignalRestartCountIncrease)
	if restartIncrease.Source.ID != "metric-checkout-restarts" {
		t.Fatalf("restart evidence source = %q", restartIncrease.Source.ID)
	}
	if restartIncrease.Attributes["start_value"] != "0" || restartIncrease.Attributes["end_value"] != "3" {
		t.Fatalf("restart attributes = %#v", restartIncrease.Attributes)
	}

	http5xxSpike := evidenceBySignal(t, evidence, investigation.EvidenceSignalHTTP5xxSpike)
	if http5xxSpike.Source.ID != "metric-checkout-5xx-rate" {
		t.Fatalf("5xx evidence source = %q", http5xxSpike.Source.ID)
	}

	relevantLog := evidenceBySignal(t, evidence, investigation.EvidenceSignalRelevantLog)
	if relevantLog.Source.ID != "log-checkout-oom" {
		t.Fatalf("log evidence source = %q", relevantLog.Source.ID)
	}
}

func TestDeterministicCollector_CollectEvidence_DoesNotEmitOldDeployment(t *testing.T) {
	collector := NewDeterministicCollector()
	incident := loadFixture(t)
	incident.Deployment.DeployedAt = incident.Alert.StartedAt.Add(-3 * time.Hour)

	evidence, err := collector.CollectEvidence(context.Background(), incident, fixtureClassification())
	if err != nil {
		t.Fatalf("CollectEvidence returned error: %v", err)
	}

	if hasSignal(evidence, investigation.EvidenceSignalRecentDeployment) {
		t.Fatalf("did not expect recent deployment evidence in %#v", evidence)
	}
}

func TestDeterministicCollector_CollectEvidence_SortsMetricPointsByTimestamp(t *testing.T) {
	collector := NewDeterministicCollector()
	incident := investigation.Incident{
		ID:    "inc-1",
		Alert: validAlert(),
		Metrics: []investigation.MetricSeries{
			{
				ID:   "metric-memory",
				Name: "container_memory_usage_mb",
				Unit: "MB",
				Points: []investigation.MetricPoint{
					{Timestamp: validTime().Add(10 * time.Minute), Value: 1200},
					{Timestamp: validTime(), Value: 300},
				},
			},
		},
	}

	evidence, err := collector.CollectEvidence(context.Background(), incident, fixtureClassification())
	if err != nil {
		t.Fatalf("CollectEvidence returned error: %v", err)
	}

	memoryIncrease := evidenceBySignal(t, evidence, investigation.EvidenceSignalMemoryIncrease)
	if memoryIncrease.Attributes["start_value"] != "300" {
		t.Fatalf("start_value = %q", memoryIncrease.Attributes["start_value"])
	}
	if memoryIncrease.Attributes["end_value"] != "1200" {
		t.Fatalf("end_value = %q", memoryIncrease.Attributes["end_value"])
	}
}

func TestDeterministicCollector_CollectEvidence_DetectsPercentageMemoryIncrease(t *testing.T) {
	collector := NewDeterministicCollector()
	incident := investigation.Incident{
		ID:    "inc-1",
		Alert: validAlert(),
		Metrics: []investigation.MetricSeries{
			{
				ID:   "metric-memory-percent",
				Name: "memory_usage",
				Unit: "%",
				Points: []investigation.MetricPoint{
					{Timestamp: validTime(), Value: 61},
					{Timestamp: validTime().Add(5 * time.Minute), Value: 97},
				},
			},
		},
	}

	evidence, err := collector.CollectEvidence(context.Background(), incident, fixtureClassification())
	if err != nil {
		t.Fatalf("CollectEvidence returned error: %v", err)
	}

	memoryIncrease := evidenceBySignal(t, evidence, investigation.EvidenceSignalMemoryIncrease)
	if memoryIncrease.Source.ID != "metric-memory-percent" {
		t.Fatalf("memory source = %q", memoryIncrease.Source.ID)
	}
	if memoryIncrease.Attributes["start_value"] != "61" || memoryIncrease.Attributes["end_value"] != "97" {
		t.Fatalf("memory attributes = %#v", memoryIncrease.Attributes)
	}
}

func TestDeterministicCollector_CollectEvidence_DetectsRestartFromBackOffEvent(t *testing.T) {
	collector := NewDeterministicCollector()
	incident := investigation.Incident{
		ID:    "inc-1",
		Alert: validAlert(),
		KubernetesEvents: []investigation.KubernetesEvent{
			{
				ID:           "kube-backoff-1",
				Timestamp:    validTime(),
				Type:         "Warning",
				Reason:       "BackOff",
				Pod:          "checkout-api-abc",
				Container:    "checkout-api",
				Message:      "Back-off restarting failed container.",
				RestartCount: 3,
			},
		},
	}

	evidence, err := collector.CollectEvidence(context.Background(), incident, fixtureClassification())
	if err != nil {
		t.Fatalf("CollectEvidence returned error: %v", err)
	}

	restartIncrease := evidenceBySignal(t, evidence, investigation.EvidenceSignalRestartCountIncrease)
	if restartIncrease.Source.ID != "kube-backoff-1" {
		t.Fatalf("restart source = %q", restartIncrease.Source.ID)
	}
	if restartIncrease.Attributes["restart_count"] != "3" {
		t.Fatalf("restart attributes = %#v", restartIncrease.Attributes)
	}
}

func TestDeterministicCollector_CollectEvidence_ReturnsContextErrorWhenCanceled(t *testing.T) {
	collector := NewDeterministicCollector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := collector.CollectEvidence(ctx, loadFixture(t), fixtureClassification())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func loadFixture(t *testing.T) investigation.Incident {
	t.Helper()

	path := filepath.Join("..", "..", "dataset", "incidents", "oom_after_deployment.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var incident investigation.Incident
	if err := json.Unmarshal(raw, &incident); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return incident
}

func fixtureClassification() investigation.Classification {
	return investigation.Classification{
		AlertType: investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		Service:   "checkout-api",
		Priority:  investigation.PriorityP1,
	}
}

func validAlert() investigation.Alert {
	return investigation.Alert{
		ID:        "alert-1",
		Name:      "Checkout memory high",
		Type:      investigation.AlertTypeMemory,
		Severity:  investigation.SeverityCritical,
		Service:   "checkout-api",
		StartedAt: validTime(),
	}
}

func validTime() time.Time {
	return time.Date(2026, 6, 28, 9, 45, 0, 0, time.UTC)
}

func hasSignal(evidence []investigation.Evidence, signal investigation.EvidenceSignal) bool {
	for _, item := range evidence {
		if item.Signal == signal {
			return true
		}
	}
	return false
}

func evidenceBySignal(t *testing.T, evidence []investigation.Evidence, signal investigation.EvidenceSignal) investigation.Evidence {
	t.Helper()

	for _, item := range evidence {
		if item.Signal == signal {
			return item
		}
	}
	t.Fatalf("missing evidence signal %q in %#v", signal, evidence)
	return investigation.Evidence{}
}
