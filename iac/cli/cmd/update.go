package cmd

import "fmt"

// Update handles the "iac update" command and routes to subcommands.
func Update(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac update [status|trigger|enable|disable]")
	}

	switch args[0] {
	case "status":
		// TODO: Implement in Phase 7 (T052)
		fmt.Println("Auto-update status: (not yet implemented — see Phase 7: US4)")
		return nil
	case "trigger":
		// TODO: Implement in Phase 7 (T053)
		fmt.Println("Manual update trigger: (not yet implemented — see Phase 7: US4)")
		return nil
	case "enable":
		// TODO: Implement in Phase 7 (T054)
		fmt.Println("Enable auto-updates: (not yet implemented — see Phase 7: US4)")
		return nil
	case "disable":
		// TODO: Implement in Phase 7 (T054)
		fmt.Println("Disable auto-updates: (not yet implemented — see Phase 7: US4)")
		return nil
	default:
		return fmt.Errorf("unknown update subcommand: %s", args[0])
	}
}
