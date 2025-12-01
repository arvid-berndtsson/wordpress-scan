# WP Hunter Architecture

```
┌────────────┐      ┌───────────────────────┐      ┌──────────────────────┐
│  Config    │─────▶│    CLI (cobra)        │─────▶│  Artifact Writers     │
│ (file/env) │      │ init / scan / report  │      │ (scan*, detections,  │
└────────────┘      └─────────┬─────────────┘      │  summary, NDJSON)    │
                              │                    └──────────────────────┘
                              ▼
                      ┌───────────────┐
                      │ Detector Hub  │
                      │ (registry +   │
                      │  pipelines)   │
                      └──────┬────────┘
                             │
        ┌────────────────────┴────────────────────┐
        ▼                                         ▼
┌──────────────┐                         ┌─────────────────────┐
│ Version      │                         │ wpprobe Runner      │
│ Detector     │                         │ (plugin/theme enum) │
└──────────────┘                         └─────────────────────┘
```

## Layers
1. **Config Loader (`internal/config`)** – merges `wphunter.config.yml`, environment variables (new `WPHUNTER_*` aliases), and CLI flags into a validated runtime struct (targets, modes, detectors, outputs).
2. **CLI (`internal/cli`)** – Cobra commands (`init`, `scan`, `report`) consuming the runtime config, emitting NDJSON events, and coordinating detectors/wpprobe.
3. **Detector Runtime (`internal/detector`)** – registry + factories for built-in detectors. Currently ships with `version` detector, with interfaces ready for plugin/theme/supply-chain modules.
4. **wpprobe Runner (`internal/wpprobe`)** – thin wrapper that ensures the `wpprobe` binary exists and executes scans with the desired mode/threads.
5. **Artifact Writers** – helper functions that produce placeholder artifacts (dry-run), detection JSON arrays, and summary files.

## Execution Flow (scan)
1. Load + validate config.
2. Materialize targets into a temporary file.
3. Emit `scan-start` event.
4. Run wpprobe for each requested format (`json`, `csv`) OR produce placeholders during `--dry-run`.
5. Instantiate detectors from the registry and run them per target (skipped during dry-run).
6. Write detection artifacts + summary, emit `detection` events for each finding, then `scan-finished` when complete.

## Extensibility Hooks
- **New Detectors:** register via `detector.DefaultRegistry`. Future work: dynamic registry fed via config, Go plugins, or external commands.
- **Outputs:** `writeDetectionsArtifact` and `writeSummary` accept raw structs – easy to extend with Markdown/HTML exporters.
- **Deployments:** CLI remains environment-agnostic; recipes under `deployments/` handle GitHub Actions, containers, and future scheduler integrations.

## Security & Observability
- NDJSON event stream surfaces metrics (`targets`, `mode`, `artifact path`, `detection severity`).
- Exit codes differentiate config, runtime, and reporting failures for automation.
- Future: add OpenTelemetry spans or structured stats exported via summary files.
