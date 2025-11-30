# Piphos

A lightweight CLI utility for tracking dynamic IP addresses.

It provides an easy way to detect your current public IP address and store it in a cloud service for remote access.

## Overview

Piphos helps you track public IP addresses for multiple hosts by storing hostname-to-IP mappings.
This is useful for finding machines with dynamic IPs without paying for dynamic DNS services.

Piphos consists of two main components:

- **Beacons**: Services that detect your public IP address (like icanhazip.com, AWS checkip)
- **Tenders**: Services that store your IP addresses (like GitHub Gists)

This allows you to easily track and access your homelab from anywhere, even when your ISP changes your IP address.

Currently, only private GitHub Gists are supported for storing mappings, so a GitHub account is necessary.

The utility can be extended to support a variety of beacons (services for discovering public IP address) and tenders (services for storing data).

## Infomercial

Tired of playing IP address hide-and-seek with your homelab?

If you're a homelab enthusiast, you know the pain: you're at work and you want to quickly check something on one of your homelabs, but you can't remote in because your ISP changed your IP address.

If you are the family IT person, you also know the pain: your sibling sends you a text for support with their laptop, and you have to text back and forth to get their IP address so that you can login and support them.

You don't have to be paying for static IPs and VPSs or fight with dynamic DNS setup.

You can let your computer tell you the information you want.

Perfect for:

- Homelab Heroes: Access your servers without expensive static IPs
- Family IT Support: Remote into family computers from anywhere
- Remote Workers: Access your home office setup while traveling

## Quick Start

### At home-server:

Run one time:

```bash
piphos push
```

Or create a cron job to automatically update your IP:

```bash
# Update IP every 30 minutes
*/30 * * * * /path/to/executable/piphos push > /dev/null 2>&1  # you need to specify where the piphos executable is located
```

### From anywhere:

```bash
piphos pull  # Shows: "home-server: 203.0.113.42"
ssh admin@203.0.113.42  # You're in!
```

## Installation

### Downloading a binary


Download the latest binary for your OS and system architecture from the [releases page](https://github.com/kappapee/piphos/releases) and place it in your `$PATH`.

### Using go install

```bash
go install github.com/kappapee/piphos/cmd/piphos@latest
```

This will install the latest version of piphos to your `$GOPATH/bin` directory. Make sure `$GOPATH/bin` is in your `$PATH`.

### Building from source

```bash
git clone https://github.com/kappapee/piphos.git
cd piphos
make build
```

## Getting started

1. Set your GitHub token (token with gist permissions):
   ```bash
   export PIPHOS_GITHUB_TOKEN=ghp_your_token_here
   ```

2. Check your public IP:
   ```bash
   piphos ping
   ```

3. Store your hostname and IP:
   ```bash
   piphos push
   ```

4. View all tracked hosts:
   ```bash
   piphos pull
   ```

## Commands

### ping

Detects your current public IP address.

**Usage**: `piphos ping [-beacon=PROVIDER]`

**Flags**:
- `-beacon string` - Beacon provider to use (default "haz")
  - Options: "haz" (icanhazip.com), "aws" (checkip.amazonaws.com)

**Example**:
```bash
$ piphos ping
203.0.113.42

$ piphos ping -beacon=aws
203.0.113.42
```

### pull

Retrieves all hostname-to-IP mappings from storage.

**Usage**: `piphos pull [-tender=PROVIDER]`

**Flags**:
- `-tender string` - Storage provider to use (default "gh")
  - Options: "gh" (GitHub Gists)

**Requirements**:
- `PIPHOS_GITHUB_TOKEN` environment variable

**Example**:
```bash
$ piphos pull
laptop: 203.0.113.42
desktop: 198.51.100.17
```

### push

Updates the current hostname's IP address in storage.

**Usage**: `piphos push [-tender=PROVIDER]`

**Flags**:
- `-tender string` - Storage provider to use (default "gh")

**Requirements**:
- `PIPHOS_GITHUB_TOKEN` environment variable

**Example**:
```bash
$ piphos push
203.0.113.42
```

### Available Services

#### Beacon Services

| Name | Identifier | URL | Description |
|------|------------|-----|-------------|
| icanhazip | `haz` | https://ipv4.icanhazip.com | Simple IP detection service |
| AWS CheckIP | `aws` | https://checkip.amazonaws.com | Amazon's IP detection service |

#### Tender Services

| Name | Identifier | Requirements | Description |
|------|------------|-------------|-------------|
| GitHub Gists | `gh` | Personal Access Token with `gist` scope | Stores IPs in private gists |

## Configuration

### Environment Variables

- **PIPHOS_GITHUB_TOKEN**: GitHub personal access token with gist permissions (required for push/pull commands)

## Storage Format

Piphos stores data in a private GitHub Gist with the description "_piphos_".
The gist contains a single JSON file mapping hostnames to IP addresses:

```json
{
  "laptop": "203.0.113.42",
  "desktop": "198.51.100.17"
}
```

## Acknowledgments

- Thanks to the various IP detection services for providing free APIs
- GitHub for providing the Gists API
- The Go team and community for excellent tooling and libraries
- [boot.dev](https://www.boot.dev/) for creating an amazing learning platform
- The documentation and this README have been created in conversations with Claude Sonnet 4.5

## License

MIT License - see LICENSE file for details
