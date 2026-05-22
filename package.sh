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
mkdir -p "${BIN_DIR}/usr/share/man/man1"
cp build/gemhelp "${BIN_DIR}/"
cp gemhelp.1 "${BIN_DIR}/usr/share/man/man1/"
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
for file in go.mod go.sum Makefile Justfile build.sh package.sh LICENSE README.md AGENTS.md PKGBUILD gemhelp.1; do
	if [ -f "$file" ]; then
		cp "$file" "${SRC_DIR}/"
	fi
done

# Copy cmd and packaging directories
cp -r cmd "${SRC_DIR}/"
cp -r packaging "${SRC_DIR}/"

# Compress source directory into build/
tar -C build -czf "build/gemhelp-${VERSION}.tar.gz" "gemhelp-${VERSION}"
tar -C build -cf - "gemhelp-${VERSION}" | xz -c > "build/gemhelp-${VERSION}.tar.xz"
tar -C build -cf - "gemhelp-${VERSION}" | lzma -c > "build/gemhelp-${VERSION}.tar.lzma"
rm -rf "${SRC_DIR}"

# 6. Package Ms. Pac-Man (ms) Packages (using lzma compression)
echo "Packaging Ms. Pac-Man (ms) packages..."

# Determine ms version format (r<rev_count>.<commit>)
GIT_REV_COUNT=$(git rev-list --count HEAD 2>/dev/null || echo "0")
if [ "$GIT_REV_COUNT" -ne 0 ]; then
	VERSION_MS="r${GIT_REV_COUNT}.${VERSION}"
else
	VERSION_MS="${VERSION}"
fi

MS_GEMHELP_DIR="build/ms-gemhelp"
MS_TLDR_DIR="build/ms-gemhelp-tldr"
MS_WIKI_DIR="build/ms-gemhelp-wiki"
MS_MAN_DIR="build/ms-gemhelp-man"

rm -rf "${MS_GEMHELP_DIR}" "${MS_TLDR_DIR}" "${MS_WIKI_DIR}" "${MS_MAN_DIR}"
mkdir -p "${MS_GEMHELP_DIR}/usr/bin" "${MS_GEMHELP_DIR}/usr/share/man/man1" "${MS_TLDR_DIR}" "${MS_WIKI_DIR}" "${MS_MAN_DIR}"

# A. gemhelp package
cp build/gemhelp "${MS_GEMHELP_DIR}/usr/bin/gemhelp"
cp gemhelp.1 "${MS_GEMHELP_DIR}/usr/share/man/man1/gemhelp.1"
sed "s/@VERSION@/${VERSION_MS}/g" packaging/ms/gemhelp/metadata.yaml.in > "${MS_GEMHELP_DIR}/metadata.yaml"

# B. gemhelp-tldr package
cp packaging/ms/gemhelp-tldr/metadata.yaml "${MS_TLDR_DIR}/metadata.yaml"

# C. gemhelp-wiki package
cp packaging/ms/gemhelp-wiki/metadata.yaml "${MS_WIKI_DIR}/metadata.yaml"

# D. gemhelp-man package
cp packaging/ms/gemhelp-man/metadata.yaml "${MS_MAN_DIR}/metadata.yaml"

# Compress them to .ms.tar.lzma
tar -C "${MS_GEMHELP_DIR}" -cf - metadata.yaml usr | lzma -c > "build/gemhelp.ms.tar.lzma"
tar -C "${MS_TLDR_DIR}" -cf - metadata.yaml | lzma -c > "build/gemhelp-tldr.ms.tar.lzma"
tar -C "${MS_WIKI_DIR}" -cf - metadata.yaml | lzma -c > "build/gemhelp-wiki.ms.tar.lzma"
tar -C "${MS_MAN_DIR}" -cf - metadata.yaml | lzma -c > "build/gemhelp-man.ms.tar.lzma"

# Clean up ms temp directories
rm -rf "${MS_GEMHELP_DIR}" "${MS_TLDR_DIR}" "${MS_WIKI_DIR}" "${MS_MAN_DIR}"

echo "Packaging completed successfully."
echo "Binary archives: build/gemhelp-bin-v${VERSION}.tar.{gz,xz,lzma}"
echo "Source archives: build/gemhelp-${VERSION}.tar.{gz,xz,lzma}"
echo "Ms. Pac-Man archives: build/gemhelp.ms.tar.lzma, build/gemhelp-tldr.ms.tar.lzma, build/gemhelp-wiki.ms.tar.lzma, build/gemhelp-man.ms.tar.lzma"

