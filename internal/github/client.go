package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Client wraps GitHub GraphQL API operations
type Client struct {
	token  string
	host   string
	client *http.Client
}

// SubscriptionState represents the subscription state
type SubscriptionState string

const (
	StateSubscribed   SubscriptionState = "SUBSCRIBED"
	StateUnsubscribed SubscriptionState = "UNSUBSCRIBED"
	StateIgnored      SubscriptionState = "IGNORED"
)

// IssueOrPullRequest represents the common fields
type IssueOrPullRequest struct {
	ID                 string
	Number             int
	Title              string
	URL                string
	TypeName           string
	ViewerSubscription SubscriptionState
	ViewerCanSubscribe bool
}

// GraphQLError represents a GraphQL error response
type GraphQLError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// GraphQLResponse wraps a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

// NewClient creates a new GitHub API client using gh authentication
func NewClient() (*Client, error) {
	// Get token from gh CLI
	token, host, err := getAuthToken()
	if err != nil {
		return nil, err
	}

	return &Client{
		token:  token,
		host:   host,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func getAuthToken() (string, string, error) {
	// Use gh auth token to get the current token
	cmd := exec.Command("gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("GitHub authentication required.\nRun: gh auth login")
	}

	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", "", fmt.Errorf("GitHub authentication required.\nRun: gh auth login")
	}

	return token, "github.com", nil
}

// graphqlRequest performs a GraphQL request
func (c *Client) graphqlRequest(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	body := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://api.%s/graphql", c.host)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitHub authentication required.\nRun: gh auth login")
	}

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("rate limit exceeded or insufficient permissions")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("GraphQL error: %s", strings.Join(msgs, "; "))
	}

	if result != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("failed to parse data: %w", err)
		}
	}

	return nil
}

// GetIssueOrPullRequest fetches an issue or pull request by number
func (c *Client) GetIssueOrPullRequest(ctx context.Context, owner, repo string, number int) (*IssueOrPullRequest, error) {
	query := `
	query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			issueOrPullRequest(number: $number) {
				__typename
				... on Issue {
					id
					number
					title
					url
					viewerSubscription
					viewerCanSubscribe
				}
				... on PullRequest {
					id
					number
					title
					url
					viewerSubscription
					viewerCanSubscribe
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	var data struct {
		Repository struct {
			IssueOrPullRequest *struct {
				ID                 string `json:"id"`
				Number             int    `json:"number"`
				Title              string `json:"title"`
				URL                string `json:"url"`
				TypeName           string `json:"__typename"`
				ViewerSubscription string `json:"viewerSubscription"`
				ViewerCanSubscribe bool   `json:"viewerCanSubscribe"`
			} `json:"issueOrPullRequest"`
		} `json:"repository"`
	}

	if err := c.graphqlRequest(ctx, query, variables, &data); err != nil {
		return nil, err
	}

	if data.Repository.IssueOrPullRequest == nil {
		return nil, fmt.Errorf("issue or pull request #%d not found in %s/%s", number, owner, repo)
	}

	item := data.Repository.IssueOrPullRequest
	return &IssueOrPullRequest{
		ID:                 item.ID,
		Number:             item.Number,
		Title:              item.Title,
		URL:                item.URL,
		TypeName:           item.TypeName,
		ViewerSubscription: SubscriptionState(item.ViewerSubscription),
		ViewerCanSubscribe: item.ViewerCanSubscribe,
	}, nil
}

// UpdateSubscription updates the subscription state
func (c *Client) UpdateSubscription(ctx context.Context, id string, state SubscriptionState) error {
	mutation := `
	mutation($id: ID!, $state: SubscriptionState!) {
		updateSubscription(input: {
			subscribableId: $id
			state: $state
		}) {
			subscribable {
				id
				viewerSubscription
			}
		}
	}`

	variables := map[string]interface{}{
		"id":    id,
		"state": string(state),
	}

	var data struct {
		UpdateSubscription struct {
			Subscribable struct {
				ID                 string `json:"id"`
				ViewerSubscription string `json:"viewerSubscription"`
			} `json:"subscribable"`
		} `json:"updateSubscription"`
	}

	if err := c.graphqlRequest(ctx, mutation, variables, &data); err != nil {
		return err
	}

	// Verify the state was applied
	actual := SubscriptionState(data.UpdateSubscription.Subscribable.ViewerSubscription)
	if actual != state {
		return fmt.Errorf("subscription state mismatch: expected %s, got %s", state, actual)
	}

	return nil
}

// DetectRepository detects the repository from the current directory using gh CLI
func DetectRepository() (owner, repo string, err error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "owner,name", "-q", ".owner.login + \"/\" + .name")
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("cannot determine repository.\nUse --repo OWNER/REPO or run from within a repository")
	}

	result := strings.TrimSpace(string(out))
	parts := strings.SplitN(result, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format detected")
	}

	return parts[0], parts[1], nil
}
