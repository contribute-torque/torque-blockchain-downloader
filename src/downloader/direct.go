package downloader

import (
	"time"

	"github.com/cavaliercoder/grab"
)

// Direct implements an HTTP/HTTPS direct blockchain downloader that
// supports download resume if available
type Direct struct {
	DownloadSource string
}

// Download the blockchain to the specified destination over HTTP/HTTPS
func (dl Direct) Download(
	destination string,
	progressChan chan Progress) error {

	client := grab.NewClient()
	req, err := grab.NewRequest(destination, dl.DownloadSource)
	if err != nil {
		return err
	}

	// start download
	resp := client.Do(req)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			progressChan <- Progress{
				BytesCompleted: resp.BytesComplete(),
				BytesTotal:     resp.Size,
			}
		case <-resp.Done:
			// Downoad completed, close progress channel
			close(progressChan)
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		return err
	}
	return nil
}
