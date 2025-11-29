# WordPress Vulnerability Scanner

A template repository for automatically scanning WordPress websites for known vulnerabilities using [wpprobe](https://github.com/Chocapikk/wpprobe) and GitHub Actions.

## Features

- 🔍 Automated vulnerability scanning using wpprobe
- ⏰ Scheduled daily scans via GitHub Actions
- 🎯 Multiple scan modes: stealthy, bruteforce, and hybrid
- 📊 Results saved as JSON and CSV formats
- 🚨 Automatic issue creation when vulnerabilities are detected
- 🔧 Manual trigger with custom URLs via workflow_dispatch

## Setup

### 1. Use This Template

Click the "Use this template" button to create a new repository from this template.

### 2. Configure Scan Targets

Create a `targets.txt` file in the root of your repository with one URL per line:

```
https://example.com
https://myblog.example.org
https://shop.example.net
```

You can use `targets.txt.example` as a starting point:

```bash
cp targets.txt.example targets.txt
# Edit targets.txt with your WordPress site URLs
```

### 3. Enable GitHub Actions

Ensure GitHub Actions is enabled in your repository settings.

### 4. Configure Permissions (Optional)

If you want automatic issue creation for detected vulnerabilities:

1. Go to Settings → Actions → General
2. Scroll to "Workflow permissions"
3. Select "Read and write permissions"
4. Check "Allow GitHub Actions to create and approve pull requests"

## Usage

### Automatic Scans

The workflow runs automatically every day at 02:00 UTC. You can modify the schedule in `.github/workflows/wordpress-scan.yml`.

### Manual Scans

You can manually trigger a scan from the Actions tab:

1. Go to the "Actions" tab in your repository
2. Select "WordPress Vulnerability Scan" workflow
3. Click "Run workflow"
4. Optionally:
   - Enter comma-separated URLs to scan (overrides targets.txt)
   - Choose scan mode: stealthy, bruteforce, or hybrid

## Worker CLI (Experimental)

For worker platforms that cannot run GitHub Actions, an experimental CLI is now available.

1. Copy `worker.config.example.yml` to `worker.config.yml` and edit your targets or set the `WORKER_TARGETS` environment variable.
2. Build the binary locally: `go build ./cmd/wp-worker-cli`.
3. Validate the environment: `./wp-worker-cli init --dry-run`.
4. Run a scan (omit `--dry-run` once `wpprobe` is installed on the worker): `./wp-worker-cli scan --dry-run`.
5. Generate summaries from an artifact: `./wp-worker-cli report --input scan-results/scan_<timestamp>.json`.

See `docs/worker-contract.md` for the complete worker contract and configuration details.

### Scan Modes

- **stealthy**: Uses REST API enumeration (low footprint, fewer requests)
- **bruteforce**: Aggressively checks plugin directories (more thorough)
- **hybrid**: Combines both methods (recommended for best coverage)

## Viewing Results

### Artifacts

After each scan, results are uploaded as artifacts:

1. Go to the "Actions" tab
2. Click on the completed workflow run
3. Download the "scan-results" artifact
4. Extract and view the JSON or CSV files

### Issues

When vulnerabilities are detected, an issue is automatically created with:
- Summary of findings
- Link to detailed results
- Action items for remediation

## Scan Results Format

Results are saved in two formats:

**JSON**: Detailed machine-readable format with full vulnerability data
**CSV**: Human-readable format for quick review in spreadsheets

## Customization

### Adjusting Scan Frequency

Edit the cron schedule in `.github/workflows/wordpress-scan.yml`:

```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 02:00 UTC
```

### Adjusting Thread Count

Modify the `-t` parameter in the workflow file to change concurrent scan threads:

```yaml
wpprobe scan -f /tmp/targets.txt --mode "$SCAN_MODE" -o "scan-results/scan_${TIMESTAMP}.json" -t 10
```

### Artifact Retention

Adjust retention period in the workflow file (default: 90 days):

```yaml
retention-days: 90
```

## About wpprobe

wpprobe is a fast WordPress plugin enumeration tool that detects installed plugins and matches them against known vulnerabilities. It offers:

- 🚀 Fast, multithreaded scanning
- 🔒 Stealthy detection via REST API
- 📦 Detection of 3,000+ plugins (stealthy) or 10,000+ (bruteforce)
- 🛡️ CVE matching against Wordfence database

## Security Considerations

- **Do not commit actual URLs** to public repositories
- Use repository secrets for sensitive information if needed
- Review scan results carefully before sharing
- This tool is for authorized security testing only

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This template is provided as-is for security scanning purposes.

## Acknowledgments

- [wpprobe](https://github.com/Chocapikk/wpprobe) by Chocapikk
- Vulnerability data from Wordfence