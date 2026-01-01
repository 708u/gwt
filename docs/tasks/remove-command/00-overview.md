# gwt remove コマンド実装

## 概要

`gwt remove <branch>` で worktree とブランチを同時に削除するコマンドを追加する。
`gwt add` の逆操作として、worktree の削除とブランチの削除を一つのコマンドで行う。

## 使用例

```bash
gwt remove feature/foo           # worktree と branch を削除
gwt remove -f feature/foo        # 強制削除
gwt remove --dry-run feature/foo # 削除内容をプレビュー
```

## フラグ

| Flag | Short | Description |
|------|-------|-------------|
| --force | -f | uncommitted changes や unmerged branch でも強制削除 |
| --dry-run | | 削除内容を表示するだけで実行しない |

## 動作フロー

```txt
1. ブランチ名から worktree パスを特定
   └─ git worktree list --porcelain で全 worktree を取得
   └─ ブランチ名に一致する worktree パスを抽出

2. 削除前の検証
   └─ worktree が存在するか
   └─ 現在のディレクトリが削除対象内でないか

3. 削除実行
   └─ git worktree remove [-f] <path>
   └─ git branch -d/-D <branch>
```

## 実装 Phase

| Phase | 内容 | ファイル |
|-------|------|---------|
| [Phase 1](./01-phase-git-runner.md) | GitRunner 拡張 | git.go |
| [Phase 2](./02-phase-filesystem.md) | FileSystem 拡張 | fs.go |
| [Phase 3](./03-phase-remove-command.md) | RemoveCommand 実装 | remove.go |
| [Phase 4](./04-phase-cli.md) | CLI 統合 | cmd/gwt/main.go |
| [Phase 5](./05-phase-test.md) | テスト | remove_test.go, mock_*.go |

## 変更対象ファイル一覧

| File | Changes |
|------|---------|
| git.go | WorktreeFindByBranch, WorktreeRemove, BranchDelete 追加 |
| fs.go | Abs メソッド追加 |
| remove.go | 新規作成 |
| remove_test.go | 新規作成 |
| cmd/gwt/main.go | removeCmd 追加 |
| internal/testutil/mock_git.go | モック拡張 |
| internal/testutil/mock_fs.go | Abs モック追加 |
