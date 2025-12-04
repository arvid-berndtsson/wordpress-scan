# WP Hunter

WP Hunter is an offensive/defensive WordPress reconnaissance toolkit that ships as a standalone CLI, container image, and set of deployment recipes. It orchestrates multiple detectors (wpprobe, version fingerprinting, future plugin/theme checks) and emits machine-readable artifacts for red, blue, and purple team workflows.

## Highlights
- 🔎 **Modular detection pipeline** – run built-in detectors (`version`, wpprobe) and extend via the new detector runtime.
- ⚙️ **Production-ready CLI** – layered config (`wphunter.config.yml`, env, flags), NDJSON events, summary artifacts, and dry-run support.
- 🚀 **Deploy anywhere** – binaries, multi-arch container, or copy/paste GitHub Action from `deployments/github/`.
- 📊 **Actionable outputs** – JSON/CSV scan data, detector findings, and summaries ready for issue trackers or SIEM ingestion.
- 🗺️ **Open roadmap** – documented vision/roadmap, with plans for diffing, authenticated checks, notifier hooks, and community detectors.

## Quick Start

```bash
# 1. Build the CLI (or download from Releases once published)
go build -o bin/wphunter ./cmd/wphunter

# 2. Copy and edit the config + targets
cp wphunter.config.example.yml wphunter.config.yml
cp wphunter.targets.example wphunter.targets.txt # optional, used by deployments
# edit targets, modes, detectors, output directory

# 3. Validate the environment (installs only run-time deps)
./bin/wphunter init --config wphunter.config.yml

# 4. Run a scan (remove --dry-run for live detectors + wpprobe)
./bin/wphunter scan --config wphunter.config.yml --dry-run

# 5. Generate a quick summary against a JSON artifact
./bin/wphunter report --input scan-results/scan_<timestamp>.json
```

Detectors require live targets, so they are automatically skipped during `--dry-run`. Set `--detectors ""` (or `WPHUNTER_DETECTORS=`) to disable them entirely. When enabled, findings are written to `detections_<timestamp>.json` and streamed via NDJSON events.

## Configuration

`wphunter.config.yml` controls targets, scan modes, threads, formats, and detectors:

```yaml
mode: hybrid
threads: 12
outputDir: scan-results
formats: [json, csv]
detectors: [version]
targets:
  - https://example.com
  - https://myblog.example.org
summaryFile: scan-results/summary.json
```

You can override any field via environment variables (new `WPHUNTER_*` names with legacy `WORKER_*` fallbacks) or CLI flags:

```bash
WPHUNTER_TARGETS="https://one.test,https://two.test" \
WPHUNTER_DETECTORS="version" \
./bin/wphunter scan --formats json
```

## Detectors
- `version` *(new)*: downloads each target homepage and extracts the WordPress generator meta tag, reporting the detected core version.
- `wpprobe`: leverages [wpprobe](https://github.com/Chocapikk/wpprobe) for plugin/theme enumeration using stealthy, bruteforce, or hybrid strategies.

Future detectors (see `docs/roadmap.md`) will include authenticated probes, misconfiguration checks, and differential analysis.

## Deployments & Integrations
- **GitHub Actions:** copy `deployments/github/wp-hunter-template.yml` into your own repo. The workflow pulls prebuilt binaries/containers instead of rebuilding Go code.
- **Workers/Fleets:** consult `docs/worker-contract.md` and `docs/worker-install.md` for install, upgrade, and release-note procedures across Linux amd64/arm64 hosts.
- **Containers:** build via `goreleaser` or `docker build -f Dockerfile.goreleaser .` to obtain a minimal distroless image.

## Packaging & Releases
Reproducible builds are defined in `.goreleaser.yaml`.

1. Update `cli-version.json` with the next semantic version.
2. Tag the repo (e.g., `git tag v0.2.0`).
3. Export the version for Goreleaser: `export CLI_VERSION=$(jq -r .version cli-version.json)`.
4. Run `goreleaser release --clean` (or `goreleaser build --snapshot --clean` for local smoke tests).

Artifacts include multi-arch tarballs, `checksums.txt`, and a Buildx multi-arch image (`ghcr.io/<org>/wphunter:<version>`). Each bundle ships docs and the example config for quick onboarding.

## Documentation
- `docs/product-vision.md` – positioning, personas, and architectural pillars.
- `docs/roadmap.md` – near-term and quarterly goals.
- `docs/architecture.md` – component overview and detector pipeline.
- `docs/worker-contract.md` – automation interface (inputs, outputs, exit codes).
- `docs/worker-install.md` – install/upgrade guide + release-note template.

## Contributing & Community
1. Run `go test ./...` and `gofmt -w internal cmd` before submitting PRs.
2. Propose new detectors or integrations via GitHub issues (`kind/detector`, `kind/deployment`).
3. Share red/blue/purple-team playbooks using discussions to help the project grow.

Built for authorized security testing only. Always ensure you have permission before scanning any target.
