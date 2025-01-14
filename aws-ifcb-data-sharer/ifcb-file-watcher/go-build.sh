#!/bin/bash
# Linux compile
GOOS=linux GOARCH=amd64 go build -o ifcb-file-watcher

# Windows compile
GOOS=windows GOARCH=amd64 go build -o ifcb-file-watcher-windows

# Mac compile
GOOS=darwin GOARCH=arm64 go build -o ifcb-file-watcher-macos