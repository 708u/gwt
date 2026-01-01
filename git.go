package gwt

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// GitExecutor abstracts git command execution for testability.
// Commands are fixed to "git" - only subcommands and args are passed.
type GitExecutor interface {
	// Run executes git with args and returns stdout.
	Run(args ...string) ([]byte, error)
}

type osGitExecutor struct {
	dir string
}

func (e osGitExecutor) Run(args ...string) ([]byte, error) {
	fullArgs := append([]string{"-C", e.dir}, args...)
	return exec.Command("git", fullArgs...).Output()
}

// GitRunner provides git operations using GitExecutor.
type GitRunner struct {
	Executor GitExecutor
	Stdout   io.Writer
}

// NewGitRunner creates a new GitRunner with the default executor.
func NewGitRunner(dir string) *GitRunner {
	return &GitRunner{
		Executor: osGitExecutor{dir: dir},
		Stdout:   os.Stdout,
	}
}

type worktreeAddOptions struct {
	createBranch bool
}

// WorktreeAddOption is a functional option for WorktreeAdd.
type WorktreeAddOption func(*worktreeAddOptions)

// WithCreateBranch creates a new branch when adding the worktree.
func WithCreateBranch() WorktreeAddOption {
	return func(o *worktreeAddOptions) {
		o.createBranch = true
	}
}

// WorktreeAdd creates a new worktree at the specified path.
func (g *GitRunner) WorktreeAdd(path, branch string, opts ...WorktreeAddOption) error {
	var o worktreeAddOptions
	for _, opt := range opts {
		opt(&o)
	}

	var output []byte
	var err error
	if o.createBranch {
		output, err = g.Executor.Run("worktree", "add", "-b", branch, path)
	} else {
		output, err = g.Executor.Run("worktree", "add", path, branch)
	}
	if len(output) > 0 {
		fmt.Fprint(g.Stdout, string(output))
	}
	return err
}

// BranchExists checks if a branch exists in the local repository.
func (g *GitRunner) BranchExists(branch string) bool {
	_, err := g.Executor.Run("rev-parse", "--verify", "refs/heads/"+branch)
	return err == nil
}

// WorktreeListBranches returns a list of branch names currently checked out in worktrees.
func (g *GitRunner) WorktreeListBranches() ([]string, error) {
	output, err := g.Executor.Run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var branches []string
	for line := range strings.SplitSeq(string(output), "\n") {
		if branch, ok := strings.CutPrefix(line, "branch refs/heads/"); ok {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// WorktreeFindByBranch returns the worktree path for the given branch.
// Returns an error if the branch is not checked out in any worktree.
func (g *GitRunner) WorktreeFindByBranch(branch string) (string, error) {
	out, err := g.Executor.Run("worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	// porcelain format:
	// worktree /path/to/worktree
	// HEAD abc123
	// branch refs/heads/branch-name
	// (blank line)

	lines := strings.Split(string(out), "\n")
	var currentPath string
	for _, line := range lines {
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			currentPath = path
		}
		if branchName, ok := strings.CutPrefix(line, "branch refs/heads/"); ok {
			if branchName == branch {
				return currentPath, nil
			}
		}
	}

	return "", fmt.Errorf("branch %q is not checked out in any worktree", branch)
}

type worktreeRemoveOptions struct {
	force bool
}

// WorktreeRemoveOption is a functional option for WorktreeRemove.
type WorktreeRemoveOption func(*worktreeRemoveOptions)

// WithForceRemove forces worktree removal even if there are uncommitted changes.
func WithForceRemove() WorktreeRemoveOption {
	return func(o *worktreeRemoveOptions) {
		o.force = true
	}
}

// WorktreeRemove removes the worktree at the given path.
// By default fails if there are uncommitted changes. Use WithForceRemove() to force.
func (g *GitRunner) WorktreeRemove(path string, opts ...WorktreeRemoveOption) error {
	var o worktreeRemoveOptions
	for _, opt := range opts {
		opt(&o)
	}

	args := []string{"worktree", "remove"}
	if o.force {
		args = append(args, "-f")
	}
	args = append(args, path)

	out, err := g.Executor.Run(args...)
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	g.Stdout.Write(out)
	return nil
}

type branchDeleteOptions struct {
	force bool
}

// BranchDeleteOption is a functional option for BranchDelete.
type BranchDeleteOption func(*branchDeleteOptions)

// WithForceDelete forces branch deletion even if not fully merged.
func WithForceDelete() BranchDeleteOption {
	return func(o *branchDeleteOptions) {
		o.force = true
	}
}

// BranchDelete deletes a local branch.
// By default uses -d (safe delete). Use WithForceDelete() to use -D (force delete).
func (g *GitRunner) BranchDelete(branch string, opts ...BranchDeleteOption) error {
	var o branchDeleteOptions
	for _, opt := range opts {
		opt(&o)
	}

	flag := "-d"
	if o.force {
		flag = "-D"
	}

	out, err := g.Executor.Run("branch", flag, branch)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}
	g.Stdout.Write(out)
	return nil
}
