package gwt

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/708u/gwt/internal/testutil"
)

func TestAddCommand_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		branch      string
		config      *Config
		sync        bool
		setupFS     func(t *testing.T) *testutil.MockFS
		setupGit    func(t *testing.T, captured *[]string) *testutil.MockGitExecutor
		wantErr     bool
		errContains string
		wantBFlag   bool
		checkPath   string
		wantSynced  bool
	}{
		{
			name:   "new_branch",
			branch: "feature/test",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree", Symlinks: []string{".envrc"}},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{CapturedArgs: captured}
			},
			wantErr:   false,
			wantBFlag: true,
		},
		{
			name:   "existing_branch",
			branch: "existing",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree", Symlinks: []string{".envrc"}},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					ExistingBranches: []string{"existing"},
					Worktrees:        []testutil.MockWorktree{{Path: "/repo/main", Branch: "main"}},
					CapturedArgs:     captured,
				}
			},
			wantErr:   false,
			wantBFlag: false,
		},
		{
			name:   "directory_exists",
			branch: "feature/test",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{
					ExistingPaths: []string{"/repo/main-worktree/feature/test"},
				}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{}
			},
			wantErr:     true,
			errContains: "directory already exists",
		},
		{
			name:   "empty_name",
			branch: "",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{}
			},
			wantErr:     true,
			errContains: "branch name is required",
		},
		{
			name:   "branch_checked_out",
			branch: "already-used",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					Worktrees: []testutil.MockWorktree{{Path: "/repo/already-used", Branch: "already-used"}},
				}
			},
			wantErr:     true,
			errContains: "already checked out",
		},
		{
			name:   "worktree_add_error",
			branch: "feature/test",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					WorktreeAddErr: errors.New("worktree add failed"),
				}
			},
			wantErr:     true,
			errContains: "failed to create worktree",
		},
		{
			name:      "slash_in_branch_name",
			branch:    "feature/foo",
			config:    &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/worktrees"},
			checkPath: "/worktrees/feature/foo",
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{CapturedArgs: captured}
			},
			wantErr:   false,
			wantBFlag: true,
		},
		{
			name:   "sync_with_changes",
			branch: "feature/sync",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree", Symlinks: []string{".envrc"}},
			sync:   true,
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					CapturedArgs: captured,
					HasChanges:   true,
				}
			},
			wantErr:    false,
			wantBFlag:  true,
			wantSynced: true,
		},
		{
			name:   "sync_no_changes",
			branch: "feature/sync-no-changes",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree", Symlinks: []string{".envrc"}},
			sync:   true,
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					CapturedArgs: captured,
					HasChanges:   false,
				}
			},
			wantErr:    false,
			wantBFlag:  true,
			wantSynced: false,
		},
		{
			name:   "sync_stash_push_error",
			branch: "feature/sync-push-err",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree"},
			sync:   true,
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					HasChanges:   true,
					StashPushErr: errors.New("stash push failed"),
				}
			},
			wantErr:     true,
			errContains: "failed to stash changes",
		},
		{
			name:   "sync_stash_apply_error",
			branch: "feature/sync-apply-err",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree"},
			sync:   true,
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					HasChanges:    true,
					StashApplyErr: errors.New("stash apply failed"),
				}
			},
			wantErr:     true,
			errContains: "failed to apply changes",
		},
		{
			name:   "sync_disabled_with_changes",
			branch: "feature/no-sync",
			config: &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/main-worktree", Symlinks: []string{".envrc"}},
			sync:   false,
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T, captured *[]string) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{
					CapturedArgs: captured,
					HasChanges:   true,
				}
			},
			wantErr:    false,
			wantBFlag:  true,
			wantSynced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var captured []string

			mockFS := tt.setupFS(t)
			mockGit := tt.setupGit(t, &captured)

			cmd := &AddCommand{
				FS:     mockFS,
				Git:    &GitRunner{Executor: mockGit},
				Config: tt.config,
				Sync:   tt.sync,
			}

			batchResult, err := cmd.Run([]string{tt.branch})

			if tt.wantErr {
				// Check batch-level error or individual item error
				if err != nil {
					if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
					}
					return
				}
				// Check individual item error
				if len(batchResult.Added) == 0 {
					t.Fatal("expected error, got empty result")
				}
				if batchResult.Added[0].Err == nil {
					t.Fatal("expected error in result, got nil")
				}
				if tt.errContains != "" && !strings.Contains(batchResult.Added[0].Err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", batchResult.Added[0].Err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if batchResult.HasErrors() {
				t.Fatalf("unexpected errors in batch result: %v", batchResult.Added[0].Err)
			}

			if tt.wantBFlag && !slices.Contains(captured, "-b") {
				t.Errorf("expected -b flag, got: %v", captured)
			}

			if !tt.wantBFlag && len(captured) > 0 && slices.Contains(captured, "-b") {
				t.Errorf("unexpected -b flag, got: %v", captured)
			}

			if tt.checkPath != "" && !slices.Contains(captured, tt.checkPath) {
				t.Errorf("expected path %q in args, got: %v", tt.checkPath, captured)
			}

			result := batchResult.Added[0].AddResult
			if result.ChangesSynced != tt.wantSynced {
				t.Errorf("ChangesSynced = %v, want %v", result.ChangesSynced, tt.wantSynced)
			}
		})
	}
}

