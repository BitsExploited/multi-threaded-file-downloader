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

}
