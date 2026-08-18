// Command custodian-cli is the official command line client for Custodian.
package main

import (
	"fmt"
	"os"

	"github.com/CocoCopi/custodian/cmd/custodian-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
