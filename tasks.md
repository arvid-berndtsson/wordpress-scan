# WP Hunter Backlog

## Core Platform
- [x] Publish product vision and roadmap docs (`docs/product-vision.md`, `docs/roadmap.md`).
- [x] Rebrand CLI/binary to `wphunter` with version flag + config rename.
- [x] Layered config (file/env/flags) with detector selection + env aliasing.
- [x] Package via Goreleaser (multi-arch binaries + Docker image) tracked by `cli-version.json`.
- [x] Document install/upgrade and automation contract for worker fleets.
- [ ] Add `doctor` subcommand to validate dependencies, network reachability, and wpprobe DB freshness.
- [ ] Implement Markdown/HTML summary exports for quick sharing.

## Detector & Pipeline Expansion
- [x] Ship `version` detector and detection artifact output.
- [ ] Add plugin/theme misconfiguration detector (directory listing, backup leaks).
- [ ] Integrate vulnerability feed diffing + regression detection.
- [ ] Introduce authenticated detector mode (cookie/token support).
- [ ] Provide diff command comparing two artifacts to highlight changes.
- [ ] Build notification hooks (Slack/Webhook) keyed to severity thresholds.

## Deployments & Integrations
- [ ] Add `deployments/github/` workflow pulling released binaries + matrix fan-out.
- [ ] Provide container/k8s cronjob examples for large estates.
- [ ] Create migration guide for former template users adopting wphunter releases.
- [ ] Add CI pipeline (`.github/workflows/ci.yml`) with go test + formatting + snapshot build.
- [ ] Ship composite GitHub Action/wrapper for invoking `wphunter` with artifact uploads.

## Community & Docs
- [ ] Publish red/blue/purple playbooks detailing example workflows.
- [ ] Add CONTRIBUTING guide + issue templates with detector/deployment labels.
- [ ] Produce changelog automation + release-note checklist tied to `docs/roadmap.md`.
