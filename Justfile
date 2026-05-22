version := `git rev-parse --short HEAD 2>/dev/null || echo "NOCOMMIT"`
go := `command -v go || echo "./go/bin/go"`

default: build

# Private bootstrap target
[private]
bootstrap:
    @if [ ! -f "{{go}}" ]; then \
        echo "Go not found in PATH. Bootstrapping Go 1.26.3..."; \
        UNAME_S=$(uname -s | tr '[:upper:]' '[:lower:]'); \
        UNAME_M=$(uname -m); \
        ARCH="amd64"; \
        if [ "$UNAME_M" = "aarch64" ] || [ "$UNAME_M" = "arm64" ]; then ARCH="arm64"; \
        elif [ "$UNAME_M" = "i386" ] || [ "$UNAME_M" = "i686" ]; then ARCH="386"; fi; \
        curl -sSL -o go_bootstrap.tar.gz "https://go.dev/dl/go1.26.3.${UNAME_S}-${ARCH}.tar.gz" || wget -qO go_bootstrap.tar.gz "https://go.dev/dl/go1.26.3.${UNAME_S}-${ARCH}.tar.gz"; \
        tar -xzf go_bootstrap.tar.gz; \
        rm go_bootstrap.tar.gz; \
    fi

build: bootstrap
    @mkdir -p build
    @echo "Building gemhelp..."
    @{{go}} build -o build/gemhelp ./cmd/gemhelp
    @echo "Creating symlinks for man, tldr, and wiki..."
    @cd build && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki

test: bootstrap
    @{{go}} test -v ./cmd/gemhelp

package: bootstrap
    @echo "Running shellcheck on shell scripts..."
    @shellcheck build.sh package.sh
    @echo "Running checkbashisms on shell scripts..."
    @checkbashisms build.sh package.sh
    @echo "Cleaning workspace..."
    @rm -rf build
    @mkdir -p build
    @echo "Compiling binaries..."
    @{{go}} build -o build/gemhelp ./cmd/gemhelp
    @cd build && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki
    @echo "Packaging binaries..."
    @mkdir -p build/gemhelp-bin-v{{version}}/usr/share/man/man1
    @cp build/gemhelp build/gemhelp-bin-v{{version}}/
    @cp gemhelp.1 build/gemhelp-bin-v{{version}}/usr/share/man/man1/
    @cd build/gemhelp-bin-v{{version}} && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki
    @tar -C build -czf build/gemhelp-bin-v{{version}}.tar.gz gemhelp-bin-v{{version}}
    @tar -C build -cf - gemhelp-bin-v{{version}} | xz -c > build/gemhelp-bin-v{{version}}.tar.xz
    @tar -C build -cf - gemhelp-bin-v{{version}} | lzma -c > build/gemhelp-bin-v{{version}}.tar.lzma
    @rm -rf build/gemhelp-bin-v{{version}}
    @echo "Packaging source code..."
    @mkdir -p build/gemhelp-{{version}}
    @cp go.mod go.sum Makefile Justfile build.sh package.sh LICENSE README.md AGENTS.md PKGBUILD gemhelp.1 build/gemhelp-{{version}}/
    @cp -r cmd packaging build/gemhelp-{{version}}/
    @tar -C build -czf build/gemhelp-{{version}}.tar.gz gemhelp-{{version}}
    @tar -C build -cf - gemhelp-{{version}} | xz -c > build/gemhelp-{{version}}.tar.xz
    @tar -C build -cf - gemhelp-{{version}} | lzma -c > build/gemhelp-{{version}}.tar.lzma
    @rm -rf build/gemhelp-{{version}}
    @echo "Packaging Ms. Pac-Man (ms) packages..."
    @VERSION_VAL="{{version}}"; \
    GIT_REV_COUNT=$(git rev-list --count HEAD 2>/dev/null || echo "0"); \
    if [ "$GIT_REV_COUNT" -ne 0 ]; then \
        VERSION_MS="r${GIT_REV_COUNT}.${VERSION_VAL}"; \
    else \
        VERSION_MS="${VERSION_VAL}"; \
    fi; \
    MS_GEMHELP_DIR="build/ms-gemhelp"; \
    MS_TLDR_DIR="build/ms-gemhelp-tldr"; \
    MS_WIKI_DIR="build/ms-gemhelp-wiki"; \
    MS_MAN_DIR="build/ms-gemhelp-man"; \
    rm -rf "$MS_GEMHELP_DIR" "$MS_TLDR_DIR" "$MS_WIKI_DIR" "$MS_MAN_DIR"; \
    mkdir -p "$MS_GEMHELP_DIR/usr/bin" "$MS_GEMHELP_DIR/usr/share/man/man1" "$MS_TLDR_DIR" "$MS_WIKI_DIR" "$MS_MAN_DIR"; \
    cp build/gemhelp "$MS_GEMHELP_DIR/usr/bin/gemhelp"; \
    cp gemhelp.1 "$MS_GEMHELP_DIR/usr/share/man/man1/gemhelp.1"; \
    sed "s/@VERSION@/$VERSION_MS/g" packaging/ms/gemhelp/metadata.yaml.in > "$MS_GEMHELP_DIR/metadata.yaml"; \
    cp packaging/ms/gemhelp-tldr/metadata.yaml "$MS_TLDR_DIR/metadata.yaml"; \
    cp packaging/ms/gemhelp-wiki/metadata.yaml "$MS_WIKI_DIR/metadata.yaml"; \
    cp packaging/ms/gemhelp-man/metadata.yaml "$MS_MAN_DIR/metadata.yaml"; \
    tar -C "$MS_GEMHELP_DIR" -cf - metadata.yaml usr | lzma -c > build/gemhelp.ms.tar.lzma; \
    tar -C "$MS_TLDR_DIR" -cf - metadata.yaml | lzma -c > build/gemhelp-tldr.ms.tar.lzma; \
    tar -C "$MS_WIKI_DIR" -cf - metadata.yaml | lzma -c > build/gemhelp-wiki.ms.tar.lzma; \
    tar -C "$MS_MAN_DIR" -cf - metadata.yaml | lzma -c > build/gemhelp-man.ms.tar.lzma; \
    rm -rf "$MS_GEMHELP_DIR" "$MS_TLDR_DIR" "$MS_WIKI_DIR" "$MS_MAN_DIR"
    @echo "Packaging completed successfully."

clean:
    rm -rf build
    rm -rf go
