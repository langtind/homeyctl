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
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Class           string                `json:"class"`
	Zone            string                `json:"zone"`
	CapabilitiesObj map[string]Capability `json:"capabilitiesObj"`
}

type Capability struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
	Title string      `json:"title"`
}

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices",
	Long:  `List, view, control, and delete Homey devices.`,
}

var devicesMatchFilter string

var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devices",
	Long: `List all devices, optionally filtered by name.

Examples:
  homeyctl devices list
  homeyctl devices list --match "kitchen"
  homeyctl devices list --match "light"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetDevices()
		if err != nil {
			return err
		}

		var devices map[string]Device
		if err := json.Unmarshal(data, &devices); err != nil {
			return fmt.Errorf("failed to parse devices: %w", err)
		}

		// Filter devices if --match is provided
		var filtered []Device
		for _, d := range devices {
			if devicesMatchFilter == "" || strings.Contains(strings.ToLower(d.Name), strings.ToLower(devicesMatchFilter)) {
				filtered = append(filtered, d)
			}
		}

		if isTableFormat() {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCLASS\tID")
			fmt.Fprintln(w, "----\t-----\t--")
			for _, d := range filtered {
				fmt.Fprintf(w, "%s\t%s\t%s\n", d.Name, d.Class, d.ID)
			}
			w.Flush()
			return nil
		}

		out, _ := json.MarshalIndent(filtered, "", "  ")
		fmt.Println(string(out))
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
  homeyctl devices set "PultLED" onoff true
  homeyctl devices set "PultLED" dim 0.5
  homeyctl devices set "Aksels rom" target_temperature 22`,
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

var devicesSetSettingCmd = &cobra.Command{
	Use:   "set-setting <name-or-id> <setting-key> <value>",
	Short: "Set device setting",
	Long: `Set a device setting value.

Device settings are different from capabilities - they configure device behavior
rather than control it. Common settings include:
  - zone_activity_disabled: Exclude sensor from zone activity detection
  - climate_exclude: Exclude device from climate control

Examples:
  homeyctl devices set-setting "Motion Sensor" zone_activity_disabled true
  homeyctl devices set-setting "Thermostat" climate_exclude false`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]
		settingKey := args[1]
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

		var deviceID, deviceName string
		for _, d := range devices {
			if d.ID == nameOrID || strings.EqualFold(d.Name, nameOrID) {
				deviceID = d.ID
				deviceName = d.Name
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

		settings := map[string]interface{}{
			settingKey: value,
		}

		if err := apiClient.SetDeviceSetting(deviceID, settings); err != nil {
			if strings.Contains(err.Error(), "Missing Scopes") {
				return fmt.Errorf(`permission denied: changing device settings requires 'homey.device' scope

OAuth tokens only support 'homey.device.control' (for on/off, dim, etc.),
not full device access needed for settings.

To change device settings, create an API key at my.homey.app:
  1. Go to https://my.homey.app
  2. Select your Homey → Settings → API Keys
  3. Create a new API key (it will have full access)
  4. Run: homeyctl config set-token <your-api-key>`)
			}
			return err
		}

		fmt.Printf("Set %s setting %s = %v\n", deviceName, settingKey, value)
		return nil
	},
}

