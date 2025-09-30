# piphos

**piphos** is a command-line tool for managing dynamic IP addresses in homelabs. It provides an easy way to detect your current public IP address and store it in cloud services for remote access.

## 🎯 Motivation

**Tired of playing IP address hide-and-seek with your homelab?**

If you're a homelab enthusiast, you know the pain: you're at work and you want to quickly check something on one of your homelabs, but you can't remote in because your ISP changed your IP address.

If you are the family IT person, you also know the pain: your sibling sends you a text for support with their laptop, and you have to text back and forth to get their IP address so that you can login and support them.

You don't have to be paying for static IPs and VPSs or fight with dynamic DNS setup.

You can let your computer tell you the information you want.

### Quick Start

**At home:**
```bash
piphos push -t github  # or setup as a cron job to automate
```

**From anywhere:**
```bash
piphos pull -t github  # Shows: "home-server: 203.0.113.42"
ssh admin@203.0.113.42  # You're in!
```

**Perfect for:**
- 🏠 **Homelab Heroes**: Access your servers without expensive static IPs
- 👨‍👩‍👧‍👦 **Family IT Support**: Remote into family computers from anywhere
- 🎒 **Remote Workers**: Access your home office setup while traveling

## Overview

piphos consists of two main components:

- **Beacons**: Services that detect your public IP address (like icanhazip.com, AWS checkip)
- **Tenders**: Services that store and retrieve your IP addresses (like GitHub Gists)

This allows you to easily track and access your homelab from anywhere, even when your ISP changes your IP address.

## Features

- 🌐 **Multiple IP Detection Services**: Support for various beacon services
- 💾 **Cloud Storage**: Store IP addresses in GitHub Gists
- 🔄 **Automatic Discovery**: Reuses existing gists to avoid clutter
- ⚙️ **Flexible Configuration**: JSON-based configuration with sensible defaults
- 🛡️ **Secure**: Token-based authentication with validation
- 🖥️ **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Using go install (Recommended)

```bash
go install github.com/kappapee/piphos@latest
```

This will install the latest version of piphos to your `$GOPATH/bin` directory. Make sure `$GOPATH/bin` is in your `$PATH`.

### From Source

```bash
git clone https://github.com/kappapee/piphos.git
cd piphos
go build -o piphos
```

### Binary Installation

