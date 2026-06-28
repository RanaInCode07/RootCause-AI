package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

func TestRunInvestigate_WritesReportAndEvaluation(t *testing.T) {
	outputDir := t.TempDir()
	incidentPath := filepath.Join("..", "..", "dataset", "incidents", "oom_after_deployment.json")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"investigate", "--output-dir", outputDir, incidentPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}

	reportPath := filepath.Join(outputDir, "report.json")
	evaluationPath := filepath.Join(outputDir, "evaluation.json")

	report := readJSON[investigation.Report](t, reportPath)
	if report.RootCause.Code != investigation.RootCauseMemoryLeakAfterDeployment {
		t.Fatalf("root cause = %q", report.RootCause.Code)
	}
	if report.RootCause.Confidence != 1 {
		t.Fatalf("confidence = %v", report.RootCause.Confidence)
	}

	evaluation := readJSON[investigation.EvaluationResult](t, evaluationPath)
	if evaluation.Status != investigation.EvaluationStatusPass {
		t.Fatalf("evaluation status = %q", evaluation.Status)
	}
	if len(evaluation.MissingEvidenceIDs) != 0 {
		t.Fatalf("missing evidence = %#v", evaluation.MissingEvidenceIDs)
	}

	output := stdout.String()
	if !strings.Contains(output, reportPath) {
		t.Fatalf("stdout = %q, want report path", output)
	}
	if !strings.Contains(output, evaluationPath) {
		t.Fatalf("stdout = %q, want evaluation path", output)
	}
}

func TestRunInvestigate_RequiresIncidentPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"investigate"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "incident path is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRun_ReturnsUsageForUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"unknown"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func readJSON[T any](t *testing.T, path string) T {
	t.Helper()

	raw := mustReadFile(t, path)
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return value
}
