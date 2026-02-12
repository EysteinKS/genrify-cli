package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"genrify/internal/playlist"
	"github.com/spf13/cobra"
)

func newStartCmd(root *Root) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Interactive menu for Spotify operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if root.newSpotifyClient == nil {
				return fmt.Errorf("missing spotify client factory")
			}
			c, err := root.newSpotifyClient(root.Cfg)
			if err != nil {
				return WrapLoginError(err)
			}

			// Proactively ensure we're logged in so the menu works immediately.
			if _, err := c.GetMe(cmd.Context()); err != nil {
				werr := WrapLoginError(err)
				if errors.Is(werr, ErrNotLoggedIn) {
					if root.doLogin == nil {
						return fmt.Errorf("missing login handler")
					}
					ctx, cancel := context.WithTimeout(cmd.Context(), LoginTimeout)
					defer cancel()

					if _, err := root.doLogin(ctx, root.Cfg); err != nil {
						return err
					}
					// Retry after login.
					if _, err := c.GetMe(cmd.Context()); err != nil {
						return WrapLoginError(err)
					}
				} else {
					return werr
				}
			}

			prompter := root.Prompter
			if prompter == nil {
				prompter = NewPrompter()
			}
			if root.runInteractiveLoop == nil {
				return fmt.Errorf("missing interactive loop")
			}
			return root.runInteractiveLoop(cmd.Context(), c, prompter)
		},
	}
	return cmd
}

func runInteractiveLoop(ctx context.Context, client SpotifyClient, prompter Prompter) error {
	mainMenu := []string{
		"List playlists",
		"Show playlist tracks",
		"Create new playlist",
		"Add tracks to playlist",
		"Merge playlists",
		"Exit",
	}

	for {
		_, choice, err := prompter.PromptSelect("Choose an action", mainMenu)
		if err != nil {
			return err
		}

		switch choice {
		case "List playlists":
			if err := interactiveListPlaylists(ctx, client, prompter); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "Show playlist tracks":
			if err := interactiveShowTracks(ctx, client, prompter); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "Create new playlist":
			if err := interactiveCreatePlaylist(ctx, client, prompter); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "Add tracks to playlist":
			if err := interactiveAddTracks(ctx, client, prompter); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "Merge playlists":
			if err := interactiveMergePlaylists(ctx, client, prompter); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "Exit":
			fmt.Println("Goodbye!")
			return nil
		}
		fmt.Println()
	}
}

