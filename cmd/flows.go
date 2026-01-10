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

Examples:
  homey flows cards --type trigger
  homey flows cards --type action --filter "PultLED"`,
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
	flowsCmd.AddCommand(flowsTriggerCmd)
	flowsCmd.AddCommand(flowsDeleteCmd)
	flowsCmd.AddCommand(flowsCardsCmd)

	flowsCardsCmd.Flags().String("type", "action", "Card type: trigger, condition, action")
	flowsCardsCmd.Flags().String("filter", "", "Filter cards by name or ID")
}
