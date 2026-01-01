package gwt

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// RemoveCommand removes git worktrees with their associated branches.
type RemoveCommand struct {
	FS     FileSystem
	Git    *GitRunner
	Config *Config
	Stdout io.Writer
	Stderr io.Writer
}

// RemoveOptions configures the remove operation.
type RemoveOptions struct {
	Force  bool
	DryRun bool
}

// NewRemoveCommand creates a new RemoveCommand with the given config.
func NewRemoveCommand(cfg *Config) *RemoveCommand {
	return &RemoveCommand{
		FS:     osFS{},
		Git:    NewGitRunner(cfg.WorktreeSourceDir),
		Config: cfg,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Run removes the worktree and branch for the given branch name.
// cwd is the current working directory (absolute path) passed from CLI layer.
func (c *RemoveCommand) Run(branch string, cwd string, opts RemoveOptions) error {
	if branch == "" {
		return fmt.Errorf("branch name is required")
	}
	if c.Config.WorktreeSourceDir == "" {
		return fmt.Errorf("worktree source directory is not configured")
	}

	wtPath, err := c.Git.WorktreeFindByBranch(branch)
	if err != nil {
		return err
	}

	if strings.HasPrefix(cwd, wtPath) {
		return fmt.Errorf("cannot remove: current directory is inside worktree %s", wtPath)
	}

	if opts.DryRun {
		fmt.Fprintf(c.Stdout, "Would remove worktree: %s\n", wtPath)
		fmt.Fprintf(c.Stdout, "Would delete branch: %s\n", branch)
		return nil
	}

	var wtOpts []WorktreeRemoveOption
	if opts.Force {
		wtOpts = append(wtOpts, WithForceRemove())
	}
	if err := c.Git.WorktreeRemove(wtPath, wtOpts...); err != nil {
		return err
	}

	var branchOpts []BranchDeleteOption
	if opts.Force {
		branchOpts = append(branchOpts, WithForceDelete())
	}
	if err := c.Git.BranchDelete(branch, branchOpts...); err != nil {
		return err
	}

	fmt.Fprintf(c.Stdout, "Removed worktree and branch: %s\n", branch)
	return nil
}
