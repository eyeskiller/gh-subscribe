package target

import (
	"testing"
)

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		owner      string
		repo       string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{
			name:       "valid number with context",
			input:      "123",
			owner:      "cli",
			repo:       "cli",
			wantOwner:  "cli",
			wantRepo:   "cli",
			wantNumber: 123,
		},
		{
			name:    "number without context",
			input:   "123",
			owner:   "",
			repo:    "",
			wantErr: true,
		},
		{
			name:    "invalid number zero",
			input:   "0",
			owner:   "cli",
			repo:    "cli",
			wantErr: true,
		},
		{
			name:    "invalid number negative",
			input:   "-1",
			owner:   "cli",
			repo:    "cli",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := Parse(tt.input, tt.owner, tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if target.Owner != tt.wantOwner {
					t.Errorf("Owner = %v, want %v", target.Owner, tt.wantOwner)
				}
				if target.Repo != tt.wantRepo {
					t.Errorf("Repo = %v, want %v", target.Repo, tt.wantRepo)
				}
				if target.Number != tt.wantNumber {
					t.Errorf("Number = %v, want %v", target.Number, tt.wantNumber)
				}
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{
			name:       "valid issue URL",
			input:      "https://github.com/cli/cli/issues/14009",
			wantOwner:  "cli",
			wantRepo:   "cli",
			wantNumber: 14009,
		},
		{
			name:       "valid pull request URL",
			input:      "https://github.com/cli/cli/pull/456",
			wantOwner:  "cli",
			wantRepo:   "cli",
			wantNumber: 456,
		},
		{
			name:       "URL with www prefix",
			input:      "https://www.github.com/owner/repo/issues/123",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantNumber: 123,
		},
		{
			name:    "unsupported host",
			input:   "https://gitlab.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "invalid path format",
			input:   "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "invalid type in path",
			input:   "https://github.com/owner/repo/discussions/123",
			wantErr: true,
		},
		{
			name:    "invalid number in URL",
			input:   "https://github.com/owner/repo/issues/abc",
			wantErr: true,
		},
		{
			name:    "zero number in URL",
			input:   "https://github.com/owner/repo/issues/0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := Parse(tt.input, "", "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if target.Owner != tt.wantOwner {
					t.Errorf("Owner = %v, want %v", target.Owner, tt.wantOwner)
				}
				if target.Repo != tt.wantRepo {
					t.Errorf("Repo = %v, want %v", target.Repo, tt.wantRepo)
				}
				if target.Number != tt.wantNumber {
					t.Errorf("Number = %v, want %v", target.Number, tt.wantNumber)
				}
			}
		})
	}
}

func TestParseOwnerRepoNumber(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{
			name:       "valid format",
			input:      "cli/cli#14009",
			wantOwner:  "cli",
			wantRepo:   "cli",
			wantNumber: 14009,
		},
		{
			name:       "valid format with hyphens",
			input:      "my-org/my-repo#123",
			wantOwner:  "my-org",
			wantRepo:   "my-repo",
			wantNumber: 123,
		},
		{
			name:       "valid format with underscores",
			input:      "my_org/my_repo#456",
			wantOwner:  "my_org",
			wantRepo:   "my_repo",
			wantNumber: 456,
		},
		{
			name:    "missing number",
			input:   "cli/cli#",
			wantErr: true,
		},
		{
			name:    "missing repo",
			input:   "cli#123",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "zero number",
			input:   "cli/cli#0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := Parse(tt.input, "", "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if target.Owner != tt.wantOwner {
					t.Errorf("Owner = %v, want %v", target.Owner, tt.wantOwner)
				}
				if target.Repo != tt.wantRepo {
					t.Errorf("Repo = %v, want %v", target.Repo, tt.wantRepo)
				}
				if target.Number != tt.wantNumber {
					t.Errorf("Number = %v, want %v", target.Number, tt.wantNumber)
				}
			}
		})
	}
}

func TestContainsRepo(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "https://github.com/cli/cli/issues/14009", want: true},
		{input: "https://github.com/cli/cli/pull/456", want: true},
		{input: "http://github.com/owner/repo/issues/123", want: true},
		{input: "cli/cli#14009", want: true},
		{input: "my-org/my-repo#123", want: true},
		{input: "123", want: false},
		{input: "456", want: false},
		{input: "", want: false},
		{input: "  https://github.com/cli/cli/issues/1  ", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ContainsRepo(tt.input)
			if got != tt.want {
				t.Errorf("ContainsRepo(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTargetString(t *testing.T) {
	target := &Target{
		Owner:  "cli",
		Repo:   "cli",
		Number: 14009,
	}

	want := "cli/cli#14009"
	if got := target.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("", "owner", "repo")
	if err == nil {
		t.Error("Parse() should fail on empty input")
	}
}
