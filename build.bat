@echo off
REM Set the target architecture and operating system
set GOARCH=amd64
set GOOS=windows
set CGO_ENABLED=0

REM For modern processors
set GOAMD64=v3

REM Build the Go application with optimization flags
go build -ldflags="-s -w" -gcflags="all=-B -l=120" -trimpath -o "bin/hashimg.exe" "cmd/hashimg/main.go"

REM Optionally, you can specify the output binary name
REM go build -o myapp -ldflags="-s -w" -gcflags="-B" -trimpath

echo Build completed with optimizations.
