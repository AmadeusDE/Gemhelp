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
    @{{go}} build -o build/gemhelp ./src
    @echo "Creating symlinks for man, tldr, and wiki..."
    @cd build && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki

test: bootstrap
    @{{go}} test -v ./src

package: bootstrap
    @echo "Running shellcheck on shell scripts..."
    @shellcheck build.sh package.sh
    @echo "Running checkbashisms on shell scripts..."
    @checkbashisms build.sh package.sh
    @echo "Cleaning workspace..."
    @rm -rf build
    @mkdir -p build
    @echo "Compiling binaries..."
    @{{go}} build -o build/gemhelp ./src
    @cd build && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki
    @echo "Packaging binaries..."
    @mkdir -p build/gemhelp-bin-v{{version}}
    @cp build/gemhelp build/gemhelp-bin-v{{version}}/
    @cd build/gemhelp-bin-v{{version}} && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki
    @tar -C build -czf build/gemhelp-bin-v{{version}}.tar.gz gemhelp-bin-v{{version}}
    @tar -C build -cf - gemhelp-bin-v{{version}} | xz -c > build/gemhelp-bin-v{{version}}.tar.xz
    @tar -C build -cf - gemhelp-bin-v{{version}} | lzma -c > build/gemhelp-bin-v{{version}}.tar.lzma
    @rm -rf build/gemhelp-bin-v{{version}}
    @echo "Packaging source code..."
    @mkdir -p build/gemhelp-{{version}}
    @cp go.mod go.sum Makefile Justfile build.sh package.sh LICENSE README.md AGENTS.md PKGBUILD build/gemhelp-{{version}}/
    @cp -r src build/gemhelp-{{version}}/
    @tar -C build -czf build/gemhelp-{{version}}.tar.gz gemhelp-{{version}}
    @tar -C build -cf - gemhelp-{{version}} | xz -c > build/gemhelp-{{version}}.tar.xz
    @tar -C build -cf - gemhelp-{{version}} | lzma -c > build/gemhelp-{{version}}.tar.lzma
    @rm -rf build/gemhelp-{{version}}
    @echo "Packaging completed successfully."

clean:
    rm -rf build
    rm -rf go
