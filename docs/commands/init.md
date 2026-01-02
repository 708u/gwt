# init subcommand

Initialize gwt configuration in the current directory.

## Usage

```txt
gwt init [flags]
```

## Flags

| Flag      | Short | Description                        |
|-----------|-------|------------------------------------|
| `--force` | `-f`  | Overwrite existing configuration   |

## Behavior

- Creates `.gwt/` directory if it doesn't exist
- Generates `.gwt/settings.toml` with default configuration template
- If `settings.toml` already exists, skips creation (unless `--force` is used)

### Generated Configuration

The generated `settings.toml` contains:

```toml
# gwt project configuration
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
```

### Force Option

With `--force`, overwrites the existing configuration file.

```bash
gwt init --force
```

## Examples

```txt
# Initialize gwt in current directory
gwt init
Created .gwt/settings.toml

# Running again without force skips
gwt init
Skipped .gwt/settings.toml (already exists)

# Force overwrite existing configuration
gwt init --force
Created .gwt/settings.toml (overwritten)
```

## Exit Code

- 0: Configuration created or skipped successfully
- 1: Failed to create configuration
