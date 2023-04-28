#!/bin/bash
DATE=`date +RELEASE.%Y-%m-%dT%H-%M-%SZ`; env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.VERSION=${DATE}" -o bin/badger.linux.amd64
DATE=`date +RELEASE.%Y-%m-%dT%H-%M-%SZ`; env GOOS=linux GOARCH=arm64 go build -ldflags "-X main.VERSION=${DATE}" -o bin/badger.linux.arm64
