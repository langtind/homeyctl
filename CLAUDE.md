# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build
make build        # or: go build -o homey .

# Test
make test         # or: go test ./...

# Format
make fmt          # requires: go install mvdan.cc/gofumpt@latest

# Lint
make lint         # requires: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install dev tools
make tools

# Run directly
go run .

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

Config is loaded in `PersistentPreRunE` on rootCmd. Commands that don't need API access (config, version, help, ai) skip loading. Environment variables prefixed with `HOMEY_` override config file values.

## Quick Context

Run `homey ai` to get full documentation for AI assistants - includes flow format, examples, and common patterns.

## Key Learnings

### Flow JSON Format
- **Droptoken format uses pipe (`|`)**: `homey:device:<id>|measure_temperature`
- NOT colon: `homey:device:<id>:measure_temperature` ❌
- The CLI validates this and warns if wrong format is detected

### Homey API Behavior
- **PUT does partial/merge updates** - only fields you send are changed
- To remove conditions/actions, explicitly set empty array: `"conditions": []`
- Omitting a field keeps its existing value

### Output Format
- All list commands return **flat JSON arrays** for easy parsing
- Example: `homey flows list | jq '.[] | select(.name | test("pult";"i"))'`

## Release Process

```bash
# Tag triggers GoReleaser + auto Homebrew tap update
git tag v0.x.x && git push origin v0.x.x
```

GoReleaser builds for darwin/linux/windows (amd64+arm64) and updates `langtind/homebrew-tap`.
