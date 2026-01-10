package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type Device struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Class           string                 `json:"class"`
	Zone            string                 `json:"zone"`
	CapabilitiesObj map[string]Capability  `json:"capabilitiesObj"`
}

type Capability struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
	Title string      `json:"title"`
}

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices",
	Long:  `List, view, and control Homey devices.`,
}

var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetDevices()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var devices map[string]Device
			if err := json.Unmarshal(data, &devices); err != nil {
				return fmt.Errorf("failed to parse devices: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCLASS\tID")
			fmt.Fprintln(w, "----\t-----\t--")
			for _, d := range devices {
				fmt.Fprintf(w, "%s\t%s\t%s\n", d.Name, d.Class, d.ID)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

var devicesGetCmd = &cobra.Command{
	Use:   "get <name-or-id>",
	Short: "Get device details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		// First get all devices to find by name
		data, err := apiClient.GetDevices()
		if err != nil {
			return err
		}

		var devices map[string]Device
		if err := json.Unmarshal(data, &devices); err != nil {
			return fmt.Errorf("failed to parse devices: %w", err)
		}

		// Find device by name or ID
		var device *Device
		for _, d := range devices {
			if d.ID == nameOrID || strings.EqualFold(d.Name, nameOrID) {
				device = &d
				break
			}
		}

		if device == nil {
			return fmt.Errorf("device not found: %s", nameOrID)
		}

		if isTableFormat() {
			fmt.Printf("Name:  %s\n", device.Name)
			fmt.Printf("Class: %s\n", device.Class)
			fmt.Printf("ID:    %s\n", device.ID)
			fmt.Println("\nCapabilities:")

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  CAPABILITY\tVALUE")
			fmt.Fprintln(w, "  ----------\t-----")
			for _, cap := range device.CapabilitiesObj {
				fmt.Fprintf(w, "  %s\t%v\n", cap.ID, cap.Value)
			}
			w.Flush()
			return nil
		}

		out, _ := json.MarshalIndent(device, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var devicesSetCmd = &cobra.Command{
	Use:   "set <name-or-id> <capability> <value>",
	Short: "Set device capability",
	Long: `Set a device capability value.

Examples:
  homey devices set "PultLED" onoff true
  homey devices set "PultLED" dim 0.5
  homey devices set "Aksels rom" target_temperature 22`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]
		capability := args[1]
		valueStr := args[2]

		// Find device ID
		data, err := apiClient.GetDevices()
		if err != nil {
			return err
		}

		var devices map[string]Device
		if err := json.Unmarshal(data, &devices); err != nil {
			return fmt.Errorf("failed to parse devices: %w", err)
		}

		var deviceID string
		for _, d := range devices {
			if d.ID == nameOrID || strings.EqualFold(d.Name, nameOrID) {
				deviceID = d.ID
				break
			}
		}

		if deviceID == "" {
			return fmt.Errorf("device not found: %s", nameOrID)
		}

		// Parse value
		var value interface{}
		if valueStr == "true" {
			value = true
		} else if valueStr == "false" {
			value = false
		} else {
			// Try as number
			var num float64
			if _, err := fmt.Sscanf(valueStr, "%f", &num); err == nil {
				value = num
			} else {
				value = valueStr
			}
		}

		if err := apiClient.SetCapability(deviceID, capability, value); err != nil {
			return err
		}

		fmt.Printf("Set %s.%s = %v\n", nameOrID, capability, value)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(devicesCmd)
	devicesCmd.AddCommand(devicesListCmd)
	devicesCmd.AddCommand(devicesGetCmd)
	devicesCmd.AddCommand(devicesSetCmd)
}