Download the latest binary from the [releases page](https://github.com/kappapee/piphos/releases) and place it in your `$PATH`.

## Quick Start

1. **Create a GitHub Personal Access Token**
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Create a token with `gist` permissions
   - Copy the token (starts with `ghp_`, `gho_`, or `github_pat_`)

2. **Create Configuration File**

   The configuration file location depends on your operating system:
   - **Linux**: `~/.config/piphos/config.json`
   - **macOS**: `~/Library/Application Support/piphos/config.json`
   - **Windows**: `%APPDATA%\piphos\config.json`

   **Linux/Unix:**
   ```bash
   mkdir -p ~/.config/piphos
   cat > ~/.config/piphos/config.json << EOF
   {
     "token": "your_github_token_here",
     "tender": "github",
     "beacon": "haz"
   }
   EOF
   ```

   **macOS:**
   ```bash
   mkdir -p ~/Library/Application\ Support/piphos
   cat > ~/Library/Application\ Support/piphos/config.json << EOF
   {
     "token": "your_github_token_here",
     "tender": "github",
     "beacon": "haz"
   }
   EOF
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:APPDATA\piphos"
   @"
   {
     "token": "your_github_token_here",
     "tender": "github",
     "beacon": "haz"
   }
   "@ | Out-File -FilePath "$env:APPDATA\piphos\config.json" -Encoding UTF8
   ```

3. **Check Your IP**
   ```bash
   piphos check
   ```

4. **Store Your IP**
   ```bash
   piphos push -t github
   ```

5. **Retrieve Stored IPs**
   ```bash
   piphos pull -t github
   ```

## Configuration

piphos uses a JSON configuration file stored in your system's configuration directory:

- **Linux**: `~/.config/piphos/config.json`
- **macOS**: `~/Library/Application Support/piphos/config.json`
- **Windows**: `%APPDATA%\piphos\config.json`

### Configuration Options

```json
{
  "hostname": "my-homelab",
  "token": "ghp_xxxxxxxxxxxxxxxxxxxx",
  "beacon": "haz",
  "tender": "github",
  "piphos_gist_id": ""
}
```

| Field | Description | Required | Default |
|-------|-------------|----------|---------|
| `hostname` | Identifier for this machine | No | System hostname |
| `token` | Authentication token for tender service | Yes | - |
| `beacon` | Preferred beacon service | No | Random selection |
| `tender` | Preferred tender service | Yes | - |
| `piphos_gist_id` | Gist ID (auto-populated) | No | - |

## Usage

### Commands

#### `check` - Detect Public IP

Detects and displays your current public IP address using a beacon service.

```bash
# Use configured or randomly selected beacon
piphos check

# Use specific beacon
piphos check -b aws
piphos check -b haz
```

#### `push` - Store IP Address

Detects your public IP and stores it in a tender service.

```bash
# Push to configured tender
piphos push -t github

# Push using specific beacon and tender
piphos push -t github -b aws
```

#### `pull` - Retrieve Stored IPs

Retrieves and displays all stored IP addresses from a tender service.

```bash
# Pull from configured tender
piphos pull -t github
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
| GitHub Gists | `github` | Personal Access Token with `gist` scope | Stores IPs in private gists |

## Examples

### Basic Usage

```bash
# Check your current IP
piphos check
# Output: 203.0.113.1

# Store your IP in GitHub
piphos push -t github
# Creates or updates a gist with your hostname and IP

# View all stored IPs
piphos pull -t github
# Output: host:my-homelab, IP:203.0.113.1
```

### Advanced Usage

```bash
# Use specific beacon for IP detection
piphos check -b aws

# Push to GitHub using AWS beacon
piphos push -t github -b aws

# Check IP from a different beacon
piphos check -b haz
```

### Automation

Create a cron job to automatically update your IP:

```bash
# Update IP every 30 minutes
*/30 * * * * /usr/local/bin/piphos push -t github >/dev/null 2>&1
```

## Error Handling

piphos provides detailed error messages for common issues:

- **Invalid token format**: Token doesn't match expected format for the service
- **Missing configuration**: Required configuration fields are not set
- **Network errors**: Beacon or tender services are unreachable

## Security Considerations

- **Token Storage**: Tokens are stored in plaintext in the configuration file. Ensure proper file permissions (600).
- **Private Gists**: By default, piphos creates private gists that are not publicly accessible.
- **HTTPS Only**: All network communication uses HTTPS for security.

## Development

### Building from Source

```bash
go mod tidy
go build -o piphos .
```

### Code Structure

- `main.go`: Entry point and command routing
- `config.go`: Configuration management
- `handlers.go`: Command handlers
- `beacons.go`: IP detection services
- `tenders.go`: Storage services
- `utils.go`: Utility functions

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Adding New Services

#### Adding a Beacon

1. Add the beacon configuration to `BeaconConfig` in `beacons.go`
2. Add a constant identifier
3. Update documentation

#### Adding a Tender

1. Add the tender configuration to `TenderConfig` in `tenders.go`
2. Implement authentication logic in `setupTender`
3. Add token validation in `utils.go`
4. Update documentation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to the various IP detection services for providing free APIs
- GitHub for providing the Gists API
- The Go community for excellent tooling and libraries
- [boot.dev](https://www.boot.dev/) for creating an amazing learning platform

## Troubleshooting

### Common Issues

**Q: "unable to load configuration file" error**

A: Create the configuration directory and file manually:

```bash
mkdir -p ~/.config/piphos
echo '{"token":"your_token","tender":"github"}' > ~/.config/piphos/config.json
```

**Q: "invalid GitHub token format" error**

A: Ensure your GitHub token starts with `ghp_`, `gho_`, or `github_pat_` and has `gist` permissions.

**Q: "no piphos records on tender" error**

A: Run `piphos push -t github` first to create initial data.

**Q: Network timeout errors**

A: Check your internet connection and try a different beacon service.

---

For more help, please [open an issue](https://github.com/kappapee/piphos/issues) on GitHub.

This document was generated in part from conversations with Claude Sonnet 4.
