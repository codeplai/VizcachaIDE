package domain

// UpdateStatus is where the update of the IDE stands.
type UpdateStatus string

// UpdateStatus values.
const (
	UpdateIdle        UpdateStatus = "idle"        // nothing checked yet in this session
	UpdateChecking    UpdateStatus = "checking"    // asking for the latest release
	UpdateUpToDate    UpdateStatus = "upToDate"    // the IDE is the latest version
	UpdateAvailable   UpdateStatus = "available"   // a newer version exists, not downloaded yet
	UpdateDownloading UpdateStatus = "downloading" // DownloadedBytes of TotalBytes so far
	UpdateReady       UpdateStatus = "ready"       // downloaded and verified: it can be installed
	UpdateFailed      UpdateStatus = "failed"      // Error says why; the IDE keeps working
)

// Release is a published version of the IDE.
type Release struct {
	Version     string `json:"version"`     // "2.3.0"
	Notes       string `json:"notes"`       // release notes (Markdown)
	NotesURL    string `json:"notesUrl"`    // the release page
	PublishedAt string `json:"publishedAt"` // RFC 3339
	// Asset is the file for this installation (variant, system, installed or portable).
	Asset     string `json:"asset"`
	AssetSize int64  `json:"assetSize"`
}

// UpdateState is what the frontend shows about updates (event update:state).
type UpdateState struct {
	Status          UpdateStatus `json:"status"`
	Current         string       `json:"current"` // the running version
	Latest          *Release     `json:"latest"`  // nil until a newer release is known
	DownloadedBytes int64        `json:"downloadedBytes"`
	TotalBytes      int64        `json:"totalBytes"`
	// Installs is true when Install runs the new installer and restarts the IDE (Windows,
	// installed copy); otherwise Install shows the downloaded file in its folder.
	Installs  bool   `json:"installs"`
	CheckedAt string `json:"checkedAt"` // RFC 3339 of the last successful check, "" if never
	Error     string `json:"error"`     // the reason of UpdateFailed
}
