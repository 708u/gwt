package gwt

import (
	"fmt"
	"path/filepath"
)

const settingsTemplate = `# gwt project configuration
# See: https://github.com/708u/gwt-worktree

# Symlink patterns to create in new worktrees
# Example: symlinks = [".envrc", ".tool-versions", "node_modules"]
symlinks = []

# Worktree destination base directory (default: ../<repo-name>-worktree)
# worktree_destination_base_dir = "../my-worktrees"

# Worktree source directory (default: current directory)
# worktree_source_dir = "."

# Default source branch for new worktrees
# default_source = "main"
`

// InitCommand initializes gwt configuration in a directory.
type InitCommand struct {
	FS FileSystem
}

// InitOptions holds options for the init command.
type InitOptions struct {
	Force bool
}

// InitResult holds the result of the init command.
type InitResult struct {
	ConfigDir   string
	SettingsPath string
	Created     bool
	Skipped     bool
	Overwritten bool
}

// InitFormatOptions holds formatting options for InitResult.
type InitFormatOptions struct {
	Verbose bool
}

// NewInitCommand creates a new InitCommand with default dependencies.
func NewInitCommand() *InitCommand {
	return &InitCommand{
		FS: osFS{},
	}
}

// Run executes the init command.
func (c *InitCommand) Run(dir string, opts InitOptions) (InitResult, error) {
	configDirPath := filepath.Join(dir, configDir)
	settingsPath := filepath.Join(configDirPath, configFileName)

	result := InitResult{
		ConfigDir:    configDirPath,
		SettingsPath: settingsPath,
	}

	// Check if settings file already exists
	_, err := c.FS.Stat(settingsPath)
	exists := err == nil || !c.FS.IsNotExist(err)

	if exists && !opts.Force {
		result.Skipped = true
		return result, nil
	}

	// Create config directory
	if err := c.FS.MkdirAll(configDirPath, 0755); err != nil {
		return result, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write settings file
	if err := c.FS.WriteFile(settingsPath, []byte(settingsTemplate), 0644); err != nil {
		return result, fmt.Errorf("failed to write settings file: %w", err)
	}

	result.Created = true
	if exists {
		result.Overwritten = true
	}

	return result, nil
}

// Format formats the result for output.
func (r InitResult) Format(opts InitFormatOptions) FormatResult {
	var stdout string

	relPath := filepath.Join(configDir, configFileName)

	if r.Skipped {
		stdout = fmt.Sprintf("Skipped %s (already exists)\n", relPath)
	} else if r.Overwritten {
		stdout = fmt.Sprintf("Created %s (overwritten)\n", relPath)
	} else if r.Created {
		stdout = fmt.Sprintf("Created %s\n", relPath)
	}

	return FormatResult{
		Stdout: stdout,
	}
}
