package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eyeskiller/gh-subscribe/internal/github"
	"github.com/eyeskiller/gh-subscribe/internal/output"
	"github.com/eyeskiller/gh-subscribe/internal/target"
	"github.com/spf13/cobra"
)

var (
	addRepo string
)

var addCmd = &cobra.Command{
	Use:   "add <target>",
	Short: "Subscribe to an issue or pull request",
	Long: `Subscribe to receive notifications for an issue or pull request.

You can specify the target as:
  - An issue/PR number (requires repository context)
  - A full GitHub URL
  - owner/repo#number format

The command checks if you're already subscribed and avoids unnecessary API calls.`,
	Example: `  gh subscribe add 123
  gh subscribe add 123 --repo cli/cli
  gh subscribe add https://github.com/cli/cli/issues/14009
  gh subscribe add https://github.com/cli/cli/pull/456
  gh subscribe add cli/cli#14009`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addRepo, "repo", "R", "", "Repository in owner/repo format")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
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
		owner, repo, err := resolveRepo(addRepo)
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

	// Check if already subscribed
	if item.ViewerSubscription == github.StateSubscribed {
		formatter.PrintSuccess(output.SubscribeMessage(t, true))
		return nil
	}

	// Check if user can subscribe
	if !item.ViewerCanSubscribe {
		formatter.PrintError("You cannot subscribe to this conversation")
		return fmt.Errorf("cannot subscribe")
	}

	// Subscribe
	if err := client.UpdateSubscription(ctx, item.ID, github.StateSubscribed); err != nil {
		formatter.PrintError(err.Error())
		return fmt.Errorf("failed to subscribe")
	}

	formatter.PrintSuccess(output.SubscribeMessage(t, false))
	return nil
}

func resolveRepo(repoFlag string) (owner, repo string, err error) {
	// Priority: explicit flag > GH_REPO env > detect from directory
	if repoFlag != "" {
		parts := splitRepoFlag(repoFlag)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid repository format: %s\nExpected: OWNER/REPO", repoFlag)
		}
		return parts[0], parts[1], nil
	}

	// Check GH_REPO environment variable
	if ghRepo := os.Getenv("GH_REPO"); ghRepo != "" {
		parts := splitRepoFlag(ghRepo)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GH_REPO format: %s", ghRepo)
		}
		return parts[0], parts[1], nil
	}

	// Detect from current directory
	return github.DetectRepository()
}

func splitRepoFlag(s string) []string {
	parts := make([]string, 0, 2)
	idx := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	parts = append(parts, s[:idx], s[idx+1:])
	return parts
}
