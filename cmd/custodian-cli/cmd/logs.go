package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logsCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs <name-or-id>",
	Short: "Stream live logs for an app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := NewClient()
		if err != nil {
			return err
		}
		u, err := url.Parse(c.URL)
		if err != nil {
			return err
		}
		u.Scheme = wsScheme(u.Scheme)
		u.Path = "/api/v1/services/" + args[0] + "/logs"
		u.RawQuery = "token=" + url.QueryEscape(c.Token)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
		if err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		defer func() { _ = conn.Close() }()
		fmt.Println("streaming logs (Ctrl-C to stop)...")

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return nil // connection closed or interrupted
			}
			var entry struct {
				Stream    string `json:"stream"`
				Message   string `json:"message"`
				Timestamp string `json:"timestamp"`
			}
			if err := json.Unmarshal(data, &entry); err == nil {
				fmt.Printf("[%s] %s", entry.Stream, entry.Message)
			} else {
				fmt.Println(string(data))
			}
		}
	},
}

func wsScheme(httpScheme string) string {
	if httpScheme == "https" {
		return "wss"
	}
	return "ws"
}
