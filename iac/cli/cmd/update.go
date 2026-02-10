package cmd

import "fmt"

// Update handles the "iac update" command and routes to subcommands.
func Update(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac update [status|trigger|enable|disable]")
	}

	switch args[0] {
	case "status":
		return updateStatus(args[1:])
	case "trigger":
		return updateTrigger(args[1:])
	case "enable":
		return updateEnable(args[1:])
	case "disable":
		return updateDisable(args[1:])
	default:
		return fmt.Errorf("unknown update subcommand: %s", args[0])
	}
}
