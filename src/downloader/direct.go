package downloader

import (
	"fmt"
	"os"
	"time"

	"github.com/cavaliercoder/grab"
)

// Direct implements an HTTP/HTTPS direct blockchain downloader that
// supports download resume if available
type Direct struct {
}

// Download the blockchain to the specified destination over HTTP/HTTPS
func (dl Direct) Download(
	destination string,
	progressChan chan Progress) error {

	fmt.Println("Downloading the latest blockchain file via direct download")

	downloadPath := "http://stellite.live.local/Stellite-Blockchain-Export-Block-109766.raw"
	fmt.Println("Download from", downloadPath)

	client := grab.NewClient()
	req, _ := grab.NewRequest(destination, downloadPath)

	// start download
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())
			fmt.Println("Progress")
			progressChan <- Progress{
				BytesCompleted: resp.BytesComplete(),
				BytesTotal:     resp.Size,
			}
		case <-resp.Done:
			fmt.Println("Done")
			close(progressChan)
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)

	return nil
}
