package cli

import (
	"github.com/alecthomas/kong"
)

// configGroupTitles contains the titles of flag groups that should only appear in root help, not in subcommand help.
var configGroupTitles = map[string]bool{
	"Logging configuration":   true,
	"Transport configuration": true,
}

// FilteredHelpPrinter is a custom Kong help printer that hides configuration flags from subcommand help
// while keeping them visible in root-level help. This prevents duplication of global config flags
// in every subcommand's help output.
func FilteredHelpPrinter(options kong.HelpOptions, context *kong.Context) error {
	// Temporarily hide config flags by setting their Hidden field to true.
	// This removes them from command signatures while keeping them visible in the Flags section.
	// We'll restore the original values after generating help.
	flags := context.Flags()
	originalHiddenStates := make(map[*kong.Flag]bool, len(flags))

	for _, flag := range flags {
		if flag.Group != nil && configGroupTitles[flag.Group.Title] {
			originalHiddenStates[flag] = flag.Hidden
			// For subcommands, hide the flags completely.
			// For root help, we still hide them to prevent them from appearing in command signatures,
			// but they'll still show in the Flags section due to how Kong generates help.
			flag.Hidden = true
		}
	}

	// Ensure we restore the original hidden states even if DefaultHelpPrinter returns an error.
	defer func() {
		for flag, originalState := range originalHiddenStates {
			flag.Hidden = originalState
		}
	}()

	// Use Kong's default help printer with the modified flag states.
	return kong.DefaultHelpPrinter(options, context)
}
