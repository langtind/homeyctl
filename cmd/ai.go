package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Output context for AI assistants",
	Long:  `Prints documentation and examples to help AI assistants use homey-cli effectively.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(aiContext)
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
}

const aiContext = `# homey-cli - AI Assistant Context

## Overview
CLI for controlling Homey smart home via local API. Requires configuration first.

## Setup
` + "```" + `bash
homey config set-host <homey-ip>    # e.g., 192.168.1.100
homey config set-token <api-token>  # From Homey Developer Tools
homey config show                   # Verify configuration
` + "```" + `

## Available Commands

### Devices
` + "```" + `bash
homey devices list                  # List all devices
homey devices get <id>              # Get device details
homey devices capability <id> <capability> <value>  # Control device
` + "```" + `

### Flows
` + "```" + `bash
homey flows list                    # List all flows
homey flows get <name-or-id>        # Get flow details
homey flows create <file.json>      # Create flow from JSON
homey flows update <name> <file>    # Update existing flow (merge)
homey flows trigger <name-or-id>    # Trigger a flow by name or ID
homey flows delete <name-or-id>     # Delete a flow
` + "```" + `

### Zones & Users
` + "```" + `bash
homey zones list                    # List all zones
homey users list                    # List all users
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

` + "`homey flows update`" + ` does a **partial/merge update**:
- Only fields you include will be changed
- Omitted fields keep their existing values
- To remove conditions/actions, explicitly set empty array: ` + "`\"conditions\": []`" + `

` + "```" + `bash
# Rename a flow
echo '{"name": "New Name"}' > rename.json
homey flows update "Old Name" rename.json

# Remove all conditions from a flow
echo '{"conditions": []}' > clear.json
homey flows update "My Flow" clear.json
` + "```" + `

## Workflow Tips

1. **Get device IDs first**: Run ` + "`homey devices list`" + ` to find device IDs
2. **Get user IDs**: Run ` + "`homey users list`" + ` for presence triggers
3. **Check capabilities**: Run ` + "`homey devices get <id>`" + ` to see available capabilities
4. **Validate before creating**: The CLI validates flow JSON and warns about common mistakes
5. **Test flows**: Use ` + "`homey flows trigger \"Flow Name\"`" + ` to test manually
`
