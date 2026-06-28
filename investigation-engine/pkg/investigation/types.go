package investigation

import "time"

type AlertType string

const (
	AlertTypeMemory AlertType = "memory"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
)

type Priority string

const (
	PriorityP0 Priority = "P0"
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
)

type RootCauseCode string

const (
	RootCauseMemoryLeakAfterDeployment RootCauseCode = "memory_leak_after_deployment"
)

type EvidenceType string

const (
	EvidenceTypeDeployment      EvidenceType = "deployment"
	EvidenceTypeMetric          EvidenceType = "metric"
	EvidenceTypeKubernetesEvent EvidenceType = "kubernetes_event"
	EvidenceTypeLog             EvidenceType = "log"
)

type Incident struct {
	ID               string            `json:"id"`
	Metadata         IncidentMetadata  `json:"metadata"`
	Alert            Alert             `json:"alert"`
	Deployment       Deployment        `json:"deployment"`
	Metrics          []MetricSeries    `json:"metrics"`
	KubernetesEvents []KubernetesEvent `json:"kubernetes_events"`
	Logs             []LogEntry        `json:"logs"`
	GroundTruth      GroundTruth       `json:"ground_truth"`
}

type IncidentMetadata struct {
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Environment string    `json:"environment"`
	Region      string    `json:"region"`
	DetectedAt  time.Time `json:"detected_at"`
}

type Alert struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      AlertType         `json:"type"`
	Severity  Severity          `json:"severity"`
	Service   string            `json:"service"`
	StartedAt time.Time         `json:"started_at"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type Deployment struct {
	ID              string    `json:"id"`
	Service         string    `json:"service"`
	Version         string    `json:"version"`
	PreviousVersion string    `json:"previous_version"`
	CommitSHA       string    `json:"commit_sha"`
	Author          string    `json:"author"`
	DeployedAt      time.Time `json:"deployed_at"`
	Summary         string    `json:"summary"`
}

type MetricSeries struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Unit   string        `json:"unit"`
	Points []MetricPoint `json:"points"`
}

type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type KubernetesEvent struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`
	Reason       string    `json:"reason"`
	Namespace    string    `json:"namespace"`
	Pod          string    `json:"pod"`
	Container    string    `json:"container"`
	Message      string    `json:"message"`
	RestartCount int       `json:"restart_count"`
}

type LogEntry struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	Level      string            `json:"level"`
	Service    string            `json:"service"`
	Message    string            `json:"message"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type GroundTruth struct {
	RootCauseCode RootCauseCode `json:"root_cause_code"`
	Summary       string        `json:"summary"`
	EvidenceIDs   []string      `json:"evidence_ids"`
}

type Classification struct {
	AlertType AlertType `json:"alert_type"`
	Severity  Severity  `json:"severity"`
	Service   string    `json:"service"`
	Priority  Priority  `json:"priority"`
}

type Evidence struct {
	ID         string            `json:"id"`
	Type       EvidenceType      `json:"type"`
	Source     EvidenceSource    `json:"source"`
	ObservedAt time.Time         `json:"observed_at"`
	Summary    string            `json:"summary"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type EvidenceSource struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type Hypothesis struct {
	ID                    string        `json:"id"`
	RootCauseCode         RootCauseCode `json:"root_cause_code"`
	Summary               string        `json:"summary"`
	SupportingEvidenceIDs []string      `json:"supporting_evidence_ids"`
	MissingEvidence       []string      `json:"missing_evidence"`
}

type ScoredHypothesis struct {
	Hypothesis         Hypothesis `json:"hypothesis"`
	Confidence         float64    `json:"confidence"`
	SupportingEvidence []Evidence `json:"supporting_evidence"`
	MissingEvidence    []string   `json:"missing_evidence"`
}

type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Summary   string    `json:"summary"`
}

type RootCausePrediction struct {
	Code                  RootCauseCode `json:"code"`
	Summary               string        `json:"summary"`
	Confidence            float64       `json:"confidence"`
	SupportingEvidenceIDs []string      `json:"supporting_evidence_ids"`
	MissingEvidence       []string      `json:"missing_evidence"`
}

type Report struct {
	IncidentID     string              `json:"incident_id"`
	GeneratedAt    time.Time           `json:"generated_at"`
	Summary        string              `json:"summary"`
	Timeline       []TimelineEvent     `json:"timeline"`
	RootCause      RootCausePrediction `json:"root_cause"`
	Evidence       []Evidence          `json:"evidence"`
	Alternatives   []ScoredHypothesis  `json:"alternative_hypotheses"`
	Recommendation string              `json:"recommendation"`
}

type EvaluationStatus string

const (
	EvaluationStatusPass EvaluationStatus = "PASS"
	EvaluationStatusFail EvaluationStatus = "FAIL"
)

type EvaluationResult struct {
	Status                 EvaluationStatus `json:"status"`
	ExpectedRootCauseCode  RootCauseCode    `json:"expected_root_cause_code"`
	PredictedRootCauseCode RootCauseCode    `json:"predicted_root_cause_code"`
	Confidence             float64          `json:"confidence"`
	MatchedEvidenceIDs     []string         `json:"matched_evidence_ids"`
	MissingEvidenceIDs     []string         `json:"missing_evidence_ids"`
}

type ReasoningInput struct {
	IncidentID     string         `json:"incident_id"`
	Classification Classification `json:"classification"`
	Evidence       []Evidence     `json:"evidence"`
	Hypotheses     []Hypothesis   `json:"hypotheses"`
}

type ReasoningResult struct {
	Notes               string       `json:"notes"`
	SuggestedHypotheses []Hypothesis `json:"suggested_hypotheses"`
}

type SummaryInput struct {
	Report Report `json:"report"`
}

type SummaryResult struct {
	Text string `json:"text"`
}

type PlanningInput struct {
	Incident       Incident       `json:"incident"`
	Classification Classification `json:"classification"`
	Evidence       []Evidence     `json:"evidence"`
}

type PlanningResult struct {
	RequiredEvidence []string `json:"required_evidence"`
	NextSteps        []string `json:"next_steps"`
}