var devicesGetSettingsCmd = &cobra.Command{
	Use:   "get-settings <name-or-id>",
	Short: "Get device settings",
	Long: `Get all settings for a device.

This shows configurable settings like zone_activity_disabled, climate_exclude,
and driver-specific settings.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		// Find device ID
		data, err := apiClient.GetDevices()
		if err != nil {
			return err
		}

		var devices map[string]Device
		if err := json.Unmarshal(data, &devices); err != nil {
			return fmt.Errorf("failed to parse devices: %w", err)
		}

		var deviceID, deviceName string
		for _, d := range devices {
			if d.ID == nameOrID || strings.EqualFold(d.Name, nameOrID) {
				deviceID = d.ID
				deviceName = d.Name
				break
			}
		}

		if deviceID == "" {
			return fmt.Errorf("device not found: %s", nameOrID)
		}

		settings, err := apiClient.GetDeviceSettings(deviceID)
		if err != nil {
			return err
		}

		if isTableFormat() {
			var settingsMap map[string]interface{}
			if err := json.Unmarshal(settings, &settingsMap); err != nil {
				return fmt.Errorf("failed to parse settings: %w", err)
			}

			fmt.Printf("Settings for %s:\n\n", deviceName)
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SETTING\tVALUE")
			fmt.Fprintln(w, "-------\t-----")
			for key, val := range settingsMap {
				fmt.Fprintf(w, "%s\t%v\n", key, val)
			}
			w.Flush()
			return nil
		}

		outputJSON(settings)
		return nil
	},
}

var devicesValuesCmd = &cobra.Command{
	Use:   "values <name-or-id>",
	Short: "Get all capability values for a device",
	Long: `Get all current capability values for a device.

Useful for multi-sensors and devices with many capabilities.

Examples:
  homeyctl devices values "PultLED"
  homeyctl devices values "Multisensor 6"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		data, err := apiClient.GetDevices()
		if err != nil {
			return err
		}

		var devices map[string]Device
		if err := json.Unmarshal(data, &devices); err != nil {
			return fmt.Errorf("failed to parse devices: %w", err)
		}

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
			fmt.Printf("Values for %s:\n\n", device.Name)
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "CAPABILITY\tVALUE")
			fmt.Fprintln(w, "----------\t-----")
			for _, cap := range device.CapabilitiesObj {
				fmt.Fprintf(w, "%s\t%v\n", cap.ID, cap.Value)
			}
			w.Flush()
			return nil
		}

		// JSON output - just the values
		values := make(map[string]interface{})
		for _, cap := range device.CapabilitiesObj {
			values[cap.ID] = cap.Value
		}
		out, _ := json.MarshalIndent(map[string]interface{}{
			"id":     device.ID,
			"name":   device.Name,
			"values": values,
		}, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var devicesOnCmd = &cobra.Command{
	Use:   "on <name-or-id>",
	Short: "Turn device on",
	Long: `Turn a device on (shorthand for 'devices set <name> onoff true').

Examples:
  homeyctl devices on "Living Room Light"
  homeyctl devices on "Aksels rom"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setDeviceOnOff(args[0], true)
	},
}

var devicesOffCmd = &cobra.Command{
	Use:   "off <name-or-id>",
	Short: "Turn device off",
	Long: `Turn a device off (shorthand for 'devices set <name> onoff false').

Examples:
  homeyctl devices off "Living Room Light"
  homeyctl devices off "Aksels rom"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setDeviceOnOff(args[0], false)
	},
}

func setDeviceOnOff(nameOrID string, on bool) error {
	data, err := apiClient.GetDevices()
	if err != nil {
		return err
	}

	var devices map[string]Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return fmt.Errorf("failed to parse devices: %w", err)
	}

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

	// Check if device supports onoff
	if _, hasOnOff := device.CapabilitiesObj["onoff"]; !hasOnOff {
		return fmt.Errorf("device '%s' does not support on/off", device.Name)
	}

	if err := apiClient.SetCapability(device.ID, "onoff", on); err != nil {
		return err
	}

	state := "on"
	if !on {
		state = "off"
	}
	fmt.Printf("Turned %s %s\n", device.Name, state)
	return nil
}

var devicesRenameCmd = &cobra.Command{
	Use:   "rename <name-or-id> <new-name>",
	Short: "Rename a device",
	Long: `Rename a device.

Examples:
  homeyctl devices rename "Old Name" "New Name"
  homeyctl devices rename abc123-device-id "New Name"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]
		newName := args[1]

		// Get all devices to find by name
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

		updates := map[string]interface{}{
			"name": newName,
		}

		if err := apiClient.UpdateDevice(device.ID, updates); err != nil {
			return err
		}

		fmt.Printf("Renamed device '%s' to '%s'\n", device.Name, newName)
		return nil
	},
}

var devicesDeleteCmd = &cobra.Command{
	Use:   "delete <name-or-id>",
	Short: "Delete a device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		// Get all devices to find by name
		data, err := apiClient.GetDevices()
		if err != nil {
			return err
		}

		var devices map[string]Device
		if err := json.Unmarshal(data, &devices); err != nil {
			return fmt.Errorf("failed to parse devices: %w", err)
		}

		// Find device by name or ID
		for _, d := range devices {
			if d.ID == nameOrID || strings.EqualFold(d.Name, nameOrID) {
				if err := apiClient.DeleteDevice(d.ID); err != nil {
					return err
				}
				fmt.Printf("Deleted device: %s\n", d.Name)
				return nil
			}
		}

		return fmt.Errorf("device not found: %s", nameOrID)
	},
}

func init() {
	rootCmd.AddCommand(devicesCmd)
	devicesCmd.AddCommand(devicesListCmd)
	devicesListCmd.Flags().StringVar(&devicesMatchFilter, "match", "", "Filter devices by name (case-insensitive)")
	devicesCmd.AddCommand(devicesGetCmd)
	devicesCmd.AddCommand(devicesValuesCmd)
	devicesCmd.AddCommand(devicesSetCmd)
	devicesCmd.AddCommand(devicesOnCmd)
	devicesCmd.AddCommand(devicesOffCmd)
	devicesCmd.AddCommand(devicesSetSettingCmd)
	devicesCmd.AddCommand(devicesGetSettingsCmd)
	devicesCmd.AddCommand(devicesRenameCmd)
	devicesCmd.AddCommand(devicesDeleteCmd)
}
