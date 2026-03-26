#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "Building backend binary for Linux amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o transfer-server ./cmd/server

echo "Done: $(pwd)/transfer-server"
