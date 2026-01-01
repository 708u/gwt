# Phase 1: GitRunner 拡張

## 対象ファイル

- `git.go`

## 追加するメソッド

### WorktreeFindByBranch

ブランチ名から worktree パスを返す。

```go
// WorktreeFindByBranch returns the worktree path for the given branch.
// Returns an error if the branch is not checked out in any worktree.
func (g *GitRunner) WorktreeFindByBranch(branch string) (string, error)
```

実装:

```go
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
        if strings.HasPrefix(line, "worktree ") {
            currentPath = strings.TrimPrefix(line, "worktree ")
        }
        if strings.HasPrefix(line, "branch refs/heads/") {
            branchName := strings.TrimPrefix(line, "branch refs/heads/")
            if branchName == branch {
                return currentPath, nil
            }
        }
    }

    return "", fmt.Errorf("branch %q is not checked out in any worktree", branch)
}
```

### WorktreeRemove

worktree を削除する。

```go
// WorktreeRemove removes the worktree at the given path.
// If force is true, removes even if there are uncommitted changes.
func (g *GitRunner) WorktreeRemove(path string, force bool) error
```

実装:

```go
func (g *GitRunner) WorktreeRemove(path string, force bool) error {
    args := []string{"worktree", "remove"}
    if force {
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
```

### BranchDelete

ローカルブランチを削除する。

```go
// BranchDelete deletes a local branch.
// If force is true, uses -D (force delete); otherwise uses -d (safe delete).
func (g *GitRunner) BranchDelete(branch string, force bool) error
```

実装:

```go
func (g *GitRunner) BranchDelete(branch string, force bool) error {
    flag := "-d"
    if force {
        flag = "-D"
    }

    out, err := g.Executor.Run("branch", flag, branch)
    if err != nil {
        return fmt.Errorf("failed to delete branch: %w", err)
    }
    g.Stdout.Write(out)
    return nil
}
```

## 動作確認

```bash
go test -run TestGitRunner -v ./...
```
