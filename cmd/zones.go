package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type Zone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Parent string `json:"parent"`
	Icon   string `json:"icon"`
}

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Manage zones",
	Long:  `List and view Homey zones.`,
}

var zonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetZones()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var zones map[string]Zone
			if err := json.Unmarshal(data, &zones); err != nil {
				return fmt.Errorf("failed to parse zones: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tICON\tID")
			fmt.Fprintln(w, "----\t----\t--")
			for _, z := range zones {
				fmt.Fprintf(w, "%s\t%s\t%s\n", z.Name, z.Icon, z.ID)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(zonesCmd)
	zonesCmd.AddCommand(zonesListCmd)
}
