# Genrify Development Plan

This document outlines the planned improvements for the genrify CLI tool across three stages: Refactoring, Testing, and New Feature implementation.

---

## Stage 1: Refactoring

**Goal:** Clean up the codebase for improved readability, maintainability, and separation of concerns.

### 1.1 Extract Shared Logic from CLI Commands

| Task | Description | Files |
|------|-------------|-------|
| 1.1.1 | Extract duplicate playlist listing/filtering logic used in both `playlists.go` (command mode) and `start.go` (interactive mode) into a shared helper function | `internal/cli/playlists.go`, `internal/cli/start.go`, `internal/cli/helpers.go` |
| 1.1.2 | Extract track display formatting (URI, name, artists) into a reusable formatter | `internal/cli/helpers.go` |
| 1.1.3 | Consolidate `truncate()` function (currently in `start.go`) into `helpers.go` | `internal/cli/start.go`, `internal/cli/helpers.go` |

### 1.2 Improve Package Structure

| Task | Description | Files |
|------|-------------|-------|
| 1.2.1 | Move `newSpotifyClient()` from `spotify_client.go` to a dedicated factory or wire it through dependency injection to avoid tight coupling with `Root` | `internal/cli/spotify_client.go` |
| 1.2.2 | Consider extracting interactive menu handlers into a separate `internal/cli/interactive/` package for better separation from CLI commands | `internal/cli/start.go` |
| 1.2.3 | Add godoc comments to all exported types and functions | All files |

### 1.3 Improve Error Handling

| Task | Description | Files |
|------|-------------|-------|
| 1.3.1 | Create custom error types for CLI-level errors (e.g., `NotLoggedInError`, `InvalidInputError`) | New: `internal/cli/errors.go` |
| 1.3.2 | Wrap errors with context consistently using `fmt.Errorf("context: %w", err)` pattern | All files |
| 1.3.3 | Add error handling for ignored errors (e.g., `filterPrompt.Run()` in `start.go` line 73) | `internal/cli/start.go` |

### 1.4 Code Quality Improvements

| Task | Description | Files |
|------|-------------|-------|
| 1.4.1 | Extract magic numbers into named constants (e.g., page sizes 50, 100, timeout durations) | `internal/spotify/client.go`, `internal/cli/login.go`, `internal/cli/start.go` |
| 1.4.2 | Remove unused variable assignment `_ = host` in `oauth.go` line 151 | `internal/auth/oauth.go` |
| 1.4.3 | Make version string configurable (currently hardcoded as "genrify 0.1" in multiple places) | `internal/cli/version.go`, `internal/config/config.go` |
| 1.4.4 | Add input validation helpers (e.g., validate playlist ID format before API calls) | `internal/cli/helpers.go` |

### 1.5 Interface Refinements

| Task | Description | Files |
|------|-------------|-------|
| 1.5.1 | Define a `SpotifyClient` interface for the methods used in CLI, enabling easier testing and mocking | New: `internal/cli/interfaces.go` |
| 1.5.2 | Define a `Prompter` interface to abstract interactive prompts for testability | New: `internal/cli/interfaces.go` |

---

## Stage 2: Testing

**Goal:** Achieve comprehensive test coverage across all packages.

### 2.1 Current Test Coverage Analysis

| Package | Current State | Coverage |
|---------|--------------|----------|
| `internal/spotify` | Has `client_test.go` and `token_manager_test.go` | Partial |
| `internal/auth` | No tests | None |
| `internal/cli` | No tests | None |
| `internal/config` | No tests | None |

### 2.2 Spotify Package Tests

| Task | Description | Files |
|------|-------------|-------|
| 2.2.1 | Add tests for `CreatePlaylist()` method | `internal/spotify/client_test.go` |
| 2.2.2 | Add tests for `ListPlaylistTracks()` with pagination | `internal/spotify/client_test.go` |
| 2.2.3 | Add tests for error handling (4xx, 5xx responses) | `internal/spotify/client_test.go` |
| 2.2.4 | Add tests for `paging.go` - `collectPaged()` with various edge cases | New: `internal/spotify/paging_test.go` |
| 2.2.5 | Add tests for `errors.go` - `decodeAPIError()` | New: `internal/spotify/errors_test.go` |

### 2.3 Auth Package Tests

| Task | Description | Files |
|------|-------------|-------|
| 2.3.1 | Add unit tests for `pkce.go` - `newPKCE()` and `randomURLSafe()` | New: `internal/auth/pkce_test.go` |
| 2.3.2 | Add unit tests for `token.go` - `Token.IsZero()` and `Token.Expired()` | New: `internal/auth/token_test.go` |
| 2.3.3 | Add unit tests for `store.go` - `Store.Load()`, `Store.Save()` using temp directories | New: `internal/auth/store_test.go` |
| 2.3.4 | Add integration tests for `oauth.go` - mock server tests for `exchangeCode()` and `Refresh()` | New: `internal/auth/oauth_test.go` |

