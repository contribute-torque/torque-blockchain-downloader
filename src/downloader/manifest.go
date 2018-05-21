package downloader

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Manifest contains the links to the blockchain download as well
// as metadata for verification
type Manifest struct {
	Magnet string `json:"magnet"`
	Direct string `json:"direct"`
	Sha512 string `json:"sha512"`
	Bytes  int64  `json:"bytes"`
	Block  int    `json:"block"`
}

// GetManifest checks a well-know address on ZeroNet for the manifest JSON
// file and returns the contents
func GetManifest(manifestURL string) (Manifest, error) {
	var manifest Manifest
	// The manifest address is hardcoded because it is a well-know location
	resp, err := http.Get(manifestURL)
	if err != nil {
		return manifest, err
	}

	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return manifest, err
	}

	if manifest.Magnet == "" || manifest.Direct == "" || manifest.Sha512 == "" {
		return manifest,
			errors.New("Invalid manifest received. Magnet, direct or SHA value is missing")
	}

	return manifest, nil
}
