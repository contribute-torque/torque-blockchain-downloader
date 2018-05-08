package downloader

// Progress holds information about the current download progress
type Progress struct {
	BytesCompleted int64
	BytesTotal     int64
}

// Downloader defines the behaviour for a blockchain downloader
type Downloader interface {
	// Download the blockchain to the specified destination
	Download(destination string, progressChan chan Progress) error
}
