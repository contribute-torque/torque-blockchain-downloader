package downloader

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/anacrolix/torrent"
)

// Torrent implements downloading of the blockchain using a torrent
type Torrent struct {
	DownloadSource string
	AllowSeed      bool
}

// Download the blockchain to the specified destination over torrent
func (dl Torrent) Download(
	destination string,
	progressChan chan Progress) error {

	defer close(progressChan)
	client, err := torrent.NewClient(&torrent.Config{
		Seed:        dl.AllowSeed,
		DisableIPv6: true,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	tor, err := client.AddMagnet(dl.DownloadSource)
	if err != nil {
		return err
	}
	<-tor.GotInfo()
	torInfo := tor.Info()
	reader := tor.NewReader()
	tor.DownloadAll()

	// Read 4kb at a time
	data := make([]byte, 4096)
	bytesRead := 0
	lastReport := time.Now()
	for {
		data = data[:cap(data)]
		n, err := reader.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return err
		}
		data = data[:n]
		bytesRead += n
		if time.Since(lastReport).Seconds() > 1 {
			lastReport = time.Now()
			progressChan <- Progress{
				BytesCompleted: int64(bytesRead),
				BytesTotal:     torInfo.TotalLength(),
			}
		}
	}
	// Download is done!
	client.WaitAll()

	// Now we move the downloaded file to the destination path
	// NOTE: The torrent must (for now) ONLY include the blockchain file
	// TODO: Fix this downloaded path thing
	//downloadedPath := "./" + tor.Files()[0].Path()
	downloadedPath := "./" + torInfo.Name
	return os.Rename(downloadedPath, destination)
}
