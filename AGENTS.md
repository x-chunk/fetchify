# fetchify

A fetch tool for Linux written in Go: it prints facts about the machine
next to an ASCII-art logo. Module path: `go.xchunk.org/fetchify`.

Read `plan.md` for the plan of the next version and `goals.md` for
deferred work and deliberate non-goals before changing the design.

## Architecture

The program is three layers under a thin `main`. Imports only go downward,
and `sysinfo` and `render` do not know about each other.

```
main.go            flags, wiring, the only code that writes output
internal/ui/       sysinfo.Info -> render.Block: content, wording, styles
internal/sysinfo/  collects machine facts as plain data
internal/render/   styled text, layout operations, the ANSI Engine
```

Pipeline:

```
sysinfo.Collect() -> ui.Build(info) -> render.Engine.Append -> one Write
```

Design rules:

- `sysinfo` returns data only: no formatting, no colors, no fallback
  strings. A fact that cannot be determined is a zero value. It never
  fails and never panics when a tool or a file is missing.
- Code that parses the output of an external command is a pure function
  with tests. The code that runs the command takes a `runner` so tests can
  replace it.
- `ui` is a pure function of `sysinfo.Info`. It collects nothing and
  writes nothing. What is shown and in which order is the `sections` slice
  in `ui.go`; a new line of output is a new `section`.
- Styles are named by role in `ui/theme.go`. Do not use a color directly
  in a section.
- `render` is the only package that produces escape sequences. Its values
  are immutable, its text is sanitized on creation, and its Engine is
  stateless and never reads the terminal or the environment.
- Only `main` writes to the output, with a single `Write` call.

## Migration to beautify

`internal/render` is a temporary stand-in for
`go.xchunk.org/beautify`, which is not implemented yet. It follows
the model of `beautify/dra` so that it can be deleted once the library
exists. Read `plan.md` in the beautify repository before changing it, and
do not grow it beyond what fetchify needs today.

Expected replacements:

| fetchify today | beautify |
|----------------|----------|
| `render.Span`, `render.Line`, `render.Block` | `dra.Model` |
| `render.Text`, `render.Styled` | `dra.Factory.Text`, `dra.Factory.Styled` |
| `render.Style`, `render.Color` | `dra.Style`, `dra.Color` |
| `render.Stack`, `render.Join`, `render.Repeat` | `dra.Stack`, `dra.Join`, `dra.Repeat` |
| `render.Engine`, `render.Profile` | `dra.Engine`, `dra.Profile` |
| the write in `main.run` | `beautify.Composer` |
| `ui.Build` | a `beautify.Template` |
| `ui.keyValue` and `entry` | `blocks.KeyValue` and `blocks.Pair` |
| `ui.header` rule, `ui.logo`, the logo and column row | `blocks.Rule`, `blocks.Text`, `blocks.Row` |

`sysinfo` is not affected by the migration. Keep it free of anything
related to presentation so that this stays true.

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
