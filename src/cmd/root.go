package cmd

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/donovansolms/stellite-blockchain-downloader/src/downloader"
	"github.com/spf13/cobra"
	"gopkg.in/cheggaaa/pb.v1"
)

// workingDir is the path we're executing from
var workingDir string
var downloadOnly bool
var disableSeed bool
var verifyImport bool
var forceCleanImport bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stellite-blockchain-downloader",
	Short: "Stellite blockchain download and importer",
	Long: `
stellite-blockchain-downloader downloads the latest available blockchain
export and imports it using the 'stellite-blockchain-import' tool.

The tool supports the following download methods:
1. Direct HTTP/HTTPS
2. Torrent`,

	// By default the root command executes a download and import of the
	// latest blockchain file
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf(`
  __ _____ ___ _   _   _ _____ ___
/' _/_   _| __| | | | | |_   _| __|
'._'. | | | _|| |_| |_| | | | | _|
|___/ |_| |___|___|___|_| |_| |___|
                  BLOCKCHAIN DOWNLOADER
			`)
		// Clear all the spaces
		fmt.Println("")

		if downloadOnly == false {
			// We need to check if the 'stellite-blockchain-import' tool
			// is available before doing anything else
			// The tool has a well known name, so just check for that
			_, err := os.Stat(filepath.Join(cmd.Flag("import-tool-path").Value.String()))
			if os.IsNotExist(err) {
				fmt.Printf(`
The blockchain import tool 'stellite-blockchain-import' does
not exist in the current directory.

Please execute this tool from the same path as the 'stellite-blockchain-import'
tool or set the flag --import-tool-path to the correct location
`)
				fmt.Print("Press any key to continue...")
				_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
				os.Exit(0)
			}
		}

		// Get the manifest file for the download locations. The manifest file
		// includes the direct download location, the torrent magnet link, the
		// block height of the download and the sha512 sub for validation
		manifest, err := downloader.GetManifest(
			cmd.Flag("manifest-url").Value.String())
		if err != nil {
			fmt.Println("Could not retrieve download manifest, please check your connection:", err)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}

		// TODO: Check ZeroNet address for JSON manifest file
		// The manifest file contains the torrent magnet link, the direct download
		// address, a SHA512 hash of the file and the block number
		// For now we're just using a JSON file on stellite.live

		// Check if the selected download path exists and is a directory
		destinationDir := cmd.Flag("destination-dir").Value.String()
		destinationDir, err = filepath.Abs(destinationDir)
		if err != nil {
			fmt.Println("Could not determine destination directory:", err)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}
		fileInfo, err := os.Stat(destinationDir)
		if err != nil {
			fmt.Println("Could not read destination directory:", err)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}
		if fileInfo.IsDir() == false {
			fmt.Printf(
				"Destination directory '%s' must be a directory\n",
				destinationDir)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}

		// Set up the download method
		var downloadHandler downloader.Downloader
		switch strings.ToLower(cmd.Flag("method").Value.String()) {
		case "direct":
			downloadHandler = downloader.Direct{
				DownloadSource: manifest.Direct,
			}
			fmt.Println("Starting direct download from", manifest.Direct)
		case "torrent":
			downloadHandler = downloader.Torrent{
				DownloadSource: manifest.Magnet,
				AllowSeed:      !disableSeed,
			}
			fmt.Printf("Starting torrent download from %s ", manifest.Magnet)
			if disableSeed {
				fmt.Printf("- not seeding")
			}
			fmt.Printf("\n")
		default:
			fmt.Printf(
				"Download method '%s' is not a valid method. Available methods are 'direct' and 'torrent'\n",
				cmd.Flag("method").Value.String())
		}

		// progressChan receives progress updates from the selected downloader
		// and is used to display the progress
		progressChan := make(chan downloader.Progress)
		progressBar := pb.New64(manifest.Bytes)
		progressBar.SetUnits(pb.U_BYTES)
		progressBar.Start()

		// We receive the progress via a channel from the downloader
		go func() {
			for progress := range progressChan {
				progressBar.Set64(progress.BytesCompleted)
			}
		}()

		// Create a definitive download location to use for the import
		destinationPath := filepath.Join(destinationDir, "stellite-blockchain.raw")
		err = downloadHandler.Download(destinationPath, progressChan)
		if err != nil {
			fmt.Printf("Download failed: %s\n", err)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}
		// Just in case the progress bar hasn't updated yet, set to 100%
		// since we're done
		progressBar.Set64(manifest.Bytes)
		progressBar.Update()
		progressBar.Finish()

		fmt.Printf("Download saved to %v \n", destinationPath)
		fmt.Println("Verifying downloaded file...")
		// Check if the download SHA512 matches the manifest's
		verified, err := verifyHash(destinationPath, manifest.Sha512)
		if err != nil || verified == false {
			fmt.Printf("Unable to verify downloaded file: %s\n", err)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}
		fmt.Println("Downloaded file verified using the SHA512 hash")

		if downloadOnly {
			fmt.Printf(`
You selected the download only option. You may use the
downloaded blockchain file now.

The location of the downloaded file is:
`)
			fmt.Printf("Download saved to %v \n\n", destinationPath)
			os.Exit(0)
		}

		fmt.Println("Importing downloaded blockchain file...")

		// Import the blockchain file by executing the 'stellite-blockchain-import'
		// tool and print the output of the tool as we continue
		importArgs := []string{
			"--input-file",
			destinationPath,
		}
		if forceCleanImport {
			importArgs = append(importArgs, "--resume")
			importArgs = append(importArgs, "0")
		}
		importArgs = append(importArgs, "--verify")
		if verifyImport {
			importArgs = append(importArgs, "1")
		} else {
			importArgs = append(importArgs, "0")
		}

		importCommand := exec.Command(
			cmd.Flag("import-tool-path").Value.String(),
			importArgs...)

		stdoutIn, _ := importCommand.StdoutPipe()
		stderrIn, _ := importCommand.StderrPipe()

		var errStdout, errStderr error
		var stdoutBuf, stderrBuf bytes.Buffer
		stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
		stderr := io.MultiWriter(os.Stderr, &stderrBuf)
		err = importCommand.Start()
		if err != nil {
			fmt.Printf("Unable to start the import tool: %s\n", err)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}

		go func() {
			_, errStdout = io.Copy(stdout, stdoutIn)
		}()

		go func() {
			_, errStderr = io.Copy(stderr, stderrIn)
		}()

		err = importCommand.Wait()
		if err != nil {
			fmt.Printf("Failed to import the downloaded blockchain: %s\n", err)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}
		if errStdout != nil || errStderr != nil {
			fmt.Printf("Unable to capture the import tool's ourput: %s, %s\n",
				errStdout,
				errStderr)
			fmt.Print("Press any key to continue...")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(0)
		}

		err = os.Remove(destinationPath)
		if err != nil {
			fmt.Printf("The downloaded file '%s' could not be removed: %s\n",
				destinationPath,
				err,
			)
		}

		fmt.Printf(`
  __ _____ ___ _   _   _ _____ ___
/' _/_   _| __| | | | | |_   _| __|
'._'. | | | _|| |_| |_| | | | | _|
|___/ |_| |___|___|___|_| |_| |___|
                  BLOCKCHAIN DOWNLOADER
			`)
		fmt.Printf(`
Imported downloaded blockchain file successfully.
You may now start 'stellited' or your wallet.

Thank you for using the Stellite Blockchain Downloader.

`)
		fmt.Print("Press any key to continue...")
		_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(0)
	},
}

