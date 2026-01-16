package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Output context for AI assistants",
	Long:  `Prints documentation and examples to help AI assistants use homeyctl effectively.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(aiContext)
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
}

const aiContext = `# homeyctl - AI Assistant Context

## Overview

CLI for controlling Homey smart home via local and cloud API. This document helps AI assistants use homeyctl effectively and safely.

## IMPORTANT: Token Security for AI

AI assistants should use **restricted tokens** to prevent accidental damage.

### Creating a Scoped Token (Human runs this)

` + "```bash" + `
# Read-only token (RECOMMENDED for AI)
homeyctl token create "AI Bot" --preset readonly --no-save

# Control token (can control devices, trigger flows)
homeyctl token create "AI Bot" --preset control --no-save
` + "```" + `

### Token Presets

| Preset | Can Read | Can Control | Can Delete | Use Case |
|--------|----------|-------------|------------|----------|
| readonly | Yes | No | No | Safe AI exploration |
| control | Yes | Yes | No | AI automation |
| full | Yes | Yes | Yes | Full access (dangerous) |

### Access Denied Response

With a readonly token, write operations return:
` + "```" + `
Error: 403 Missing Scopes
` + "```" + `
This is expected - inform the user they need a higher access token.

---

## Quick Reference

### Read-Only Commands (Safe with readonly token)

` + "```bash" + `
# Devices
homeyctl devices list                    # List all devices
homeyctl devices list --match "kitchen"  # Filter by name
homeyctl devices get "Device"            # Get device details
homeyctl devices values "Device"         # Get capability values
homeyctl devices get-settings "Device"   # Get device settings

# Zones
homeyctl zones list                      # List zones
homeyctl zones get "Zone"                # Get zone details
homeyctl zones icons                     # List available icons

# Flows
homeyctl flows list                      # List flows
homeyctl flows get "Flow"                # Get flow details
homeyctl flows cards --type trigger      # List available triggers
homeyctl flows cards --type condition    # List conditions
homeyctl flows cards --type action       # List actions
homeyctl flows folders list              # List flow folders

# Users & Presence
homeyctl users list                      # List users
homeyctl users me                        # Current user
homeyctl users get "User"                # Get user details
homeyctl presence get me                 # Get presence status
homeyctl presence get "User"
homeyctl presence asleep get me          # Get sleep status

# Apps
homeyctl apps list                       # List installed apps
homeyctl apps get "App"                  # Get app details
homeyctl apps usage "App"                # Resource usage
homeyctl apps settings list "App"        # List app settings

# Smart Home Status
homeyctl moods list                      # List moods
homeyctl moods get "Mood"                # Get mood details
homeyctl dashboards list                 # List dashboards
homeyctl dashboards get "Dashboard"
homeyctl weather current                 # Current weather
homeyctl weather forecast                # Weather forecast

# Energy
homeyctl energy live                     # Live power usage
homeyctl energy report day               # Today's report
homeyctl energy report week              # Weekly report
homeyctl energy report month             # Monthly report
homeyctl energy report year              # Yearly report
homeyctl energy price                    # Electricity prices
homeyctl energy currency                 # Currency setting

# Insights & History
homeyctl insights list                   # List insight logs
homeyctl insights get "log-id"           # Get historical data
homeyctl insights get "id" --resolution lastWeek

# System
homeyctl system info                     # System information
homeyctl system name get                 # Get Homey name
homeyctl variables list                  # List logic variables
homeyctl variables get "var"             # Get variable value
homeyctl notify list                     # List notifications
homeyctl notify owners                   # Notification sources
homeyctl snapshot                        # System overview
` + "```" + `

### Control Commands (Require control or full token)

` + "```bash" + `
# Device Control
homeyctl devices on "Device"             # Turn on
homeyctl devices off "Device"            # Turn off
homeyctl devices set "Device" dim 0.5    # Set capability
homeyctl devices set "Device" target_temperature 22

# Presence Control
homeyctl presence set me home            # Set home
homeyctl presence set me away            # Set away
homeyctl presence asleep set me asleep   # Set asleep
homeyctl presence asleep set me awake    # Set awake

# Flow Control
homeyctl flows trigger "Flow"            # Trigger a flow

# Mood Control
homeyctl moods set "Mood"                # Activate mood

# Variables
homeyctl variables set "var" value       # Set variable

# Notifications
homeyctl notify send "Message"           # Send notification
` + "```" + `

### Write/Delete Commands (Require full token)

` + "```bash" + `
# Devices
homeyctl devices rename "Old" "New"
homeyctl devices move "Device" "Zone"
homeyctl devices delete "Device"
homeyctl devices hide "Device"
homeyctl devices unhide "Device"
homeyctl devices set-icon "Device" icon.png
homeyctl devices set-note "Device" "Note"
homeyctl devices set-setting "Device" key value

# Zones
homeyctl zones create "Zone"
homeyctl zones rename "Old" "New"
homeyctl zones move "Zone" "Parent"
homeyctl zones set-icon "Zone" "icon"
homeyctl zones delete "Zone"

