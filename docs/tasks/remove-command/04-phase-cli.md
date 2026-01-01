# Phase 4: CLI 統合

## 対象ファイル

- `cmd/gwt/main.go`

## removeCmd 定義

```go
var removeCmd = &cobra.Command{
    Use:   "remove <branch>",
    Short: "Remove a worktree and its branch",
    Long: `Remove a git worktree and delete its associated branch.

The branch name is used to locate the worktree.
By default, fails if there are uncommitted changes or the branch is not merged.
Use --force to override these checks.`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        force, _ := cmd.Flags().GetBool("force")
        dryRun, _ := cmd.Flags().GetBool("dry-run")

        cwd, err := os.Getwd()
        if err != nil {
            return fmt.Errorf("failed to get current directory: %w", err)
        }

        result, err := gwt.LoadConfig(cwd)
        if err != nil {
            return fmt.Errorf("failed to load config: %w", err)
        }
        for _, w := range result.Warnings {
            fmt.Fprintln(os.Stderr, "warning:", w)
        }

        return gwt.NewRemoveCommand(result.Config).Run(args[0], gwt.RemoveOptions{
            Force:  force,
            DryRun: dryRun,
        })
    },
}
```

## init() への追加

```go
func init() {
    removeCmd.Flags().BoolP("force", "f", false, "Force removal even with uncommitted changes or unmerged branch")
    removeCmd.Flags().Bool("dry-run", false, "Show what would be removed without making changes")
    rootCmd.AddCommand(removeCmd)
}
```

## 使用例

```bash
# 基本的な削除
gwt remove feature/test

# 強制削除
gwt remove -f feature/test

# プレビュー
gwt remove --dry-run feature/test

# 組み合わせ
gwt remove -f --dry-run feature/test
```

## 動作確認

```bash
make build
./out/gwt remove --help
```