// verifyHash verifies the given SHA512 hash of a file with the given hash
func verifyHash(filepath string, sha512hash string) (bool, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hasher := sha512.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, err
	}

	if hex.EncodeToString(hasher.Sum(nil)) == sha512hash {
		return true, nil
	}
	return false, errors.New("File SHA512 hash does not match expected hash")
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
	cobra.MousetrapHelpText = ""

	// Declare err to not shadow workingDir
	var err error
	workingDir, err = os.Executable()
	if err != nil {
		fmt.Println(err)
		fmt.Print("Press any key to continue...")
		_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(0)
	}
	workingDir = filepath.Dir(workingDir)

	rootCmd.Flags().BoolVar(
		&downloadOnly,
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
		"torrent",
		"set the download method. Available 'direct' or 'torrent'")

	rootCmd.Flags().String(
		"manifest-url",
		"https://stellite.live/downloads/blockchain-download.manifest",
		"set the manifest URL")

	rootCmd.Flags().BoolVar(
		&verifyImport,
		"with-import-verification",
		false,
		"if --verify 1 should be used on import")

	rootCmd.Flags().BoolVar(
		&disableSeed,
		"disable-seed",
		false,
		"if we are allowed to seed the torrent while downloading")

	rootCmd.Flags().BoolVar(
		&forceCleanImport,
		"force",
		false,
		"if we should overwrite the current chain")

	importToolPAth := filepath.Join(workingDir, "stellite-blockchain-import")
	if runtime.GOOS == "windows" {
		importToolPAth = filepath.Join(workingDir, "stellite-blockchain-import.exe")
	}
	rootCmd.Flags().String(
		"import-tool-path",
		importToolPAth,
		"set the path to the import tool")
}
