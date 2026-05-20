#!/bin/sh
set -e

# POSIX shell script to compile gemhelp, bootstrapping Go 1.26.3 if not in PATH.

# Check if go is in path
if command -v go >/dev/null 2>&1; then
	GO_BIN="go"
else
	# Bootstrap Go if not present
	if [ ! -f "go/bin/go" ]; then
		echo "Go not found in PATH. Bootstrapping Go 1.26.3..."
		UNAME_S=$(uname -s | tr '[:upper:]' '[:lower:]')
		UNAME_M=$(uname -m)
		ARCH="amd64"
		if [ "$UNAME_M" = "aarch64" ] || [ "$UNAME_M" = "arm64" ]; then
			ARCH="arm64"
		elif [ "$UNAME_M" = "i386" ] || [ "$UNAME_M" = "i686" ]; then
			ARCH="386"
		fi

		URL="https://go.dev/dl/go1.26.3.${UNAME_S}-${ARCH}.tar.gz"
		echo "Downloading bootstrap Go from ${URL}..."
		if command -v wget >/dev/null 2>&1; then
			wget -qO go_bootstrap.tar.gz "${URL}"
		elif command -v curl >/dev/null 2>&1; then
			curl -sSL -o go_bootstrap.tar.gz "${URL}"
		else
			echo "Error: Neither wget nor curl found. Cannot download Go." >&2
			exit 1
		fi

		tar -xzf go_bootstrap.tar.gz
		rm go_bootstrap.tar.gz
	fi
	GO_BIN="./go/bin/go"
fi

mkdir -p build
echo "Building gemhelp..."
$GO_BIN build -o build/gemhelp ./src

echo "Creating symlinks for man, tldr, and wiki..."
(cd build && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki)

echo "Build successful."