### 2.4 Config Package Tests

| Task | Description | Files |
|------|-------------|-------|
| 2.4.1 | Add unit tests for `Load()` with various environment variable combinations | New: `internal/config/config_test.go` |
| 2.4.2 | Test default values and error cases (missing SPOTIFY_CLIENT_ID) | New: `internal/config/config_test.go` |
| 2.4.3 | Test `splitScopes()` with various input formats | New: `internal/config/config_test.go` |

### 2.5 CLI Package Tests

| Task | Description | Files |
|------|-------------|-------|
| 2.5.1 | Add unit tests for `helpers.go` - `joinArtistNames()`, `normalizeTrackURI()` | New: `internal/cli/helpers_test.go` |
| 2.5.2 | Add integration tests for CLI commands using test fixtures | New: `internal/cli/playlists_test.go` |
| 2.5.3 | Add tests for command argument validation | New: `internal/cli/playlists_test.go` |
| 2.5.4 | Test `Root` command initialization | New: `internal/cli/root_test.go` |

### 2.6 Test Infrastructure

| Task | Description | Files |
|------|-------------|-------|
| 2.6.1 | Create shared test utilities (mock Spotify server, test fixtures) | New: `internal/testutil/mock_server.go` |
| 2.6.2 | Move `memStore` from `token_manager_test.go` to shared test utilities | `internal/spotify/token_manager_test.go`, New: `internal/testutil/stores.go` |
| 2.6.3 | Add test coverage reporting to CI/dev workflow | Update Makefile or add scripts |

---

## Stage 3: Playlist Merge Feature

**Goal:** Implement functionality to merge multiple playlists matching a regex pattern into a new playlist, with optional deletion of source playlists.

### 3.1 Feature Overview

```
genrify playlists merge --pattern "Workout.*" --name "All Workouts" [--delete-sources]
```

**Flow:**
1. Search all user playlists for names matching the regex pattern
2. Display matched playlists for user confirmation
3. Create a new playlist with the specified name
4. Collect all tracks from matched playlists
5. Add all tracks to the new playlist (with deduplication option)
6. Verify all tracks were added successfully
7. Prompt user to keep or delete source playlists
8. If confirmed, delete source playlists

### 3.2 Spotify Client Extensions

| Task | Description | Files |
|------|-------------|-------|
| 3.2.1 | Add `DeletePlaylist(ctx, playlistID)` method (uses `DELETE /playlists/{id}/followers`) | `internal/spotify/client.go` |
| 3.2.2 | Add `GetPlaylist(ctx, playlistID)` method for verification | `internal/spotify/client.go` |
| 3.2.3 | Add `RemoveTracksFromPlaylist(ctx, playlistID, uris)` method (for potential future use) | `internal/spotify/client.go` |

### 3.3 Playlist Service Layer

| Task | Description | Files |
|------|-------------|-------|
| 3.3.1 | Create `PlaylistService` to handle complex playlist operations | New: `internal/playlist/service.go` |
| 3.3.2 | Implement `FindPlaylistsByPattern(pattern string) ([]SimplifiedPlaylist, error)` | New: `internal/playlist/service.go` |
| 3.3.3 | Implement `MergePlaylists(sourceIDs []string, targetName string, opts MergeOptions) (*MergeResult, error)` | New: `internal/playlist/service.go` |
| 3.3.4 | Implement `VerifyPlaylistContents(playlistID string, expectedURIs []string) (bool, []string, error)` | New: `internal/playlist/service.go` |
| 3.3.5 | Implement `DeletePlaylists(playlistIDs []string) error` | New: `internal/playlist/service.go` |

### 3.4 Types and Options

| Task | Description | Files |
|------|-------------|-------|
| 3.4.1 | Define `MergeOptions` struct (deduplicate, public, description) | New: `internal/playlist/types.go` |
| 3.4.2 | Define `MergeResult` struct (newPlaylistID, trackCount, duplicatesRemoved, etc.) | New: `internal/playlist/types.go` |
| 3.4.3 | Define `VerificationResult` struct for track verification | New: `internal/playlist/types.go` |

### 3.5 CLI Command Implementation

| Task | Description | Files |
|------|-------------|-------|
| 3.5.1 | Add `playlists merge` subcommand with flags: `--pattern`, `--name`, `--description`, `--public`, `--deduplicate`, `--delete-sources`, `--dry-run` | `internal/cli/playlists.go` |
| 3.5.2 | Implement confirmation prompts showing matched playlists before proceeding | `internal/cli/playlists.go` |
| 3.5.3 | Implement progress output during merge operation | `internal/cli/playlists.go` |
| 3.5.4 | Implement verification step with clear success/failure output | `internal/cli/playlists.go` |
| 3.5.5 | Implement final prompt for source playlist deletion | `internal/cli/playlists.go` |

