# Genrify Code Style Guide

This document defines the coding conventions and style guidelines for the `genrify` project. All contributors should follow these rules to maintain consistency and quality.

---

## General Principles

1. **Clarity over cleverness**: Write clear, idiomatic Go code that's easy to understand
2. **Testability**: Design code to be testable; use interfaces for dependencies
3. **Error handling**: Always wrap errors with context; use sentinel errors for detectable conditions
4. **Documentation**: Document all exported types, functions, and packages with godoc comments
5. **Consistency**: Follow existing patterns in the codebase

---

## Go Formatting

- Use `gofmt` or `goimports` to format all code
- Maximum line length: **120 characters** (soft guideline)
- Use tabs for indentation (Go standard)
- Group imports into standard library, third-party, and local packages:
  ```go
  import (
      "context"
      "fmt"
      
      "github.com/spf13/cobra"
      
      "genrify/internal/auth"
      "genrify/internal/spotify"
  )
  ```

---

## Package Structure

### Package Naming
- Use short, lowercase, single-word package names
- Avoid generic names like `util`, `common`, `helpers` as package names
- Package name should describe what the package provides (e.g., `auth`, `spotify`, `cli`)

### Internal Packages
```
internal/
├── auth/          # OAuth, token management, storage
├── buildinfo/     # App name, version, user-agent
├── cli/           # Cobra commands, interactive mode, CLI helpers
├── config/        # Configuration loading from environment
├── playlist/      # (Future) Playlist service layer
├── spotify/       # Spotify API client
└── testutil/      # (Future) Shared test utilities
```

### File Organization
- Group related functionality in the same file
- Split large files when they exceed ~300-400 lines
- Test files go in the same package: `client.go` → `client_test.go`
- Common file names:
  - `types.go` - type definitions
  - `errors.go` - error types and helpers
  - `interfaces.go` - interface definitions
  - `constants.go` - package constants

---

## Naming Conventions

### Variables
- Use **camelCase** for local variables and parameters
- Use **short names** for limited scope (e.g., `c` for client in a 5-line function)
- Use **descriptive names** for broader scope (e.g., `playlistID`, `userAgent`)
- Avoid single-letter names except for: `i`, `j` (loop indices), `w` (io.Writer), `r` (io.Reader/http.Request), `err`

### Functions
- Use **camelCase** for unexported functions
- Use **PascalCase** for exported functions
- Start function names with verbs: `Get`, `List`, `Create`, `Delete`, `Normalize`, `Format`
- Constructor functions: `New`, `NewWithOptions`, `NewFoo`
- Factory functions: `newSpotifyClient`, `newLoginCmd`

### Types
- Use **PascalCase** for all type names
- Avoid stuttering: `spotify.Client` not `spotify.SpotifyClient`
- Interface names should describe behavior:
  - Single-method interfaces: `Runner`, `Closer`, `Store`
  - Multi-method interfaces: `SpotifyClient`, `Prompter`, `TokenManager`

### Constants
- Use **PascalCase** for exported constants
- Use **camelCase** for unexported constants
- Group related constants:
  ```go
  const (
      DefaultPlaylistLimit = 50
      DefaultTrackLimit    = 100
  )
  ```

---

## Error Handling

### Error Creation
- Always wrap errors with context using `fmt.Errorf("operation: %w", err)`
- Use `%w` verb for error wrapping (Go 1.13+)
- Add operation context at each layer:
  ```go
  if err := c.ListPlaylists(ctx); err != nil {
      return fmt.Errorf("list playlists: %w", err)
  }
  ```

### Sentinel Errors
- Define sentinel errors in `internal/cli/errors.go` for detectable conditions
- Use `errors.Is()` to check for sentinel errors
- Example:
  ```go
  var ErrNotLoggedIn = errors.New("not logged in; run genrify login")
  
  if errors.Is(err, ErrNotLoggedIn) {
      // Handle not-logged-in case
  }
  ```

### Custom Error Types
- Use custom error types when you need structured error data:
  ```go
  type APIError struct {
      Status  int
      Message string
  }
  
  func (e APIError) Error() string {
      return fmt.Sprintf("spotify api error: http %d: %s", e.Status, e.Message)
  }
  ```

