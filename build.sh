#!/bin/sh
echo "Building amd64 version..."
CGO_ENABLED=0 GOOS=linux go build -o thinkbacklight -a -ldflags '-s -w -extldflags "-static"' thinkbacklight.go
echo "Done."
echo "comperessing binary with UPX"
upx --ultra-brute --best thinkbacklight
ls -lh thinkbacklight

