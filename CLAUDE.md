# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gwt is a Go CLI tool that simplifies git worktree workflows by automating
related operations (branch creation, symlinks, etc.) in a single command.

## Project Structure

```txt
cmd/gwt/         # CLI entrypoint (uses cobra)
internal/testutil/  # Test mocks for FileSystem and GitExecutor
*.go (root)      # Core library: commands, config, abstractions
```

- `cmd/gwt`: CLI layer. Parses arguments and delegates to library.
- Root package (`gwt`): Business logic as reusable library.
  - Command structs (e.g., `AddCommand`) with injected dependencies
  - `Config`: Configuration loading from TOML files
  - Abstraction interfaces (`FileSystem`, `GitExecutor`) for testability
- `internal/testutil`: Mock implementations for unit testing

## Design Principles

- Flat package structure: avoid deep nesting, keep packages at root level
- Prefer lower implementation cost over performance optimization (aiming for minimal package)
- Keep dependencies minimal
- Add complexity only when necessary

## Common Commands

```bash
make build                        # Build binary to out/gwt
go test ./...                     # Run unit tests
go test -tags=integration ./...   # Run integration tests
```

## Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Error handling: return errors rather than panicking
- Naming: use camelCase for unexported, PascalCase for exported identifiers

## Git Commit

- Use Conventional Commits format
- Example: `feat: add new feature`, `fix: resolve bug`, `docs: update README`
