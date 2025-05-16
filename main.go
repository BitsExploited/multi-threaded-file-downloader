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

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.Writer.Write(p) 
	if err == nil {
		pw.Downloader.Mutex.Lock()
		pw.Downloader.Downloaded += int64(n)
		pw.Downloader.Mutex.Unlock()
	}
	return
}

func (d *Downloader) Download() error {
	// Get file size
	resp, err := http.Head(d.URL)
	if err != nil {
		return fmt.Errorf("failed to get file size: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	d.TotalSize = resp.ContentLength
	if d.TotalSize <= 0 {
		return fmt.Errorf("invalid file size: %d", d.TotalSize)
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(d.OutPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create temporary files for parts
	partFiles := make([]*os.File, d.Concurrency)
	for i := 0; i < d.Concurrency; i++ {
		partFiles[i], err = os.CreateTemp("", "download_part_")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %v", err)
		}
		defer os.Remove(partFiles[i].Name())
		defer partFiles[i].Close()
	}

	// Calculate part sizes
	partSize := d.TotalSize / int64(d.Concurrency)
	parts := make([]DownloadPart, d.Concurrency)
	for i := 0; i < d.Concurrency; i++ {
		start := int64(i) * partSize
		end := start + partSize - 1
		if i == d.Concurrency-1 {
			end = d.TotalSize - 1
		}
		parts[i] = DownloadPart{
			Index: i,
			Start: start,
			End:   end,
		}
	}

	done := make(chan struct{})
	go d.reportProgress(done)

	var wg sync.WaitGroup
	errChan := make(chan error, d.Concurrency)

	for i, part := range parts {
		wg.Add(1)
		go func(part DownloadPart, partFile *os.File) {
			defer wg.Done()
			if err := d.downloadPart(part, partFile); err != nil {
				errChan <- fmt.Errorf("failed to download part %d: %v", part.Index, err)
			}
		}(part, partFiles[i])
	}

	wg.Wait()
	close(errChan)
	close(done) 

	if len(errChan) > 0 {
		return <-errChan
	}

	outFile, err := os.Create(d.OutPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	for _, partFile := range partFiles {
		if _, err := partFile.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek part file: %v", err)
		}
		if _, err := io.Copy(outFile, partFile); err != nil {
			return fmt.Errorf("failed to write part to output: %v", err)
		}
	}

	fmt.Printf("\nDownload completed: %s\n", d.OutPath)
	return nil
}


func (d *Downloader) downloadPart(part DownloadPart, partFile *os.File) error {
	req, err := http.NewRequest("GET", d.URL, nil) 
	if err  != nil {
		return err
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", part.Start, part.End)
	req.Header.Set("Range", rangeHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	progressWriter := &ProgressWriter{
		Writer:     partFile,
		Downloader: d,
		PartIndex:  part.Index,
	}

	_, err = io.Copy(progressWriter, resp.Body)
	return err

}

func (d *Downloader) reportProgress(done chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			d.Mutex.Lock()
			percent := float64(d.Downloaded) / float64(d.TotalSize) * 100
			fmt.Printf("\rDownloading: %.2f%% (%d / %d bytes)", percent, d.Downloaded, d.TotalSize)
			d.Mutex.Unlock()
		}
	}
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
