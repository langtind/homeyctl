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
CLI for controlling Homey smart home via local and cloud API.

## IMPORTANT: Scoped Tokens for AI Bots

AI assistants should use **restricted tokens** to prevent accidental damage.

### Creating a Scoped Token

The user (human) should run this command to create a token for you:
` + "```" + `bash
homeyctl token create "AI Bot" --preset readonly --no-save
` + "```" + `

This outputs a token that only has READ access. The AI cannot:
- Control devices
- Delete devices/flows/zones
- Trigger flows
- Modify anything

### Available Presets

| Preset | Access Level | Use Case |
|--------|--------------|----------|
| readonly | Read only | Safe for AI exploration |
| control | Read + control | AI can control devices, trigger flows |
| full | Full access | Same as owner (dangerous) |

### Using the Token

After creating, configure homeyctl with the scoped token:
` + "```" + `bash
homeyctl config set-token <the-token-from-above>
` + "```" + `

### Verifying Your Access Level

If you try an operation you don't have access to, you'll get:
` + "```" + `
Error: 403 Missing Scopes
` + "```" + `

This is expected behavior with a readonly token.

## Quick Setup (Full Access)

For users who want full access:
` + "```" + `bash
homeyctl login
` + "```" + `

## Connection Modes

homeyctl supports local (LAN/VPN) and cloud connections:

` + "```" + `bash
# Set connection mode
homeyctl config set-mode auto   # Prefer local, fallback to cloud (default)
homeyctl config set-mode local  # Always use local
homeyctl config set-mode cloud  # Always use cloud

# Discover Homey on local network (mDNS)
homeyctl config discover        # Returns JSON array of found devices
homeyctl config discover --timeout 10  # Increase search time

# Discovery returns:
# [{"address": "http://10.0.1.1:4859", "homeyId": "abc123", "host": "10.0.1.1", "port": 4859}]

# Configure local connection
homeyctl config set-local http://192.168.1.50 "local-api-key"

# Configure cloud connection
homeyctl config set-cloud "cloud-token"

# View current config
homeyctl config show
` + "```" + `

## Available Commands

### Devices
` + "```" + `bash
homeyctl devices list                  # List all devices
homeyctl devices list --match "kitchen" # Filter by name
homeyctl devices get <id>              # Get device details
homeyctl devices values <name-or-id>   # Get all capability values
homeyctl devices on <name-or-id>       # Turn device on (shorthand)
homeyctl devices off <name-or-id>      # Turn device off (shorthand)
homeyctl devices set <id> <capability> <value>  # Control device
homeyctl devices get-settings <id>     # Get device settings
homeyctl devices set-setting <id> <key> <value>  # Set device setting
homeyctl devices delete <name-or-id>   # Delete a device
` + "```" + `

**Note on Device Settings**: Settings (like ` + "`zone_activity_disabled`" + `) require full ` + "`homey.device`" + ` scope.
OAuth tokens only support ` + "`homey.device.control`" + `. For settings access, create an API key
at https://my.homey.app (Select Homey → Settings → API Keys).

### Flows
` + "```" + `bash
homeyctl flows list                    # List all flows
homeyctl flows list --match "night"    # Filter by name
homeyctl flows get <name-or-id>        # Get flow details
homeyctl flows create <file.json>      # Create flow from JSON
homeyctl flows update <name> <file>    # Update existing flow (merge)
homeyctl flows trigger <name-or-id>    # Trigger a flow by name or ID
homeyctl flows delete <name-or-id>     # Delete a flow
` + "```" + `

### Zones & Users
` + "```" + `bash
homeyctl zones list                    # List all zones
homeyctl zones delete <name-or-id>     # Delete a zone
homeyctl users list                    # List all users
` + "```" + `

### Tokens
` + "```" + `bash
homeyctl token list                    # List existing tokens
homeyctl token create "Name" --preset readonly  # Create scoped token
homeyctl token create "Name" --preset control   # Create control token
homeyctl token delete <id>             # Delete a token
` + "```" + `

