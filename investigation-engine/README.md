# Autonomous Investigation Engine

Autonomous Investigation Engine (AIE) is a deterministic backend engine for starting production incident investigations before an engineer opens dashboards.

Version 1 is intentionally small. It does not integrate with Grafana, Kubernetes, GitHub, Slack, databases, queues, auth, or cloud services. It reads simulated incident JSON, extracts structured evidence, generates explainable hypotheses, calculates confidence, and evaluates the predicted root cause against known ground truth.

## V1 Goal

Correctly solve simulated production incidents.

The first supported incident type is:

- OOMKilled after deployment

The engine optimizes for correctness, explainability, low AI cost, and replaceable components. It should prove the investigation model before any production integrations are added.

## Core Flow

1. Load a simulated incident JSON file.
2. Classify the alert using deterministic logic.
3. Collect structured evidence from deployment, metrics, Kubernetes events, and logs.
4. Generate hypotheses from deterministic rules.
5. Calculate confidence with explicit supporting and missing evidence.
6. Produce a JSON report.
7. Evaluate prediction accuracy against ground truth.

## Running an Investigation

Input is an incident JSON file. Output is stored as JSON files.

The incident JSON contract is defined in:

```text
dataset/schema/incident.schema.json
```

Important V1 input rules:

- `incident_window` is required for evidence correlation.
- `deployment` is required as a field but may be `null`.
- `metrics`, `kubernetes_events`, and `logs` are required as fields but may be empty arrays.
- `ground_truth` is used only by the evaluator, not by the investigation engine.

```bash
go run ./cmd/aie investigate dataset/incidents/oom_after_deployment.json
```

By default, output is written to:

```text
outputs/<incident-id>/report.json
outputs/<incident-id>/evaluation.json
```

Use `--output-dir` to choose an exact output directory:

```bash
go run ./cmd/aie investigate --output-dir /tmp/aie-run dataset/incidents/oom_after_deployment.json
```

`report.json` is the investigation output. `evaluation.json` compares the report with the fixture ground truth.

## Non-Goals

Version 1 does not include:

- Authentication
- Database persistence
- Billing
- Organizations or teams
- Frontend
- Slack integration
- Kubernetes integration
- Grafana, Prometheus, or Loki integration
- GitHub integration
- Kafka
- Cloud deployment
- Microservices
- Networking
- Real APIs
- LLM integration

## Design Principles

- Use deterministic logic before AI.
- Never send raw logs or large datasets to AI.
- Extract evidence before reasoning.
- Make every conclusion cite supporting evidence.
- Keep each module replaceable.
- Prefer simple interfaces at boundaries.
- Avoid premature abstractions inside stable local code.
- Keep V1 testable with local files and `go test`.

## Documentation

- [Architecture](docs/Architecture.md)
- [Folder Structure](docs/FolderStructure.md)
- [Roadmap](docs/Roadmap.md)
