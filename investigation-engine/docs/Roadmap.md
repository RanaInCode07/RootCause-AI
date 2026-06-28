# Roadmap

## Version 1: Deterministic OOM Investigation

Goal: correctly identify OOMKilled after deployment in simulated incidents.

Scope:

- Define incident JSON schema.
- Add one OOMKilled-after-deployment incident fixture.
- Implement deterministic classifier.
- Implement structured evidence collection.
- Implement rule-based hypothesis generation.
- Implement confidence calculation.
- Implement JSON report generation.
- Implement evaluator.
- Add unit tests for every module.
- Add an end-to-end evaluation test.

Success criteria:

- Evaluator returns PASS for the OOMKilled fixture.
- Predicted root cause matches ground truth.
- Confidence is high only when supporting evidence is present.
- Report cites evidence for every conclusion.

## Version 1.1: More Simulated Incident Types

Add additional curated fixtures and deterministic rules:

- CrashLoopBackOff because of missing secret
- Database latency
- Redis timeout
- Kafka consumer lag
- Certificate expiration
- DNS failure
- Feature flag rollout issue
- CPU spike because of infinite loop
- Dependency timeout

Success criteria:

- Evaluation reports per-incident and aggregate accuracy.
- Low-confidence cases list missing evidence instead of guessing.
- Rules remain explainable and testable.

## Version 1.2: Dataset Quality and Evaluation Depth

Improve evaluation without adding real integrations:

- Add negative cases.
- Add noisy logs and irrelevant events.
- Add overlapping symptoms.
- Add partial evidence incidents.
- Track false positives and false negatives.
- Add confidence calibration checks.

Success criteria:

- The engine does not overfit a single happy path.
- Confidence decreases when key evidence is missing.

## Version 2: Plugin Interfaces

Introduce replaceable data-source adapters while keeping simulated adapters as default.

Potential plugins:

- Deployment plugin
- Metrics plugin
- Logs plugin
- Kubernetes plugin
- GitHub plugin
- Tracing plugin
- Runbook plugin

Success criteria:

- Plugins return the same internal evidence contracts used by V1.
- Core hypothesis and confidence logic remains unchanged.

## Version 3: Optional AI Assistance

Add AI only where deterministic logic is insufficient.

Possible AI-backed interfaces:

- `ReasoningEngine`
- `Summarizer`
- `Planner`

Rules:

- Do not send raw logs directly to AI.
- Send extracted evidence only.
- Keep provider-specific code behind adapters.
- Make AI output optional and auditable.
- Keep deterministic fallback behavior.

Success criteria:

- AI improves ambiguous investigations without reducing explainability.
- The engine remains usable with AI disabled.

