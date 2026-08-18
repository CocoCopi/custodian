package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	tokensCmd.AddCommand(tokensCreateCmd, tokensListCmd, tokensDeleteCmd)
	rootCmd.AddCommand(tokensCmd)
}

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Manage API tokens",
}

var tokensCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		body, _ := json.Marshal(map[string]string{"name": args[0]})
		resp, err := c.post("/api/v1/tokens", body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			return responseError(resp)
		}
		var out struct {
			Plaintext string `json:"plaintext"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return err
		}
		fmt.Println("token created (shown once):")
		fmt.Println(out.Plaintext)
		return nil
	},
}

var tokensListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		resp, err := c.get("/api/v1/tokens")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return responseError(resp)
		}
		var out struct {
			Tokens []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Prefix    string `json:"prefix"`
				CreatedAt string `json:"created_at"`
			} `json:"tokens"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return err
		}
		for _, t := range out.Tokens {
			fmt.Printf("%-24s %-24s %s\n", t.Name, t.Prefix, t.ID)
		}
		return nil
	},
}

var tokensDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Revoke an API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		resp, err := c.delete("/api/v1/tokens/" + args[0])
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			return responseError(resp)
		}
		fmt.Println("token revoked")
		return nil
	},
}
