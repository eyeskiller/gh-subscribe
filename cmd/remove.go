package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lubosberan/gh-subscribe/internal/github"
	"github.com/lubosberan/gh-subscribe/internal/output"
	"github.com/lubosberan/gh-subscribe/internal/target"
	"github.com/spf13/cobra"
)

var (
	removeRepo string
)

var removeCmd = &cobra.Command{
	Use:   "remove <target>",
	Short: "Unsubscribe from an issue or pull request",
	Long: `Unsubscribe from an issue or pull request.

Note: Even after unsubscribing, you may still receive notifications when 
participating in the conversation or when mentioned. Use 'ignore' to 
suppress all notifications.

You can specify the target as:
  - An issue/PR number (requires repository context)
  - A full GitHub URL
  - owner/repo#number format`,
	Example: `  gh subscribe remove 123
  gh subscribe remove 123 --repo cli/cli
  gh subscribe remove https://github.com/cli/cli/pull/456`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().StringVarP(&removeRepo, "repo", "R", "", "Repository in owner/repo format")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	formatter := output.NewFormatter(os.Stdout)

	// Parse target - if it contains repo info (URL or owner/repo#number), no detection needed
	var t *target.Target
	var err error

	if target.ContainsRepo(args[0]) {
		// Parse directly - no repo detection needed
		t, err = target.Parse(args[0], "", "")
	} else {
		// Need repo context - detect first
		owner, repo, err := resolveRepo(removeRepo)
		if err != nil {
			formatter.PrintError(err.Error())
			return fmt.Errorf("invalid target")
		}
		t, err = target.Parse(args[0], owner, repo)
	}

	if err != nil {
		formatter.PrintError(err.Error())
		return fmt.Errorf("invalid target")
	}

	// Create GitHub client
	client, err := github.NewClient()
	if err != nil {
		formatter.PrintError(err.Error())
		return fmt.Errorf("authentication failed")
	}

	// Get current subscription state
	item, err := client.GetIssueOrPullRequest(ctx, t.Owner, t.Repo, t.Number)
	if err != nil {
		formatter.PrintError(err.Error())
		return fmt.Errorf("failed to fetch item")
	}

	// Check if already unsubscribed
	if item.ViewerSubscription == github.StateUnsubscribed {
		formatter.PrintSuccess(output.UnsubscribeMessage(t, true))
		return nil
	}

	// Unsubscribe
	if err := client.UpdateSubscription(ctx, item.ID, github.StateUnsubscribed); err != nil {
		formatter.PrintError(err.Error())
		return fmt.Errorf("failed to unsubscribe")
	}

	formatter.PrintSuccess(output.UnsubscribeMessage(t, false))
	return nil
}