### Energy
` + "```" + `bash
homeyctl energy live                   # Live power usage
homeyctl energy report day             # Today's energy report
homeyctl energy report week            # This week's report
homeyctl energy report month --date 2025-12  # December report
homeyctl energy price                  # Show dynamic electricity prices
homeyctl energy price set 0.50         # Set fixed price (e.g., Norgespris)
homeyctl energy price type             # Show current price type
homeyctl energy price type fixed       # Switch to fixed pricing
` + "```" + `

### Insights
` + "```" + `bash
homeyctl insights list                 # List all insight logs
homeyctl insights get <log-id>         # Get historical data
` + "```" + `

### Snapshot
` + "```" + `bash
homeyctl snapshot                      # Get system, zones, devices overview
homeyctl snapshot --include-flows      # Also include flows
` + "```" + `

## Flow JSON Format

### Simple Flow Example
` + "```" + `json
{
  "name": "Turn on lights when arriving",
  "trigger": {
    "id": "homey:manager:presence:user_enter",
    "args": { "user": "user-uuid-here" }
  },
  "conditions": [
    {
      "id": "homey:manager:logic:lt",
      "args": { "value": 20 },
      "droptoken": "homey:device:<device-id>|measure_temperature"
    }
  ],
  "actions": [
    {
      "id": "homey:device:<device-id>:thermostat_mode_heat",
      "args": { "mode": "heat" }
    }
  ]
}
` + "```" + `

## Critical: Droptoken Format

When referencing device capabilities in conditions, use pipe (|) separator:
` + "```" + `
CORRECT: "homey:device:abc123|measure_temperature"
WRONG:   "homey:device:abc123:measure_temperature"
` + "```" + `

## ID Format Reference

| Type | Format | Example |
|------|--------|---------|
| Device action | homey:device:<id>:<capability> | homey:device:abc123:on |
| Manager trigger | homey:manager:<manager>:<event> | homey:manager:presence:user_enter |
| Logic condition | homey:manager:logic:<operator> | homey:manager:logic:lt |
| Droptoken | homey:device:<id>\|<capability> | homey:device:abc123\|measure_temperature |

## Common Triggers
- homey:manager:presence:user_enter - User arrives home
- homey:manager:presence:user_leave - User leaves home
- homey:manager:time:time - At specific time
- homey:device:<id>:<capability>_changed - Device state changes

## Common Conditions
- homey:manager:logic:lt - Less than (use with droptoken)
- homey:manager:logic:gt - Greater than (use with droptoken)
- homey:manager:logic:eq - Equals (use with droptoken)

## Flow Update Behavior

` + "`homeyctl flows update`" + ` does a **partial/merge update**:
- Only fields you include will be changed
- Omitted fields keep their existing values
- To remove conditions/actions, explicitly set empty array: ` + "`\"conditions\": []`" + `

` + "```" + `bash
# Rename a flow
echo '{"name": "New Name"}' > rename.json
homeyctl flows update "Old Name" rename.json

# Remove all conditions from a flow
echo '{"conditions": []}' > clear.json
homeyctl flows update "My Flow" clear.json
` + "```" + `

## Output Format

All list commands return flat JSON arrays for easy parsing:
` + "```" + `bash
# Find flow by name
homeyctl flows list | jq '.[] | select(.name | test("pult";"i"))'

# Get all enabled flows
homeyctl flows list | jq '.[] | select(.enabled)'

# Get device IDs by name
homeyctl devices list | jq '.[] | select(.name | test("office";"i")) | .id'
` + "```" + `

## Workflow Tips

1. **Get device IDs first**: Run ` + "`homeyctl devices list`" + ` to find device IDs
2. **Get user IDs**: Run ` + "`homeyctl users list`" + ` for presence triggers
3. **Check capabilities**: Run ` + "`homeyctl devices get <id>`" + ` to see available capabilities
4. **Validate before creating**: The CLI validates flow JSON and warns about common mistakes
5. **Test flows**: Use ` + "`homeyctl flows trigger \"Flow Name\"`" + ` to test manually

## Utility Commands

` + "```" + `bash
homeyctl version                       # Show version info
homeyctl help                          # Show all commands
homeyctl <command> --help              # Show help for specific command
` + "```" + `
`