func interactiveMergePlaylists(ctx context.Context, client SpotifyClient, prompter Prompter) error {
	pattern, err := prompter.PromptString("Regex pattern to match playlists", "")
	if err != nil {
		return err
	}
	name, err := prompter.PromptString("Name for the new merged playlist", "")
	if err != nil {
		return err
	}
	desc, err := prompter.PromptString("Description (optional)", "")
	if err != nil {
		return err
	}

	_, pubChoice, err := prompter.PromptSelect("Make new playlist public?", []string{"No", "Yes"})
	if err != nil {
		return err
	}
	public := pubChoice == "Yes"

	_, dedupChoice, err := prompter.PromptSelect("Deduplicate tracks?", []string{"No", "Yes"})
	if err != nil {
		return err
	}
	deduplicate := dedupChoice == "Yes"

	svc := playlist.NewService(client)
	matched, err := svc.FindPlaylistsByPattern(ctx, pattern)
	if err != nil {
		if err == playlist.ErrNoPlaylistsMatched {
			fmt.Println("No playlists matched the pattern")
			return nil
		}
		return err
	}

	fmt.Printf("\nMatched %d playlist(s):\n", len(matched))
	for _, p := range matched {
		fmt.Printf("%s\t%s\n", p.ID, p.Name)
	}

	_, proceed, err := prompter.PromptSelect("Proceed with merge?", []string{"No", "Yes"})
	if err != nil {
		return err
	}
	if proceed != "Yes" {
		return ErrCancelled
	}

	sourceIDs := make([]string, 0, len(matched))
	for _, p := range matched {
		sourceIDs = append(sourceIDs, p.ID)
	}

	res, err := svc.MergePlaylists(ctx, sourceIDs, name, playlist.MergeOptions{
		Deduplicate: deduplicate,
		Public:      public,
		Description: desc,
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nCreated playlist: %s\n", res.NewPlaylistID)
	fmt.Printf("Tracks added: %d (duplicates removed: %d)\n", res.TrackCount, res.DuplicatesRemoved)
	if res.Verified {
		fmt.Println("Verification: OK")
	} else {
		fmt.Printf("Verification: FAILED (missing %d track(s))\n", len(res.MissingURIs))
		return fmt.Errorf("verification failed")
	}

	_, deleteChoice, err := prompter.PromptSelect("Delete source playlists?", []string{"No", "Yes"})
	if err != nil {
		return err
	}
	if deleteChoice == "Yes" {
		if err := svc.DeletePlaylists(ctx, sourceIDs); err != nil {
			return err
		}
		fmt.Println("Deleted source playlists")
	}

	return nil
}

func interactiveListPlaylists(ctx context.Context, client SpotifyClient, prompter Prompter) error {
	filter, err := prompter.PromptString("Filter by name (press Enter to skip)", "")
	if err != nil {
		return err
	}
	filter = strings.TrimSpace(filter)

	limit, err := prompter.PromptInt("Max playlists to display", DefaultPlaylistLimit)
	if err != nil {
		return err
	}

	fetchMax := limit
	if filter != "" {
		fetchMax = 0 // fetch all for filtering
	}

	pls, err := client.ListCurrentUserPlaylists(ctx, fetchMax)
	if err != nil {
		return err
	}

	filtered := filterPlaylistsByName(pls, filter)
	printed := 0
	fmt.Println("\nID\t\tName\t\t\t\tTracks\tOwner")
	fmt.Println(strings.Repeat("-", 80))
	for _, p := range filtered {
		fmt.Printf("%s\t%-30s\t%d\t%s\n", p.ID, truncate(p.Name, 30), p.Tracks.Total, p.Owner.ID)
		printed++
		if limit > 0 && printed >= limit {
			break
		}
	}
	fmt.Printf("\nShowing %d playlist(s)\n", printed)
	return nil
}

func interactiveShowTracks(ctx context.Context, client SpotifyClient, prompter Prompter) error {
	playlistID, err := prompter.PromptString("Playlist ID", "")
	if err != nil {
		return err
	}
	playlistID, err = normalizePlaylistID(playlistID)
	if err != nil {
		return err
	}

	limit, err := prompter.PromptInt("Max tracks to display", DefaultTrackLimit)
	if err != nil {
		return err
	}

	tracks, err := client.ListPlaylistTracks(ctx, playlistID, limit)
	if err != nil {
		return err
	}

	fmt.Println("\nURI\t\t\t\t\tName\t\t\t\tArtists")
	fmt.Println(strings.Repeat("-", 100))
	for _, t := range tracks {
		fmt.Printf("%s\t%-35s\t%s\n", t.URI, truncate(t.Name, 35), truncate(joinArtistNames(t.Artists), 30))
	}
	fmt.Printf("\nShowing %d track(s)\n", len(tracks))
	return nil
}

func interactiveCreatePlaylist(ctx context.Context, client SpotifyClient, prompter Prompter) error {
	name, err := prompter.PromptString("Playlist name", "")
	if err != nil {
		return err
	}

	desc, err := prompter.PromptString("Description (optional)", "")
	if err != nil {
		return err
	}

	_, pubChoice, err := prompter.PromptSelect("Make playlist public?", []string{"No", "Yes"})
	if err != nil {
		return err
	}
	public := pubChoice == "Yes"

	me, err := client.GetMe(ctx)
	if err != nil {
		return err
	}

	pl, err := client.CreatePlaylist(ctx, me.ID, name, desc, public)
	if err != nil {
		return err
	}

	fmt.Printf("\nPlaylist created: %s (%s)\n", pl.Name, pl.ID)
	return nil
}

func interactiveAddTracks(ctx context.Context, client SpotifyClient, prompter Prompter) error {
	playlistID, err := prompter.PromptString("Playlist ID", "")
	if err != nil {
		return err
	}
	playlistID, err = normalizePlaylistID(playlistID)
	if err != nil {
		return err
	}

	tracksInput, err := prompter.PromptString("Track URIs or URLs (comma-separated)", "")
	if err != nil {
		return err
	}

	rawTracks := strings.Split(tracksInput, ",")
	uris := make([]string, 0, len(rawTracks))
	for _, raw := range rawTracks {
		u, err := normalizeTrackURI(raw)
		if err != nil {
			fmt.Printf("Warning: skipping invalid track %q: %v\n", raw, err)
			continue
		}
		uris = append(uris, u)
	}

	if len(uris) == 0 {
		return fmt.Errorf("no valid tracks to add")
	}

	snapshot, err := client.AddTracksToPlaylist(ctx, playlistID, uris)
	if err != nil {
		return err
	}

	fmt.Printf("\nAdded %d track(s). Snapshot: %s\n", len(uris), snapshot)
	return nil
}
