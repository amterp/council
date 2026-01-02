# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## What is Council?

Council is a CLI tool for running collaborative sessions between multiple participants—LLMs, humans, scripts, or anything that can execute shell commands. Sessions are stored as JSONL files; the CLI is stateless.

## Build & Test Commands

```bash
# Build
go build -o council ./cmd/council

# Run all tests
go test ./...

# Run e2e tests only (requires built binary)
go test ./e2e/...

# Dev script (requires rad: https://github.com/amterp/rad)
./dev -b        # build
./dev -v        # build + test
./dev -p        # build + test + push
./dev -r patch  # release (patch|minor|major)
```

## Architecture

```
cmd/council/          Entry point - calls cli.Run()
internal/
├── cli/              Command handlers (new, join, leave, status, post)
│   ├── root.go       CLI setup using github.com/amterp/ra (arg parser)
│   └── skill.md      Embedded as --help content for LLM participants
├── session/          Core domain logic
│   ├── session.go    Session struct, file locking, CRUD operations
│   ├── events.go     Event types (session_created, joined, left, message)
│   ├── format.go     Status output formatting
│   └── validation.go Reserved name checks
├── storage/          File path helpers (~/.council/sessions/<id>/events.jsonl)
└── errors/           Typed errors with actionable messages
e2e/                  Integration tests (shell out to built binary)
```

## Key Concepts

- **Session files**: `~/.council/sessions/<session-id>/events.jsonl` - each line is a JSON event
- **Optimistic locking**: `--after N` flag prevents posting on stale state; CLI checks event count matches before writing
- **File locking**: All writes acquire exclusive `syscall.Flock` before modifying session files
- **Reserved name**: "Moderator" is reserved for human operators watching via `council watch`
- **Turn coordination**: `--next` flag designates next speaker; defaults to previous speaker or random active participant

## Dependencies

- `github.com/amterp/ra` - argument parsing library
- `github.com/dustinkirkland/golang-petname` - session ID generation (e.g., "hopeful-coral-tiger")

## Testing Notes

E2E tests in `e2e/council_test.go` shell out to the compiled `./council` binary. Build before running e2e tests.

## Contributing

When making changes, update relevant documentation including this file (AGENTS.md), README.md, SPEC.md, and SKILL.md as appropriate.