### Error Checking
- Check errors immediately; don't defer error handling
- Don't ignore errors with `_` unless there's a good reason (document it)
- Return errors to the caller; avoid logging in libraries (log in main/CLI only)

---

## Interfaces and Testability

### Interface Design
- Keep interfaces small (1-5 methods)
- Define interfaces in the **consumer** package, not the provider
- Accept interfaces, return structs
- Example:
  ```go
  // In internal/cli/interfaces.go (consumer)
  type SpotifyClient interface {
      GetMe(ctx context.Context) (spotify.User, error)
      ListCurrentUserPlaylists(ctx context.Context, max int) ([]spotify.SimplifiedPlaylist, error)
  }
  
  // In internal/spotify/client.go (provider)
  type Client struct { /* ... */ }
  func (c *Client) GetMe(ctx context.Context) (spotify.User, error) { /* ... */ }
  ```

### Dependency Injection
- Pass dependencies as function parameters or struct fields
- Avoid global state (except for build-time constants)
- Use constructor functions to wire dependencies:
  ```go
  func newSpotifyClient(cfg config.Config) (*spotify.Client, error) {
      store := auth.NewStore(cfg.TokenCacheAppKey)
      // ...
  }
  ```

---

## Testing

### Test File Organization
- Test files live in the same package: `client_test.go`
- Use table-driven tests for multiple cases:
  ```go
  func TestNormalizeTrackURI(t *testing.T) {
      tests := []struct {
          name    string
          input   string
          want    string
          wantErr bool
      }{
          {"spotify URI", "spotify:track:abc123", "spotify:track:abc123", false},
          {"open.spotify.com URL", "https://open.spotify.com/track/abc123", "spotify:track:abc123", false},
          {"empty input", "", "", true},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              got, err := normalizeTrackURI(tt.input)
              if (err != nil) != tt.wantErr {
                  t.Errorf("normalizeTrackURI() error = %v, wantErr %v", err, tt.wantErr)
                  return
              }
              if got != tt.want {
                  t.Errorf("normalizeTrackURI() = %v, want %v", got, tt.want)
              }
          })
      }
  }
  ```

### Test Helpers
- Use `t.Helper()` in test helper functions
- Create shared mocks/fakes in `internal/testutil/`
- Use `httptest.NewServer()` for testing HTTP clients

### Coverage
- Aim for **80%+ coverage** on new code
- Focus on testing happy paths and common error cases
- Don't test trivial getters/setters

---

## Constants and Magic Numbers

### Extract to Named Constants
- Avoid magic numbers; use named constants in `constants.go`:
  ```go
  const (
      DefaultPlaylistLimit = 50
      DefaultHTTPTimeout   = 30 * time.Second
  )
  ```

### Package-Level vs Local Constants
- Package-level constants for values used across files
- Local `const` blocks for values used in a single file/function

---

## Version and Build Info

### Centralized Build Info
- All version/app metadata lives in `internal/buildinfo/buildinfo.go`
- Override `Version` at build time:
  ```bash
  go build -ldflags "-X genrify/internal/buildinfo.Version=1.0.0" ./cmd/genrify
  ```

### User-Agent
- Use `buildinfo.UserAgent` for all HTTP requests
- Format: `genrify/version`

---

## CLI Conventions

### Command Structure
- Use Cobra for CLI commands
- Group related commands under subcommands: `genrify playlists list`
- Keep `RunE` functions short; extract logic to helper functions

### Flags
- Use long flags with defaults: `--filter`, `--limit`
- Provide short descriptions in flag help text
- Mark required flags: `cmd.MarkFlagRequired("name")`

### Output
- Use `cmd.Println()` and `cmd.Printf()` instead of `fmt.Println()` for testability
- Write to `cmd.OutOrStdout()` for table/list output
- Write errors to `cmd.ErrOrStderr()`
- Format output consistently:
  - Tab-separated columns for tables
  - Human-readable messages for interactive mode

### Interactive Mode
- Use `Prompter` interface for all user input
- Default implementation: `promptuiPrompter` (wraps `promptui`)
- Handle cancellation (Ctrl+C) gracefully with `ErrCancelled`

---

## Comments and Documentation

### Package Documentation
- Every package should have a package comment:
  ```go
  // Package spotify provides a client for the Spotify Web API.
  //
  // It handles authentication, token refresh, and API requests.
  package spotify
  ```

