package target

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Target represents a resolved GitHub issue or pull request
type Target struct {
	Owner  string
	Repo   string
	Number int
}

// Parse parses a target from various input formats:
// - "123" (requires repo context)
// - "owner/repo#123"
// - "https://github.com/owner/repo/issues/123"
// - "https://github.com/owner/repo/pull/123"
func Parse(input string, defaultOwner, defaultRepo string) (*Target, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty target")
	}

	// Try URL format first
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return parseURL(input)
	}

	// Try owner/repo#number format
	if strings.Contains(input, "#") {
		return parseOwnerRepoNumber(input)
	}

	// Try plain number format
	if num, err := strconv.Atoi(input); err == nil {
		if defaultOwner == "" || defaultRepo == "" {
			return nil, fmt.Errorf("repository context required for number-only input")
		}
		if num <= 0 {
			return nil, fmt.Errorf("invalid issue/PR number: %d", num)
		}
		return &Target{
			Owner:  defaultOwner,
			Repo:   defaultRepo,
			Number: num,
		}, nil
	}

	return nil, fmt.Errorf("invalid target format: %s", input)
}

func parseURL(input string) (*Target, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Host != "github.com" && u.Host != "www.github.com" {
		return nil, fmt.Errorf("unsupported host: %s (only github.com supported)", u.Host)
	}

	// Parse path: /owner/repo/issues/123 or /owner/repo/pull/123
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid GitHub URL format")
	}

	owner := parts[0]
	repo := parts[1]
	typePath := parts[2]
	numberStr := parts[3]

	if typePath != "issues" && typePath != "pull" {
		return nil, fmt.Errorf("URL must point to an issue or pull request")
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil || number <= 0 {
		return nil, fmt.Errorf("invalid issue/PR number in URL")
	}

	return &Target{
		Owner:  owner,
		Repo:   repo,
		Number: number,
	}, nil
}

func parseOwnerRepoNumber(input string) (*Target, error) {
	// Match pattern: owner/repo#number
	pattern := regexp.MustCompile(`^([a-zA-Z0-9\-_.]+)/([a-zA-Z0-9\-_.]+)#(\d+)$`)
	matches := pattern.FindStringSubmatch(input)

	if matches == nil {
		return nil, fmt.Errorf("invalid format, expected owner/repo#number")
	}

	number, _ := strconv.Atoi(matches[3])
	if number <= 0 {
		return nil, fmt.Errorf("invalid issue/PR number")
	}

	return &Target{
		Owner:  matches[1],
		Repo:   matches[2],
		Number: number,
	}, nil
}

// String returns a human-readable representation
func (t *Target) String() string {
	return fmt.Sprintf("%s/%s#%d", t.Owner, t.Repo, t.Number)
}
