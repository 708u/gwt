package gwt

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// AddCommand creates git worktrees with symlinks.
type AddCommand struct {
	FS     FileSystem
	Git    *GitRunner
	Config *Config
	Sync   bool
}

// AddOptions holds options for the add command.
type AddOptions struct {
	Sync bool
}

// NewAddCommand creates a new AddCommand with the given config.
func NewAddCommand(cfg *Config, opts AddOptions) *AddCommand {
	return &AddCommand{
		FS:     osFS{},
		Git:    NewGitRunner(cfg.WorktreeSourceDir),
		Config: cfg,
		Sync:   opts.Sync,
	}
}

// SymlinkResult holds information about a symlink operation.
type SymlinkResult struct {
	Src     string
	Dst     string
	Skipped bool
	Reason  string
}

// AddResult holds the result of an add operation.
type AddResult struct {
	Branch        string
	WorktreePath  string
	Symlinks      []SymlinkResult
	GitOutput     []byte
	ChangesSynced bool
}

// AddedWorktree holds the result of a single add operation with optional error.
type AddedWorktree struct {
	AddResult
	Err error
}

// AddBatchResult aggregates results from multiple add operations.
type AddBatchResult struct {
	Added         []AddedWorktree
	ChangesSynced bool
}

// HasErrors returns true if any add operations failed.
func (r AddBatchResult) HasErrors() bool {
	for _, wt := range r.Added {
		if wt.Err != nil {
			return true
		}
	}
	return false
}

// ErrorCount returns the number of failed add operations.
func (r AddBatchResult) ErrorCount() int {
	count := 0
	for _, wt := range r.Added {
		if wt.Err != nil {
			count++
		}
	}
	return count
}

// SuccessCount returns the number of successful add operations.
func (r AddBatchResult) SuccessCount() int {
	return len(r.Added) - r.ErrorCount()
}

// AddFormatOptions configures add output formatting.
type AddFormatOptions struct {
	Verbose bool
	Print   []string // ["path"], ["path", "branch"], or empty for default
}

// ValidPrintFields contains valid --print field names.
var ValidPrintFields = []string{"path"}

// ValidatePrintFields validates the given print fields.
func ValidatePrintFields(fields []string) error {
	for _, f := range fields {
		if !slices.Contains(ValidPrintFields, f) {
			return fmt.Errorf("invalid print field: %s (valid: %s)",
				f, strings.Join(ValidPrintFields, ", "))
		}
	}
	return nil
}

// Format formats the AddResult for display.
func (r AddResult) Format(opts AddFormatOptions) FormatResult {
	// Print overrides default output
	if len(opts.Print) > 0 {
		return r.formatPrint(opts.Print)
	}
	return r.formatDefault(opts)
}

// formatPrint outputs only the specified fields.
func (r AddResult) formatPrint(fields []string) FormatResult {
	var stdout strings.Builder
	for _, field := range fields {
		switch field {
		case "path":
			stdout.WriteString(r.WorktreePath)
			stdout.WriteString("\n")
		}
	}
	return FormatResult{Stdout: stdout.String()}
}

// formatDefault outputs the default or verbose format.
func (r AddResult) formatDefault(opts AddFormatOptions) FormatResult {
	var stdout, stderr strings.Builder

	var createdCount int
	for _, s := range r.Symlinks {
		if s.Skipped {
			stderr.WriteString(fmt.Sprintf("warning: %s\n", s.Reason))
		} else {
			createdCount++
		}
	}

	if opts.Verbose {
		if len(r.GitOutput) > 0 {
			stdout.Write(r.GitOutput)
		}
		stdout.WriteString(fmt.Sprintf("Created worktree at %s\n", r.WorktreePath))
		for _, s := range r.Symlinks {
			if !s.Skipped {
				stdout.WriteString(fmt.Sprintf("Created symlink: %s -> %s\n", s.Dst, s.Src))
			}
		}
		if r.ChangesSynced {
			stdout.WriteString("Synced uncommitted changes\n")
		}
	}

	var syncInfo string
	if r.ChangesSynced {
		syncInfo = ", synced"
	}
	stdout.WriteString(fmt.Sprintf("gwt add: %s (%d symlinks%s)\n", r.Branch, createdCount, syncInfo))

	return FormatResult{Stdout: stdout.String(), Stderr: stderr.String()}
}

// Format formats the AddBatchResult for display.
func (r AddBatchResult) Format(opts AddFormatOptions) FormatResult {
	if len(opts.Print) > 0 {
		return r.formatPrint(opts.Print)
	}
	return r.formatDefault(opts)
}

func (r AddBatchResult) formatPrint(fields []string) FormatResult {
	var stdout strings.Builder
	for _, wt := range r.Added {
		if wt.Err != nil {
			continue
		}
		for _, field := range fields {
			switch field {
			case "path":
				stdout.WriteString(wt.WorktreePath)
				stdout.WriteString("\n")
			}
		}
	}
	return FormatResult{Stdout: stdout.String()}
}

