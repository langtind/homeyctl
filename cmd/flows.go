package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type Flow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Triggerable bool   `json:"triggerable"`
	Broken      bool   `json:"broken"`
}

type AdvancedFlow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Triggerable bool   `json:"triggerable"`
	Broken      bool   `json:"broken"`
}

var flowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "Manage flows",
	Long:  `List, trigger, create, and delete Homey flows.`,
}

var flowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all flows",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get both normal and advanced flows
		normalData, err := apiClient.GetFlows()
		if err != nil {
			return err
		}

		advancedData, err := apiClient.GetAdvancedFlows()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var normalFlows map[string]Flow
			var advancedFlows map[string]AdvancedFlow
			json.Unmarshal(normalData, &normalFlows)
			json.Unmarshal(advancedData, &advancedFlows)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tENABLED\tID")
			fmt.Fprintln(w, "----\t----\t-------\t--")

			for _, f := range normalFlows {
				enabled := "yes"
				if !f.Enabled {
					enabled = "no"
				}
				fmt.Fprintf(w, "%s\tsimple\t%s\t%s\n", f.Name, enabled, f.ID)
			}

			for _, f := range advancedFlows {
				enabled := "yes"
				if !f.Enabled {
					enabled = "no"
				}
				fmt.Fprintf(w, "%s\tadvanced\t%s\t%s\n", f.Name, enabled, f.ID)
			}

			w.Flush()
			return nil
		}

		result := map[string]json.RawMessage{
			"flows":         normalData,
			"advancedFlows": advancedData,
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var flowsTriggerCmd = &cobra.Command{
	Use:   "trigger <name-or-id>",
	Short: "Trigger a flow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		// Get all flows to find by name
		normalData, _ := apiClient.GetFlows()
		advancedData, _ := apiClient.GetAdvancedFlows()

		var normalFlows map[string]Flow
		var advancedFlows map[string]AdvancedFlow
		json.Unmarshal(normalData, &normalFlows)
		json.Unmarshal(advancedData, &advancedFlows)

		// Try normal flows first
		for _, f := range normalFlows {
			if f.ID == nameOrID || strings.EqualFold(f.Name, nameOrID) {
				if err := apiClient.TriggerFlow(f.ID); err != nil {
					return err
				}
				fmt.Printf("Triggered flow: %s\n", f.Name)
				return nil
			}
		}

		// Try advanced flows
		for _, f := range advancedFlows {
			if f.ID == nameOrID || strings.EqualFold(f.Name, nameOrID) {
				if err := apiClient.TriggerAdvancedFlow(f.ID); err != nil {
					return err
				}
				fmt.Printf("Triggered advanced flow: %s\n", f.Name)
				return nil
			}
		}

		return fmt.Errorf("flow not found: %s", nameOrID)
	},
}

var flowsGetCmd = &cobra.Command{
	Use:   "get <name-or-id>",
	Short: "Get flow details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		normalData, _ := apiClient.GetFlows()
		advancedData, _ := apiClient.GetAdvancedFlows()

		var normalFlows map[string]json.RawMessage
		var advancedFlows map[string]json.RawMessage
		json.Unmarshal(normalData, &normalFlows)
		json.Unmarshal(advancedData, &advancedFlows)

		// Try normal flows first
		for id, raw := range normalFlows {
			var f Flow
			json.Unmarshal(raw, &f)
			if id == nameOrID || strings.EqualFold(f.Name, nameOrID) {
				outputJSON(raw)
				return nil
			}
		}

		// Try advanced flows
		for id, raw := range advancedFlows {
			var f AdvancedFlow
			json.Unmarshal(raw, &f)
			if id == nameOrID || strings.EqualFold(f.Name, nameOrID) {
				outputJSON(raw)
				return nil
			}
		}

		return fmt.Errorf("flow not found: %s", nameOrID)
	},
}

var flowsCreateCmd = &cobra.Command{
	Use:   "create <json-file>",
	Short: "Create a new flow",
	Long: `Create a new flow from a JSON file.

Use --advanced flag to create an advanced flow.

DISCOVERING IDs:
  - Device IDs:     homey devices list
  - User IDs:       homey users list
  - Flow card IDs:  homey flows cards --type trigger|condition|action
  - Zone IDs:       homey zones list

FLOW JSON STRUCTURE (group fields are auto-added if missing):
{
  "name": "Flow Name",
  "trigger": {
    "id": "homey:manager:presence:user_enter",
    "args": {
      "user": {"id": "<user-id>", "name": "<user-name>"}
    }
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

COMMON TRIGGER IDs:
  homey:manager:presence:user_enter      - A specific user came home
  homey:manager:presence:user_leave      - A user left home
  homey:manager:presence:first_user_enter - First person came home
  homey:manager:presence:last_user_left  - Last person left home

COMMON CONDITION IDs:
  homey:manager:logic:lt                 - Value is less than (use droptoken)
  homey:manager:logic:gt                 - Value is greater than (use droptoken)
  homey:manager:logic:eq                 - Value equals (use droptoken)
  homey:device:<id>:on                   - Device is on/off

COMMON ACTION IDs:
  homey:device:<id>:on                   - Turn device on
  homey:device:<id>:off                  - Turn device off
  homey:device:<id>:toggle               - Toggle device
  homey:device:<id>:dim                  - Set dim level (args: {"dim": 0.5})
  homey:device:<id>:target_temperature_set - Set temperature (args: {"target_temperature": 22})

DROPTOKENS (for logic conditions):
  Reference device capability values using: homey:device:<device-id>|<capability>
  Common capabilities: measure_temperature, measure_humidity, measure_power, onoff

Examples:
  homey flows create flow.json
  homey flows create --advanced advanced-flow.json
  cat flow.json | homey flows create -`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		advanced, _ := cmd.Flags().GetBool("advanced")

		var data []byte
		var err error

		if args[0] == "-" {
			data, err = os.ReadFile("/dev/stdin")
		} else {
			data, err = os.ReadFile(args[0])
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		var flow map[string]interface{}
		if err := json.Unmarshal(data, &flow); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		// Normalize simple flow structure (add required group fields)
		if !advanced {
			normalizeSimpleFlow(flow)
		}

		var result json.RawMessage
		if advanced {
			result, err = apiClient.CreateAdvancedFlow(flow)
		} else {
			result, err = apiClient.CreateFlow(flow)
		}
		if err != nil {
			return err
		}

		var created struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		json.Unmarshal(result, &created)

		flowType := "flow"
		if advanced {
			flowType = "advanced flow"
		}
		fmt.Printf("Created %s: %s (ID: %s)\n", flowType, created.Name, created.ID)
		return nil
	},
}