### 3.6 Interactive Mode Extension

| Task | Description | Files |
|------|-------------|-------|
| 3.6.1 | Add "Merge playlists" option to interactive menu | `internal/cli/start.go` |
| 3.6.2 | Implement `interactiveMergePlaylists()` function with step-by-step prompts | `internal/cli/start.go` |

### 3.7 Testing for Merge Feature

| Task | Description | Files |
|------|-------------|-------|
| 3.7.1 | Add unit tests for regex pattern matching | New: `internal/playlist/service_test.go` |
| 3.7.2 | Add unit tests for track deduplication logic | New: `internal/playlist/service_test.go` |
| 3.7.3 | Add integration tests with mock Spotify server | New: `internal/playlist/service_test.go` |
| 3.7.4 | Add tests for verification logic | New: `internal/playlist/service_test.go` |
| 3.7.5 | Add CLI command tests | `internal/cli/playlists_test.go` |

### 3.8 Error Handling for Merge Feature

| Task | Description | Files |
|------|-------------|-------|
| 3.8.1 | Handle case where no playlists match the pattern | `internal/playlist/service.go` |
| 3.8.2 | Handle partial failures during track addition (rollback strategy) | `internal/playlist/service.go` |
| 3.8.3 | Handle rate limiting from Spotify API with exponential backoff | `internal/spotify/client.go` |
| 3.8.4 | Provide clear error messages for permission issues (can't delete others' playlists) | `internal/playlist/service.go` |

---

## Implementation Order

### Phase 1: Foundation (Stage 1 - Refactoring)
1. ✅ Complete tasks 1.1.x - Extract shared logic
2. ✅ Complete tasks 1.3.x - Improve error handling  
3. ✅ Complete tasks 1.4.x - Code quality improvements
4. ✅ Complete tasks 1.5.x - Define interfaces

### Phase 2: Confidence (Stage 2 - Testing)
1. ✅ Complete tasks 2.5.1 - Test existing helpers
2. ✅ Complete tasks 2.3.x - Auth package tests
3. ✅ Complete tasks 2.4.x - Config package tests
4. ✅ Complete tasks 2.2.x - Complete Spotify package tests
5. ✅ Complete tasks 2.6.x - Test infrastructure

### Phase 3: Feature (Stage 3 - Playlist Merge)
1. ✅ Complete tasks 3.2.x - Spotify client extensions
2. ✅ Complete tasks 3.4.x - Types and options
3. ✅ Complete tasks 3.3.x - Playlist service layer
4. ✅ Complete tasks 3.5.x - CLI command
5. ✅ Complete tasks 3.6.x - Interactive mode
6. ✅ Complete tasks 3.7.x - Testing
7. ✅ Complete tasks 3.8.x - Error handling

---

## File Structure After All Stages

```
genrify/
├── cmd/genrify/main.go
├── go.mod
├── go.sum
├── README.md
├── plan.md
├── internal/
│   ├── auth/
│   │   ├── oauth.go
│   │   ├── oauth_test.go          # New
│   │   ├── pkce.go
│   │   ├── pkce_test.go           # New
│   │   ├── store.go
│   │   ├── store_test.go          # New
│   │   ├── token.go
│   │   └── token_test.go          # New
│   ├── cli/
│   │   ├── errors.go              # New
│   │   ├── helpers.go
│   │   ├── helpers_test.go        # New
│   │   ├── interfaces.go          # New
│   │   ├── login.go
│   │   ├── playlists.go           # Modified (merge command)
│   │   ├── playlists_test.go      # New
│   │   ├── root.go
│   │   ├── root_test.go           # New
│   │   ├── spotify_client.go
│   │   ├── start.go               # Modified (merge option)
│   │   └── version.go
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go         # New
│   ├── playlist/                  # New package
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── types.go
│   ├── spotify/
│   │   ├── client.go              # Modified (new methods)
│   │   ├── client_test.go
│   │   ├── errors.go
│   │   ├── errors_test.go         # New
│   │   ├── paging.go
│   │   ├── paging_test.go         # New
│   │   ├── token_manager.go
│   │   ├── token_manager_test.go
│   │   └── types.go
│   └── testutil/                  # New package
│       ├── mock_server.go
│       └── stores.go
```

---

## Notes

- All new code should follow existing patterns and style
- Prefer table-driven tests for Go test files
- Consider adding a Makefile with targets: `build`, `test`, `coverage`, `lint`
- Consider adding GitHub Actions CI workflow
- Update README.md with new `playlists merge` command documentation
