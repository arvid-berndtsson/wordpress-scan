# WP Hunter Roadmap

## Q1 2025
- Version detector GA: identify WordPress core versions via content analysis and `wp-json` metadata.
- Expand `scan` summaries with severity counts, delta detection, and Markdown exports.
- Release GitHub Action template (deployments/github) that pulls published binaries instead of rebuilding Go code.

## Q2 2025
- Add plugin/theme misconfiguration detectors (directory listing, exposed backups, vulnerable plugin versions via CVE DB).
- Introduce authenticated mode using supplied cookies/tokens for deeper enumeration.
- Deliver Slack/Webhook notifiers and integrate with issue trackers directly from the CLI.

## Q3 2025
- Implement diffing/regression workflows (`wphunter diff --previous <artifact>`).
- Ship distributed scan coordinator (chunk large target lists, manage worker queues, upload results to S3/GCS).
- Launch community plugin SDK with examples (custom detectors compiled as Go plugins or external commands).

## Stretch Goals
- Browser-assisted detector harness (headless Chrome) for DOM-reliant fingerprints.
- Managed vulnerability feed with signed updates for air-gapped environments.
- Threat-intel sharing format (STIX/TAXII export) for purple-team exercises.

Track progress via GitHub milestones and keep this file updated whenever scope changes.
