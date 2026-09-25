#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

LDFLAGS="-s -w -H windowsgui"

echo "==> Building collector (Windows amd64) -> dist/360ti-hwinfo.exe"
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/360ti-hwinfo.exe .
echo "==> Building collector (Windows 386) -> dist/360ti-hwinfo-386.exe"
GOOS=windows GOARCH=386 go build -ldflags "$LDFLAGS" -o dist/360ti-hwinfo-386.exe .

echo "==> Done:"
ls -lh dist/