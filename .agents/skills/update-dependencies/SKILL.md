# Update Workflow: Nix Inputs, Go Modules, and Go Version

`trump-center` is packaged with `gomod2nix`, so Go module changes require regenerating `gomod2nix.toml`. The following areas require a coordinated update. Run everything inside `nix develop` (or via direnv).

## Nix inputs (`flake.lock`)

1. Run `nix flake update`.
1. Confirm `nix develop` still evaluates. nixpkgs-unstable may rename or deprecate attributes (e.g. `stdenv.isLinux` → `stdenv.hostPlatform.isLinux`); fix `flake.nix` in the same change if needed.
1. Run `treefmt --fail-on-change`, `make build`, `make test`, `make lint`, and `nix flake check`.
1. Commit with `build(flake.lock): nix flake update`.

## Go modules (`go.mod`, `go.sum`, `gomod2nix.toml`)

1. Change the dependency (e.g. `go get github.com/hajimehoshi/ebiten/v2@<version>`).
1. Run `make update`. It runs `go mod tidy` and then `gomod2nix` to regenerate `gomod2nix.toml`.
1. Commit `go.mod`, `go.sum`, and `gomod2nix.toml` together. A stale `gomod2nix.toml` breaks the Nix build even when `go build` succeeds.
1. Run `make build`, `make test`, `make lint`, and `nix flake check`.

## Go version bumps

When changing the Go version, update all of the following together:

- The `go` directive in `go.mod`
- `pkgs.go` in `flake.nix` if a pinned version (e.g. `pkgs.go_1_XX`) is required, and `flake.lock` if nixpkgs must be updated to provide it
- The Go version noted in `AGENTS.md`
- Any `go-version` in `.github/workflows/`, if workflows exist

Then run `make update` and the full verification set above.

Adding, removing, or upgrading dependencies is an irreversible-adjacent operation (see `.agents/skills/irreversible/SKILL.md`). Confirm with the user before changing dependencies they did not ask for.
