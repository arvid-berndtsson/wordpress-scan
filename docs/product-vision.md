# WP Hunter Vision

WP Hunter aims to be the definitive WordPress reconnaissance and validation platform for red, blue, and purple teams. Instead of being a GitHub Actions template, it operates as a standalone, batteries-included scanner/automation toolkit with optional deployment recipes (GitHub Actions, containers, worker fleets).

## Personas & Outcomes
- **Red Teams:** Automate pretext enumeration, plugin/theme fingerprinting, version detection, and differential findings to accelerate foothold discovery.
- **Blue Teams:** Continuously validate production WordPress estates, raise issues, and create evidence packages for patch validation.
- **Purple Teams:** Script collaborative exercises that combine attacker-style reconnaissance with defensive monitoring/alerting hooks.

## Product Pillars
1. **Multi-vector Detection** – Modular detectors (version, plugins, misconfigurations, brute-force, authenticated checks) that can run locally or in distributed workers.
2. **Actionable Reporting** – NDJSON streams, Markdown/HTML summaries, and structured artifacts for issue trackers and SIEM pipelines.
3. **Deployment Agility** – First-class support for binaries, Docker images, GitHub/GitLab workers, and on-prem schedulers.
4. **Community-first** – Open roadmap, plugin API, sample datasets, and battle-tested defaults that make the project star-worthy and approachable.

## Architecture Direction
- **Core CLI (`wphunter`)** orchestrates detectors, handles config layering, produces artifacts, and exposes subcommands (`init`, `scan`, `report`, future `diff`, `doctor`).
- **Detector Runtime** (new `internal/detector` package) allows shipping first-party detectors and community plugins.
- **Deployment Recipes** live under `deployments/` (GitHub template, container manifests, Terraform snippets) rather than forcing the entire repo to be a template.
- **Data Contracts** are documented in `docs/` to standardize inputs/outputs regardless of where the tool runs.

## Release Cadence & Community Plan
- Ship signed binaries/Docker images via Goreleaser each minor release.
- Maintain `docs/roadmap.md` with quarterly objectives and label GitHub issues (`kind/detector`, `kind/deployment`, `good-first-issue`).
- Publish example walkthroughs (red-team recon, blue-team continuous validation) and highlight contributions to grow stars.

## Near-term Goals
1. Harden the CLI + detector pipeline (version detection, wpprobe integration, summary artifacts).
2. Provide GitHub Action + container recipes under `deployments/`.
3. Document install/upgrade paths, worker contracts, and release governance.
4. Launch a public roadmap and invite feedback via discussions/issues.
