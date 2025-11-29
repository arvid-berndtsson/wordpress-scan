# Worker Enablement TODOs

## CLI Worker Enablement
- [ ] Document worker contract covering inputs, outputs, exit codes, and secrets (`docs/worker-contract.md`).
- [ ] Build `cmd/wp-worker-cli` with `scan`, `report`, and `init` commands wrapping `wpprobe` and emitting NDJSON.
- [ ] Implement configuration resolution (file, env, CLI) plus validation for worker environments.
- [ ] Add reproducible packaging via Goreleaser (multi-arch binaries + Docker image) and track version in `cli-version.json`.
- [ ] Create unit/smoke tests for the CLI (mock targets, dry-run) and wire them into CI.
- [ ] Document install/upgrade flows and release notes for worker fleets.

## GitHub Worker Templates
- [ ] Add `worker-templates/github/` with a workflow that downloads the CLI release artifact.
- [ ] Provide matrix-ready workflow sample for parallel target scans with cached vulnerability DB.
- [ ] Ship composite action plus issue/PR templates for invoking the worker CLI and uploading artifacts.
- [ ] Write README snippet explaining secrets, permissions, and schedule customization for templates.
- [ ] Add automated integration test workflow exercising the template against fixtures.
- [ ] Create migration checklist for current `.github/workflows/wordpress-scan.yml` users.

## Functionality Expansion
- [ ] Implement post-processing (severity counts, regression detection) to enable gating and auto-issues.
- [ ] Add Slack/webhook notifications and storage backends (S3/GCS) for streaming artifacts.
- [ ] Support differential scans comparing last artifacts to skip unchanged targets.
- [ ] Introduce sharding/chunking for large target lists with backpressure controls.
- [ ] Add `worker doctor` command to validate network, rate limits, and dependencies pre-run.
- [ ] Maintain signed vulnerability-feed cache for air-gapped workers.
