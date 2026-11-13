package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/lubosberan/gh-subscribe/internal/github"
	"github.com/lubosberan/gh-subscribe/internal/target"
)

// Formatter handles terminal output formatting
type Formatter struct {
	out      io.Writer
	useColor bool
}

// NewFormatter creates a new output formatter
func NewFormatter(out io.Writer) *Formatter {
	// Detect color support: respect NO_COLOR and check if stdout is a terminal
	noColor := os.Getenv("NO_COLOR") != ""
	isTerminal := isTerminal(out)

	return &Formatter{
		out:      out,
		useColor: !noColor && isTerminal,
	}
}

// StatusResult holds all info needed for status display
type StatusResult struct {
	Owner  string
	Repo   string
	Type   string
	Number int
	Title  string
	URL    string
	State  github.SubscriptionState
}

// PrintSuccess prints a success message
func (f *Formatter) PrintSuccess(msg string) {
	if f.useColor {
		fmt.Fprintf(f.out, "\033[32m✓\033[0m %s\n", msg)
	} else {
		fmt.Fprintf(f.out, "✓ %s\n", msg)
	}
}

// PrintError prints an error message
func (f *Formatter) PrintError(msg string) {
	if f.useColor {
		fmt.Fprintf(f.out, "\033[31mError:\033[0m %s\n", msg)
	} else {
		fmt.Fprintf(f.out, "Error: %s\n", msg)
	}
}

// PrintStatus prints subscription status in human-readable format
func (f *Formatter) PrintStatus(result *StatusResult) {
	fmt.Fprintf(f.out, "Repository: %s/%s\n", result.Owner, result.Repo)
	fmt.Fprintf(f.out, "Type:       %s\n", result.Type)
	fmt.Fprintf(f.out, "Number:     %d\n", result.Number)
	fmt.Fprintf(f.out, "Title:      %s\n", result.Title)

	stateStr := string(result.State)
	if result.State == "" {
		stateStr = "UNKNOWN"
	}
	fmt.Fprintf(f.out, "Status:     %s\n", stateStr)
	fmt.Fprintf(f.out, "URL:        %s\n", result.URL)
}

// PrintStatusJSON prints subscription status in JSON format
func (f *Formatter) PrintStatusJSON(result *StatusResult) error {
	stateStr := string(result.State)
	if result.State == "" {
		stateStr = "UNKNOWN"
	}

	data := map[string]interface{}{
		"repository": fmt.Sprintf("%s/%s", result.Owner, result.Repo),
		"type":       result.Type,
		"number":     result.Number,
		"title":      result.Title,
		"url":        result.URL,
		"status":     stateStr,
	}

	encoder := json.NewEncoder(f.out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// SubscribeMessage returns the appropriate message for a subscribe action
func SubscribeMessage(t *target.Target, alreadySubscribed bool) string {
	if alreadySubscribed {
		return fmt.Sprintf("Already subscribed to %s", t.String())
	}
	return fmt.Sprintf("Subscribed to %s", t.String())
}

// UnsubscribeMessage returns the appropriate message for an unsubscribe action
func UnsubscribeMessage(t *target.Target, alreadyUnsubscribed bool) string {
	if alreadyUnsubscribed {
		return fmt.Sprintf("Already unsubscribed from %s", t.String())
	}
	return fmt.Sprintf("Unsubscribed from %s", t.String())
}

// IgnoreMessage returns the appropriate message for an ignore action
func IgnoreMessage(t *target.Target, alreadyIgnored bool) string {
	if alreadyIgnored {
		return fmt.Sprintf("Already ignoring %s", t.String())
	}
	return fmt.Sprintf("Ignoring %s", t.String())
}

func isTerminal(w io.Writer) bool {
	// Check if the writer is an *os.File and is a terminal
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}
