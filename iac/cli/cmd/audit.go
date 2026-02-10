package cmd

import "fmt"

// Audit handles the "iac audit" command and routes to subcommands.
func Audit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand. Usage: iac audit [security]")
	}

	switch args[0] {
	case "security":
		// TODO: Implement in Phase 6 (T048)
		fmt.Println("Security audit: (not yet implemented — see Phase 6: US3)")
		return nil
	default:
		return fmt.Errorf("unknown audit subcommand: %s", args[0])
	}
}