func TestAddResult_Format(t *testing.T) {
	t.Parallel()

	result := AddResult{
		Branch:       "feature/test",
		WorktreePath: "/worktrees/feature/test",
		Symlinks: []SymlinkResult{
			{Src: "/repo/.envrc", Dst: "/worktrees/feature/test/.envrc"},
		},
		ChangesSynced: false,
	}

	tests := []struct {
		name       string
		opts       AddFormatOptions
		wantStdout string
		wantStderr string
	}{
		{
			name:       "default_output",
			opts:       AddFormatOptions{},
			wantStdout: "gwt add: feature/test (1 symlinks)\n",
			wantStderr: "",
		},
		{
			name:       "print_path",
			opts:       AddFormatOptions{Print: []string{"path"}},
			wantStdout: "/worktrees/feature/test\n",
			wantStderr: "",
		},
		{
			name:       "print_ignores_verbose",
			opts:       AddFormatOptions{Verbose: true, Print: []string{"path"}},
			wantStdout: "/worktrees/feature/test\n",
			wantStderr: "",
		},
		{
			name:       "verbose_output",
			opts:       AddFormatOptions{Verbose: true},
			wantStdout: "Created worktree at /worktrees/feature/test\nCreated symlink: /worktrees/feature/test/.envrc -> /repo/.envrc\ngwt add: feature/test (1 symlinks)\n",
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := result.Format(tt.opts)

			if got.Stdout != tt.wantStdout {
				t.Errorf("Stdout = %q, want %q", got.Stdout, tt.wantStdout)
			}
			if got.Stderr != tt.wantStderr {
				t.Errorf("Stderr = %q, want %q", got.Stderr, tt.wantStderr)
			}
		})
	}
}

func TestValidatePrintFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fields  []string
		wantErr bool
	}{
		{
			name:    "valid_path",
			fields:  []string{"path"},
			wantErr: false,
		},
		{
			name:    "empty",
			fields:  []string{},
			wantErr: false,
		},
		{
			name:    "invalid_field",
			fields:  []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "valid_and_invalid",
			fields:  []string{"path", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePrintFields(tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrintFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddCommand_createSymlinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		targets        []string
		setupFS        func(t *testing.T) *testutil.MockFS
		wantErr        bool
		errContains    string
		wantSkipped    int
		wantCreated    int
		wantReasonLike string
	}{
		{
			name:    "success",
			targets: []string{".envrc", ".tool-versions"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{
					GlobResults: map[string][]string{
						".envrc":         {".envrc"},
						".tool-versions": {".tool-versions"},
					},
				}
			},
			wantErr:     false,
			wantCreated: 2,
		},
		{
			name:    "source_not_exist",
			targets: []string{".envrc"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{
					GlobResults: map[string][]string{},
				}
			},
			wantErr:        false,
			wantSkipped:    1,
			wantReasonLike: "does not match any files",
		},
		{
			name:    "symlink_error",
			targets: []string{".envrc"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{
					GlobResults: map[string][]string{
						".envrc": {".envrc"},
					},
					SymlinkErr: errors.New("symlink failed"),
				}
			},
			wantErr:     true,
			errContains: "failed to create symlink",
		},
		{
			name:    "destination_already_exists",
			targets: []string{".claude"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{
					GlobResults: map[string][]string{
						".claude": {".claude"},
					},
					ExistingPaths: []string{"/dst/.claude"},
				}
			},
			wantErr:        false,
			wantSkipped:    1,
			wantReasonLike: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockFS := tt.setupFS(t)

			cmd := &AddCommand{
				FS: mockFS,
			}

			results, err := cmd.createSymlinks("/src", "/dst", tt.targets)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var skipped, created int
			for _, r := range results {
				if r.Skipped {
					skipped++
					if tt.wantReasonLike != "" && !strings.Contains(r.Reason, tt.wantReasonLike) {
						t.Errorf("reason %q should contain %q", r.Reason, tt.wantReasonLike)
					}
				} else {
					created++
				}
			}

			if skipped != tt.wantSkipped {
				t.Errorf("got %d skipped, want %d", skipped, tt.wantSkipped)
			}
			if created != tt.wantCreated {
				t.Errorf("got %d created, want %d", created, tt.wantCreated)
			}
		})
	}
}

func TestAddBatchResult_HasErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result AddBatchResult
		want   bool
	}{
		{
			name:   "empty",
			result: AddBatchResult{},
			want:   false,
		},
		{
			name: "success_only",
			result: AddBatchResult{
				Added: []AddedWorktree{{AddResult: AddResult{Branch: "a"}}},
			},
			want: false,
		},
		{
			name: "error_only",
			result: AddBatchResult{
				Added: []AddedWorktree{{AddResult: AddResult{Branch: "b"}, Err: errors.New("fail")}},
			},
			want: true,
		},
		{
			name: "mixed",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "a"}},
					{AddResult: AddResult{Branch: "b"}, Err: errors.New("fail")},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.result.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddBatchResult_ErrorCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result AddBatchResult
		want   int
	}{
		{
			name:   "empty",
			result: AddBatchResult{},
			want:   0,
		},
		{
			name: "no_errors",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "a"}},
					{AddResult: AddResult{Branch: "b"}},
				},
			},
			want: 0,
		},
		{
			name: "one_error",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "a"}},
					{AddResult: AddResult{Branch: "b"}, Err: errors.New("fail")},
				},
			},
			want: 1,
		},
		{
			name: "all_errors",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "a"}, Err: errors.New("fail1")},
					{AddResult: AddResult{Branch: "b"}, Err: errors.New("fail2")},
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.result.ErrorCount(); got != tt.want {
				t.Errorf("ErrorCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddBatchResult_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     AddBatchResult
		opts       AddFormatOptions
		wantStdout string
		wantStderr string
	}{
		{
			name: "single_success",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "feature/a", WorktreePath: "/wt/feature/a", Symlinks: []SymlinkResult{{Src: "/src", Dst: "/dst"}}}},
				},
			},
			opts:       AddFormatOptions{},
			wantStdout: "gwt add: feature/a (1 symlinks)\n",
			wantStderr: "",
		},
		{
			name: "multiple_success",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "feature/a", WorktreePath: "/wt/feature/a"}},
					{AddResult: AddResult{Branch: "feature/b", WorktreePath: "/wt/feature/b"}},
				},
			},
			opts:       AddFormatOptions{},
			wantStdout: "gwt add: feature/a (0 symlinks)\ngwt add: feature/b (0 symlinks)\n",
			wantStderr: "",
		},
		{
			name: "print_path_multiple",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "feature/a", WorktreePath: "/wt/feature/a"}},
					{AddResult: AddResult{Branch: "feature/b", WorktreePath: "/wt/feature/b"}},
				},
			},
			opts:       AddFormatOptions{Print: []string{"path"}},
			wantStdout: "/wt/feature/a\n/wt/feature/b\n",
			wantStderr: "",
		},
		{
			name: "mixed_success_and_error",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "feature/a", WorktreePath: "/wt/feature/a"}},
					{AddResult: AddResult{Branch: "feature/b"}, Err: errors.New("directory already exists")},
				},
			},
			opts:       AddFormatOptions{},
			wantStdout: "gwt add: feature/a (0 symlinks)\n",
			wantStderr: "error: feature/b: directory already exists\n",
		},
		{
			name: "print_path_with_error",
			result: AddBatchResult{
				Added: []AddedWorktree{
					{AddResult: AddResult{Branch: "feature/a", WorktreePath: "/wt/feature/a"}},
					{AddResult: AddResult{Branch: "feature/b"}, Err: errors.New("fail")},
				},
			},
			opts:       AddFormatOptions{Print: []string{"path"}},
			wantStdout: "/wt/feature/a\n",
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.result.Format(tt.opts)

			if got.Stdout != tt.wantStdout {
				t.Errorf("Stdout = %q, want %q", got.Stdout, tt.wantStdout)
			}
			if got.Stderr != tt.wantStderr {
				t.Errorf("Stderr = %q, want %q", got.Stderr, tt.wantStderr)
			}
		})
	}
}

