#!/bin/bash

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o file-server-mac cmd/main/main.go

# Build for Linux
GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o file-server-lin cmd/main/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o file-server-win.exe cmd/main/main.go