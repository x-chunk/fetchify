# fetchify

A fetch tool for Linux written in Go: it prints facts about the machine
in a framed card next to the logo of the distribution. Module path:
`go.xchunk.org/fetchify`.

Read `plan.md` for the design of version 2.0 and its decision log, and
`goals.md` for deferred work and deliberate non-goals, before changing
the design.

## Architecture

```
main.go            flags -> config -> collect -> ui -> one Composer
internal/cli/      flag parsing, help text, version
internal/config/   Config, defaults, default.toml, TOML loading and validation
internal/sysinfo/  facts about the machine as plain data, collected in parallel
internal/distro/   os-release parsing, the distribution table, logos
internal/ui/       the screen: theme, items, rows, card, logo, responsive layout
internal/golden/   golden-file test helper
```

Pipeline:

```
cli.Parse -> config.Load -> sysinfo.Collect(ui.Facts(cfg), ui.Options(cfg))
          -> ui.Screen(Input) -> beautify.Composer.Render
```

Imports:

- `sysinfo`, `distro`, `golden` import nothing from this module. `config`
  imports only `sysinfo` (for defaults). `cli` imports `config`. `ui`
  imports `sysinfo`, `distro`, `config` and beautify. `main` imports
  everything.
- Only `ui`, `main` and `distro` (for colors) import beautify.

Design rules:

- Never break the shell startup. A missing tool, file or sensor is never
  an error: the fact is `Absent` (does not exist here) or `Unknown`
  (could not be determined). Only invalid configuration and invalid flags
  stop the program, with exit code 2.
- `sysinfo` returns raw data only: sizes in bytes, durations, unformatted
  names. Collectors read the machine only through `sysinfo.Source`, honor
  their context and never panic on bad input; `Collect` still recovers a
  panic and gives every fact its own time limit.
- Code that parses files or command output is a pure function with
  table-driven tests. Collectors are tested against the fixture machines
  in `internal/sysinfo/testdata`, never against the machine running the
  tests, except for one smoke test.
- `ui` is a pure function of `ui.Input`. It collects, measures and writes
  nothing. Styles are named by role in `ui/theme.go`; no part of the
  screen uses a color directly.
- Presentation is built from beautify blocks and `dra` operations, with
  models made by the factory of the render context. No code produces
  escape sequences except the screen clear in `main`. If the look needs
  something beautify cannot do, add it to beautify.
- beautify keeps line breaks in text as layout, so `ui` reduces collected
  strings to one line before it shows them.
- Only `main` writes to the output.
- The screen is pinned by golden files in `internal/ui/testdata`.
  Regenerate them with `go test ./internal/ui -update` and review the
  diff by hand before committing.

## Adding an item

1. If it needs new data, add a fact to `sysinfo`: a `Fact` constant, a
   field of `Info`, a collector registered in `collectors`, fixtures and
   tests. A new facility of the machine becomes a field of `Source`.
2. Register the item in `ui/items.go` (or `ui/optin.go` for opt-in items)
   with its label, the facts it needs and a `rows` function. Return no
   rows for `Absent` and `unknownRow` for `Unknown`.
3. Add its icon to `ui/icons.go` with the Nerd Font glyph name.
4. A default item also goes into a group of `config.Defaults()` and
   `config/default.toml`; a test keeps the two equal.
5. Document it in `README.md`.

## Adding a distribution

1. Add an entry to one of the `distro/table_*.go` files: ID, name,
   palette, accent and the Nerd Font icon, plus aliases of other IDs.
2. Add `logos/<id>.txt` (at most 40x20) and `logos/<id>_small.txt` (at
   most 16x8). `${1}` to `${9}` select palette colors; `$$` is a dollar
   sign. No tabs and no trailing spaces; the tests enforce the format.
3. Record the source of both files in `logos/NOTICE`.
4. Add a real os-release file to `distro/testdata/os-release` and the
   expected entry to `TestLookupReleaseFiles`.

## Working in this repository

- Format with `gofmt`. Run `go vet ./...` and `go test ./...` before every
  commit.
- Install and run with `go install .` or `go run .` from the repository
  root; `go install go.xchunk.org/fetchify/v2@latest` is the only
  supported installation.

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
