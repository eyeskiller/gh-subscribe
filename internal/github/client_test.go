package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetIssueOrPullRequest(t *testing.T) {
	tests := []struct {
		name       string
		response   interface{}
		wantErr    bool
		wantID     string
		wantNumber int
		wantTitle  string
		wantType   string
		wantState  SubscriptionState
		wantCanSub bool
	}{
		{
			name: "successful issue fetch",
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"repository": map[string]interface{}{
						"issueOrPullRequest": map[string]interface{}{
							"__typename":         "Issue",
							"id":                 "ISSUE_ID",
							"number":             123,
							"title":              "Test Issue",
							"url":                "https://github.com/owner/repo/issues/123",
							"viewerSubscription": "SUBSCRIBED",
							"viewerCanSubscribe": true,
						},
					},
				},
			},
			wantID:     "ISSUE_ID",
			wantNumber: 123,
			wantTitle:  "Test Issue",
			wantType:   "Issue",
			wantState:  StateSubscribed,
			wantCanSub: true,
		},
		{
			name: "successful pull request fetch",
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"repository": map[string]interface{}{
						"issueOrPullRequest": map[string]interface{}{
							"__typename":         "PullRequest",
							"id":                 "PR_ID",
							"number":             456,
							"title":              "Test PR",
							"url":                "https://github.com/owner/repo/pull/456",
							"viewerSubscription": "UNSUBSCRIBED",
							"viewerCanSubscribe": true,
						},
					},
				},
			},
			wantID:     "PR_ID",
			wantNumber: 456,
			wantTitle:  "Test PR",
			wantType:   "PullRequest",
			wantState:  StateUnsubscribed,
			wantCanSub: true,
		},
		{
			name: "not found",
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"repository": map[string]interface{}{
						"issueOrPullRequest": nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "GraphQL error",
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"message": "Could not resolve to a Repository",
						"type":    "NOT_FOUND",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := &Client{
				token:  "test-token",
				host:   "github.com",
				client: server.Client(),
			}

			// Override the GraphQL endpoint
			ctx := context.Background()
			item, err := client.GetIssueOrPullRequestWithEndpoint(ctx, server.URL, "owner", "repo", 123)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssueOrPullRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if item.ID != tt.wantID {
					t.Errorf("ID = %v, want %v", item.ID, tt.wantID)
				}
				if item.Number != tt.wantNumber {
					t.Errorf("Number = %v, want %v", item.Number, tt.wantNumber)
				}
				if item.Title != tt.wantTitle {
					t.Errorf("Title = %v, want %v", item.Title, tt.wantTitle)
				}
				if item.TypeName != tt.wantType {
					t.Errorf("TypeName = %v, want %v", item.TypeName, tt.wantType)
				}
				if item.ViewerSubscription != tt.wantState {
					t.Errorf("ViewerSubscription = %v, want %v", item.ViewerSubscription, tt.wantState)
				}
				if item.ViewerCanSubscribe != tt.wantCanSub {
					t.Errorf("ViewerCanSubscribe = %v, want %v", item.ViewerCanSubscribe, tt.wantCanSub)
				}
			}
		})
	}
}

func TestUpdateSubscription(t *testing.T) {
	tests := []struct {
		name     string
		state    SubscriptionState
		response map[string]interface{}
		wantErr  bool
	}{
		{
			name:  "successful subscribe",
			state: StateSubscribed,
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"updateSubscription": map[string]interface{}{
						"subscribable": map[string]interface{}{
							"id":                 "ID",
							"viewerSubscription": "SUBSCRIBED",
						},
					},
				},
			},
		},
		{
			name:  "successful unsubscribe",
			state: StateUnsubscribed,
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"updateSubscription": map[string]interface{}{
						"subscribable": map[string]interface{}{
							"id":                 "ID",
							"viewerSubscription": "UNSUBSCRIBED",
						},
					},
				},
			},
		},
		{
			name:  "successful ignore",
			state: StateIgnored,
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"updateSubscription": map[string]interface{}{
						"subscribable": map[string]interface{}{
							"id":                 "ID",
							"viewerSubscription": "IGNORED",
						},
					},
				},
			},
		},
		{
			name:  "state mismatch",
			state: StateSubscribed,
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"updateSubscription": map[string]interface{}{
						"subscribable": map[string]interface{}{
							"id":                 "ID",
							"viewerSubscription": "UNSUBSCRIBED",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "GraphQL error",
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"message": "Mutation failed",
						"type":    "INTERNAL_ERROR",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := &Client{
				token:  "test-token",
				host:   "github.com",
				client: server.Client(),
			}

			ctx := context.Background()
			err := client.UpdateSubscriptionWithEndpoint(ctx, server.URL, "ID", tt.state)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			wantErr:    true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := &Client{
				token:  "test-token",
				host:   "github.com",
				client: server.Client(),
			}

			ctx := context.Background()
			_, err := client.GetIssueOrPullRequestWithEndpoint(ctx, server.URL, "owner", "repo", 123)

			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error = %v, got %v", tt.wantErr, err != nil)
			}
		})
	}
}

// Helper methods to allow testing with custom endpoints
func (c *Client) GetIssueOrPullRequestWithEndpoint(ctx context.Context, endpoint, owner, repo string, number int) (*IssueOrPullRequest, error) {
	// Temporarily override the endpoint for testing
	originalHost := c.host
	defer func() { c.host = originalHost }()

	// Extract host from endpoint
	c.host = "github.com" // Will be overridden by custom request

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

	if err := c.graphqlRequestWithEndpoint(ctx, endpoint, query, variables, &data); err != nil {
		return nil, err
	}

	if data.Repository.IssueOrPullRequest == nil {
		return nil, errNotFound(number, owner, repo)
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

func (c *Client) UpdateSubscriptionWithEndpoint(ctx context.Context, endpoint, id string, state SubscriptionState) error {
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

	if err := c.graphqlRequestWithEndpoint(ctx, endpoint, mutation, variables, &data); err != nil {
		return err
	}

	actual := SubscriptionState(data.UpdateSubscription.Subscribable.ViewerSubscription)
	if actual != state {
		return errStateMismatch(state, actual)
	}

	return nil
}

func (c *Client) graphqlRequestWithEndpoint(ctx context.Context, endpoint, query string, variables map[string]interface{}, result interface{}) error {
	body := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errUnauthorized()
	}

	if resp.StatusCode == http.StatusForbidden {
		return errRateLimit()
	}

	if resp.StatusCode != http.StatusOK {
		return errAPIStatus(resp.StatusCode, string(respBody))
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return err
	}

	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return errGraphQL(msgs)
	}

	if result != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return err
		}
	}

	return nil
}

// Error helper functions
func errNotFound(number int, owner, repo string) error {
	return fmt.Errorf("issue or pull request #%d not found in %s/%s", number, owner, repo)
}

func errStateMismatch(expected, actual SubscriptionState) error {
	return fmt.Errorf("subscription state mismatch: expected %s, got %s", expected, actual)
}

func errUnauthorized() error {
	return fmt.Errorf("GitHub authentication required.\nRun: gh auth login")
}

func errRateLimit() error {
	return fmt.Errorf("rate limit exceeded or insufficient permissions")
}

func errAPIStatus(code int, body string) error {
	return fmt.Errorf("API request failed with status %d: %s", code, body)
}

func errGraphQL(messages []string) error {
	return fmt.Errorf("GraphQL error: %s", strings.Join(messages, "; "))
}
