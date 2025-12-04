# Deployment Recipes

WP Hunter is no longer a template repository. Instead, use these deployment recipes to run the published CLI in your own environments.

## GitHub Actions (Recommended)
1. Copy `deployments/github/wp-hunter-template.yml` into `.github/workflows/wp-hunter.yml` in **your** scan repository.
2. Provide targets in one of two ways:
   - Commit `wphunter.targets.txt` (one URL per line, keep repo private or encrypt file).
   - Supply inline targets through the workflow dispatch input (`https://a.test,https://b.test`).
3. Trigger the workflow manually or on a schedule. The job:
   - Prepares `.wphunter/targets.txt` from your inputs.
   - Pulls `ghcr.io/<org>/wphunter:<tag>` and runs `wphunter scan` with your settings.
   - Uploads `scan-results/` (JSON/CSV artifacts, detections, summary).
4. Monitor NDJSON logs in the workflow output for detector findings (`detection` events).

**Checklist**
- Keep the workflow repository private if targets are sensitive.
- Set `scan_mode` to `stealthy` for low-noise schedules and `bruteforce` for deep audits.
- Use GitHub Environments + secrets if you plan to inject credentials (future authenticated detectors).

## Self-Hosted Linux Worker
1. Install the binary or container per `docs/worker-install.md`.
2. Store `wphunter.config.yml` alongside your scheduler config (systemd timer, cron, Jenkins, etc.).
3. Run `wphunter init` as part of provisioning to verify dependencies and config.
4. Schedule `wphunter scan --config /path/to/wphunter.config.yml` at desired intervals. Export artifacts to an S3 bucket or internal share.
5. Pipe stdout to your SIEM/log shipper to capture NDJSON events.

## Purple-Team Exercises
1. Create multiple config files (e.g., `configs/red.yaml`, `configs/blue.yaml`) with different detector mixes.
2. Run scans in parallel using separate output directories to compare offensive vs defensive views.
3. Use `wphunter report --input scan-results/*.json` to create summary artifacts for exercise debriefs.
4. Feed detections into your detection engineering backlog; the raw JSON makes it easy to write regression tests.

## Legacy Template Migration
If you previously relied on the old GitHub template:
1. Archive the template repo or convert it into a “scan orchestrator” repository.
2. Remove the legacy `.github/workflows/wordpress-scan.yml` file.
3. Add `deployments/github/wp-hunter-template.yml` instead, pointing to released binaries/containers.
4. Replace `targets.txt` with `wphunter.targets.txt` (optional but recommended for clarity).
5. Update docs/readmes to reference WP Hunter instead of the template.

## Operational Tips
- **Targets hygiene:** keep `.gitignore` entries for `wphunter.config.yml` and `wphunter.targets.txt` so secrets never leak.
- **Artifacts:** download/upload `scan-results/` to long-term storage immediately (GitHub artifact retention defaults to 90 days).
- **Scaling:** split very large target inventories across multiple configs; future releases will add automated sharding.
- **Testing:** use `--dry-run --formats json --detectors ""` in CI to validate configs without touching live targets.

Always ensure you have permission to scan each target. Misuse may violate laws or terms of service.
