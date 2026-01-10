# homey-cli

A command-line interface for controlling [Homey](https://homey.app) smart home devices via the local API.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap langtind/tap
brew install homey-cli
```

### Download Binary

Download from [Releases](https://github.com/langtind/homey-cli/releases) and add to your PATH.

### Build from Source

```bash
go install github.com/langtind/homey-cli@latest
```

## Configuration

First, get your Homey's local IP and API token from the Homey Developer Tools.

```bash
# Set your Homey's IP address
homey config set-host 192.168.1.100

# Set your API token
homey config set-token <your-token>

# Verify configuration
homey config show
```

Configuration is stored in `~/.config/homey-cli/config.toml`.

## Usage

### Devices

```bash
# List all devices
homey devices list

# Get device details
homey devices get "Living Room Light"

# Control devices
homey devices set "Living Room Light" onoff true
homey devices set "Living Room Light" dim 0.5
homey devices set "Thermostat" target_temperature 22
```

### Flows

```bash
# List all flows
homey flows list

# Trigger a flow
homey flows trigger "Good Morning"

# List available flow cards
homey flows cards --type action
homey flows cards --type trigger --filter motion
```

### Zones

```bash
homey zones list
```

### Apps

```bash
# List installed apps
homey apps list

# Restart an app
homey apps restart com.some.app
```

### Variables

```bash
# List logic variables
homey variables list

# Get/set variable
homey variables get "my_variable"
homey variables set "my_variable" 42

# Create/delete variable
homey variables create "new_var" number 0
homey variables delete "new_var"
```

### Notifications

```bash
# Send notification to Homey timeline
homey notifications send "Hello from CLI"

# List notifications
homey notifications list
```

### System

```bash
# System info
homey system info

# Reboot Homey (use with caution)
homey system reboot
```

## Output Formats

```bash
# JSON output (default)
homey devices list

# Table output
homey devices list --format table

# Set default format
homey config set-format table
```

## Environment Variables

All config options can be set via environment variables:

```bash
export HOMEY_HOST=192.168.1.100
export HOMEY_TOKEN=your-token
export HOMEY_FORMAT=table
```