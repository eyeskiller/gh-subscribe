package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/eyeskiller/gh-subscribe/internal/github"
	"github.com/eyeskiller/gh-subscribe/internal/target"
)

func TestPrintSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	f := &Formatter{out: buf, useColor: false}

	f.PrintSuccess("Test message")

	got := buf.String()
	want := "✓ Test message\n"
	if got != want {
		t.Errorf("PrintSuccess() = %q, want %q", got, want)
	}
}

func TestPrintSuccessColor(t *testing.T) {
	buf := &bytes.Buffer{}
	f := &Formatter{out: buf, useColor: true}

	f.PrintSuccess("Test message")

	got := buf.String()
	if !strings.Contains(got, "\033[32m") {
		t.Errorf("PrintSuccess() should contain color codes when useColor=true")
	}
	if !strings.Contains(got, "Test message") {
		t.Errorf("PrintSuccess() should contain the message")
	}
}

func TestPrintError(t *testing.T) {
	buf := &bytes.Buffer{}
	f := &Formatter{out: buf, useColor: false}

	f.PrintError("Test error")

	got := buf.String()
	want := "Error: Test error\n"
	if got != want {
		t.Errorf("PrintError() = %q, want %q", got, want)
	}
}

func TestPrintStatus(t *testing.T) {
	buf := &bytes.Buffer{}
	f := &Formatter{out: buf, useColor: false}

	result := &StatusResult{
		Owner:  "cli",
		Repo:   "cli",
		Type:   "Issue",
		Number: 14009,
		Title:  "Test Issue",
		URL:    "https://github.com/cli/cli/issues/14009",
		State:  github.StateSubscribed,
	}

	f.PrintStatus(result)

	got := buf.String()
	if !strings.Contains(got, "Repository: cli/cli") {
		t.Error("Status should contain repository")
	}
	if !strings.Contains(got, "Type:       Issue") {
		t.Error("Status should contain type")
	}
	if !strings.Contains(got, "Number:     14009") {
		t.Error("Status should contain number")
	}
	if !strings.Contains(got, "Title:      Test Issue") {
		t.Error("Status should contain title")
	}
	if !strings.Contains(got, "Status:     SUBSCRIBED") {
		t.Error("Status should contain subscription state")
	}
	if !strings.Contains(got, "URL:        https://github.com/cli/cli/issues/14009") {
		t.Error("Status should contain URL")
	}
}

func TestPrintStatusUnknownState(t *testing.T) {
	buf := &bytes.Buffer{}
	f := &Formatter{out: buf, useColor: false}

	result := &StatusResult{
		Owner:  "cli",
		Repo:   "cli",
		Type:   "Issue",
		Number: 123,
		Title:  "Test",
		URL:    "https://github.com/cli/cli/issues/123",
		State:  "", // Unknown state
	}

	f.PrintStatus(result)

	got := buf.String()
	if !strings.Contains(got, "Status:     UNKNOWN") {
		t.Errorf("Unknown state should be displayed as UNKNOWN, got: %s", got)
	}
}

func TestPrintStatusJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	f := &Formatter{out: buf, useColor: false}

	result := &StatusResult{
		Owner:  "cli",
		Repo:   "cli",
		Type:   "Issue",
		Number: 14009,
		Title:  "Test Issue",
		URL:    "https://github.com/cli/cli/issues/14009",
		State:  github.StateSubscribed,
	}

	err := f.PrintStatusJSON(result)
	if err != nil {
		t.Fatalf("PrintStatusJSON() error = %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if data["repository"] != "cli/cli" {
		t.Errorf("repository = %v, want cli/cli", data["repository"])
	}
	if data["type"] != "Issue" {
		t.Errorf("type = %v, want Issue", data["type"])
	}
	if data["number"] != float64(14009) {
		t.Errorf("number = %v, want 14009", data["number"])
	}
	if data["title"] != "Test Issue" {
		t.Errorf("title = %v, want Test Issue", data["title"])
	}
	if data["url"] != "https://github.com/cli/cli/issues/14009" {
		t.Errorf("url = %v", data["url"])
	}
	if data["status"] != "SUBSCRIBED" {
		t.Errorf("status = %v, want SUBSCRIBED", data["status"])
	}
}

func TestPrintStatusJSONUnknownState(t *testing.T) {
	buf := &bytes.Buffer{}
	f := &Formatter{out: buf, useColor: false}

	result := &StatusResult{
		Owner:  "cli",
		Repo:   "cli",
		Type:   "Issue",
		Number: 123,
		Title:  "Test",
		URL:    "https://github.com/cli/cli/issues/123",
		State:  "", // Unknown state
	}

	err := f.PrintStatusJSON(result)
	if err != nil {
		t.Fatalf("PrintStatusJSON() error = %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if data["status"] != "UNKNOWN" {
		t.Errorf("Unknown state should be displayed as UNKNOWN in JSON, got: %v", data["status"])
	}
}

func TestSubscribeMessage(t *testing.T) {
	tgt := &target.Target{Owner: "cli", Repo: "cli", Number: 123}

	got := SubscribeMessage(tgt, false)
	want := "Subscribed to cli/cli#123"
	if got != want {
		t.Errorf("SubscribeMessage() = %q, want %q", got, want)
	}

	got = SubscribeMessage(tgt, true)
	want = "Already subscribed to cli/cli#123"
	if got != want {
		t.Errorf("SubscribeMessage() = %q, want %q", got, want)
	}
}

func TestUnsubscribeMessage(t *testing.T) {
	tgt := &target.Target{Owner: "cli", Repo: "cli", Number: 123}

	got := UnsubscribeMessage(tgt, false)
	want := "Unsubscribed from cli/cli#123"
	if got != want {
		t.Errorf("UnsubscribeMessage() = %q, want %q", got, want)
	}

	got = UnsubscribeMessage(tgt, true)
	want = "Already unsubscribed from cli/cli#123"
	if got != want {
		t.Errorf("UnsubscribeMessage() = %q, want %q", got, want)
	}
}

func TestIgnoreMessage(t *testing.T) {
	tgt := &target.Target{Owner: "cli", Repo: "cli", Number: 123}

	got := IgnoreMessage(tgt, false)
	want := "Ignoring cli/cli#123"
	if got != want {
		t.Errorf("IgnoreMessage() = %q, want %q", got, want)
	}

	got = IgnoreMessage(tgt, true)
	want = "Already ignoring cli/cli#123"
	if got != want {
		t.Errorf("IgnoreMessage() = %q, want %q", got, want)
	}
}
