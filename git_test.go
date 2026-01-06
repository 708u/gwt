package twig

import (
	"testing"

	"github.com/708u/twig/internal/testutil"
)

func TestGitRunner_IsBranchContentIdentical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		branch        string
		target        string
		contentMerged map[string][]string
		want          bool
	}{
		{
			name:   "content merged",
			branch: "feat/squashed",
			target: "main",
			contentMerged: map[string][]string{
				"main": {"feat/squashed"},
			},
			want: true,
		},
		{
			name:   "content not merged",
			branch: "feat/new",
			target: "main",
			contentMerged: map[string][]string{
				"main": {},
			},
			want: false,
		},
		{
			name:          "empty content merged map",
			branch:        "feat/any",
			target:        "main",
			contentMerged: map[string][]string{},
			want:          false,
		},
		{
			name:   "multiple branches content merged",
			branch: "feat/b",
			target: "main",
			contentMerged: map[string][]string{
				"main": {"feat/a", "feat/b", "feat/c"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := &testutil.MockGitExecutor{
				ContentMergedBranches: tt.contentMerged,
			}
			runner := &GitRunner{Executor: mockGit}

			got, err := runner.IsBranchContentIdentical(tt.branch, tt.target)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitRunner_IsBranchMerged_WithSquashMerge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		branch        string
		target        string
		merged        map[string][]string
		contentMerged map[string][]string
		want          bool
	}{
		{
			name:   "traditional merge detected",
			branch: "feat/merged",
			target: "main",
			merged: map[string][]string{
				"main": {"feat/merged"},
			},
			contentMerged: map[string][]string{},
			want:          true,
		},
		{
			name:   "squash merge detected via content check",
			branch: "feat/squashed",
			target: "main",
			merged: map[string][]string{
				"main": {}, // Not in --merged output
			},
			contentMerged: map[string][]string{
				"main": {"feat/squashed"}, // But content is in target
			},
			want: true,
		},
		{
			name:   "not merged at all",
			branch: "feat/new",
			target: "main",
			merged: map[string][]string{
				"main": {},
			},
			contentMerged: map[string][]string{
				"main": {},
			},
			want: false,
		},
		{
			name:   "traditional merge takes precedence",
			branch: "feat/both",
			target: "main",
			merged: map[string][]string{
				"main": {"feat/both"},
			},
			contentMerged: map[string][]string{
				"main": {"feat/both"},
			},
			want: true,
		},
		{
			name:   "different target branch",
			branch: "feat/dev-feature",
			target: "develop",
			merged: map[string][]string{
				"develop": {},
			},
			contentMerged: map[string][]string{
				"develop": {"feat/dev-feature"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := &testutil.MockGitExecutor{
				MergedBranches:        tt.merged,
				ContentMergedBranches: tt.contentMerged,
			}
			runner := &GitRunner{Executor: mockGit}

			got, err := runner.IsBranchMerged(tt.branch, tt.target)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
