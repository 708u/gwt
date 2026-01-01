# Phase 5: テスト

## 対象ファイル

- `remove_test.go` (新規作成)
- `internal/testutil/mock_git.go` (拡張)
- `internal/testutil/mock_fs.go` (拡張)

## MockGitExecutor 拡張

```go
// 既存フィールドに追加
type MockGitExecutor struct {
    // ... existing fields ...
    WorktreeRemoveErr error
    BranchDeleteErr   error
}
```

Run メソッドで worktree remove / branch -d/-D を処理:

```go
func (m *MockGitExecutor) Run(args ...string) ([]byte, error) {
    // ... existing logic ...

    // worktree remove
    if len(args) >= 2 && args[0] == "worktree" && args[1] == "remove" {
        if m.WorktreeRemoveErr != nil {
            return nil, m.WorktreeRemoveErr
        }
        return nil, nil
    }

    // branch -d/-D
    if len(args) >= 2 && args[0] == "branch" && (args[1] == "-d" || args[1] == "-D") {
        if m.BranchDeleteErr != nil {
            return nil, m.BranchDeleteErr
        }
        return []byte(fmt.Sprintf("Deleted branch %s\n", args[2])), nil
    }

    return nil, nil
}
```

## MockFS 拡張

```go
type MockFS struct {
    // ... existing fields ...
    AbsFunc func(path string) (string, error)
}

func (m MockFS) Abs(path string) (string, error) {
    if m.AbsFunc != nil {
        return m.AbsFunc(path)
    }
    return filepath.Abs(path)
}
```

## remove_test.go

### テストケース

1. **正常系: worktree とブランチの削除**

```go
func TestRemoveCommand_Run_Success(t *testing.T) {
    // worktree が存在し、正常に削除できるケース
}
```

2. **正常系: dry-run でプレビュー表示**

```go
func TestRemoveCommand_Run_DryRun(t *testing.T) {
    // dry-run で実際には削除しないケース
}
```

3. **異常系: ブランチが worktree に紐づいていない**

```go
func TestRemoveCommand_Run_BranchNotInWorktree(t *testing.T) {
    // WorktreeFindByBranch がエラーを返すケース
}
```

4. **異常系: 現在ディレクトリが削除対象内**

```go
func TestRemoveCommand_Run_InsideWorktree(t *testing.T) {
    // cwd が worktree パス内にあるケース
}
```

5. **異常系: worktree 削除失敗**

```go
func TestRemoveCommand_Run_WorktreeRemoveFails(t *testing.T) {
    // WorktreeRemove がエラーを返すケース
}
```

6. **異常系: ブランチ削除失敗**

```go
func TestRemoveCommand_Run_BranchDeleteFails(t *testing.T) {
    // BranchDelete がエラーを返すケース
}
```

7. **正常系: force オプション付き**

```go
func TestRemoveCommand_Run_Force(t *testing.T) {
    // force=true で削除するケース
}
```

## 動作確認

```bash
go test -v ./...
go test -tags=integration -v ./...
```
