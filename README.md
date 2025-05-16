# Multi-Threaded File Downloader (Go)

A high-speed, multi-threaded file downloader built in **Go**, designed to split a file into multiple chunks and download them in parallel — boosting download performance significantly.

---

## Features

-  **Multi-threaded downloading** using Goroutines
-  File is split into chunks and downloaded in parallel
-  Automatic file merging after download
-  Simple CLI usage

---

# Requirements
- Golang

## Project Structure
```bash
file-downloader/
├── cmd/
│ └── app/
│ └── mulDownloader.go
├── go.mod
└── README.md
```

### 🔧 Build

```bash
go build -o main ./cmd/app/mulDownload.go
```

# Usage

```bash 
./main <the_url> <output_path> <no_of_threads>
```
