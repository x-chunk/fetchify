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
