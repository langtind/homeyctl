package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Present  bool   `json:"present"`
	Asleep   bool   `json:"asleep"`
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  `List and view Homey users.`,
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := apiClient.GetUsers()
		if err != nil {
			return err
		}

		if isTableFormat() {
			var users map[string]User
			if err := json.Unmarshal(data, &users); err != nil {
				return fmt.Errorf("failed to parse users: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tROLE\tPRESENT\tID")
			fmt.Fprintln(w, "----\t----\t-------\t--")
			for _, u := range users {
				present := "no"
				if u.Present {
					present = "yes"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", u.Name, u.Role, present, u.ID)
			}
			w.Flush()
			return nil
		}

		outputJSON(data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersListCmd)
}
