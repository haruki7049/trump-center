# Pull Request & Commit Workflow for `trump-center`

This skill defines the procedures for code verification, commit creation, and pull request submission.

## 1. Mandatory Verification Steps

Run all commands inside `nix develop` (or via direnv). Before committing or opening a PR, execute the following commands and ensure all pass cleanly:

| Task | Command | Description |
| :--- | :--- | :--- |
| **Check All Formatting (treefmt)** | `treefmt --fail-on-change` | Verifies formatting across Nix, Go, GitHub Actions, Markdown, and shell files |
| **Format All Files (treefmt)** | `make fmt` | Auto-formats all files in the repository using treefmt |
| **Build** | `make build` | Builds `cmd/trump-center` into `target/bin/` |
| **Run All Tests** | `make test` | Runs `go test ./...` |
| **Lint** | `make lint` | Runs `go vet ./...` |
| **Nix Build / Check** | `nix flake check` | Builds the `gomod2nix` package; run when `flake.nix`, `flake.lock`, `go.mod`, `go.sum`, or `gomod2nix.toml` change |

If a change affects rendering or input handling, also launch the game with `make run` and confirm the behavior when a display is available. State explicitly if this was not possible.

## 2. Commit & PR Title Conventions

Use Conventional Commits style prefixes, optionally with a scope:

- `feat:` New gameplay, scenes, UI, or assets.
- `fix:` Bug fixes.
- `build:` Updates to `flake.nix`, `flake.lock`, `go.mod`, `gomod2nix.toml`, `Makefile`, `scripts/`, or CI workflows.
- `refactor:` Code restructuring without changing behavior.
- `docs:` Updates to README, AGENTS.md, skills, or code documentation.
- `test:` Adding or updating tests.

**Do NOT include issue numbers (e.g., `(#24)` or `#24`) in commit messages or PR titles.** Issue linkage must be done exclusively in the PR description using explicit issue-closing keywords (e.g. `Closes #24`).

**Language**: Write all commit messages, PR titles, PR descriptions, and repository documentation strictly in English.

## 3. PR Description Requirements

Ensure the PR description includes:

- **Summary**: Concise overview of changes.
- **Linked Issue / Closes Statement**: Always include an explicit issue-closing keyword (e.g. `Closes #16`, `Fixes #12`, or `Resolves #5`) when resolving an open issue.
- **Verification**: Explicitly list executed verification commands (`treefmt --fail-on-change`, `make test`, etc.) and their success status.
- **Breaking Changes**: Highlight any breaking changes (e.g. to build or run procedures).

## 4. Strict Safety & Approval Rules

- **NO AUTONOMOUS CHANGES TO MAIN**: AI agents **MUST NEVER** merge PRs, execute `git merge` into `main`, or commit / push to `main` on their own initiative. Direct commits to `main` are allowed only when the user explicitly instructs it for that specific change.
- **NEVER PROPOSE COMMITS OR PUSHES UNPROMPTED**: AI agents **MUST NEVER** prompt the user to commit or push unprompted. When instructed by the user or when preparing pull requests on topic branches, agents may execute `git commit` and `git push` directly.
- **Mandatory Human Approval**: AI agents may create branches, create commits, push topic branches, propose PRs, format code, and run test suites, but the final action of merging changes into `main` rests strictly with the human maintainer.
- **Explicit Milestone Assignment Only**: AI agents **MUST NEVER** automatically attach or set GitHub Milestones on Pull Requests or Issues unless explicitly requested or instructed by the user.
