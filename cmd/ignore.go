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
	ignoreRepo string
)

var ignoreCmd = &cobra.Command{
	Use:   "ignore <target>",
	Short: "Ignore notifications for an issue or pull request",
	Long: `Ignore all notifications for an issue or pull request.

This suppresses all notifications for the conversation, including mentions
and direct participation. This is stronger than unsubscribing.

Subscription states explained:
  - SUBSCRIBED: receive all conversation notifications
  - UNSUBSCRIBED: notifications may still arrive when participating or mentioned
  - IGNORED: suppress all notifications for the conversation

You can specify the target as:
  - An issue/PR number (requires repository context)
  - A full GitHub URL
  - owner/repo#number format`,
	Example: `  gh subscribe ignore 123
  gh subscribe ignore 123 --repo cli/cli
  gh subscribe ignore https://github.com/cli/cli/pull/456`,
	Args: cobra.ExactArgs(1),
	RunE: runIgnore,
}

func init() {
	ignoreCmd.Flags().StringVarP(&ignoreRepo, "repo", "R", "", "Repository in owner/repo format")
	rootCmd.AddCommand(ignoreCmd)
}

func runIgnore(cmd *cobra.Command, args []string) error {
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
		owner, repo, err := resolveRepo(ignoreRepo)
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

	// Check if already ignored
	if item.ViewerSubscription == github.StateIgnored {
		formatter.PrintSuccess(output.IgnoreMessage(t, true))
		return nil
	}

	// Ignore
	if err := client.UpdateSubscription(ctx, item.ID, github.StateIgnored); err != nil {
		formatter.PrintError(err.Error())
		return fmt.Errorf("failed to ignore")
	}

	formatter.PrintSuccess(output.IgnoreMessage(t, false))
	return nil
}