# Flows
homeyctl flows create flow.json
homeyctl flows update "Flow" changes.json
homeyctl flows delete "Flow"
homeyctl flows folders create "Folder"
homeyctl flows folders update "Folder" --name "New"
homeyctl flows folders delete "Folder"

# Users
homeyctl users create "User"
homeyctl users delete "User"

# Moods
homeyctl moods create "Mood"
homeyctl moods update "Mood" --name "New"
homeyctl moods delete "Mood"

# Dashboards
homeyctl dashboards create "Dashboard"
homeyctl dashboards update "Dashboard" --name "New"
homeyctl dashboards delete "Dashboard"

# Apps
homeyctl apps install com.app.id
homeyctl apps uninstall com.app.id
homeyctl apps enable com.app.id
homeyctl apps disable com.app.id
homeyctl apps restart com.app.id
homeyctl apps settings set "App" key value

# Variables
homeyctl variables create "var" number 0
homeyctl variables delete "var"

# Notifications
homeyctl notify delete <id>
homeyctl notify clear

# Insights
homeyctl insights delete "log-id"
homeyctl insights clear "log-id"

# Energy
homeyctl energy price set 0.50
homeyctl energy price type fixed
homeyctl energy delete --force

# System
homeyctl system name set "Name"
homeyctl system reboot --force
` + "```" + `

---

## Flow JSON Format

### Simple Flow Example

` + "```json" + `
{
  "name": "Turn on heater when cold",
  "trigger": {
    "id": "homey:manager:presence:user_enter",
    "args": { "user": {"id": "user-uuid", "name": "User"} }
  },
  "conditions": [
    {
      "id": "homey:manager:logic:lt",
      "droptoken": "homey:device:abc123|measure_temperature",
      "args": { "comparator": 20 }
    }
  ],
  "actions": [
    { "id": "homey:device:def456:on", "args": {} },
    { "id": "homey:device:def456:target_temperature_set", "args": { "target_temperature": 22 } }
  ]
}
` + "```" + `

### CRITICAL: Droptoken Format

When referencing device capabilities in conditions, use **pipe (|)** separator:

` + "```" + `
CORRECT: "homey:device:abc123|measure_temperature"
WRONG:   "homey:device:abc123:measure_temperature"
` + "```" + `

### ID Format Reference

| Type | Format | Example |
|------|--------|---------|
| Device action | homey:device:<id>:<capability> | homey:device:abc123:on |
| Manager trigger | homey:manager:<mgr>:<event> | homey:manager:presence:user_enter |
| Logic condition | homey:manager:logic:<op> | homey:manager:logic:lt |
| Droptoken | homey:device:<id>\|<cap> | homey:device:abc123\|measure_temperature |

### Common Triggers

- ` + "`homey:manager:presence:user_enter`" + ` - User arrives home
- ` + "`homey:manager:presence:user_leave`" + ` - User leaves home
- ` + "`homey:manager:time:time`" + ` - At specific time
- ` + "`homey:device:<id>:<cap>_changed`" + ` - Device state changes

### Common Conditions

- ` + "`homey:manager:logic:lt`" + ` - Less than (use with droptoken)
- ` + "`homey:manager:logic:gt`" + ` - Greater than
- ` + "`homey:manager:logic:eq`" + ` - Equals

### Flow Update Behavior

` + "`homeyctl flows update`" + ` does **partial/merge updates**:
- Only included fields are changed
- Omitted fields keep existing values
- To clear: ` + "`\"conditions\": []`" + `

---

## Output & Parsing

All list commands return flat JSON arrays:

` + "```bash" + `
# Find device by name
homeyctl devices list | jq '.[] | select(.name | test("light";"i"))'

# Get enabled flows
homeyctl flows list | jq '.[] | select(.enabled)'

# Get device IDs
homeyctl devices list | jq -r '.[].id'
` + "```" + `

---

## Workflow Tips

1. **Get IDs first**: Use ` + "`homeyctl devices list`" + ` or ` + "`homeyctl users list`" + `
2. **Check capabilities**: Use ` + "`homeyctl devices get <id>`" + ` to see what a device can do
3. **List flow cards**: Use ` + "`homeyctl flows cards --type action`" + ` to find available actions
4. **Test flows**: Use ` + "`homeyctl flows trigger \"Name\"`" + ` to test manually
5. **Verify access**: If you get 403, inform user they need a different token preset

---

## Connection Setup

` + "```bash" + `
# Quick login (opens browser)
homeyctl login

# Set connection mode
homeyctl config set-mode auto    # Prefer local, fallback cloud
homeyctl config set-mode local   # Always local
homeyctl config set-mode cloud   # Always cloud

# Discover Homey on network
homeyctl config discover

# Manual setup
homeyctl config set-local http://192.168.1.50 <token>
homeyctl config set-cloud <token>

# View config
homeyctl config show
` + "```" + `

---

## Help Commands

` + "```bash" + `
homeyctl help                    # All commands
homeyctl <command> --help        # Command help
homeyctl version                 # Version info
` + "```" + `
`
