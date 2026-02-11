package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"genrify/internal/playlist"
	"github.com/spf13/cobra"
)

func newPlaylistsCmd(root *Root) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playlists",
		Short: "Playlist operations",
	}

	cmd.AddCommand(newPlaylistsListCmd(root))
	cmd.AddCommand(newPlaylistsTracksCmd(root))
	cmd.AddCommand(newPlaylistsCreateCmd(root))
	cmd.AddCommand(newPlaylistsAddCmd(root))
	cmd.AddCommand(newPlaylistsMergeCmd(root))

	return cmd
}

func newPlaylistsListCmd(root *Root) *cobra.Command {
	var filter string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List playlists (requires login)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSpotifyClient(root.Cfg)
			if err != nil {
				return WrapLoginError(err)
			}

			fetchMax := limit
			if strings.TrimSpace(filter) != "" {
				fetchMax = 0 // fetch all so filtering is accurate
			}

			pls, err := c.ListCurrentUserPlaylists(cmd.Context(), fetchMax)
			if err != nil {
				return fmt.Errorf("list playlists: %w", err)
			}

			filtered := filterPlaylistsByName(pls, filter)
			printed := 0
			for _, p := range filtered {
				cmd.Println(formatPlaylistRow(p))
				printed++
				if limit > 0 && printed >= limit {
					break
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter by playlist name substring (case-insensitive)")
	cmd.Flags().IntVar(&limit, "limit", DefaultPlaylistLimit, "Max playlists to print (0 = no limit)")
	return cmd
}

func newPlaylistsTracksCmd(root *Root) *cobra.Command {
	var limit int
	var urisOnly bool

	cmd := &cobra.Command{
		Use:   "tracks <playlist-id>",
		Short: "List tracks in a playlist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSpotifyClient(root.Cfg)
			if err != nil {
				return WrapLoginError(err)
			}

			playlistID, err := normalizePlaylistID(args[0])
			if err != nil {
				return fmt.Errorf("%w: %v", ErrInvalidInput, err)
			}

			tracks, err := c.ListPlaylistTracks(cmd.Context(), playlistID, limit)
			if err != nil {
				return fmt.Errorf("list tracks: %w", err)
			}

			for _, t := range tracks {
				if urisOnly {
					cmd.Println(t.URI)
					continue
				}
				cmd.Println(formatTrackRow(t))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", DefaultTrackLimit, "Max tracks to print (0 = no limit)")
	cmd.Flags().BoolVar(&urisOnly, "uris", false, "Only print track URIs")
	return cmd
}

func newPlaylistsCreateCmd(root *Root) *cobra.Command {
	var name string
	var description string
	var public bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new playlist for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSpotifyClient(root.Cfg)
			if err != nil {
				return WrapLoginError(err)
			}

			me, err := c.GetMe(cmd.Context())
			if err != nil {
				return fmt.Errorf("get current user: %w", err)
			}

			pl, err := c.CreatePlaylist(cmd.Context(), me.ID, name, description, public)
			if err != nil {
				return fmt.Errorf("create playlist: %w", err)
			}
			cmd.Printf("%s\t%s\n", pl.ID, pl.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Playlist name")
	cmd.Flags().StringVar(&description, "description", "", "Playlist description")
	cmd.Flags().BoolVar(&public, "public", false, "Make playlist public")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newPlaylistsAddCmd(root *Root) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <playlist-id> <track-uri-or-url> [more...]",
		Short: "Add tracks to a playlist",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSpotifyClient(root.Cfg)
			if err != nil {
				return WrapLoginError(err)
			}

			playlistID, err := normalizePlaylistID(args[0])
			if err != nil {
				return fmt.Errorf("%w: %v", ErrInvalidInput, err)
			}
			uris := make([]string, 0, len(args)-1)
			for _, a := range args[1:] {
				u, err := normalizeTrackURI(a)
				if err != nil {
					return fmt.Errorf("%w: %v", ErrInvalidInput, err)
				}
				uris = append(uris, u)
			}

			snapshot, err := c.AddTracksToPlaylist(cmd.Context(), playlistID, uris)
			if err != nil {
				return fmt.Errorf("add tracks: %w", err)
			}
			cmd.Println(snapshot)
			return nil
		},
	}
	return cmd
}

func newPlaylistsMergeCmd(root *Root) *cobra.Command {
	var pattern string
	var name string
	var description string
	var public bool
	var deduplicate bool
	var deleteSources bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge multiple playlists matching a regex into a new playlist",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newSpotifyClient(root.Cfg)
			if err != nil {
				return WrapLoginError(err)
			}
			svc := playlist.NewService(c)

			matched, err := svc.FindPlaylistsByPattern(cmd.Context(), pattern)
			if err != nil {
				if errors.Is(err, playlist.ErrNoPlaylistsMatched) {
					return fmt.Errorf("no playlists matched pattern %q", pattern)
				}
				return fmt.Errorf("find playlists: %w", err)
			}

			cmd.Printf("Matched %d playlist(s):\n", len(matched))
			for _, p := range matched {
				cmd.Println(formatPlaylistRow(p))
			}

			ok, err := confirm(cmd, "Proceed with merge?")
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("%w", ErrCancelled)
			}

			if dryRun {
				cmd.Println("Dry run: no changes made")
				return nil
			}

			sourceIDs := make([]string, 0, len(matched))
			for _, p := range matched {
				if p.ID != "" {
					sourceIDs = append(sourceIDs, p.ID)
				}
			}

			cmd.Println("Merging...")
			res, err := svc.MergePlaylists(cmd.Context(), sourceIDs, name, playlist.MergeOptions{
				Deduplicate: deduplicate,
				Public:      public,
				Description: description,
			})
			if err != nil {
				return fmt.Errorf("merge playlists: %w", err)
			}

			cmd.Printf("Created playlist: %s\n", res.NewPlaylistID)
			cmd.Printf("Tracks added: %d (duplicates removed: %d)\n", res.TrackCount, res.DuplicatesRemoved)
			if res.Verified {
				cmd.Println("Verification: OK")
			} else {
				cmd.Printf("Verification: FAILED (missing %d track(s))\n", len(res.MissingURIs))
				for i := 0; i < len(res.MissingURIs) && i < 10; i++ {
					cmd.Printf("- %s\n", res.MissingURIs[i])
				}
				return fmt.Errorf("verification failed")
			}

			if deleteSources {
				del, err := confirm(cmd, "Delete source playlists?")
				if err != nil {
					return err
				}
				if del {
					ids := make([]string, 0, len(matched))
					for _, p := range matched {
						ids = append(ids, p.ID)
					}
					if err := svc.DeletePlaylists(cmd.Context(), ids); err != nil {
						return err
					}
					cmd.Println("Deleted source playlists")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&pattern, "pattern", "", "Regex pattern to match playlist names")
	cmd.Flags().StringVar(&name, "name", "", "Name of the new merged playlist")
	cmd.Flags().StringVar(&description, "description", "", "Description for the new playlist")
	cmd.Flags().BoolVar(&public, "public", false, "Make the new playlist public")
	cmd.Flags().BoolVar(&deduplicate, "deduplicate", false, "Remove duplicate track URIs")
	cmd.Flags().BoolVar(&deleteSources, "delete-sources", false, "Prompt to delete source playlists after merge")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be merged without making changes")
	_ = cmd.MarkFlagRequired("pattern")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func confirm(cmd *cobra.Command, prompt string) (bool, error) {
	cmd.Printf("%s [y/N]: ", prompt)
	r := bufio.NewReader(cmd.InOrStdin())
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	s := strings.ToLower(strings.TrimSpace(line))
	return s == "y" || s == "yes", nil
}
