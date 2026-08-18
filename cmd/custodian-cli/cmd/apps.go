package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type servicePayload struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RepoURL   string `json:"repo_url"`
	Branch    string `json:"branch"`
	BuildType string `json:"build_type"`
	Image     string `json:"image"`
	Status    string `json:"status"`
}

func init() {
	appsCmd.AddCommand(appsCreateCmd, appsListCmd, appsGetCmd, appsDeleteCmd)
	rootCmd.AddCommand(appsCmd)
}

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage applications",
}

var appsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		blueprint, _ := cmd.Flags().GetString("blueprint")
		var blueprintBody string
		if blueprint != "" {
			data, err := os.ReadFile(blueprint)
			if err != nil {
				return fmt.Errorf("read blueprint: %w", err)
			}
			blueprintBody = string(data)
		}
		repo, _ := cmd.Flags().GetString("repo")
		buildType, _ := cmd.Flags().GetString("build-type")

		body, _ := json.Marshal(map[string]any{
			"name":       args[0],
			"repo_url":   repo,
			"build_type": buildType,
			"blueprint":  blueprintBody,
		})
		resp, err := c.post("/api/v1/services", body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			return responseError(resp)
		}
		var svc servicePayload
		if err := json.NewDecoder(resp.Body).Decode(&svc); err != nil {
			return err
		}
		fmt.Printf("created %s (id=%s, status=%s)\n", svc.Name, svc.ID, svc.Status)
		return nil
	},
}

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		resp, err := c.get("/api/v1/services")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return responseError(resp)
		}
		var out struct {
			Services []servicePayload `json:"services"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return err
		}
		if len(out.Services) == 0 {
			fmt.Println("no apps yet — run `custodian apps create <name>`")
			return nil
		}
		fmt.Printf("%-24s %-12s %-14s %s\n", "NAME", "STATUS", "BUILD", "ID")
		for _, s := range out.Services {
			fmt.Printf("%-24s %-12s %-14s %s\n", s.Name, s.Status, s.BuildType, s.ID)
		}
		return nil
	},
}

var appsGetCmd = &cobra.Command{
	Use:   "get <name-or-id>",
	Short: "Show an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		resp, err := c.get("/api/v1/services/" + args[0])
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return responseError(resp)
		}
		var svc servicePayload
		if err := json.NewDecoder(resp.Body).Decode(&svc); err != nil {
			return err
		}
		out, _ := json.MarshalIndent(svc, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var appsDeleteCmd = &cobra.Command{
	Use:   "delete <name-or-id>",
	Short: "Delete an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		resp, err := c.delete("/api/v1/services/" + args[0])
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			return responseError(resp)
		}
		fmt.Println("deleted")
		return nil
	},
}

func init() {
	appsCreateCmd.Flags().String("repo", "", "git repository URL")
	appsCreateCmd.Flags().String("build-type", "dockerfile", "dockerfile | buildpacks | static")
	appsCreateCmd.Flags().String("blueprint", "", "path to custodian.yaml blueprint")
}

// --- shared HTTP helpers ---

func (c *Client) get(path string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, c.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	return http.DefaultClient.Do(req)
}

func (c *Client) post(path string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodPost, c.URL+path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func (c *Client) delete(path string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodDelete, c.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	return http.DefaultClient.Do(req)
}

func responseError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)
	var out struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(data, &out)
	if out.Error != "" {
		return fmt.Errorf("server: %s", out.Error)
	}
	return fmt.Errorf("server returned %s", resp.Status)
}