// normalizeSimpleFlow adds required fields that Homey expects
func normalizeSimpleFlow(flow map[string]interface{}) {
	// Add group to conditions
	if conditions, ok := flow["conditions"].([]interface{}); ok {
		for _, c := range conditions {
			if cond, ok := c.(map[string]interface{}); ok {
				if _, hasGroup := cond["group"]; !hasGroup {
					cond["group"] = "group1"
				}
				if _, hasInverted := cond["inverted"]; !hasInverted {
					cond["inverted"] = false
				}
			}
		}
	}

	// Add group to actions
	if actions, ok := flow["actions"].([]interface{}); ok {
		for _, a := range actions {
			if action, ok := a.(map[string]interface{}); ok {
				if _, hasGroup := action["group"]; !hasGroup {
					action["group"] = "then"
				}
			}
		}
	}
}

var flowsDeleteCmd = &cobra.Command{
	Use:   "delete <name-or-id>",
	Short: "Delete a flow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		// Get all flows to find by name
		normalData, _ := apiClient.GetFlows()
		advancedData, _ := apiClient.GetAdvancedFlows()

		var normalFlows map[string]Flow
		var advancedFlows map[string]AdvancedFlow
		json.Unmarshal(normalData, &normalFlows)
		json.Unmarshal(advancedData, &advancedFlows)

		// Try normal flows first
		for _, f := range normalFlows {
			if f.ID == nameOrID || strings.EqualFold(f.Name, nameOrID) {
				if err := apiClient.DeleteFlow(f.ID); err != nil {
					return err
				}
				fmt.Printf("Deleted flow: %s\n", f.Name)
				return nil
			}
		}

		// Try advanced flows
		for _, f := range advancedFlows {
			if f.ID == nameOrID || strings.EqualFold(f.Name, nameOrID) {
				if err := apiClient.DeleteAdvancedFlow(f.ID); err != nil {
					return err
				}
				fmt.Printf("Deleted advanced flow: %s\n", f.Name)
				return nil
			}
		}

		return fmt.Errorf("flow not found: %s", nameOrID)
	},
}

var flowsCardsCmd = &cobra.Command{
	Use:   "cards",
	Short: "List available flow cards",
	Long: `List available flow cards (triggers, conditions, actions).

Use this to discover card IDs for creating flows.

Card types:
  trigger   - Events that start a flow (user arrives, time, device changes)
  condition - Checks that must pass (temperature < X, device is on)
  action    - Things to do (turn on device, send notification)

The card ID format is: homey:<owner>:<card-name>
  - homey:manager:presence:user_enter (system trigger)
  - homey:device:<device-id>:on (device action)
  - homey:manager:logic:lt (logic condition)

Examples:
  homey flows cards --type trigger
  homey flows cards --type trigger | jq '.[] | select(.id | contains("presence"))'
  homey flows cards --type action | jq '.[] | select(.id | contains("<device-id>"))'
  homey flows cards --type condition | jq '.[] | select(.id | contains("logic"))'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cardType, _ := cmd.Flags().GetString("type")
		filter, _ := cmd.Flags().GetString("filter")

		var data json.RawMessage
		var err error

		switch cardType {
		case "trigger":
			data, err = apiClient.GetFlowTriggers()
		case "condition":
			data, err = apiClient.GetFlowConditions()
		case "action":
			data, err = apiClient.GetFlowActions()
		default:
			return fmt.Errorf("invalid card type: %s (use: trigger, condition, action)", cardType)
		}

		if err != nil {
			return err
		}

		if isTableFormat() {
			var cards []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			}
			if err := json.Unmarshal(data, &cards); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TITLE\tID")
			fmt.Fprintln(w, "-----\t--")

			for _, c := range cards {
				if filter != "" && !strings.Contains(strings.ToLower(c.ID), strings.ToLower(filter)) &&
					!strings.Contains(strings.ToLower(c.Title), strings.ToLower(filter)) {
					continue
				}
				fmt.Fprintf(w, "%s\t%s\n", c.Title, c.ID)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(flowsCmd)
	flowsCmd.AddCommand(flowsListCmd)
	flowsCmd.AddCommand(flowsGetCmd)
	flowsCmd.AddCommand(flowsCreateCmd)
	flowsCmd.AddCommand(flowsTriggerCmd)
	flowsCmd.AddCommand(flowsDeleteCmd)
	flowsCmd.AddCommand(flowsCardsCmd)

	flowsCreateCmd.Flags().Bool("advanced", false, "Create an advanced flow")
	flowsCardsCmd.Flags().String("type", "action", "Card type: trigger, condition, action")
	flowsCardsCmd.Flags().String("filter", "", "Filter cards by name or ID")
}
