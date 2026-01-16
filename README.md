# homeyctl

![homeyctl banner](banner.png)

A command-line interface for controlling [Homey](https://homey.app) smart home devices via local and cloud API.

> **Note:** The binary is named `homeyctl` to avoid conflicts with Athom's official `homey` CLI tool used for app development.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install langtind/tap/homeyctl
```

### Download Binary

Download from [Releases](https://github.com/langtind/homeyctl/releases) and add to your PATH.

### Build from Source

```bash
go install github.com/langtind/homeyctl@latest
```

## Quick Start

```bash
# Login with your Athom account (opens browser)
homeyctl login

# List your devices
homeyctl devices list

# Control a device
homeyctl devices on "Living Room Light"

# Get system overview
homeyctl snapshot
```

## Configuration

### Connection Modes

homeyctl supports local (LAN) and cloud connections:

```bash
homeyctl config set-mode auto    # Prefer local, fallback to cloud (default)
homeyctl config set-mode local   # Always use local
homeyctl config set-mode cloud   # Always use cloud
```

### Auto-Discovery

```bash
homeyctl config discover         # Find Homey on local network (mDNS)
```

### Manual Configuration

```bash
# Local connection
homeyctl config set-local http://192.168.1.50 <token>

# Cloud connection
homeyctl config set-cloud <token>

# View current config
homeyctl config show
```

### Creating API Tokens

```bash
# For AI bots (read-only, safe)
homeyctl token create "AI Bot" --preset readonly --no-save

# For automation (can control devices)
homeyctl token create "Automation" --preset control --no-save
```

Available presets: `readonly`, `control`, `full`

Configuration is stored in `~/.config/homeyctl/config.toml`.

---

## Command Reference

### Devices

The most commonly used commands for controlling your smart home.

```bash
# List and search
homeyctl devices list                        # List all devices
homeyctl devices list --match "kitchen"      # Filter by name
homeyctl devices get "Device Name"           # Get device details
homeyctl devices values "Device Name"        # Get all capability values

# Control
homeyctl devices on "Living Room Light"      # Turn on
homeyctl devices off "Living Room Light"     # Turn off
homeyctl devices set "Light" dim 0.5         # Set capability value
homeyctl devices set "Thermostat" target_temperature 22

# Management
homeyctl devices rename "Old Name" "New Name"
homeyctl devices move "Device" "New Zone"
homeyctl devices delete "Device"
homeyctl devices hide "Device"               # Hide from UI
homeyctl devices unhide "Device"             # Show in UI

# Customization
homeyctl devices set-icon "Device" icon.png  # Custom icon
homeyctl devices set-note "Device" "Note"    # Add a note

# Settings (separate from capabilities)
homeyctl devices get-settings "Motion Sensor"
homeyctl devices set-setting "Motion Sensor" motion_sensitivity high
```

### Zones

Organize your home into zones.

```bash
homeyctl zones list                          # List all zones
homeyctl zones get "Living Room"             # Get zone details
homeyctl zones create "New Zone"             # Create zone
homeyctl zones create "Bedroom" --parent "Upstairs"  # Nested zone
homeyctl zones rename "Old" "New"            # Rename
homeyctl zones move "Zone" "New Parent"      # Move to different parent
homeyctl zones delete "Zone"                 # Delete

# Icons
homeyctl zones icons                         # List available icons
homeyctl zones set-icon "Zone" "bedroom"     # Set zone icon
```

### Flows

Automate your home with flows.

```bash
# List and view
homeyctl flows list                          # List all flows
homeyctl flows list --match "morning"        # Filter by name
homeyctl flows get "Flow Name"               # Get flow details

# Control
homeyctl flows trigger "Good Morning"        # Trigger manually

# Create and modify
homeyctl flows create flow.json              # Create from JSON
homeyctl flows create --advanced flow.json   # Create advanced flow
homeyctl flows update "Flow" changes.json    # Update (merge)
homeyctl flows delete "Flow"                 # Delete

# Flow cards (for creating flows)
homeyctl flows cards --type trigger          # List triggers
homeyctl flows cards --type condition        # List conditions
homeyctl flows cards --type action           # List actions
```

#### Flow Folders

Organize flows into folders.

```bash
homeyctl flows folders list                  # List folders
homeyctl flows folders get "Folder"          # Get folder details
homeyctl flows folders create "New Folder"   # Create
homeyctl flows folders update "Folder" --name "New Name"
homeyctl flows folders delete "Folder"       # Delete
```

### Presence

Track and control user presence (home/away) and sleep status.

```bash
# Check presence
homeyctl presence get me                     # Your status
homeyctl presence get "User Name"            # Other user

# Set presence
homeyctl presence set me home                # Mark as home
homeyctl presence set me away                # Mark as away
homeyctl presence set "User" home

# Sleep status
homeyctl presence asleep get me
homeyctl presence asleep set me asleep       # Mark as sleeping
homeyctl presence asleep set me awake        # Mark as awake
```

### Moods

Control room moods and ambiances.

```bash
homeyctl moods list                          # List all moods
homeyctl moods get "Relaxed"                 # Get mood details
homeyctl moods set "Movie Night"             # Activate a mood
homeyctl moods create "New Mood"             # Create
homeyctl moods update "Mood" --name "New"    # Update
homeyctl moods delete "Mood"                 # Delete
```

### Weather

Get weather information from Homey.

```bash
homeyctl weather current                     # Current conditions
homeyctl weather forecast                    # Hourly forecast
```

### Energy

Monitor energy consumption and electricity prices.

```bash
# Live usage
homeyctl energy live                         # Current power consumption

# Reports
homeyctl energy report day                   # Today
homeyctl energy report day --date 2025-01-10
homeyctl energy report week                  # This week
homeyctl energy report month --date 2025-01  # January
homeyctl energy report year                  # This year
homeyctl energy report year 2024             # Specific year

# Electricity prices
homeyctl energy price                        # Show prices
homeyctl energy price set 0.50               # Set fixed price
homeyctl energy price type                   # Show price type
homeyctl energy price type fixed             # Use fixed pricing
homeyctl energy price type dynamic           # Use dynamic pricing

# Management
homeyctl energy currency                     # Show currency
homeyctl energy delete --force               # Delete all reports
```

### Apps

Manage installed Homey apps.

```bash
# List and view
homeyctl apps list                           # List all apps
homeyctl apps get "App Name"                 # Get app details
homeyctl apps usage "App Name"               # Resource usage

# Control
homeyctl apps restart com.app.id             # Restart app
homeyctl apps enable com.app.id              # Enable
homeyctl apps disable com.app.id             # Disable

# Install/Uninstall
homeyctl apps install com.app.id             # Install from store
homeyctl apps install com.app.id --channel test  # Test channel
homeyctl apps uninstall com.app.id           # Uninstall

# Settings
homeyctl apps settings list "App"            # List settings
homeyctl apps settings set "App" key value   # Set setting
```

### Users

Manage Homey users.

```bash
homeyctl users list                          # List all users
homeyctl users get "User Name"               # Get user details
homeyctl users me                            # Get current user
homeyctl users create "New User"             # Create user
homeyctl users delete "User"                 # Delete user
```

### Dashboards

Manage Homey dashboards.

```bash
homeyctl dashboards list                     # List dashboards
homeyctl dashboards get "Dashboard"          # Get details
homeyctl dashboards create "New Dashboard"   # Create
homeyctl dashboards update "Dashboard" --name "New Name"
homeyctl dashboards delete "Dashboard"       # Delete
```

### Notifications

Send and manage timeline notifications.

```bash
homeyctl notify send "Hello from CLI"        # Send notification
homeyctl notify list                         # List notifications
homeyctl notify delete <id>                  # Delete one
homeyctl notify clear                        # Clear all
homeyctl notify owners                       # List sources
```

### Insights

Access historical data and logs.

```bash
# List and view
homeyctl insights list                       # List all logs
homeyctl insights get "log-id"               # Get data (last 24h)
homeyctl insights get "log-id" --resolution lastWeek
homeyctl insights get "log-id" --resolution lastMonth

# Management
homeyctl insights delete "log-id"            # Delete log
homeyctl insights clear "log-id"             # Clear entries only
```

Resolutions: `last24Hours`, `lastWeek`, `lastMonth`, `lastYear`, `last2Years`

### Variables

Manage logic variables for flows.

```bash
homeyctl variables list                      # List all
homeyctl variables get "my_var"              # Get value
homeyctl variables set "my_var" 42           # Set value
homeyctl variables create "new_var" number 0 # Create
homeyctl variables delete "my_var"           # Delete
```

### System

System information and control.

```bash
homeyctl system info                         # System information
homeyctl system name get                     # Get Homey name
homeyctl system name set "My Homey"          # Set Homey name
homeyctl system users                        # List system users
homeyctl system reboot --force               # Reboot Homey
```

### Snapshot

Get a quick overview of your system.

```bash
homeyctl snapshot                            # System, zones, devices
homeyctl snapshot --include-flows            # Include flows
```

---

## Output Formats

```bash
# JSON output (default, good for scripting)
homeyctl devices list

# Table output (human readable)
homeyctl devices list --format table

# Set default format
homeyctl config set-format table
```

### Parsing JSON with jq

```bash
# Find devices by name
homeyctl devices list | jq '.[] | select(.name | test("light";"i"))'

# Get all enabled flows
homeyctl flows list | jq '.[] | select(.enabled)'

# Get device IDs in a zone
homeyctl devices list | jq '.[] | select(.zone == "zone-id") | .id'
```

---

## Creating Flows

Create flows from JSON files. The CLI validates your JSON and warns about common mistakes.

### Simple Flow Example

```json
{
  "name": "Heat office on arrival",
  "trigger": {
    "id": "homey:manager:presence:user_enter",
    "args": { "user": {"id": "<user-id>", "name": "User"} }
  },
  "conditions": [
    {
      "id": "homey:manager:logic:lt",
      "droptoken": "homey:device:<device-id>|measure_temperature",
      "args": {"comparator": 20}
    }
  ],
  "actions": [
    {"id": "homey:device:<device-id>:on", "args": {}},
    {"id": "homey:device:<device-id>:target_temperature_set", "args": {"target_temperature": 23}}
  ]
}
```

### Important: Droptoken Format

When referencing device capabilities in conditions, use pipe (`|`) as separator:

```
CORRECT: "homey:device:abc123|measure_temperature"
WRONG:   "homey:device:abc123:measure_temperature"
```

### Flow Update Behavior

`homeyctl flows update` does a **partial/merge update**:
- Only fields you include will be changed
- Omitted fields keep their existing values
- To remove conditions/actions, explicitly set empty array: `"conditions": []`

---

## AI Assistant Support

Get context for AI assistants (Claude, ChatGPT, etc.):

```bash
homeyctl ai
```

This outputs documentation, examples, and flow JSON format - perfect for AI chat or project context.

---

## Environment Variables

All config options can be set via environment variables (prefix `HOMEY_`):

```bash
export HOMEY_MODE=auto              # auto, local, or cloud
export HOMEY_LOCAL_ADDRESS=http://192.168.1.50
export HOMEY_LOCAL_TOKEN=your-local-token
export HOMEY_FORMAT=table           # json or table
```

