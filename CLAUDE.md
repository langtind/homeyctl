# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build the CLI
go build -o homey .

# Run directly
go run .

# Build with version info (for releases)
go build -ldflags "-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o homey .

# Run tests
go test ./...

# Run a single test
go test -run TestName ./path/to/package
```

## Architecture

This is a CLI tool for controlling Homey smart home devices via the local Homey API. Built with Cobra for command handling and Viper for configuration.

### Package Structure

- `cmd/` - Cobra command definitions. Each file defines a command group (devices, flows, zones, etc.) with subcommands. Commands follow the pattern: `homey <resource> <action> [args]`
- `internal/client/` - HTTP client for Homey's REST API. All API calls go through `Client.doRequest()` which handles auth headers and error responses
- `internal/config/` - Configuration management using Viper. Config stored in `~/.config/homey-cli/config.toml`

### Adding New Commands

1. Create a new file in `cmd/` (e.g., `cmd/newresource.go`)
2. Define the parent command and subcommands using Cobra
3. Use `apiClient` (from root.go) for API calls
4. Support both JSON and table output via `isTableFormat()` and `outputJSON()`
5. Register commands in `init()` by adding to `rootCmd`

### API Client Pattern

The client returns `json.RawMessage` for GET requests, allowing commands to parse only what they need. Commands that modify state return `error` only.

### Configuration

Config is loaded in `PersistentPreRunE` on rootCmd. Commands that don't need API access (config, version, help) skip loading. Environment variables prefixed with `HOMEY_` override config file values.
