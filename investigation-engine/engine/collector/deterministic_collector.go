package collector

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

const recentDeploymentWindow = time.Hour

type DeterministicCollector struct{}

func NewDeterministicCollector() DeterministicCollector {
	return DeterministicCollector{}
}

func (DeterministicCollector) CollectEvidence(ctx context.Context, incident investigation.Incident, classification investigation.Classification) ([]investigation.Evidence, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	evidence := make([]investigation.Evidence, 0, 8)

	if item, ok := collectRecentDeployment(incident, classification); ok {
		evidence = append(evidence, item)
	}

	for _, series := range incident.Metrics {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if item, ok := collectMetricEvidence(series); ok {
			evidence = append(evidence, item)
		}
	}

	for _, event := range incident.KubernetesEvents {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if item, ok := collectOOMKilledEvent(event, classification); ok {
			evidence = append(evidence, item)
		}
	}

	for _, log := range incident.Logs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if item, ok := collectRelevantLog(log, classification); ok {
			evidence = append(evidence, item)
		}
	}

	return evidence, nil
}

func collectRecentDeployment(incident investigation.Incident, classification investigation.Classification) (investigation.Evidence, bool) {
	if incident.Deployment.DeployedAt.IsZero() {
		return investigation.Evidence{}, false
	}
	if !sameService(incident.Deployment.Service, classification.Service) {
		return investigation.Evidence{}, false
	}

	alertTime := incident.Alert.StartedAt
	if alertTime.IsZero() {
		alertTime = incident.Metadata.DetectedAt
	}
	if alertTime.IsZero() {
		return investigation.Evidence{}, false
	}

	age := alertTime.Sub(incident.Deployment.DeployedAt)
	if age < 0 || age > recentDeploymentWindow {
		return investigation.Evidence{}, false
	}

	minutes := int(age.Minutes())
	return investigation.Evidence{
		ID:         "evidence-recent-deployment-" + incident.Deployment.ID,
		Type:       investigation.EvidenceTypeDeployment,
		Signal:     investigation.EvidenceSignalRecentDeployment,
		Source:     investigation.EvidenceSource{Kind: "deployment", ID: incident.Deployment.ID},
		ObservedAt: incident.Deployment.DeployedAt,
		Summary: fmt.Sprintf("%s version %s was deployed %d minutes before the alert",
			incident.Deployment.Service,
			incident.Deployment.Version,
			minutes,
		),
		Attributes: map[string]string{
			"service":              incident.Deployment.Service,
			"version":              incident.Deployment.Version,
			"previous_version":     incident.Deployment.PreviousVersion,
			"minutes_before_alert": strconv.Itoa(minutes),
		},
	}, true
}

func collectMetricEvidence(series investigation.MetricSeries) (investigation.Evidence, bool) {
	points := sortedMetricPoints(series.Points)
	if len(points) < 2 {
		return investigation.Evidence{}, false
	}

	start := points[0]
	end := points[len(points)-1]
	delta := end.Value - start.Value
	if delta <= 0 {
		return investigation.Evidence{}, false
	}

	name := strings.ToLower(series.Name)
	switch {
	case strings.Contains(name, "memory") && isSignificantIncrease(start.Value, end.Value, 512, 1.5):
		return metricEvidence(
			series,
			investigation.EvidenceSignalMemoryIncrease,
			start,
			end,
		), true
	case strings.Contains(name, "restart") && end.Value > start.Value:
		return metricEvidence(
			series,
			investigation.EvidenceSignalRestartCountIncrease,
			start,
			end,
		), true
	case isHTTP5xxMetric(name) && isSignificantIncrease(start.Value, end.Value, 10, 2):
		return metricEvidence(
			series,
			investigation.EvidenceSignalHTTP5xxSpike,
			start,
			end,
		), true
	default:
		return investigation.Evidence{}, false
	}
}

func collectOOMKilledEvent(event investigation.KubernetesEvent, classification investigation.Classification) (investigation.Evidence, bool) {
	if !strings.EqualFold(event.Reason, "OOMKilled") {
		return investigation.Evidence{}, false
	}
	if !eventMatchesService(event, classification.Service) {
		return investigation.Evidence{}, false
	}

	return investigation.Evidence{
		ID:         "evidence-oom-killed-" + event.ID,
		Type:       investigation.EvidenceTypeKubernetesEvent,
		Signal:     investigation.EvidenceSignalOOMKilled,
		Source:     investigation.EvidenceSource{Kind: "kubernetes_event", ID: event.ID},
		ObservedAt: event.Timestamp,
		Summary:    fmt.Sprintf("container %s was OOMKilled in pod %s", event.Container, event.Pod),
		Attributes: map[string]string{
			"reason":        event.Reason,
			"pod":           event.Pod,
			"container":     event.Container,
			"restart_count": strconv.Itoa(event.RestartCount),
		},
	}, true
}

