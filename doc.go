// Package main implements piphos, a command-line tool for managing dynamic IP addresses
// in homelabs and remote systems.
//
// # Overview
//
// piphos provides a simple way to detect your current public IP address and store it
// in cloud services for remote access. It's particularly useful for homelab enthusiasts
// who need to track their dynamic IP addresses assigned by ISPs.
//
// The tool operates on two main concepts:
//   - Beacons: External services that detect your public IP address
//   - Tenders: Cloud services that store and retrieve your IP addresses
//
// # Architecture
//
// piphos uses a modular architecture with clearly separated concerns:
//
//	┌─────────────┐    ┌─────────────┐    ┌─────────────┐
//	│   Beacons   │    │   piphos    │    │   Tenders   │
//	│             │───▶│    Core     │───▶│             │
//	│ IP Detection│    │             │    │ IP Storage  │
//	└─────────────┘    └─────────────┘    └─────────────┘
//
// # Beacon Services
//
// Beacon services are HTTP endpoints that return your public IP address in plain text.
// piphos includes support for multiple reliable beacon services:
//
//   - AWS CheckIP (aws): https://checkip.amazonaws.com
//   - icanhazip (haz): https://ipv4.icanhazip.com
//
// If no beacon is specified, piphos automatically selects one randomly from the
// available configured options.
//
// # Tender Services
//
// Tender services provide persistent storage for IP addresses. Currently supported:
//
//   - GitHub Gists (github): Stores IP addresses in private GitHub Gists
//
// Tender services require authentication tokens and provide features like:
//   - Automatic gist discovery and reuse
//   - Private storage by default
//   - Version history through the underlying service
//   - Cross-platform access from anywhere with internet
//
// # Configuration
//
// piphos uses a JSON configuration file stored in the user's configuration directory:
//   - Linux/macOS: ~/.config/piphos/config.json
//   - Windows: %APPDATA%\piphos\config.json
//
// Example configuration:
//
//	{
//	  "hostname": "my-homelab",
//	  "token": "GITHUB_TOKEN_WITH_GIST_PERMISSIONS",
//	  "beacon": "aws",
//	  "tender": "github",
//	  "piphos_gist_id": ""
//	}
//
// # Usage Examples
//
// Basic IP detection:
//
//	piphos check
//	piphos check -b aws
//
// Store IP address:
//
//	piphos push -t github
//	piphos push -t github -b haz
//
// Retrieve stored IP addresses:
//
//	piphos pull -t github
//
// # Authentication
//
// Different tender services require different authentication methods:
//
// GitHub Gists:
//   - Personal Access Token with 'gist' scope
//   - Token must start with 'ghp_', 'gho_', or 'github_pat_'
//   - Created at: https://github.com/settings/tokens
//
// # Error Handling
//
// piphos provides comprehensive error handling with descriptive messages:
//   - Network connectivity issues
//   - Invalid configuration
//   - Service unavailability
//
// All errors are logged with context to help diagnose issues.
//
// # Extensibility
//
// The architecture is designed to be easily extensible:
//
// Adding new beacon services:
//  1. Add beacon configuration to BeaconConfig map
//  2. Add constant identifier
//  3. Update validation logic if needed
//
// Adding new tender services:
//  1. Define data structures for the service API
//  2. Add tender configuration to TenderConfig map
//  3. Implement service-specific authentication
//  4. Add token validation logic
//
// # Security Considerations
//
// piphos follows security best practices:
//   - All network communication uses HTTPS
//   - Private storage is used by default where available
//   - Configuration files should have restricted permissions (600)
//   - Tokens are never logged or exposed in error messages
//
// # Common Use Cases
//
// Remote Access Setup:
//
//	Store your home IP address in a gist, then retrieve it from anywhere
//	to establish SSH, VPN, or web connections to your homelab.
//
// Multiple Location Tracking:
//
//	Use different hostnames for different locations (home, office, cabin)
//	and track all their IP addresses in a single gist.
//
// Automated Updates:
//
//	Set up cron jobs to automatically update IP addresses when they change,
//	ensuring you always have current connectivity information.
//
// For detailed usage information, run 'piphos' without arguments to see
// the help text, or visit the project documentation.
package main
