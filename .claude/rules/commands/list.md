# list subcommand

List all worktrees in `git worktree list` compatible format.

## Usage

```txt
gwt list
```

## Behavior

- Lists all worktrees including the main worktree
- Output format is compatible with `git worktree list`
- Shows path, short commit hash, and branch name for each worktree
- Displays additional status: locked, prunable, detached HEAD

## Examples

```txt
gwt list
/Users/user/repo                                   d9ef543 [main]
/Users/user/repo-worktree/feat/add-list-command    abc1234 [feat/add-list-command]
/Users/user/repo-worktree/feat/add-move-command    def5678 [feat/add-move-command]

# Detached HEAD example
/Users/user/repo-worktree/detached                 1234abc (detached HEAD)

# Locked worktree example
/Users/user/repo-worktree/locked                   5678def [locked-branch] locked
```

## Shell Integration

Combine with fzf for quick worktree navigation:

```bash
gcd() {
  local selected
  selected=$(gwt list | awk '{print $1}' | fzf +m) &&
  cd "$selected"
}
```