func TestAddCommand_Run_Multiple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		branches     []string
		config       *Config
		sync         bool
		setupFS      func(t *testing.T) *testutil.MockFS
		setupGit     func(t *testing.T) *testutil.MockGitExecutor
		wantErr      bool
		wantErrCount int
		wantSynced   bool
	}{
		{
			name:     "multiple_success",
			branches: []string{"feature/a", "feature/b"},
			config:   &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/worktrees"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{}
			},
			wantErr:      false,
			wantErrCount: 0,
		},
		{
			name:     "partial_failure",
			branches: []string{"feature/ok", "existing"},
			config:   &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/worktrees"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{
					ExistingPaths: []string{"/repo/worktrees/existing"},
				}
			},
			setupGit: func(t *testing.T) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{}
			},
			wantErr:      false,
			wantErrCount: 1,
		},
		{
			name:     "sync_with_multiple",
			branches: []string{"feature/a", "feature/b"},
			config:   &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/worktrees"},
			sync:     true,
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{HasChanges: true}
			},
			wantErr:    false,
			wantSynced: true,
		},
		{
			name:     "empty_branches",
			branches: []string{},
			config:   &Config{WorktreeSourceDir: "/repo/main", WorktreeDestBaseDir: "/repo/worktrees"},
			setupFS: func(t *testing.T) *testutil.MockFS {
				t.Helper()
				return &testutil.MockFS{}
			},
			setupGit: func(t *testing.T) *testutil.MockGitExecutor {
				t.Helper()
				return &testutil.MockGitExecutor{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockFS := tt.setupFS(t)
			mockGit := tt.setupGit(t)

			cmd := &AddCommand{
				FS:     mockFS,
				Git:    &GitRunner{Executor: mockGit},
				Config: tt.config,
				Sync:   tt.sync,
			}

			result, err := cmd.Run(tt.branches)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ErrorCount() != tt.wantErrCount {
				t.Errorf("ErrorCount() = %d, want %d", result.ErrorCount(), tt.wantErrCount)
			}

			if result.ChangesSynced != tt.wantSynced {
				t.Errorf("ChangesSynced = %v, want %v", result.ChangesSynced, tt.wantSynced)
			}

			// Verify correct number of results
			if len(result.Added) != len(tt.branches) {
				t.Errorf("len(Added) = %d, want %d", len(result.Added), len(tt.branches))
			}
		})
	}
}
