# Repository Guidelines

## Project Structure & Module Organization
- Interpreter source centers around `compiler/`, `vm/`, and `runtime/`, with shared utilities in `pkg/`.
- Executables live under `cmd/`: `cmd/hey` for the CLI, `cmd/php-parser` for introspection, and `cmd/vm-demo` for the minimal pipeline. Build outputs land in `build/`.
- Supporting assets include `docs/` (design notes), `examples/` (PHP samples), and `tests/` plus top-level `test_*.php` scripts for regression fixtures. Docker files in `docker/` reproduce WordPress scenarios.

## Build, Test, and Development Commands
- `make build` produces the release-ready binary; `make dev` is the fast edit-compile loop.
- `make test` runs all Go tests, while `make test-parser`, `make test-lexer`, and `make test-vm` scope feedback to specific subsystems.
- `make fmt` applies gofmt and `make lint` calls golangci-lint; run `make deps` whenever module references change.
- Demos: `go run ./cmd/vm-demo` for an end-to-end trace, or `go build -o build/php-parser ./cmd/php-parser` to inspect tokens and ASTs.

## Coding Style & Naming Conventions
- Keep Go files gofmt-clean (tabs, trailing newline). Run `make fmt` or the editor equivalent before review.
- Exported APIs use `CamelCase`, internal helpers use `lowerCamelCase`, and new test files follow the `_test.go` suffix. Prefer descriptive snake_case names for PHP fixtures placed in `tests/`.
- Document new packages and exported types with brief GoDoc comments, especially when mirroring PHP behavior.

## Testing Guidelines
- Run `go test ./...` before pushing; table-driven tests following `TestXxx` naming keep coverage approachable.
- Use `make test-coverage` on VM or compiler changes and chase regressions with the focused `make test-*` targets.
- When a PHP script reproduces a bug, commit the minimal case under `tests/` and reference it from accompanying Go tests to keep failures actionable.

## Commit & Pull Request Guidelines
- Follow the Conventional Commits pattern seen in history (`feat:`, `fix:`, `docs:`, etc.) and keep messages imperative and scoped.
- PRs should summarize the PHP behavior or subsystem touched, list the commands you ran (at least `make fmt` and `make test`), and link any tracking issue. Attach CLI excerpts or PHP snippets when behavior changes are observable.
- Confirm lint, format, and test targets locally; call out exceptions explicitly in the PR description.
