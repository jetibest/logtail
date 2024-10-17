#!/bin/sh

if ! [ -e logtail.go ]
then
	echo "Execute ./compile.sh within the logtail directory."
	exit 1
fi

# cross-compile using: GOOS=linux GOARCH=amd64 go build
# for other platforms, use for example: arm64, mips, etc.

go build || exit 1

echo "Compilation complete, install with:"
echo "  "
echo "  sudo cp ./logtail /usr/local/bin/logtail"
echo "  "
