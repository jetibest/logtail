#!/bin/sh
set -e

if ! [ -e logtail.go ]
then
	echo "Execute ./compile.sh within the logtail directory."
	exit 1
fi

go build || exit 1

echo "Compilation complete, install with:"
echo "  "
echo "  sudo cp ./bin/logtail /usr/local/bin/logtail"
echo "  "
