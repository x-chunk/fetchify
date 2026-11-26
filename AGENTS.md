# fetchify

A fetch tool for Linux written in Go: it prints facts about the machine
next to an ASCII-art logo. Module path: `go.xchunk.org/fetchify`.

Read `plan.md` for the plan of the next version and `goals.md` for
deferred work and deliberate non-goals before changing the design.

## Architecture

The program is two layers under a thin `main`, built on the
`go.xchunk.org/beautify` library. Imports only go downward, and `sysinfo`
does not import beautify.

```
main.go            flags, wiring, the only code that writes output
internal/ui/       sysinfo.Info and config -> beautify.Composable: content, styles
internal/sysinfo/  collects machine facts as plain data
internal/golden/   golden-file test helper
```

Pipeline:

```
sysinfo.Collect() -> ui.Screen(input) -> beautify.Composer.Render -> one Write
```

Design rules:

- `sysinfo` returns data only: no formatting, no colors, no fallback
  strings. A fact that cannot be determined is a zero value. It never
  fails and never panics when a tool or a file is missing.
- Code that parses the output of an external command is a pure function
  with tests. The code that runs the command takes a `runner` so tests can
  replace it.
- `ui` is a pure function of `ui.Input`. It collects nothing and writes
  nothing. What is shown and in which order is the configuration; a new
  kind of row is a new item in the registry of `ui/items.go`.
- Styles are named by role in `ui/theme.go`. Do not use a color directly
  in a section.
- Presentation is built from beautify blocks and `dra` operations. Models
  are made with the factory of the render context, never with a factory
  of their own. No code in fetchify produces escape sequences; if a look
  needs something beautify cannot do, it is added to beautify.
- beautify keeps line breaks in text as layout, so `ui` reduces collected
  strings to one line before it shows them.
- Only `main` writes to the output, through one `beautify.Composer`.
- Screen output is pinned by golden files in `internal/ui/testdata`.
  Regenerate them with `go test ./internal/ui -update` and review the
  diff by hand before committing.

## Working in this repository

- Format with `gofmt`. Run `go vet ./...` and `go test ./...` before every
  commit.
- `go build -o fetchify .` must keep working from the repository root:
  `install.sh` relies on it.

## Code comments and documentation

- Write all code comments in American English, in plain but technical language.
- Write comments the way an engineer writes them in a production codebase: short, direct and specific. No filler, no chatty tone, no narrating the obvious.
- Explain why, not what. Use comments for intent, constraints, invariants, edge cases, side effects and non-obvious trade-offs.
- Code must be well documented by default. Every exported or public package, type, function, method, constant and variable gets a doc comment that follows the language's convention (for example, in Go the comment starts with the identifier name).
- In doc comments, describe parameters, return values, returned errors and, where relevant, concurrency, ownership and lifetime expectations.
- Keep comments in sync with the code. Update or delete a stale comment in the same change that makes it stale.
- Do not leave commented-out code. Do not leave a TODO without context on what is missing and why.

## Commits

- Make every change as atomic commits: one logical change per commit.
- Each commit must leave the project in a working state (it builds and tests pass).
- Do not mix unrelated changes in one commit. Keep features, bug fixes, refactors, formatting and dependency updates in separate commits.
- Write commit messages in American English, imperative mood, with a concise subject line (72 characters or fewer). Add a body only when the reason for the change is not obvious from the diff.
