# Template Usage Guide

This guide will help you get started with the WordPress Vulnerability Scanner template.

## Quick Start

### Step 1: Create Your Repository

1. Click "Use this template" button on GitHub
2. Choose a name for your repository (e.g., `my-wordpress-security`)
3. Select whether it should be public or private
4. Click "Create repository from template"

### Step 2: Add Your WordPress Sites

1. In your new repository, create a file named `targets.txt`
2. Add one WordPress site URL per line:

```
https://yoursite.com
https://yourblog.example.org
```

**Note:** The `targets.txt` file is in `.gitignore` to prevent accidental exposure of your site URLs. You can commit it if your repository is private and you want to track changes.

### Step 3: Run Your First Scan

#### Option A: Wait for Scheduled Run
The workflow is scheduled to run daily at 02:00 UTC automatically.

#### Option B: Trigger Manually
1. Go to the "Actions" tab
2. Select "WordPress Vulnerability Scan"
3. Click "Run workflow"
4. Leave defaults or customize:
   - **URLs**: Override targets.txt with comma-separated URLs
   - **Scan mode**: Choose stealthy (fast), bruteforce (thorough), or hybrid (recommended)
5. Click "Run workflow"

### Step 4: Review Results

1. After the workflow completes, go to the workflow run page
2. Scroll down to "Artifacts"
3. Download the scan results artifact
4. Extract and open the CSV or JSON files

If vulnerabilities are found, an issue will be automatically created in your repository.

## Understanding Scan Modes

### Stealthy Mode
- **Best for:** Regular scheduled scans
- **Speed:** Fastest
- **Detection:** Good (3,000+ plugins)
- **Visibility:** Low footprint, fewer requests
- **Use when:** You want minimal impact on target sites

### Bruteforce Mode
- **Best for:** Comprehensive audits
- **Speed:** Slower
- **Detection:** Best (10,000+ plugins)
- **Visibility:** High footprint, many requests
- **Use when:** You need maximum plugin detection

### Hybrid Mode (Recommended)
- **Best for:** Balanced scanning
- **Speed:** Medium
- **Detection:** Excellent
- **Visibility:** Moderate
- **Use when:** You want good coverage without excessive requests

## Interpreting Results

### CSV Format
Open in Excel, Google Sheets, or any spreadsheet software:
- **URL**: The scanned WordPress site
- **Plugin**: Detected plugin name
- **Version**: Plugin version found
- **Vulnerabilities**: Known CVEs affecting this version
- **Severity**: Risk level (Critical, High, Medium, Low)

### JSON Format
Machine-readable format for automated processing:
```json
{
  "url": "https://example.com",
  "plugins": [
    {
      "name": "plugin-name",
      "version": "1.0.0",
      "vulnerabilities": [
        {
          "cve": "CVE-2024-XXXXX",
          "severity": "high",
          "description": "..."
        }
      ]
    }
  ]
}
```

## Taking Action on Vulnerabilities

When vulnerabilities are detected:

1. **Review**: Check the scan results for details
2. **Prioritize**: Focus on Critical and High severity issues first
3. **Update**: Log into WordPress and update affected plugins
4. **Verify**: Re-run the scan to confirm fixes
5. **Document**: Close the GitHub issue once resolved

## Customizing the Schedule

Edit `.github/workflows/wordpress-scan.yml` to change the schedule:

```yaml
on:
  schedule:
    # Examples:
    - cron: '0 2 * * *'     # Daily at 02:00 UTC
    - cron: '0 2 * * 1'     # Weekly on Mondays
    - cron: '0 2 1 * *'     # Monthly on the 1st
    - cron: '0 */6 * * *'   # Every 6 hours
```

Use [crontab.guru](https://crontab.guru/) to help create cron expressions.

## Best Practices

### Security
- ✅ Keep your repository private if scanning production sites
- ✅ Use secrets for any sensitive configuration
- ✅ Only scan sites you own or have permission to test
- ✅ Review and act on findings promptly

### Performance
- ✅ Use stealthy mode for frequent scans
- ✅ Adjust thread count based on target site capacity
- ✅ Space out scans to avoid overwhelming servers
- ✅ Consider scanning during off-peak hours

### Maintenance
- ✅ Regularly review and update targets.txt
- ✅ Archive old scan results after remediation
- ✅ Keep the workflow file updated
- ✅ Monitor for wpprobe updates

## Troubleshooting

### Workflow Fails to Find Targets
- Check that `targets.txt` exists and has valid URLs
- Ensure URLs start with `http://` or `https://`
- Verify there are no blank lines or comments in wrong format

### No Results Generated
- The site might not be WordPress
- Plugins might be well-hidden or renamed
- Try a different scan mode (hybrid or bruteforce)

### Permission Errors
- Check repository settings → Actions → General
- Enable "Read and write permissions"
- Enable "Allow GitHub Actions to create issues"

### wpprobe Installation Fails
- Ensure Go 1.22+ is being used (check workflow file)
- GitHub Actions might be experiencing issues
- Check wpprobe repository for breaking changes

## Advanced Configuration

### Adjusting Thread Count
More threads = faster scan but more load on target:

```yaml
wpprobe scan -f /tmp/targets.txt --mode "$SCAN_MODE" -o "results.json" -t 20
```

Recommended:
- Small sites (1-3): `-t 5`
- Medium sites (4-10): `-t 10`
- Large sites (10+): `-t 20`

### Custom Output Formats
Modify the workflow to add additional output formats or processing steps after the scan completes.

### Integration with Security Tools
Parse the JSON output and send to:
- Slack/Discord webhooks for notifications
- Security dashboards
- SIEM systems
- Jira/Linear for issue tracking

## Getting Help

- **wpprobe Issues**: [GitHub Issues](https://github.com/Chocapikk/wpprobe/issues)
- **Template Issues**: [Create an issue](../../issues/new) in this repository
- **False Positives**: Review vulnerability details and verify plugin versions

## License & Legal

- Only scan WordPress sites you own or have explicit permission to test
- This tool is for authorized security testing only
- Unauthorized scanning may violate laws in your jurisdiction
- The template and wpprobe are provided "as-is" without warranty