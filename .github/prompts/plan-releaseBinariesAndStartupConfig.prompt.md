# Plan: GitHub Executables + Startup Config

Ship prebuilt binaries via GitHub Releases, and replace env-var-only configuration with a "first run" prompt that saves settings to a user config file. This keeps power-user env vars working, but makes "download + run" viable for non-technical users.

## Steps

1. **Add a release pipeline** that builds `genrify` for macOS/Windows/Linux and uploads artifacts to GitHub Releases (Option A: add GoReleaser; Option B: add a matrix build workflow).

2. **Extend `internal/config`** to load/save an app config file (reuse the existing per-user app dir used by `internal/auth/store.go`) and keep env vars as override.

3. **Implement a "first run / missing config" prompt path** via `internal/cli`'s existing `Prompter`, hooked into `internal/cli/root.go`'s `PersistentPreRunE`.

4. **Validate prompted config**: require Client ID, default redirect/scopes, and require TLS cert/key only when redirect is `https://...`.

5. **Update docs in `README.md`** to emphasize "download binary → run → guided setup", plus keep advanced env-var instructions.

6. **Update/add tests** around config precedence and prompt-to-persist behavior (`internal/config/config_test.go`, `internal/cli/root_test.go`).

## Further Considerations

1. **Release tooling choice**: GoReleaser (less YAML, standard) vs pure GitHub Actions (no new tool).

2. **Config location/format**: `~/.config/genrify/config.json` (or `.yaml`) alongside the existing token store directory.

3. **Security**: ensure `.certs/` and any local secrets are gitignored and not included in release assets.

## Current State Context

### Build/CI
- Local run: `go run ./cmd/genrify`; tests via `go test ./...`
- CI: `.github/workflows/ci.yml` runs `golangci-lint` + tests, uploads coverage artifact, optional Codecov
- Security scan: `.github/workflows/security.yml` (weekly + on PR/push)
- No release/distribution tooling: no `.goreleaser.*`, no `Makefile`, no release workflow
- Versioning designed for `-ldflags`: `internal/buildinfo/buildinfo.go` (defaults to `"dev"`)

### Environment Variables
- `SPOTIFY_CLIENT_ID` (required): read in `internal/config/config.go`
- `SPOTIFY_REDIRECT_URI` (default `http://localhost:8888/callback`): `internal/config/config.go`
- `SPOTIFY_SCOPES` (default playlist read/write scopes): `internal/config/config.go`
- `SPOTIFY_TLS_CERT_FILE`, `SPOTIFY_TLS_KEY_FILE` (only required when redirect is `https://...`): read in `internal/config/config.go`; enforced in `internal/cli/login.go`

### Startup/Login Flow
- Entrypoint: `cmd/genrify/main.go` → `internal/cli/root.go` → `cobra.Command.Execute()`
- Central startup hook: `internal/cli/root.go` sets `PersistentPreRunE` to call `config.NewFromEnv()` for *all commands* except `version`
- Login flow: `internal/cli/login.go` builds `auth.Config` from `config.Config` and calls `auth.StartOAuth`; token saved to store
- **Best insertion point**: replace/extend the `PersistentPreRunE` in `internal/cli/root.go` to load config from a file, and if missing/invalid, prompt then write config before continuing

### Existing Persistence
- Tokens persisted: `internal/auth/store.go` uses `os.UserConfigDir()` + `genrify/tokens.json` (mode `0600`, dir `0700`)
- App name: `internal/cli/constants.go` = `"genrify"`
- No persisted "app config" (client id/redirect/scopes) exists today; `internal/config/config.go` is env-only

### Tests Needing Updates
- `internal/config/config_test.go`: currently asserts env-var behavior; needs refactor to cover file-based config load/save + defaults + validation
- `internal/cli/root_test.go`: currently skips the "PersistentPreRunE loads config" test; once config loading becomes injectable, can be un-skipped
- Add tests around "first run prompts then persists" using mock `Prompter` (interface exists in `internal/cli/interfaces.go`)

## Implementation Options

### Option A: GoReleaser
**Pros:**
- Standard multi-OS builds, archives/checksums/changelog
- Easy `-ldflags` version injection
- Straightforward GitHub Releases publishing
- Industry standard tooling

**Cons:**
- Adds `.goreleaser.yml` + extra tool/dependency
- Requires some upfront config decisions (archive naming, snapshot builds, signing/not)

### Option B: Pure GitHub Actions
**Pros:**
- No extra tool; full control
- Minimal for single-binary repos
- Can be simpler for small projects

**Cons:**
- More YAML/maintenance
- Reimplement packaging/checksums/changelog conventions

### Option C: Source Install Only
**Pros:**
- Simplest: `go install github.com/EysteinKS/genrify-cli/cmd/genrify@<tag>`
- No CI release required

**Cons:**
- Not "prebuilt executables"
- Requires Go toolchain
- Less friendly for non-Go users (doesn't meet requirement)
