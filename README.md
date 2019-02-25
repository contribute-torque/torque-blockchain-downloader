# Stellite Blockchain Downloader

A simple tool to download and import the latest Stellite blockchain file. It uses a
torrent for faster downloads, direct HTTPS download is also provided as failover.

## Available commands

`./stellite-blockchain-downloader --help`

```
Flags:
      --data-dir string               set a custom blockchain path
  -d, --destination-dir string        directory to download to
      --disable-seed                  if we are allowed to seed the torrent while downloading
      --download-only                 download the blockchain but don't import
      --force                         if we should remove the current chain
  -h, --help                          help for stellite-blockchain-downloader
      --import-tool-path string       set the path to the import tool if in other location
      --manifest-url string           set the manifest URL
  -m, --method string                 set the download method. Available 'direct' or 'torrent' (default "torrent")
      --without-import-verification   if --dangerous-unverified-import 1 should be used on import (less safe, but much faster)
```

## Common uses

* To download and import the blockchain on first start

```./stellite-blockchain-downloader```

* To download and import the blockchain and overwrite your current one

```./stellite-blockchain-downloader --force```

* To download and import the blockchain if you store your blockchain somewhere else than the default

```./stellite-blockchain-downloader --data-dir /path/to/stellite```

* To download, import the blockchain and overwrite the current blockchain if you store your blockchain somewhere else than the default

```./stellite-blockchain-downloader --data-dir /path/to/stellite --force```

* To download and import the blockchain if torrents are blocked by your provider

```./stellite-blockchain-downloader --method direct```


## Compiling

The tool is written in Go and can be cross-compiled to Linux, Windows and MacOS.

### Linux

* Install Go

https://golang.org/dl/

* Install dep, a Go dependency manager

https://golang.github.io/dep/docs/installation.html

* Clone the repository

```
git clone https://github.com/stellitecoin/stellite-blockchain-downloader.git
cd stellite-blockchain-downloader
```

* Pull dependencies

```
dep ensure
```

* Build for all platforms

```
make
```

* Or only build for your platform

```
go build -o stellite-blockchain-downloader src/main.go
```
