package cmd

import "fmt"

// Update handles the "iac update" command and routes to subcommands.
func Update(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac update status")
	}

	switch args[0] {
	case "status":
		return updateStatus(args[1:])
	default:
		return fmt.Errorf("unknown update subcommand: %s. Available: status", args[0])
	}
}
