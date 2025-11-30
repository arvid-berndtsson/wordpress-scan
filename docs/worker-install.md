# Worker Install & Upgrade Guide

This guide explains how to distribute `wp-worker-cli` to managed worker fleets, verify builds, and communicate release notes.

## Prerequisites
- Access to the GitHub Releases page (or artifact registry) that hosts Goreleaser outputs.
- `sha256sum` or equivalent utility for checksum validation.
- A directory on each worker with at least 100 MB free space.

## Install from Tarball
1. Download the matching archive for the worker architecture (linux-amd64 or linux-arm64):
   ```bash
   VERSION=$(jq -r .version cli-version.json)
   curl -L "https://github.com/<org>/<repo>/releases/download/${VERSION}/wp-worker-cli_${VERSION}_linux_amd64.tar.gz" -o /tmp/wp-worker-cli.tar.gz
   curl -L "https://github.com/<org>/<repo>/releases/download/${VERSION}/checksums.txt" -o /tmp/checksums.txt
   ```
2. Verify the checksum:
   ```bash
   (cd /tmp && sha256sum --ignore-missing -c checksums.txt)
   ```
3. Extract the binary and docs:
   ```bash
   sudo mkdir -p /opt/wp-worker-cli
   sudo tar -xzf /tmp/wp-worker-cli.tar.gz -C /opt/wp-worker-cli
   sudo ln -sf /opt/wp-worker-cli/wp-worker-cli /usr/local/bin/wp-worker-cli
   ```
4. Copy `worker.config.example.yml` to a writable location (e.g., `/etc/wp-worker/worker.config.yml`) and adjust targets/env vars.
5. Run `wp-worker-cli init --config /etc/wp-worker/worker.config.yml` to confirm the installation.

## Install via Docker Image
1. Pull the published multi-arch image:
   ```bash
   VERSION=$(jq -r .version cli-version.json)
   docker pull ghcr.io/<org>/wp-worker-cli:${VERSION}
   ```
2. Mount a config directory and run the container:
   ```bash
   docker run --rm \
     -v "$PWD/worker.config.yml:/config/worker.config.yml" \
     -v "$PWD/scan-results:/scan-results" \
     ghcr.io/<org>/wp-worker-cli:${VERSION} \
     scan --config /config/worker.config.yml --output-dir /scan-results
   ```
3. Publish the image digest to your orchestrator so jobs pin an exact build.

## Upgrade Playbook
1. Update `cli-version.json` with the new semantic version and run Goreleaser to produce artifacts.
2. Draft release notes (template below) and circulate them to stakeholders.
3. Roll out upgrades incrementally:
   - Stage 1: Canary workers download the new tarball/image and run `wp-worker-cli init --dry-run`.
   - Stage 2: Enable scans for a subset of real targets while monitoring NDJSON logs.
   - Stage 3: Promote the version to all workers and remove the previous binary/image after 48 hours of stability.
4. Record the deployment status in your internal change log.

## Release Notes Template
```
## wp-worker-cli {{version}}

### Highlights
- ...

### Worker Actions
1. Download artifact: https://github.com/<org>/<repo>/releases/tag/{{version}}
2. Verify checksum: `sha256sum --check checksums.txt`
3. Restart worker services after updating the binary or container image.

### Compatibility
- Requires WordPress scan workflow vX.Y or later.
- No config changes required / New config keys: ...
```

Keep the latest release notes alongside `cli-version.json` so operators can quickly correlate deployed versions with documented changes.
