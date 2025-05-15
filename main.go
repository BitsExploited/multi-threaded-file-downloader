package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Downloader struct {
	URL string
	OutPath string
	Concurrency int
	TotalSize int64
	Downloaded int64
	Mutex sync.Mutex
}

type DownloadPart struct {
	Index int
	Start int64
	End int64
}

type ProgressWriter struct {
	Writer io.Writer
	Downloader *Downloader
	PartIndex int
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: donwloader <url> <output_path> [concurrency]")
		os.Exit(1)
	}

	url := os.Args[1]
	outPath := os.Args[2]
	concurrecny := 4
	if len(os.Args) > 3 {
		if c, err := strconv.Atoi(os.Args[3]); err == nil && c > 0 {
			concurrecny = c
		}
	}

	downloader := &Downloader{
		URL: url,
		OutPath: outPath,
		Concurrency: concurrecny,
	}

	if err := downloader.Download(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
