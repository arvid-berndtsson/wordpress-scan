# WP Hunter Automation Contract

This contract specifies how worker hosts (GitHub runners, self-hosted agents, containers, or bespoke schedulers) interact with the `wphunter` CLI. It replaces the old template-style assumptions with a product-first interface.

## Responsibilities
1. **Worker** downloads a signed release, provides configuration (targets, credentials, detector choices), executes `wphunter`, and ships artifacts/logs to long-term storage.
2. **WP Hunter CLI** validates configuration, orchestrates detectors (wpprobe + modular plugins), streams NDJSON events, and returns meaningful exit codes.

## Inputs

| Input | Source | Required | Notes |
| --- | --- | --- | --- |
| `targets` | `--targets`, `WPHUNTER_TARGETS`, config | ✅ | Comma/newline-separated list or file path. Normalized into a temp file automatically. |
| `mode` | `--mode`, `WPHUNTER_MODE`, config | ⛔ (default `hybrid`) | Steering parameter for wpprobe (stealthy, bruteforce, hybrid). |
| `threads` | `--threads`, `WPHUNTER_THREADS`, config | ⛔ (default `10`) | Guarded between 1 and 64. |
| `output-dir` | `--output-dir`, `WPHUNTER_OUTPUT_DIR` | ⛔ (default `./scan-results`) | Must be writable; CLI creates timestamped files. |
| `formats` | `--formats`, `WPHUNTER_FORMATS` | ⛔ (default `json,csv`) | Determines scan artifact formats. |
| `detectors` | `--detectors`, `WPHUNTER_DETECTORS` | ⛔ (default `version`) | Controls built-in detector set. Accepts comma-separated names. |
| `summary-file` | `--summary-file`, `WPHUNTER_SUMMARY_FILE` | ⛔ | Optional consolidated JSON summary path. |
| `config file` | `--config` (default `wphunter.config.yml`) | ⛔ | YAML file mirroring the fields above. |

Legacy `WORKER_*` environment variables are still honored for compatibility.

## Outputs
- `scan_<timestamp>.<format>` artifacts written to `output-dir` (JSON/CSV) with raw wpprobe findings.
- `detections_<timestamp>.json` containing detector findings (version fingerprints, future plugins, etc.).
- NDJSON events on stdout (`scan-start`, `artifact-written`, `detection`, `scan-finished`, etc.).
- Optional `summaryFile` consolidating targets, modes, detectors, and artifact paths.

## Exit Codes
| Code | Meaning |
| --- | --- |
| `0` | Scan completed successfully (findings may still exist; inspect artifacts).
| `1` | Invalid configuration (missing targets, unsupported mode, unreadable config).
| `2` | Runtime failure (wpprobe errors, network timeouts, filesystem issues).
| `3` | Post-processing/reporting failure.

Workers must treat non-zero exit codes as failed jobs.

## Environment Requirements
- **OS:** Linux amd64/arm64 (static binary). macOS binaries are on the roadmap.
- **Dependencies:** `wpprobe` binary is bundled via Goreleaser; no Go toolchain required at runtime.
- **Network:** HTTPS egress to targets and the Wordfence feed (unless using pre-fetched DBs).
- **Disk:** ≥200 MB free for temp files and artifacts.

## Secrets & Sensitive Data
- Inject credentials with env vars (resolved in config via `${VAR}` placeholders). They are redacted from NDJSON logs.
- Always use HTTPS when embedding credentials in targets.
- Store artifacts in private buckets or encrypted volumes if they contain sensitive findings.

## Logging & Observability
- **Stdout:** NDJSON events for ingestion into log pipelines.
- **Stderr:** Human-readable progress lines (prefixed with `[wphunter]`).
- **Artifacts:** JSON/CSV + detection files suitable for downstream processing.

## Workflow Sequence
1. Download/extract the desired release (binary tarball or container image) and verify checksums.
2. Create/update `wphunter.config.yml` or set `WPHUNTER_*` environment variables.
3. Run `wphunter init --config wphunter.config.yml` to verify environment readiness (skips detectors when `--dry-run`).
4. Execute `wphunter scan` with desired flags. Provide `--detectors version,plugins` etc. for custom pipelines.
5. Optionally run `wphunter report --input scan-results/scan_<timestamp>.json` for quick stats.
6. Archive/upload artifacts and summaries to centralized storage, open tickets, or trigger follow-up actions.

## Validation Rules
- At least one target is mandatory.
- Threads must stay within 1–64; detectors may impose additional limits.
- `output-dir` must be writable by the current user.
- Detectors only run when not in `--dry-run` mode (they require live targets).

## Future Extensions
- Differential scans referencing previous artifacts.
- Notification sinks (Slack/webhook/email) triggered directly from the CLI.
- Signed vulnerability feed cache distribution for offline worker fleets.
- Pluggable detector SDK so teams can ship private detectors alongside first-party ones.