func (r AddBatchResult) formatDefault(opts AddFormatOptions) FormatResult {
	var stdout, stderr strings.Builder

	for _, wt := range r.Added {
		if wt.Err != nil {
			stderr.WriteString(fmt.Sprintf("error: %s: %v\n", wt.Branch, wt.Err))
			continue
		}
		formatted := wt.AddResult.Format(opts)
		stdout.WriteString(formatted.Stdout)
		stderr.WriteString(formatted.Stderr)
	}

	return FormatResult{Stdout: stdout.String(), Stderr: stderr.String()}
}

// Run creates worktrees for the given branch names.
// With Sync enabled, changes are stashed once and applied to all new worktrees.
func (c *AddCommand) Run(names []string) (AddBatchResult, error) {
	var result AddBatchResult

	if len(names) == 0 {
		return result, fmt.Errorf("at least one branch name is required")
	}

	if c.Config.WorktreeSourceDir == "" {
		return result, fmt.Errorf("worktree source directory is not configured")
	}
	if c.Config.WorktreeDestBaseDir == "" {
		return result, fmt.Errorf("worktree destination base directory is not configured")
	}

	// Handle stash once for all worktrees
	var shouldSync bool
	if c.Sync {
		hasChanges, err := c.Git.HasChanges()
		if err != nil {
			return result, fmt.Errorf("failed to check for changes: %w", err)
		}
		shouldSync = hasChanges
		if shouldSync {
			if _, err := c.Git.StashPush("gwt sync"); err != nil {
				return result, fmt.Errorf("failed to stash changes: %w", err)
			}
		}
	}

	// Process each branch
	for _, name := range names {
		addResult, err := c.runSingle(name, shouldSync)
		added := AddedWorktree{AddResult: addResult}
		if err != nil {
			added.Err = err
			added.Branch = name
		}
		result.Added = append(result.Added, added)
	}

	// Restore stash in source (pop to clean up stash stack)
	if shouldSync {
		if _, err := c.Git.StashPop(); err != nil {
			// Log warning but don't fail - worktrees are already created
			// The stash still exists and user can manually pop it
		}
		result.ChangesSynced = true
	}

	return result, nil
}

// runSingle creates a single worktree. If shouldSync is true, applies stash to new worktree.
func (c *AddCommand) runSingle(name string, shouldSync bool) (AddResult, error) {
	var result AddResult
	result.Branch = name

	if name == "" {
		return result, fmt.Errorf("branch name is required")
	}

	wtPath := filepath.Join(c.Config.WorktreeDestBaseDir, name)
	result.WorktreePath = wtPath

	gitOutput, err := c.createWorktree(name, wtPath)
	if err != nil {
		return result, err
	}
	result.GitOutput = gitOutput

	// Apply stash to new worktree if sync is enabled
	if shouldSync {
		if _, err := c.Git.InDir(wtPath).StashApply(); err != nil {
			// Rollback: remove worktree on stash apply failure
			_, _ = c.Git.WorktreeRemove(wtPath, WithForceRemove())
			return result, fmt.Errorf("failed to apply changes to new worktree: %w", err)
		}
		result.ChangesSynced = true
	}

	symlinks, err := c.createSymlinks(
		c.Config.WorktreeSourceDir, wtPath, c.Config.Symlinks)
	if err != nil {
		return result, err
	}
	result.Symlinks = symlinks

	return result, nil
}

func (c *AddCommand) createWorktree(branch, path string) ([]byte, error) {
	if _, err := c.FS.Stat(path); err == nil {
		return nil, fmt.Errorf("directory already exists: %s", path)
	}

	var opts []WorktreeAddOption
	if c.Git.BranchExists(branch) {
		branches, err := c.Git.WorktreeListBranches()
		if err != nil {
			return nil, fmt.Errorf("failed to list worktree branches: %w", err)
		}
		if slices.Contains(branches, branch) {
			return nil, fmt.Errorf("branch %s is already checked out in another worktree", branch)
		}
	} else {
		opts = append(opts, WithCreateBranch())
	}

	output, err := c.Git.WorktreeAdd(path, branch, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}

	return output, nil
}

func (c *AddCommand) createSymlinks(
	srcDir, dstDir string, patterns []string) ([]SymlinkResult, error) {
	var results []SymlinkResult

	for _, pattern := range patterns {
		matches, err := c.FS.Glob(srcDir, pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}
		if len(matches) == 0 {
			results = append(results, SymlinkResult{
				Skipped: true,
				Reason:  fmt.Sprintf("%s does not match any files, skipping", pattern),
			})
			continue
		}

		for _, match := range matches {
			src := filepath.Join(srcDir, match)
			dst := filepath.Join(dstDir, match)

			// Skip if destination already exists (e.g., git-tracked file checked out by worktree).
			if _, err := c.FS.Stat(dst); err == nil {
				results = append(results, SymlinkResult{
					Src:     src,
					Dst:     dst,
					Skipped: true,
					Reason:  fmt.Sprintf("skipping symlink for %s (already exists)", match),
				})
				continue
			}

			if dir := filepath.Dir(dst); dir != dstDir {
				if err := c.FS.MkdirAll(dir, 0755); err != nil {
					return nil, fmt.Errorf("failed to create directory for %s: %w", match, err)
				}
			}

			if err := c.FS.Symlink(src, dst); err != nil {
				return nil, fmt.Errorf("failed to create symlink for %s: %w", match, err)
			}

			results = append(results, SymlinkResult{Src: src, Dst: dst})
		}
	}

	return results, nil
}
