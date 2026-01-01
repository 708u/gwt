# Phase 2: FileSystem 拡張

## 対象ファイル

- `fs.go`
- `internal/testutil/mock_fs.go`

## 追加するメソッド

### Abs

パスを絶対パスに変換する。現在ディレクトリが削除対象 worktree 内かどうかの検証に使用。

```go
// FileSystem interface に追加
Abs(path string) (string, error)
```

### osFS への実装

```go
func (osFS) Abs(path string) (string, error) {
    return filepath.Abs(path)
}
```

### MockFS への実装

```go
// MockFS に追加
AbsFunc func(path string) (string, error)

func (m MockFS) Abs(path string) (string, error) {
    if m.AbsFunc != nil {
        return m.AbsFunc(path)
    }
    return filepath.Abs(path)
}
```

## 動作確認

```bash
go test -run TestFileSystem -v ./...
```
