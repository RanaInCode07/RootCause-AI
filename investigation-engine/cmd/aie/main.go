package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"rootcause.ai/investigation-engine/engine"
	"rootcause.ai/investigation-engine/evaluator"
)

const (
	reportFilename     = "report.json"
	evaluationFilename = "evaluation.json"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "investigate":
		return runInvestigate(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runInvestigate(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("investigate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	outputDir := flags.String("output-dir", "", "directory where report.json and evaluation.json will be written")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "incident path is required")
		fmt.Fprintln(stderr, "usage: aie investigate [--output-dir DIR] <incident.json>")
		return 2
	}

	incidentPath := flags.Arg(0)
	result, err := engine.NewDefault().InvestigateFile(context.Background(), incidentPath)
	if err != nil {
		fmt.Fprintf(stderr, "investigate: %v\n", err)
		return 1
	}

	evaluation, err := evaluator.NewDeterministicEvaluator().Evaluate(context.Background(), result.Report, result.Incident.GroundTruth)
	if err != nil {
		fmt.Fprintf(stderr, "evaluate: %v\n", err)
		return 1
	}

	destination := *outputDir
	if destination == "" {
		destination = filepath.Join("outputs", safePathSegment(result.Incident.ID))
	}

	reportPath := filepath.Join(destination, reportFilename)
	evaluationPath := filepath.Join(destination, evaluationFilename)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		fmt.Fprintf(stderr, "create output directory: %v\n", err)
		return 1
	}
	if err := writeJSONFile(reportPath, result.Report); err != nil {
		fmt.Fprintf(stderr, "write report: %v\n", err)
		return 1
	}
	if err := writeJSONFile(evaluationPath, evaluation); err != nil {
		fmt.Fprintf(stderr, "write evaluation: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "incident_id=%s\n", result.Incident.ID)
	fmt.Fprintf(stdout, "root_cause=%s\n", result.Report.RootCause.Code)
	fmt.Fprintf(stdout, "confidence=%.2f\n", result.Report.RootCause.Confidence)
	fmt.Fprintf(stdout, "evaluation_status=%s\n", evaluation.Status)
	fmt.Fprintf(stdout, "report=%s\n", reportPath)
	fmt.Fprintf(stdout, "evaluation=%s\n", evaluationPath)
	return 0
}

func writeJSONFile(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

func safePathSegment(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "incident"
	}

	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteRune('-')
		}
	}
	return strings.Trim(builder.String(), "-")
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage:")
	fmt.Fprintln(writer, "  aie investigate [--output-dir DIR] <incident.json>")
}
