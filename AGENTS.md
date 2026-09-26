# Agent Guidelines for `trump-center`

This document defines core principles, architectural invariants, and non-negotiable safety rules for AI agents working on the `trump-center` repository.

______________________________________________________________________

## 1. Project Overview & Architecture

`trump-center` is a playing-card (trump) game written in Go on top of [Ebitengine](https://ebitengine.org/) (`github.com/hajimehoshi/ebiten/v2`), with `github.com/ebitenui/ebitenui` as a UI dependency.

- **Development Environment**: Managed with Nix (`flake.nix`, with `default.nix` / `shell.nix` via `flake-compat`), `direnv` (`.envrc`), and `treefmt-nix` for formatting Nix, Go, GitHub Actions, Markdown, and shell scripts. `go`, `gopls`, `gomod2nix`, `nushell`, and `treefmt` are available on `PATH` inside `nix develop` (or via direnv). On Linux, the devShell also provides the X11 / ALSA / libGL libraries Ebitengine needs.
- **Target Language Version**: Go `1.26.x` (`go` directive in `go.mod`), provided by `pkgs.go` in `flake.nix`.
- **Nix Packaging**: The package is built with `gomod2nix` (`pkgs.buildGoApplication`). `gomod2nix.toml` holds the module hashes and **must** be regenerated whenever `go.mod` / `go.sum` change.
- **Directory Structure**:
  - `cmd/trump-center/`: Executable entry point (`main` calls `game.Run()`).
  - `internal/game/`: Root `ebiten.Game` implementation (`Game`, `Run()`, `NewGame()`) and window constants (`constants.go`). `Game` owns the active scene and delegates `Update`/`Draw` to it.
  - `internal/game/scenes/<name>/`: Concrete scenes (currently `title`).
  - `internal/scene/`: The `Scene` interface. `Update()` returns the next `Scene`, or `nil` to stay on the current one.
  - `internal/ui/`: Lightweight UI widgets (e.g. `Button`).
  - `assets/`: Embedded game assets (`assets.go`): card images under `assets/cards/` and the DotGothic16 font (OFL) under `assets/fonts/`.
  - `scripts/*.nu`: Nushell scripts invoked by the `Makefile` (`build`, `clean`, `test`, `fmt`, `lint`, `update`). They are written in Nushell for cross-platform (including Windows) support.
  - `Makefile`: Entry points: `make build` (outputs to `target/bin/`), `make run`, `make test`, `make fmt`, `make lint`, `make clean`, `make update`.

______________________________________________________________________

## 2. Strict Safety & Operational Rules (Always Enforced)

- **NO AUTONOMOUS CHANGES TO MAIN**: AI agents **MUST NEVER** merge PRs, execute `git merge` into `main`, or commit / push to `main` on their own initiative. Work on topic branches by default. Committing or pushing to `main` directly is allowed only when the user explicitly instructs it for that specific change.
- **NEVER PROPOSE COMMITS OR PUSHES UNPROMPTED**: AI agents **MUST NEVER** prompt the user to commit or push, nor propose commit messages unprompted. When instructed by the user or when creating/updating pull requests on topic branches, agents may execute `git commit` and `git push` directly without seeking confirmation.
- **Mandatory Human Approval**: AI agents may create branches, create commits, push topic branches, propose PRs, format code, and run test suites, but the final action of merging changes into `main` rests strictly with the human maintainer.
- **Verification Before Submitting**: All changes must pass `treefmt --fail-on-change`, `make build`, `make test`, and `make lint`, run inside `nix develop` (or via direnv).
- **Conventional Commits**: Use conventional commit prefixes (`feat:`, `fix:`, `refactor:`, `docs:`, `build:`, `test:`), optionally with a scope (e.g. `build(flake.lock):`).
- **Evidence First**: Base all answers and actions on actual file contents and command output. Never speculate or assume.
- **Non-Destructive**: Never perform irreversible actions (file deletions, hard resets, remote push, removal or replacement of tracked assets) without explicit user approval.
- **Targeted Edits**: Make minimal, logical changes strictly necessary for the request. Do not modify unrelated files.
- **English-Only Documentation**: All repository documentation, agent skills, code comments, commit messages, and PR descriptions must be written strictly in English.
- **Explicit Milestone Assignment Only**: AI agents **MUST NEVER** automatically attach or set GitHub Milestones on Pull Requests or Issues unless explicitly requested or instructed by the user.

______________________________________________________________________

## 3. Status Assessment Workflow

When asked to check status, assess the situation, or understand workspace context:

1. **Local Git State**: Inspect working tree (`git status -s -b`) and recent commits (`git log -n 5 --oneline`).
1. **GitHub PRs (always display)**: List **all** open PRs (`gh pr list`) and check the current branch's PR (`gh pr status`). Never skip this step, even when the local state is clean.
1. **GitHub Issues (always display)**: List **all** open issues (`gh issue list`). Never skip this step.
1. **Environment Health**: Verify that `nix develop` evaluates, then check build and test status inside it (`treefmt --fail-on-change`, `make build`, `make test`, `make lint`).
1. **Synthesis**: Report a concise, structured status covering local state, remote GitHub state, and environment health. The report **must** include the open PR and Issue lists (number, title, and state), or explicitly state that there are none.

______________________________________________________________________

## 4. Workspace Skills

Detailed runbooks and procedural workflows are maintained as workspace skills under `.agents/skills/`:

| Trigger / Context | Skill to Read | Purpose |
| :--- | :--- | :--- |
| Deep investigation, complex code search | [`investigate`](.agents/skills/investigate/SKILL.md) | Non-destructive investigation guidelines |
| Commit conventions & policies | [`git-commit`](.agents/skills/git-commit/SKILL.md) | Commit conventions and prohibition of unprompted commit/push proposals |
| Deleting files, overwriting, git push/reset | [`irreversible`](.agents/skills/irreversible/SKILL.md) | Pre-checks and confirmation prompts |
| Testing, verifying builds or behavior | [`verify`](.agents/skills/verify/SKILL.md) | Minimal, high-signal verification steps |
| Bumping `flake.lock`, Go modules, or the Go version | [`update-dependencies`](.agents/skills/update-dependencies/SKILL.md) | Procedures for Nix input updates, Go module updates with `gomod2nix`, and Go version bumps |
| Preparing PRs, formatting, pre-submission checks | [`pr-workflow`](.agents/skills/pr-workflow/SKILL.md) | Verification command table, commit rules, and PR requirements |
| "Fresh eyes" sweep for issues not already tracked, sanity-checking a batch of fixes | [`fresh-eyes-audit`](.agents/skills/fresh-eyes-audit/SKILL.md) | Parallel, context-free repo audits to surface gaps a single continuously-informed reviewer would miss |
