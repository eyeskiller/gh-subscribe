# gh-subscribe

> Subscribe to GitHub issues and pull requests directly from your terminal. No browser, no extra tokens, no configuration.

**gh-subscribe** is a lightweight GitHub CLI extension that lets you manage subscriptions to GitHub issues and pull requests without leaving your terminal. It reuses your existing `gh` authentication and feels like a native GitHub CLI command.

[![CI](https://github.com/lubosberan/gh-subscribe/actions/workflows/ci.yml/badge.svg)](https://github.com/lubosberan/gh-subscribe/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## The Problem

GitHub CLI currently lacks dedicated commands for managing issue and pull request subscriptions. Users must open GitHub in their browser or interact with the GraphQL API manually to subscribe to conversations, unsubscribe from noisy threads, or ignore issues entirely.

This feature has been [requested by the community](https://github.com/cli/cli/issues/14009).

**gh-subscribe** solves this problem with a simple, fast, reliable CLI extension that integrates seamlessly with your existing GitHub CLI workflow.

## Features

- Subscribe to issues and pull requests
- Unsubscribe from conversations
- Ignore notifications completely
- Check subscription status
- Full URL support
- Repository auto-detection
- Machine-readable JSON output
- Reuses existing `gh` authentication
- Idempotent operations
- Clean, actionable error messages

## Installation

Install using GitHub CLI:

```bash
gh extension install lubosberan/gh-subscribe
```

### Prerequisites

- GitHub CLI (`gh`) installed and authenticated
- Valid GitHub authentication via `gh auth login`

## Quick Start

```bash
# Subscribe to an issue in the current repository
gh subscribe add 123

# Subscribe using a full URL
gh subscribe add https://github.com/cli/cli/issues/14009

# Check subscription status
gh subscribe status 123

# Unsubscribe from a conversation
gh subscribe remove 123

# Ignore all notifications for an issue
gh subscribe ignore 123
```

## Usage

### Subscribe to Issues and Pull Requests

```bash
# Subscribe to issue #123 in the current repository
gh subscribe add 123

# Subscribe to a specific repository
gh subscribe add 123 --repo cli/cli
gh subscribe add 123 -R cli/cli

# Subscribe using a full GitHub URL
gh subscribe add https://github.com/cli/cli/issues/14009
gh subscribe add https://github.com/cli/cli/pull/456

# Shorthand: equivalent to "gh subscribe add"
gh subscribe 123
```

### Unsubscribe

```bash
# Unsubscribe from an issue
gh subscribe remove 123

# Unsubscribe from a specific repository
gh subscribe remove 123 -R cli/cli

# Unsubscribe using a URL
gh subscribe remove https://github.com/cli/cli/pull/456
```

**Note:** Even after unsubscribing, you may still receive notifications when participating in the conversation or when mentioned. Use `ignore` to suppress all notifications.

### Ignore Notifications

```bash
# Ignore all notifications for an issue
gh subscribe ignore 123

# Ignore a specific repository's issue
gh subscribe ignore 123 -R cli/cli

# Ignore using a URL
gh subscribe ignore https://github.com/cli/cli/pull/456
```

### Check Subscription Status

```bash
# Check status in the current repository
gh subscribe status 123

# Check status in a specific repository
gh subscribe status 123 -R cli/cli

# Check status using a URL
gh subscribe status https://github.com/cli/cli/pull/456

# Get machine-readable JSON output
gh subscribe status 123 --json
```

Example output:

```text
Repository: cli/cli
Type:       Issue
Number:     14009
Title:      Add issue subscription commands
Status:     SUBSCRIBED
URL:        https://github.com/cli/cli/issues/14009
```

JSON output:

```json
{
  "repository": "cli/cli",
  "type": "Issue",
  "number": 14009,
  "title": "Add issue subscription commands",
  "url": "https://github.com/cli/cli/issues/14009",
  "status": "SUBSCRIBED"
}
```

## Target Formats

You can specify issues and pull requests in multiple formats:

- **Issue/PR number:** `123` (requires repository context)
- **Owner/repo with number:** `cli/cli#14009`
- **Full GitHub URL:** `https://github.com/cli/cli/issues/14009`
- **Full GitHub URL:** `https://github.com/cli/cli/pull/456`

## Repository Resolution

The extension resolves the target repository in this priority order:

1. Explicit `--repo` or `-R` flag
2. `GH_REPO` environment variable
3. Repository detected from the current working directory

## Subscription States

GitHub provides three subscription states:

- **SUBSCRIBED:** Receive all conversation notifications
- **UNSUBSCRIBED:** Notifications may still arrive when participating or mentioned
- **IGNORED:** Suppress all notifications for the conversation

The `IGNORED` state is stronger than `UNSUBSCRIBED`. Use it when you want to completely silence notifications for a conversation.

## Help and Version

```bash
# Show help
gh subscribe --help
gh subscribe add --help
gh subscribe remove --help
gh subscribe ignore --help
gh subscribe status --help

# Show version
gh subscribe --version
```

## Supported Platforms

- macOS (ARM64, AMD64)
- Linux (AMD64, ARM64)
- Windows (AMD64)

## Authentication

gh-subscribe reuses your existing GitHub CLI authentication. No additional tokens or configuration required.

To authenticate:

```bash
gh auth login
```

Verify your authentication:

```bash
gh auth status
```

## Limitations

- Currently supports GitHub.com only (GitHub Enterprise support planned)
- Does not support listing all subscriptions (GitHub API limitation)
- Requires valid `gh` authentication
- Cannot modify subscriptions for conversations you don't have access to

## Development

### Prerequisites

- Go 1.21 or later
- GitHub CLI installed and authenticated

### Building from Source

```bash
# Clone the repository
git clone https://github.com/lubosberan/gh-subscribe.git
cd gh-subscribe

# Build the extension
go build -o gh-subscribe

# Install locally
gh extension install .
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run go vet
go vet ./...

# Check formatting
gofmt -l .
```

### Project Structure

```
gh-subscribe/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and shorthand
│   ├── add.go             # Subscribe command
│   ├── remove.go          # Unsubscribe command
│   ├── ignore.go          # Ignore command
│   └── status.go          # Status command
├── internal/
│   ├── github/            # GitHub API client
│   │   ├── client.go      # GraphQL client
│   │   └── client_test.go # API tests
│   ├── target/            # Target parsing
│   │   ├── parser.go      # URL/number parser
│   │   └── parser_test.go # Parser tests
│   └── output/            # Output formatting
│       ├── formatter.go   # Terminal formatter
│       └── formatter_test.go # Formatter tests
├── .github/workflows/     # CI/CD
│   ├── ci.yml             # Continuous integration
│   └── release.yml        # Automated releases
├── main.go                # Entry point
├── go.mod                 # Go modules
├── LICENSE                # MIT License
└── README.md              # This file
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

Please ensure your code passes all tests and follows Go best practices.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Uses [go-gh](https://github.com/cli/go-gh) GitHub CLI SDK
- Inspired by the [community feature request](https://github.com/cli/cli/issues/14009)

## Disclaimer

This is an independent community extension and not an official GitHub product. It is maintained by the open-source community.

## ☕ Support My Work

<a href="https://ko-fi.com/eyeskiller" target="_blank">
  <img src="https://raw.githubusercontent.com/eyeskiller/fail2ban-notifier/master/support.png" alt="Support my independent work" style="max-width:100%;">
</a>

Help me maintain this free open-source project.

<a href="https://www.giveth.io/project/fail2ban-notifier" target="_blank">
  <img src="https://raw.githubusercontent.com/eyeskiller/fail2ban-notifier/master/support-crypto.png" alt="Support with crypto on Giveth" style="max-width:100%;">
</a>
