package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/donovansolms/stellite-blockchain-downloader/src/downloader"
	"github.com/spf13/cobra"
)

// workingDir is the path we're executing from
var workingDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stellite-blockchain-downloader",
	Short: "Stellite blockchain download and importer",
	Long: `
stellite-blockchain-downloader downloads the latest available blockchain
export and imports it using the 'stellite-blockchain-import' tool.

The tool supports the following download methods:
1. Direct HTTP/HTTPS
2. Direct via IPFS
3. Torrent`,

	// By default the root command executes a download and import of the
	// latest blockchain file
	Run: func(cmd *cobra.Command, args []string) {

		// First we need to check if the 'stellite-blockchain-import' tool
		// is available
		// The tool has a well known name, so just check for that
		/*

					_, err = os.Stat(filepath.Join(workingDir, "stellite-blockchain-import"))
					if runtime.GOOS == "windows" {
						_, err = os.Stat(filepath.Join(workingDir, "stellite-blockchain-import.exe"))
					}
					if os.IsNotExist(err) {
						fmt.Printf(`The blockchain import tool 'stellite-blockchain-import' does
			not exist in the current directory.

			Please execute this tool from the same path as the 'stellite-blockchain-import'
			tool.`)
						fmt.Println("")
						os.Exit(1)
					}
		*/
		// TODO: Remove err declaration
		var err error

		var downloadHandler downloader.Downloader

		switch strings.ToLower(cmd.Flag("method").Value.String()) {
		case "direct":
			downloadHandler = downloader.Direct{
			// TODO: Give download path here
			}
		}

		// Check if the selected download path exists and is a directory
		destinationDir := cmd.Flag("destination-dir").Value.String()
		destinationDir, err = filepath.Abs(destinationDir)
		if err != nil {
			fmt.Println("Could not determine destination directory:", err)
			os.Exit(1)
		}
		fileInfo, err := os.Stat(destinationDir)
		if err != nil {
			fmt.Println("Could not read destination directory:", err)
			os.Exit(1)
		}
		if fileInfo.IsDir() == false {
			fmt.Printf(
				"Destination directory '%s' must be a directory\n",
				destinationDir)
			os.Exit(1)
		}

		// progressChan receives progress updates from the selected downloader
		// and is used to display the progress
		progressChan := make(chan downloader.Progress)

		// TODO: Implement progress
		go func() {
			select {
			case progress := <-progressChan:
				fmt.Println("Progress")
				fmt.Println(progress)
			}
		}()

		// Create a definitive download location to use for the import
		destinationPath := filepath.Join(destinationDir, "stellite-blockchain.raw")
		err = downloadHandler.Download(destinationPath, progressChan)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// TODO import the blockchain
		// stellite-blockchain-import --input-file xxx --verify 0
		// print output
	},
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once
// to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Declare err to not shadow workingDir
	var err error
	workingDir, err = os.Executable()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	workingDir = filepath.Dir(workingDir)

	rootCmd.Flags().Bool(
		"download-only",
		false,
		"download the blockchain but don't import")

	rootCmd.Flags().StringP(
		"destination-dir",
		"d",
		workingDir,
		"directory to download to")

	rootCmd.Flags().StringP(
		"method",
		"m",
		"direct",
		"set the download method. Available 'direct', 'torrent' or 'ipfs'")

	// TODO: Add flag to set path to stellite-blockchain-import tool
}
