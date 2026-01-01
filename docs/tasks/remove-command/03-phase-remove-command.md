# Phase 3: RemoveCommand 実装

## 対象ファイル

- `remove.go` (新規作成)

## 構造体定義

```go
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
```

## コンストラクタ

```go
func NewRemoveCommand(cfg *Config) *RemoveCommand {
    return &RemoveCommand{
        FS:     osFS{},
        Git:    NewGitRunner(cfg.WorktreeSourceDir),
        Config: cfg,
        Stdout: os.Stdout,
        Stderr: os.Stderr,
    }
}
```

## Run メソッド

```go
func (c *RemoveCommand) Run(branch string, opts RemoveOptions) error {
    // 1. ブランチから worktree パスを取得
    wtPath, err := c.Git.WorktreeFindByBranch(branch)
    if err != nil {
        return err
    }

    // 2. 現在ディレクトリが worktree 内でないか検証
    cwd, err := c.FS.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %w", err)
    }

    absWtPath, err := c.FS.Abs(wtPath)
    if err != nil {
        return fmt.Errorf("failed to resolve worktree path: %w", err)
    }

    absCwd, err := c.FS.Abs(cwd)
    if err != nil {
        return fmt.Errorf("failed to resolve current directory: %w", err)
    }

    if strings.HasPrefix(absCwd, absWtPath) {
        return fmt.Errorf("cannot remove: current directory is inside worktree %s", wtPath)
    }

    // 3. dry-run の場合はプレビューのみ
    if opts.DryRun {
        fmt.Fprintf(c.Stdout, "Would remove worktree: %s\n", wtPath)
        fmt.Fprintf(c.Stdout, "Would delete branch: %s\n", branch)
        return nil
    }

    // 4. worktree 削除
    if err := c.Git.WorktreeRemove(wtPath, opts.Force); err != nil {
        return err
    }

    // 5. ブランチ削除
    if err := c.Git.BranchDelete(branch, opts.Force); err != nil {
        return err
    }

    // 6. 成功メッセージ
    fmt.Fprintf(c.Stdout, "Removed worktree and branch: %s\n", branch)
    return nil
}
```

## エラーケース

| Condition | Error Message |
|-----------|---------------|
| ブランチが worktree に紐づいていない | `branch "%s" is not checked out in any worktree` |
| 現在ディレクトリが削除対象内 | `cannot remove: current directory is inside worktree %s` |
| uncommitted changes (force なし) | git のエラーをそのまま返す |
| branch 削除失敗 | git のエラーをそのまま返す |

## 動作確認

```bash
go test -run TestRemoveCommand -v ./...
```
