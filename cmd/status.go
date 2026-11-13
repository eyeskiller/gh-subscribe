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
	statusRepo string
	statusJSON bool
)

var statusCmd = &cobra.Command{
	Use:   "status <target>",
	Short: "Check subscription status for an issue or pull request",
	Long: `Display the current subscription status for an issue or pull request.

Shows detailed information including the repository, type, number, title,
current subscription state, and URL.

Use --json for machine-readable output.

Subscription states:
  - SUBSCRIBED: receive all conversation notifications
  - UNSUBSCRIBED: notifications may still arrive when participating or mentioned
  - IGNORED: suppress all notifications for the conversation`,
	Example: `  gh subscribe status 123
  gh subscribe status 123 --repo cli/cli
  gh subscribe status https://github.com/cli/cli/pull/456
  gh subscribe status 123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().StringVarP(&statusRepo, "repo", "R", "", "Repository in owner/repo format")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
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
		owner, repo, err := resolveRepo(statusRepo)
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

	// Prepare result
	result := &output.StatusResult{
		Owner:  t.Owner,
		Repo:   t.Repo,
		Type:   item.TypeName,
		Number: item.Number,
		Title:  item.Title,
		URL:    item.URL,
		State:  item.ViewerSubscription,
	}

	// Output
	if statusJSON {
		if err := formatter.PrintStatusJSON(result); err != nil {
			formatter.PrintError(err.Error())
			return fmt.Errorf("failed to format JSON")
		}
	} else {
		formatter.PrintStatus(result)
	}

	return nil
}
