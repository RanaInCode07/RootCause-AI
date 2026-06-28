package investigation

import (
	"fmt"
	"strings"
)

func (i Incident) Validate() error {
	if blank(i.ID) {
		return fmt.Errorf("incident id is required")
	}
	if err := i.Alert.Validate(); err != nil {
		return fmt.Errorf("alert: %w", err)
	}
	if err := i.Deployment.Validate(); err != nil {
		return fmt.Errorf("deployment: %w", err)
	}
	if len(i.Metrics) == 0 {
		return fmt.Errorf("metrics are required")
	}
	for _, series := range i.Metrics {
		if err := series.Validate(); err != nil {
			return fmt.Errorf("metric series %q: %w", series.ID, err)
		}
	}
	if len(i.KubernetesEvents) == 0 {
		return fmt.Errorf("kubernetes events are required")
	}
	if len(i.Logs) == 0 {
		return fmt.Errorf("logs are required")
	}
	if err := i.GroundTruth.Validate(); err != nil {
		return fmt.Errorf("ground truth: %w", err)
	}
	return nil
}

func (a Alert) Validate() error {
	if blank(a.ID) {
		return fmt.Errorf("id is required")
	}
	if blank(a.Name) {
		return fmt.Errorf("name is required")
	}
	if blank(string(a.Type)) {
		return fmt.Errorf("type is required")
	}
	if blank(string(a.Severity)) {
		return fmt.Errorf("severity is required")
	}
	if blank(a.Service) {
		return fmt.Errorf("service is required")
	}
	if a.StartedAt.IsZero() {
		return fmt.Errorf("started_at is required")
	}
	return nil
}

func (d Deployment) Validate() error {
	if blank(d.ID) {
		return fmt.Errorf("id is required")
	}
	if blank(d.Service) {
		return fmt.Errorf("service is required")
	}
	if blank(d.Version) {
		return fmt.Errorf("version is required")
	}
	if d.DeployedAt.IsZero() {
		return fmt.Errorf("deployed_at is required")
	}
	return nil
}

func (m MetricSeries) Validate() error {
	if blank(m.ID) {
		return fmt.Errorf("id is required")
	}
	if blank(m.Name) {
		return fmt.Errorf("name is required")
	}
	if blank(m.Unit) {
		return fmt.Errorf("unit is required")
	}
	if len(m.Points) == 0 {
		return fmt.Errorf("points are required")
	}
	for idx, point := range m.Points {
		if err := point.Validate(); err != nil {
			return fmt.Errorf("point %d: %w", idx, err)
		}
	}
	return nil
}

func (m MetricPoint) Validate() error {
	if m.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}

func (g GroundTruth) Validate() error {
	if blank(string(g.RootCauseCode)) {
		return fmt.Errorf("root_cause_code is required")
	}
	if blank(g.Summary) {
		return fmt.Errorf("summary is required")
	}
	return nil
}

func (r Report) Validate() error {
	if blank(r.IncidentID) {
		return fmt.Errorf("incident_id is required")
	}
	if blank(r.Summary) {
		return fmt.Errorf("summary is required")
	}
	if blank(string(r.RootCause.Code)) {
		return fmt.Errorf("root_cause code is required")
	}
	return nil
}

func blank(value string) bool {
	return strings.TrimSpace(value) == ""
}
