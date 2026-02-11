package playlist

// MergeOptions controls merge behavior.
//
// Deduplicate removes duplicate track URIs while preserving first-seen order.
// Public and Description control the created playlist properties.
type MergeOptions struct {
	Deduplicate bool
	Public      bool
	Description string
}

// MergeResult describes the outcome of a merge operation.
type MergeResult struct {
	NewPlaylistID     string
	TrackCount        int
	DuplicatesRemoved int
	Verified          bool
	MissingURIs       []string
}

// VerificationResult describes the outcome of verifying playlist contents.
type VerificationResult struct {
	OK          bool
	MissingURIs []string
}
