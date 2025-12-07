# WP Hunter Install & Upgrade Guide

This guide covers distributing `wphunter` across managed fleets (GitHub runners, on-prem cron workers, container platforms) plus how to handle upgrades and release notes.

## Prerequisites
- Access to the GitHub Releases page (or internal artifact registry) where Goreleaser outputs are published.
- `sha256sum` (or equivalent) for checksum validation.
- At least 100 MB free disk per worker for binaries + artifacts.

## Install from Tarball
1. Download the matching archive for your architecture:
   ```bash
   VERSION=$(jq -r .version cli-version.json)
   curl -L "https://github.com/<org>/<repo>/releases/download/${VERSION}/wphunter_${VERSION}_linux_amd64.tar.gz" -o /tmp/wphunter.tar.gz
   curl -L "https://github.com/<org>/<repo>/releases/download/${VERSION}/checksums.txt" -o /tmp/checksums.txt
   ```
2. Verify checksums:
   ```bash
   (cd /tmp && sha256sum --ignore-missing -c checksums.txt)
   ```
3. Install the binary + docs:
   ```bash
   sudo mkdir -p /opt/wphunter
   sudo tar -xzf /tmp/wphunter.tar.gz -C /opt/wphunter
   sudo ln -sf /opt/wphunter/wphunter /usr/local/bin/wphunter
   ```
4. Copy `wphunter.config.example.yml` to `/etc/wphunter/config.yml` (or another writable path) and edit targets, detectors, and outputs.
5. Run `wphunter init --config /etc/wphunter/config.yml` to validate the environment, then schedule scans (systemd timer, cron, runner job, etc.).

## Install via Container
1. Pull the multi-arch image:
   ```bash
   VERSION=$(jq -r .version cli-version.json)
   docker pull ghcr.io/<org>/wphunter:${VERSION}
   ```
2. Mount config + results and run a scan:
   ```bash
   docker run --rm \
     -v "$PWD/wphunter.config.yml:/config/wphunter.config.yml" \
     -v "$PWD/scan-results:/scan-results" \
     ghcr.io/<org>/wphunter:${VERSION} \
     scan --config /config/wphunter.config.yml --output-dir /scan-results
   ```
3. Pin the image digest inside your orchestrator for reproducible deployments.

## Upgrade Playbook
1. Update `cli-version.json`, tag the repo, and run `goreleaser release --clean` to publish new assets.
2. Draft release notes (template below) summarizing detectors, new outputs, and migration steps.
3. Roll out upgrades incrementally:
   - **Stage 1:** Canary workers download the new tarball/image and run `wphunter init --dry-run`.
   - **Stage 2:** Enable scans for a subset of targets, monitor NDJSON logs and artifacts.
   - **Stage 3:** Promote to all workers; retire the previous version after ≥48 hours of stability.
4. Capture deployment status in your change log/runbook.

## Release Notes Template
```
## wphunter {{version}}

### Highlights
- ...

### Worker Actions
1. Download artifact: https://github.com/<org>/<repo>/releases/tag/{{version}}
2. Verify checksum: `sha256sum --check checksums.txt`
3. Update configs (if needed) and restart worker services.

### Compatibility
- Requires detector runtime vX.Y or greater / No config changes required / New config keys: ...
```

Keep release notes near `docs/roadmap.md` so operators can correlate deployed versions with documented changes.