func collectRelevantLog(log investigation.LogEntry, classification investigation.Classification) (investigation.Evidence, bool) {
	if !sameService(log.Service, classification.Service) {
		return investigation.Evidence{}, false
	}

	message := strings.ToLower(log.Message)
	matchedTerm := ""
	for _, term := range []string{"out-of-memory", "out of memory", "oom"} {
		if strings.Contains(message, term) {
			matchedTerm = term
			break
		}
	}
	if matchedTerm == "" {
		return investigation.Evidence{}, false
	}

	return investigation.Evidence{
		ID:         "evidence-relevant-log-" + log.ID,
		Type:       investigation.EvidenceTypeLog,
		Signal:     investigation.EvidenceSignalRelevantLog,
		Source:     investigation.EvidenceSource{Kind: "log", ID: log.ID},
		ObservedAt: log.Timestamp,
		Summary:    log.Message,
		Attributes: map[string]string{
			"level":        log.Level,
			"service":      log.Service,
			"matched_term": matchedTerm,
		},
	}, true
}

func metricEvidence(series investigation.MetricSeries, signal investigation.EvidenceSignal, start, end investigation.MetricPoint) investigation.Evidence {
	delta := end.Value - start.Value
	increasePercent := percentageIncrease(start.Value, end.Value)
	return investigation.Evidence{
		ID:         "evidence-" + string(signal) + "-" + series.ID,
		Type:       investigation.EvidenceTypeMetric,
		Signal:     signal,
		Source:     investigation.EvidenceSource{Kind: "metric", ID: series.ID},
		ObservedAt: end.Timestamp,
		Summary:    metricSummary(series, signal, start, end),
		Attributes: map[string]string{
			"metric_name":      series.Name,
			"unit":             series.Unit,
			"start_value":      formatNumber(start.Value),
			"end_value":        formatNumber(end.Value),
			"delta":            formatNumber(delta),
			"increase_percent": formatFixed(increasePercent, 2),
		},
	}
}

func metricSummary(series investigation.MetricSeries, signal investigation.EvidenceSignal, start, end investigation.MetricPoint) string {
	startValue := formatNumber(start.Value)
	endValue := formatNumber(end.Value)

	switch signal {
	case investigation.EvidenceSignalMemoryIncrease:
		return fmt.Sprintf("memory usage increased from %s %s to %s %s", startValue, series.Unit, endValue, series.Unit)
	case investigation.EvidenceSignalRestartCountIncrease:
		return fmt.Sprintf("restart count increased from %s to %s", startValue, endValue)
	case investigation.EvidenceSignalHTTP5xxSpike:
		return fmt.Sprintf("5xx rate increased from %s %s to %s %s", startValue, series.Unit, endValue, series.Unit)
	default:
		return fmt.Sprintf("%s increased from %s %s to %s %s", series.Name, startValue, series.Unit, endValue, series.Unit)
	}
}

func sortedMetricPoints(points []investigation.MetricPoint) []investigation.MetricPoint {
	sorted := append([]investigation.MetricPoint(nil), points...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})
	return sorted
}

func isSignificantIncrease(start, end, minDelta, minRatio float64) bool {
	if end-start < minDelta {
		return false
	}
	if start <= 0 {
		return end > 0
	}
	return end/start >= minRatio
}

func isHTTP5xxMetric(name string) bool {
	return strings.Contains(name, "5xx") || strings.Contains(name, "status_5")
}

func percentageIncrease(start, end float64) float64 {
	if start <= 0 {
		return 0
	}
	return ((end - start) / start) * 100
}

func eventMatchesService(event investigation.KubernetesEvent, service string) bool {
	if strings.TrimSpace(service) == "" {
		return true
	}
	return sameService(event.Container, service) || strings.Contains(strings.ToLower(event.Pod), strings.ToLower(service))
}

func sameService(left, right string) bool {
	if strings.TrimSpace(right) == "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatFixed(value float64, precision int) string {
	return strconv.FormatFloat(value, 'f', precision, 64)
}
