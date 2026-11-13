package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootRepoFlag string

var rootCmd = &cobra.Command{
	Use:   "subscribe <target>",
	Short: "Subscribe to GitHub issues and pull requests",
	Long: `Subscribe to GitHub issues and pull requests directly from your terminal.

Manage notifications for GitHub conversations without leaving your command line.
This extension feels like a native GitHub CLI command and reuses your existing gh authentication.`,
	Example: `  gh subscribe 123
  gh subscribe add 123
  gh subscribe add 123 --repo cli/cli
  gh subscribe add https://github.com/cli/cli/issues/14009
  gh subscribe remove 123
  gh subscribe ignore 123
  gh subscribe status 123`,
	Version: version,
	// Shorthand: gh subscribe 123 → equivalent to gh subscribe add 123
	Args:              cobra.MaximumNArgs(1),
	RunE:              runRoot,
	ValidArgsFunction: cobra.NoFileCompletions,
}

func runRoot(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	// Delegate to add command
	addRepo = rootRepoFlag
	return runAdd(cmd, args)
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&rootRepoFlag, "repo", "R", "", "Repository in owner/repo format")

	// Configure version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("gh-subscribe version %s\n", version))

	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func exitWithError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
