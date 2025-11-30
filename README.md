# Piphos

A lightweight utility for tracking dynamic IP addresses.

## Overview

Piphos helps you track public IP addresses for multiple hosts by storing hostname-to-IP mappings.
This is useful for finding machines with dynamic IPs without paying for dynamic DNS services.

Currently, only private GitHub Gists are supported for storing mappings, so a GitHub account is necessary.
The utility can be extended to support a variety of beacons (services for discovering public IP address) and tenders (services for storing data).

## Installation

### Downloading a binary

You can download a binary for your OS and system architecture from the Releases page.

### Using go install

```bash
go install github.com/kappapee/piphos/cmd/piphos@latest
```

### Building from source

```bash
git clone https://github.com/kappapee/piphos.git
cd piphos
make build
```

## Quick Start

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

## License

MIT License - see LICENSE file for details