### Function Documentation
- Document all exported functions with a complete sentence:
  ```go
  // ListCurrentUserPlaylists fetches playlists for the authenticated user.
  // If max is 0, it fetches all playlists with automatic pagination.
  func (c *Client) ListCurrentUserPlaylists(ctx context.Context, max int) ([]SimplifiedPlaylist, error) {
  ```

### Inline Comments
- Use inline comments to explain **why**, not **what**
- Avoid obvious comments:
  ```go
  // Bad: Set the limit to 50
  limit := 50
  
  // Good: Fetch all playlists for accurate filtering
  fetchMax := 0
  ```

---

## Git and Commits

### Commit Messages
- Use present tense: "Add feature" not "Added feature"
- Keep first line under 72 characters
- Add detail in commit body if needed:
  ```
  Add Prompter interface for testable interactive mode
  
  - Define Prompter interface in internal/cli/interfaces.go
  - Implement promptuiPrompter using promptui library
  - Refactor interactive commands to use Prompter
  ```

### Branch Naming
- Feature branches: `feature/playlist-merge`
- Bug fixes: `fix/handle-empty-playlists`
- Refactoring: `refactor/extract-helpers`

---

## Performance Considerations

### Context Usage
- Always pass `context.Context` as the first parameter
- Respect context cancellation in loops and long-running operations
- Use `context.WithTimeout()` for operations with deadlines

### Memory Allocation
- Preallocate slices when size is known: `make([]string, 0, capacity)`
- Avoid unnecessary allocations in hot paths
- Use `strings.Builder` for string concatenation in loops

### API Pagination
- Batch API requests (e.g., add tracks in batches of 100)
- Respect rate limits (handle 429 responses with retry/backoff)
- Fetch only what's needed (use `max` parameter)

---

## Future Enhancements

### TODO Comments
- Use `// TODO(username): description` for future work
- Link to GitHub issues when appropriate: `// TODO: see #123`

### Deprecation
- Mark deprecated code with `// Deprecated: use X instead`
- Keep deprecated code for at least one major version

---

## Tools and Linting

### Required Tools
- `go fmt` or `goimports` for formatting
- `go vet` for static analysis
- `golangci-lint` for comprehensive linting (recommended)

### Recommended Linters
- `errcheck` - check for unchecked errors
- `staticcheck` - advanced static analysis
- `gosec` - security checks
- `gocyclo` - cyclomatic complexity

### Pre-commit Checks
```bash
go fmt ./...
go vet ./...
go test ./...
```

---

## Examples

### Good Example: Error Handling
```go
func (c *Client) GetPlaylist(ctx context.Context, id string) (SimplifiedPlaylist, error) {
    id = strings.TrimSpace(id)
    if id == "" {
        return SimplifiedPlaylist{}, fmt.Errorf("playlist id is required")
    }
    
    var pl SimplifiedPlaylist
    if err := c.doJSON(ctx, http.MethodGet, "/playlists/"+id, nil, nil, &pl); err != nil {
        return SimplifiedPlaylist{}, fmt.Errorf("get playlist: %w", err)
    }
    return pl, nil
}
```

### Good Example: Table-Driven Test
```go
func TestFilterPlaylistsByName(t *testing.T) {
    playlists := []spotify.SimplifiedPlaylist{
        {ID: "1", Name: "Workout Mix"},
        {ID: "2", Name: "Chill Vibes"},
        {ID: "3", Name: "Workout Hits"},
    }
    
    tests := []struct {
        name   string
        filter string
        want   int
    }{
        {"no filter", "", 3},
        {"case insensitive", "workout", 2},
        {"no match", "jazz", 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := filterPlaylistsByName(playlists, tt.filter)
            if len(got) != tt.want {
                t.Errorf("got %d playlists, want %d", len(got), tt.want)
            }
        })
    }
}
```

---

## Summary

- **Format with `gofmt`**, follow Go conventions
- **Use interfaces** for testability and decoupling
- **Wrap errors** with context at every layer
- **Write tests** for all new code (table-driven when possible)
- **Document exports** with complete godoc comments
- **Extract constants** instead of magic numbers
- **Keep functions small** and focused on one task
- **Check this guide** before submitting changes

When in doubt, follow the existing patterns in the codebase or consult [Effective Go](https://go.dev/doc/effective_go).
