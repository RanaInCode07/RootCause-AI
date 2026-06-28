# Folder Structure

The project is organized around the investigation pipeline, not around delivery mechanisms like APIs or workers.

```text
investigation-engine/
  README.md
  docs/
    Architecture.md
    FolderStructure.md
    Roadmap.md
  dataset/
    incidents/
engine/
  loader/
  classifier/
  collector/
  hypothesis/
    confidence/
    report/
  evaluator/
  internal/
  pkg/
  tests/
```

## Root

The root contains project-level documentation and, later, Go module files.

Why it exists: it gives contributors one clear entry point for the product goal, constraints, and local development commands.

## `docs/`

Long-form design documentation lives here.

Why it exists: architecture and roadmap decisions should be explicit before code is added. This avoids drifting into a generic AI wrapper.

## `dataset/`

Static simulated incidents live here.

Planned initial layout:

```text
dataset/
  schema/
    incident.schema.json
  incidents/
    oom_after_deployment.json
```

Why it exists: the engine needs deterministic, repeatable cases with known ground truth. Keeping fixtures outside engine code allows the evaluator to test behavior without hidden assumptions.

The schema folder defines the JSON contract for incident fixtures. The incidents folder contains concrete cases that implement that contract.

## `engine/`

The investigation pipeline lives here. Each subpackage owns one stage.

Why it exists: classification, evidence extraction, hypothesis generation, confidence scoring, and report creation are separate responsibilities. Splitting them keeps tests focused and makes future replacements easier.

## `engine/loader/`

Loads simulated incident files into the shared incident model.

Why it exists: file-based fixtures are a V1 adapter. Later dataset sources or generated simulators can replace this loader without changing classifier, collector, or hypothesis code.

## `engine/classifier/`

Classifies alert type, severity, service, and priority.

Why it exists: classification is deterministic metadata interpretation and should not depend on evidence extraction or hypothesis logic.

## `engine/collector/`

Extracts structured evidence from incident input.

Why it exists: evidence extraction is the most important boundary in the system. Future real integrations should produce the same evidence shape without changing hypothesis rules.

## `engine/hypothesis/`

Generates candidate root causes from evidence.

Why it exists: root-cause candidates should be explainable and independently testable. V1 uses deterministic rules.

## `engine/confidence/`

Scores hypotheses and records supporting and missing evidence.

Why it exists: confidence is a separate concern from hypothesis generation. This makes it possible to tune scoring without changing what hypotheses are considered.

## `engine/report/`

Builds the final investigation report.

Why it exists: reporting is a contract boundary. It should format decisions without re-running investigation logic.

## `evaluator/`

Compares predicted root cause with ground truth.

Why it exists: the product objective is correctness against known incidents. Evaluation must stay separate from the engine to avoid leaking ground truth into investigation logic.

## `internal/`

Private shared implementation code can live here when needed.

Why it exists: Go's `internal` visibility prevents accidental external dependencies. V1 should keep this small.

## `pkg/`

Public reusable types can live here if a type must be consumed across commands or future adapters.

Why it exists: stable contracts may eventually be shared by plugins. V1 should avoid putting code here unless the contract is genuinely shared.

## `tests/`

End-to-end fixtures and evaluation tests can live here.

Why it exists: unit tests should sit next to packages, while cross-pipeline tests need a clear home.

## Tradeoffs

This structure creates more folders than a single-package prototype, but each folder maps to a product concept that must stay independently replaceable.

The structure avoids delivery-specific folders such as `api/`, `server/`, or `cmd/` until an executable interface is needed.
