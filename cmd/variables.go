package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type Variable struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

var varsCmd = &cobra.Command{
	Use:     "variables",
	Aliases: []string{"vars", "var"},
	Short:   "Manage logic variables",
	Long:    `List, create, update, and delete Homey logic variables.`,
}

var varsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetVariables()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var vars map[string]Variable
			if err := json.Unmarshal(data, &vars); err != nil {
				return fmt.Errorf("failed to parse variables: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tVALUE\tID")
			fmt.Fprintln(w, "----\t----\t-----\t--")
			for _, v := range vars {
				fmt.Fprintf(w, "%s\t%s\t%v\t%s\n", v.Name, v.Type, v.Value, v.ID)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

var varsGetCmd = &cobra.Command{
	Use:   "get <name-or-id>",
	Short: "Get variable value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		data, err := apiClient.GetVariables()
		if err != nil {
			return err
		}

		var vars map[string]Variable
		if err := json.Unmarshal(data, &vars); err != nil {
			return fmt.Errorf("failed to parse variables: %w", err)
		}

		// Find variable by name or ID
		var variable *Variable
		for _, v := range vars {
			if v.ID == nameOrID || strings.EqualFold(v.Name, nameOrID) {
				variable = &v
				break
			}
		}

		if variable == nil {
			return fmt.Errorf("variable not found: %s", nameOrID)
		}

		if isTableFormat() {
			fmt.Printf("Name:  %s\n", variable.Name)
			fmt.Printf("Type:  %s\n", variable.Type)
			fmt.Printf("Value: %v\n", variable.Value)
			fmt.Printf("ID:    %s\n", variable.ID)
			return nil
		}

		out, _ := json.MarshalIndent(variable, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var varsSetCmd = &cobra.Command{
	Use:   "set <name-or-id> <value>",
	Short: "Set variable value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]
		valueStr := args[1]

		data, err := apiClient.GetVariables()
		if err != nil {
			return err
		}

		var vars map[string]Variable
		if err := json.Unmarshal(data, &vars); err != nil {
			return fmt.Errorf("failed to parse variables: %w", err)
		}

		// Find variable by name or ID
		var variable *Variable
		for _, v := range vars {
			if v.ID == nameOrID || strings.EqualFold(v.Name, nameOrID) {
				variable = &v
				break
			}
		}

		if variable == nil {
			return fmt.Errorf("variable not found: %s", nameOrID)
		}

		// Parse value based on variable type
		var value interface{}
		switch variable.Type {
		case "boolean":
			value = valueStr == "true" || valueStr == "1" || valueStr == "yes"
		case "number":
			var num float64
			if _, err := fmt.Sscanf(valueStr, "%f", &num); err != nil {
				return fmt.Errorf("invalid number: %s", valueStr)
			}
			value = num
		default:
			value = valueStr
		}

		if err := apiClient.SetVariable(variable.ID, value); err != nil {
			return err
		}

		fmt.Printf("Set %s = %v\n", variable.Name, value)
		return nil
	},
}

var varsCreateCmd = &cobra.Command{
	Use:   "create <name> <type> <value>",
	Short: "Create a new variable",
	Long: `Create a new logic variable.

Types: string, number, boolean

Examples:
  homey variables create myvar string "hello"
  homey variables create counter number 0
  homey variables create enabled boolean true`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		varType := args[1]
		valueStr := args[2]

		// Validate type
		switch varType {
		case "string", "number", "boolean":
			// ok
		default:
			return fmt.Errorf("invalid type: %s (use: string, number, boolean)", varType)
		}

		// Parse value based on type
		var value interface{}
		switch varType {
		case "boolean":
			value = valueStr == "true" || valueStr == "1" || valueStr == "yes"
		case "number":
			var num float64
			if _, err := fmt.Sscanf(valueStr, "%f", &num); err != nil {
				return fmt.Errorf("invalid number: %s", valueStr)
			}
			value = num
		default:
			value = valueStr
		}

		result, err := apiClient.CreateVariable(name, varType, value)
		if err != nil {
			return err
		}

		var created Variable
		if err := json.Unmarshal(result, &created); err == nil {
			fmt.Printf("Created variable: %s (id: %s)\n", created.Name, created.ID)
		} else {
			fmt.Println("Variable created")
		}
		return nil
	},
}

var varsDeleteCmd = &cobra.Command{
	Use:   "delete <name-or-id>",
	Short: "Delete a variable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		data, err := apiClient.GetVariables()
		if err != nil {
			return err
		}

		var vars map[string]Variable
		if err := json.Unmarshal(data, &vars); err != nil {
			return fmt.Errorf("failed to parse variables: %w", err)
		}

		// Find variable by name or ID
		var variable *Variable
		for _, v := range vars {
			if v.ID == nameOrID || strings.EqualFold(v.Name, nameOrID) {
				variable = &v
				break
			}
		}

		if variable == nil {
			return fmt.Errorf("variable not found: %s", nameOrID)
		}

		if err := apiClient.DeleteVariable(variable.ID); err != nil {
			return err
		}

		fmt.Printf("Deleted variable: %s\n", variable.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(varsCmd)
	varsCmd.AddCommand(varsListCmd)
	varsCmd.AddCommand(varsGetCmd)
	varsCmd.AddCommand(varsSetCmd)
	varsCmd.AddCommand(varsCreateCmd)
	varsCmd.AddCommand(varsDeleteCmd)
}
