package report

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

type Clock func() time.Time

type Option func(*DeterministicGenerator)

type DeterministicGenerator struct {
	clock Clock
}

func NewDeterministicGenerator(options ...Option) DeterministicGenerator {
	generator := DeterministicGenerator{
		clock: time.Now,
	}
	for _, option := range options {
		option(&generator)
	}
	return generator
}

func WithClock(clock Clock) Option {
	return func(generator *DeterministicGenerator) {
		if clock != nil {
			generator.clock = clock
		}
	}
}

func (generator DeterministicGenerator) GenerateReport(ctx context.Context, incident investigation.Incident, classification investigation.Classification, evidence []investigation.Evidence, hypotheses []investigation.ScoredHypothesis) (investigation.Report, error) {
	if err := ctx.Err(); err != nil {
		return investigation.Report{}, err
	}
	if len(hypotheses) == 0 {
		return investigation.Report{}, fmt.Errorf("no scored hypotheses available for report")
	}

	ranked := rankedHypotheses(hypotheses)
	primary := ranked[0]
	alternatives := append([]investigation.ScoredHypothesis(nil), ranked[1:]...)

	report := investigation.Report{
		IncidentID:  incident.ID,
		GeneratedAt: generator.clock().UTC(),
		Summary:     reportSummary(incident, classification, primary),
		Timeline:    buildTimeline(incident, evidence),
		RootCause: investigation.RootCausePrediction{
			Code:                  primary.Hypothesis.RootCauseCode,
			Summary:               primary.Hypothesis.Summary,
			Confidence:            primary.Confidence,
			SupportingEvidenceIDs: cloneStrings(primary.Hypothesis.SupportingEvidenceIDs),
			MissingEvidence:       cloneStrings(primary.MissingEvidence),
		},
		Evidence:       cloneEvidence(evidence),
		Alternatives:   cloneScoredHypotheses(alternatives),
		Recommendation: recommendation(incident, primary),
	}

	if err := report.Validate(); err != nil {
		return investigation.Report{}, fmt.Errorf("validate report: %w", err)
	}
	return report, nil
}

func rankedHypotheses(hypotheses []investigation.ScoredHypothesis) []investigation.ScoredHypothesis {
	ranked := append([]investigation.ScoredHypothesis(nil), hypotheses...)
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].Confidence > ranked[j].Confidence
	})
	return ranked
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func cloneEvidence(values []investigation.Evidence) []investigation.Evidence {
	if len(values) == 0 {
		return []investigation.Evidence{}
	}
	return append([]investigation.Evidence(nil), values...)
}

func cloneScoredHypotheses(values []investigation.ScoredHypothesis) []investigation.ScoredHypothesis {
	if len(values) == 0 {
		return []investigation.ScoredHypothesis{}
	}
	return append([]investigation.ScoredHypothesis(nil), values...)
}

func reportSummary(incident investigation.Incident, classification investigation.Classification, primary investigation.ScoredHypothesis) string {
	service := classification.Service
	if service == "" {
		service = incident.Alert.Service
	}
	return fmt.Sprintf(
		"%s investigation identified %s as the leading root cause with %.2f confidence.",
		service,
		primary.Hypothesis.RootCauseCode,
		primary.Confidence,
	)
}

func buildTimeline(incident investigation.Incident, evidence []investigation.Evidence) []investigation.TimelineEvent {
	timeline := make([]investigation.TimelineEvent, 0, len(evidence)+3)
	if incident.Deployment != nil && !incident.Deployment.DeployedAt.IsZero() {
		timeline = append(timeline, investigation.TimelineEvent{
			Timestamp: incident.Deployment.DeployedAt,
			Source:    "deployment",
			Summary:   deploymentTimelineSummary(incident.Deployment),
		})
	}
	if !incident.Alert.StartedAt.IsZero() {
		timeline = append(timeline, investigation.TimelineEvent{
			Timestamp: incident.Alert.StartedAt,
			Source:    "alert",
			Summary:   incident.Alert.Name,
		})
	}
	if !incident.Metadata.DetectedAt.IsZero() {
		timeline = append(timeline, investigation.TimelineEvent{
			Timestamp: incident.Metadata.DetectedAt,
			Source:    "incident",
			Summary:   "incident detected",
		})
	}
	for _, item := range evidence {
		if item.ObservedAt.IsZero() {
			continue
		}
		timeline = append(timeline, investigation.TimelineEvent{
			Timestamp: item.ObservedAt,
			Source:    "evidence",
			Summary:   item.Summary,
		})
	}

	sort.SliceStable(timeline, func(i, j int) bool {
		return timeline[i].Timestamp.Before(timeline[j].Timestamp)
	})
	return timeline
}

func deploymentTimelineSummary(deployment *investigation.Deployment) string {
	parts := []string{"deployment"}
	if deployment == nil {
		return strings.Join(parts, " ")
	}
	if deployment.Service != "" {
		parts = append(parts, deployment.Service)
	}
	if deployment.Version != "" {
		parts = append(parts, deployment.Version)
	}
	return strings.Join(parts, " ")
}

func recommendation(incident investigation.Incident, primary investigation.ScoredHypothesis) string {
	service := incident.Alert.Service
	if service == "" && incident.Deployment != nil {
		service = incident.Deployment.Service
	}
	if service == "" {
		service = "the affected service"
	}

	if primary.Confidence >= 0.8 {
		return fmt.Sprintf("Rollback or disable the recent %s deployment and verify memory usage, pod restarts, and 5xx rate return to baseline.", service)
	}

	missing := strings.Join(primary.MissingEvidence, ", ")
	if missing == "" {
		missing = "additional supporting evidence"
	}
	return fmt.Sprintf("Collect %s for %s before taking corrective action; confidence is %.2f.", missing, service, primary.Confidence)
}
