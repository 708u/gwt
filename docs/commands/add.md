# add subcommand

Create worktrees with optional symlinks.

## Usage

```txt
gwt add <name>... [flags]
```

## Arguments

- `<name>...`: One or more branch names (required)

## Flags

| Flag              | Short | Description                                |
|-------------------|-------|--------------------------------------------|
| `--sync`          | `-s`  | Sync uncommitted changes to new worktrees  |
| `--print <field>` |       | Print specific field (path)                |

## Behavior

- Creates worktrees at `WorktreeDestBaseDir/<name>` for each branch
- If a branch already exists, uses that branch
- If a branch doesn't exist, creates a new branch with `-b` flag
- Creates symlinks from `WorktreeSourceDir` to worktrees
  based on `Config.Symlinks` patterns
- Warns when symlink patterns don't match any files
- Errors on individual branches do not stop processing of remaining branches
- Exit code 0 if all succeed, 1 if any fail

### Sync Option

With `--sync`, uncommitted changes are copied to all new worktrees:

1. Stashes current changes (once)
2. For each branch:
   - Creates the new worktree
   - Applies stash to new worktree
3. Restores changes in the source worktree (pops stash)

If a worktree creation or stash apply fails, that branch is skipped
but processing continues for remaining branches.

### Print Option

With `--print`, only the specified field is output to stdout.
This is useful for piping to other commands.

```bash
cd $(gwt add feat/x --print path)
```

Available fields:

- `path`: Worktree path

When `--print` is specified, `--verbose` is ignored.

## Examples

```bash
# Create single worktree
gwt add feature/new-feature

# Create multiple worktrees
gwt add feature/a feature/b feature/c

# Sync changes to multiple worktrees
gwt add --sync feature/a feature/b

# Print paths for scripting
gwt add --print path feature/a feature/b
```
