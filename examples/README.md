# piphos Basic Configuration and Usage Example

This directory contains a basic configuration example for piphos.

## 🚀 Quick Start Example

### Basic Configuration

Use the [example](https://github.com/kappapee/piphos/examples/config.json) configuration and customize it:

```bash
# Create config directory
mkdir -p ~/.config/piphos

# Edit example config
nano ~/.config/piphos/config.json
```

Edit the configuration file:
- Replace `"token"` with your GitHub personal access token
- Update `"hostname"` to identify your machine
- Choose your preferred `"beacon"` service

### Basic Usage

```bash
# Check current IP
piphos check

# Store IP in GitHub Gist
piphos push -t github

# Retrieve stored IPs
piphos pull -t github
```

## 🔄 Automation Example

Simple cron-based automation:

```bash
# Edit crontab
crontab -e

# Add entry to update IP every 30 minutes
*/30 * * * * /usr/local/bin/piphos push -t github >/dev/null 2>&1
```

## 🌐 Multi-Location Setup

For tracking multiple locations (home, office, etc.):

Configure each location with a unique hostname:

```json
// Home server config
{
  "hostname": "home-server",
  "token": "ghp_...",
  "tender": "github"
}

// Office router config
{
  "hostname": "office-router",
  "token": "ghp_...",
  "tender": "github"
}
```

All locations will store IPs in the same gist with different hostnames.
