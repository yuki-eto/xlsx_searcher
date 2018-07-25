#!/usr/bin/env bash

BASE_PATH="$(cd $(dirname $0) && pwd)/.."
ROOT_DIR="/go/src/github.com/yuki-eto/xlsx_searcher"

export GOARCH=amd64
for GOOS in darwin windows linux; do
	export GOOS
	go build -v -o ${BASE_PATH}/build/xs_${GOOS} ${BASE_PATH}/cmd/xs.go
done
