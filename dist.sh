#!/bin/sh
set -e

if ! [ -e logtail.go ]
then
	echo "Execute ./dist.sh within the logtail directory."
	exit 1
fi

dist() {
	GOOS=$1 GOARCH=$2 go build
	tar -czf logtail-$1-$2.tar.gz logtail
}

dist linux amd64
dist linux arm64
dist linux mips
