package loader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestJSONFileLoader_LoadIncident_LoadsAndValidatesFixture(t *testing.T) {
	var _ investigation.IncidentLoader = JSONFileLoader{}

	loader := NewJSONFileLoader()
	path := filepath.Join("..", "..", "dataset", "incidents", "oom_after_deployment.json")

	incident, err := loader.LoadIncident(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadIncident returned error: %v", err)
	}

	if incident.ID != "inc-oom-checkout-2026-06-28" {
		t.Fatalf("incident ID = %q", incident.ID)
	}
	if incident.Alert.Service != "checkout-api" {
		t.Fatalf("service = %q", incident.Alert.Service)
	}
	if incident.GroundTruth.RootCauseCode != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("ground truth = %q", incident.GroundTruth.RootCauseCode)
	}
}

func TestJSONFileLoader_LoadIncident_ReturnsValidationError(t *testing.T) {
	loader := NewJSONFileLoader()
	path := writeTempFile(t, `{"id":"inc-1"}`)

	_, err := loader.LoadIncident(context.Background(), path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "validate incident") {
		t.Fatalf("error = %q, want validation context", err.Error())
	}
}

func TestJSONFileLoader_LoadIncident_RejectsUnknownFields(t *testing.T) {
	loader := NewJSONFileLoader()
	path := writeTempFile(t, `{"id":"inc-1","unexpected":true}`)

	_, err := loader.LoadIncident(context.Background(), path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "decode incident") {
		t.Fatalf("error = %q, want decode context", err.Error())
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("error = %q, want unknown field context", err.Error())
	}
}

func TestJSONFileLoader_LoadIncident_ReturnsContextErrorWhenCanceled(t *testing.T) {
	loader := NewJSONFileLoader()
	path := filepath.Join("..", "..", "dataset", "incidents", "oom_after_deployment.json")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := loader.LoadIncident(ctx, path)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestJSONFileLoader_LoadIncident_ReturnsReadErrorForMissingFile(t *testing.T) {
	loader := NewJSONFileLoader()

	_, err := loader.LoadIncident(context.Background(), "does-not-exist.json")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "read incident file") {
		t.Fatalf("error = %q, want read context", err.Error())
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "incident.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
