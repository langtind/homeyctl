package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type App struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
	Ready   bool   `json:"ready"`
}

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage apps",
	Long:  `List, view, and restart Homey apps.`,
}

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetApps()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var apps map[string]App
			if err := json.Unmarshal(data, &apps); err != nil {
				return fmt.Errorf("failed to parse apps: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tVERSION\tENABLED\tREADY\tID")
			fmt.Fprintln(w, "----\t-------\t-------\t-----\t--")
			for _, a := range apps {
				enabled := "yes"
				if !a.Enabled {
					enabled = "no"
				}
				ready := "yes"
				if !a.Ready {
					ready = "no"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", a.Name, a.Version, enabled, ready, a.ID)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

var appsGetCmd = &cobra.Command{
	Use:   "get <name-or-id>",
	Short: "Get app details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		data, err := apiClient.GetApps()
		if err != nil {
			return err
		}

		var apps map[string]App
		if err := json.Unmarshal(data, &apps); err != nil {
			return fmt.Errorf("failed to parse apps: %w", err)
		}

		// Find app by name or ID
		var appID string
		for _, a := range apps {
			if a.ID == nameOrID || strings.EqualFold(a.Name, nameOrID) {
				appID = a.ID
				break
			}
		}

		if appID == "" {
			return fmt.Errorf("app not found: %s", nameOrID)
		}

		appData, err := apiClient.GetApp(appID)
		if err != nil {
			return err
		}

		outputJSON(appData)
		return nil
	},
}

var appsRestartCmd = &cobra.Command{
	Use:   "restart <name-or-id>",
	Short: "Restart an app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrID := args[0]

		data, err := apiClient.GetApps()
		if err != nil {
			return err
		}

		var apps map[string]App
		if err := json.Unmarshal(data, &apps); err != nil {
			return fmt.Errorf("failed to parse apps: %w", err)
		}

		// Find app by name or ID
		var app *App
		for _, a := range apps {
			if a.ID == nameOrID || strings.EqualFold(a.Name, nameOrID) {
				app = &a
				break
			}
		}

		if app == nil {
			return fmt.Errorf("app not found: %s", nameOrID)
		}

		if err := apiClient.RestartApp(app.ID); err != nil {
			return err
		}

		fmt.Printf("Restarted app: %s\n", app.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsGetCmd)
	appsCmd.AddCommand(appsRestartCmd)
}
