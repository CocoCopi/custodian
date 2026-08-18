package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the control plane",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := flagURL
		if url == "" {
			url = "http://localhost:8080"
		}
		url = strings.TrimSuffix(url, "/")

		fmt.Println("Open the following URL in your browser to sign in:")
		fmt.Println()
		fmt.Printf("  %s/api/v1/auth/login\n", url)
		fmt.Println()
		fmt.Print("Paste the token returned after login: ")

		var token string
		if _, err := fmt.Fscanln(os.Stdin, &token); err != nil {
			return fmt.Errorf("read token: %w", err)
		}
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("no token provided")
		}

		// Verify the token against the API before persisting.
		req, _ := http.NewRequest(http.MethodGet, url+"/api/v1/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("verify token: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("verification failed (%s): %s", resp.Status, strings.TrimSpace(string(body)))
		}

		if err := saveConfig(&fileConfig{URL: url, Token: token}); err != nil {
			return err
		}
		fmt.Println("authenticated and saved to ~/.custodian/config.json")
		return nil
	},
}
