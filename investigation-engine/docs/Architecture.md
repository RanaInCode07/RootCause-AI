# Architecture

## System Boundary

AIE V1 is a local Go program. It receives one simulated incident, investigates it through deterministic stages, emits a JSON report, and evaluates the result against ground truth.

The boundary is deliberately narrow:

- Input: incident JSON from `dataset/`.
- Output: investigation report JSON and evaluator result.
- No external network calls.
- No persistence layer.
- No LLM calls.
- No real infrastructure integrations.

This keeps the core engine measurable before adapters for real systems are introduced.

## High-Level Components

### Incident Dataset

The dataset contains predefined incident JSON files for a fictional company. Each incident includes alert metadata, deployment context, metric samples, Kubernetes-like events, logs, and ground truth.

Tradeoff: static fixtures are less realistic than a simulator, but they make correctness repeatable and keep V1 focused on investigation quality.

### Investigation Engine

The engine is a pipeline with explicit stages:

1. Classifier
2. Evidence collector
3. Hypothesis generator
4. Confidence scorer
5. Report generator

Each stage receives structured input and returns structured output. The engine orchestrates stages but does not hide their decisions.

Tradeoff: a pipeline is less flexible than an agent planner, but it is easier to test, explain, and evaluate. Agentic planning can be added later behind a `ReasoningEngine` interface if deterministic stages are insufficient.

### Evaluation Engine

The evaluator compares the report root cause against incident ground truth. It reports pass or fail, predicted confidence, and evidence checks.

Tradeoff: exact root-cause matching is strict, but this is useful in V1 because the dataset is curated. Later versions can support semantic matching and partial credit.

## Investigation Stages

### Stage 1: Incident Classification

Classifies alert type, severity, service, and priority from incident metadata and alert fields.

No AI is used. The classifier should be deterministic and easily unit-tested.

### Stage 2: Evidence Collection

Extracts structured facts from the incident:

- Recent deployment
- Memory increase
- CPU increase
- OOMKilled event
- Restart count
- Latency increase
- 5xx spike
- Relevant log signals

The collector should avoid interpreting root cause. It only normalizes observations into evidence.

### Stage 3: Hypothesis Generation

Generates possible root causes from rules.

Initial rule:

If there is a recent deployment, high memory, and OOMKilled events, generate `memory_leak_after_deployment`.

Rules should produce multiple hypotheses when evidence supports ambiguity.

### Stage 4: Confidence Calculation

Scores each hypothesis using supporting and missing evidence.

Confidence must be explainable:

- Every positive signal should reference evidence.
- Missing evidence should be explicit.
- High confidence can stop investigation.
- Low confidence should say what information is needed next.

### Stage 5: Report Generation

Produces a JSON report containing:

- Incident summary
- Timeline
- Predicted root cause
- Evidence
- Confidence
- Alternative hypotheses
- Recommendation

The report is the contract consumed by evaluation and, later, user-facing surfaces.

## Replaceable Boundaries

V1 should define small interfaces where future implementations are expected:

- `IncidentLoader`
- `Classifier`
- `EvidenceCollector`
- `HypothesisGenerator`
- `ConfidenceScorer`
- `ReportGenerator`
- `Evaluator`
- `ReasoningEngine`
- `Summarizer`
- `Planner`

The AI-facing interfaces are placeholders only. They should not be used in V1 behavior.

## Future Plugin Architecture

Future real integrations should replace simulated data sources without changing core reasoning:

- Deployment plugin
- Metrics plugin
- Logs plugin
- Kubernetes plugin
- GitHub plugin
- Tracing plugin
- Runbook plugin

V1 should treat incident JSON as an adapter output. Later plugins can produce the same internal evidence model from real systems.

## Architectural Decisions

### Deterministic First

The first engine version uses rules instead of AI because OOMKilled after deployment can be identified through reliable structured signals.

This reduces cost, improves testability, and makes failures debuggable.

### Evidence Before Hypotheses

The engine extracts facts before reasoning. This prevents raw logs and noisy telemetry from leaking into future AI calls and gives every conclusion a traceable basis.

### Interfaces at Boundaries Only

Interfaces belong where components must be replaceable. Internal helper functions should stay concrete until duplication or complexity justifies abstraction.

This avoids over-engineering while preserving future plugin seams.

### Local Files Before Storage

Dataset files and report output can be local files in V1.

This avoids database design before the investigation model is proven.

### JSON Contracts

Incident input and report output are JSON because they are readable, easy to version, and convenient for tests.

The tradeoff is weaker schema enforcement than Protobuf. That is acceptable for V1. If contracts stabilize and multiple services appear later, schema-first formats can be introduced.

