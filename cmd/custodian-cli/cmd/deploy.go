package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy <name-or-id>",
	Short: "Trigger a new deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		commit, _ := cmd.Flags().GetString("commit")

		path := "/api/v1/services/" + args[0] + "/deployments"
		if commit != "" {
			path += "?commit=" + commit
		}
		resp, err := c.post(path, nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusAccepted {
			return responseError(resp)
		}
		var dep struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&dep); err != nil {
			return err
		}
		fmt.Printf("deployment %s queued (status=%s)\n", dep.ID, dep.Status)
		fmt.Printf("watch logs with: custodian logs %s\n", args[0])
		return nil
	},
}

func init() {
	deployCmd.Flags().String("commit", "", "git commit SHA being deployed")
}
