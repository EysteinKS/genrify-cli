package spotify

type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type SimplifiedPlaylist struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Public        bool   `json:"public"`
	Collaborative bool   `json:"collaborative"`
	Owner         User   `json:"owner"`
	Tracks        struct {
		Total int `json:"total"`
	} `json:"tracks"`
}

type paging[T any] struct {
	Href   string `json:"href"`
	Items  []T    `json:"items"`
	Limit  int    `json:"limit"`
	Next   string `json:"next"`
	Offset int    `json:"offset"`
	Prev   string `json:"previous"`
	Total  int    `json:"total"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FullTrack struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	URI     string   `json:"uri"`
	Artists []Artist `json:"artists"`
	Album   Album    `json:"album"`
}

type playlistTrackItem struct {
	Track FullTrack `json:"track"`
}

// https://developer.spotify.com/documentation/web-api/reference/add-tracks-to-playlist

type snapshotResponse struct {
	SnapshotID string `json:"snapshot_id"`
}
