#!/bin/bash

# Set the target architecture and operating system
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

# For modern processors
export GOAMD64=v3

# Build the Go application with optimization flags
go build -ldflags="-s -w" -gcflags="all=-B -l=120" -trimpath -o "bin/hashimg" "cmd/hashimg/main.go"

# Optionally, you can specify the output binary name
# go build -o myapp -ldflags="-s -w" -gcflags="-B" -trimpath

echo "Build completed with optimizations."
