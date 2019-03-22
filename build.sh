#!/bin/bash
set -e

cd "$(dirname "$0")"
rm -rfv build
mkdir -pv build

build() {
    export "GOOS=$1"
    export "GOARCH=$2"
    export "SUFFIX=$3"

    echo "> Building for ${GOOS}-${GOARCH} as ${SUFFIX}"
    echo "-- dictlookup"
    go build -o "./build/dictlookup_${SUFFIX}" -ldflags "-s -w" "./tools/dictlookup"
    echo "-- dictparse"
    go build -o "./build/dictparse_${SUFFIX}" -ldflags "-s -w" "./tools/dictparse"
    echo "-- dictverify"
    go build -o "./build/dictverify_${SUFFIX}" -ldflags "-s -w" "./tools/dictverify"
    echo "-- dictserver"
    go build -o "./build/dictserver_${SUFFIX}" -ldflags "-s -w -X main.version=$(git describe --tags --always)" "."
}

build linux amd64 linux-x64
build linux arm linux-arm
build windows 386 windows
build darwin amd64 darwin-x64

echo "> Building index"
GOOS="" GOARCH="" go run ./tools/dictparse ./data/dictionary.txt ./build/dict

echo "> Verifying index"
GOOS="" GOARCH="" go run ./tools/dictverify ./build/dict
