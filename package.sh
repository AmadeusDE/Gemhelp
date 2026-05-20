#!/bin/sh
set -e

# 1. Run shellcheck and checkbashisms
echo "Running shellcheck on shell scripts..."
shellcheck build.sh package.sh

echo "Running checkbashisms on shell scripts..."
checkbashisms build.sh package.sh

# 2. Run clean to start from a clean state
echo "Cleaning workspace..."
rm -rf build
mkdir -p build

# 3. Build release binaries
echo "Compiling binaries..."
sh build.sh

# Define version from git commit
VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "NOCOMMIT")

# 4. Package Binary Tarballs
echo "Packaging binaries..."
BIN_DIR="build/gemhelp-bin-v${VERSION}"
mkdir -p "${BIN_DIR}"
cp build/gemhelp "${BIN_DIR}/"
# Create relative symlinks for man, tldr, and wiki
(cd "${BIN_DIR}" && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki)

# Compress binary directory into build/
tar -C build -czf "build/gemhelp-bin-v${VERSION}.tar.gz" "gemhelp-bin-v${VERSION}"
tar -C build -cf - "gemhelp-bin-v${VERSION}" | xz -c > "build/gemhelp-bin-v${VERSION}.tar.xz"
tar -C build -cf - "gemhelp-bin-v${VERSION}" | lzma -c > "build/gemhelp-bin-v${VERSION}.tar.lzma"
rm -rf "${BIN_DIR}"

# 5. Package Source Tarballs
echo "Packaging source code..."
SRC_DIR="build/gemhelp-${VERSION}"
mkdir -p "${SRC_DIR}"

# Copy root configuration and metadata files
for file in go.mod go.sum Makefile Justfile build.sh package.sh LICENSE README.md PKGBUILD; do
	if [ -f "$file" ]; then
		cp "$file" "${SRC_DIR}/"
	fi
done

# Copy src directory
cp -r src "${SRC_DIR}/"

# Compress source directory into build/
tar -C build -czf "build/gemhelp-${VERSION}.tar.gz" "gemhelp-${VERSION}"
tar -C build -cf - "gemhelp-${VERSION}" | xz -c > "build/gemhelp-${VERSION}.tar.xz"
tar -C build -cf - "gemhelp-${VERSION}" | lzma -c > "build/gemhelp-${VERSION}.tar.lzma"
rm -rf "${SRC_DIR}"

echo "Packaging completed successfully."
echo "Binary archives: build/gemhelp-bin-v${VERSION}.tar.{gz,xz,lzma}"
echo "Source archives: build/gemhelp-${VERSION}.tar.{gz,xz,lzma}"
