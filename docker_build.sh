#!/usr/bin/env bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
docker build -t bililive-go .
rm bililive-go