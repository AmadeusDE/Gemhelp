.PHONY: all build clean test package

all: build

build:
	@if command -v go >/dev/null 2>&1; then \
		echo "Building with system Go..."; \
		mkdir -p build; \
		go build -o build/gemhelp ./cmd/gemhelp; \
	else \
		if [ ! -f go/bin/go ]; then \
			echo "Go not found in PATH. Bootstrapping Go 1.26.3..."; \
			UNAME_S=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
			UNAME_M=$$(uname -m); \
			ARCH="amd64"; \
			if [ "$$UNAME_M" = "aarch64" ] || [ "$$UNAME_M" = "arm64" ]; then ARCH="arm64"; \
			elif [ "$$UNAME_M" = "i386" ] || [ "$$UNAME_M" = "i686" ]; then ARCH="386"; fi; \
			curl -sSL -o go_bootstrap.tar.gz https://go.dev/dl/go1.26.3.$$UNAME_S-$$ARCH.tar.gz || wget -qO go_bootstrap.tar.gz https://go.dev/dl/go1.26.3.$$UNAME_S-$$ARCH.tar.gz; \
			tar -xzf go_bootstrap.tar.gz; \
			rm go_bootstrap.tar.gz; \
		fi; \
		echo "Building with bootstrapped Go..."; \
		mkdir -p build; \
		./go/bin/go build -o build/gemhelp ./cmd/gemhelp; \
	fi
	@echo "Creating symlinks for man, tldr, and wiki..."
	@cd build && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki

test:
	@if command -v go >/dev/null 2>&1; then \
		go test -v ./cmd/gemhelp; \
	elif [ -f go/bin/go ]; then \
		./go/bin/go test -v ./cmd/gemhelp; \
	else \
		echo "Go not found in PATH or bootstrapped folder. Running build to bootstrap..."; \
		$(MAKE) build; \
		if command -v go >/dev/null 2>&1; then \
			go test -v ./cmd/gemhelp; \
		else \
			./go/bin/go test -v ./cmd/gemhelp; \
		fi \
	fi

package:
	@echo "Running shellcheck on shell scripts..."
	@shellcheck build.sh package.sh
	@echo "Running checkbashisms on shell scripts..."
	@checkbashisms build.sh package.sh
	@echo "Cleaning workspace..."
	@rm -rf build
	@mkdir -p build
	@echo "Compiling binaries..."
	@if command -v go >/dev/null 2>&1; then \
		go build -o build/gemhelp ./cmd/gemhelp; \
	else \
		if [ ! -f go/bin/go ]; then \
			echo "Go not found in PATH. Bootstrapping Go 1.26.3..."; \
			UNAME_S=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
			UNAME_M=$$(uname -m); \
			ARCH="amd64"; \
			if [ "$$UNAME_M" = "aarch64" ] || [ "$$UNAME_M" = "arm64" ]; then ARCH="arm64"; \
			elif [ "$$UNAME_M" = "i386" ] || [ "$$UNAME_M" = "i686" ]; then ARCH="386"; fi; \
			curl -sSL -o go_bootstrap.tar.gz https://go.dev/dl/go1.26.3.$$UNAME_S-$$ARCH.tar.gz || wget -qO go_bootstrap.tar.gz https://go.dev/dl/go1.26.3.$$UNAME_S-$$ARCH.tar.gz; \
			tar -xzf go_bootstrap.tar.gz; \
			rm go_bootstrap.tar.gz; \
		fi; \
		./go/bin/go build -o build/gemhelp ./cmd/gemhelp; \
	fi
	@cd build && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki
	@echo "Packaging binaries and source code..."
	@VERSION=$$(git rev-parse --short HEAD 2>/dev/null || echo NOCOMMIT); \
	echo "Version: $$VERSION"; \
	mkdir -p build/gemhelp-bin-v$$VERSION/usr/share/man/man1; \
	cp build/gemhelp build/gemhelp-bin-v$$VERSION/; \
	cp gemhelp.1 build/gemhelp-bin-v$$VERSION/usr/share/man/man1/; \
	cd build/gemhelp-bin-v$$VERSION && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki; \
	cd ../..; \
	tar -C build -czf build/gemhelp-bin-v$$VERSION.tar.gz gemhelp-bin-v$$VERSION; \
	tar -C build -cf - gemhelp-bin-v$$VERSION | xz -c > build/gemhelp-bin-v$$VERSION.tar.xz; \
	tar -C build -cf - gemhelp-bin-v$$VERSION | lzma -c > build/gemhelp-bin-v$$VERSION.tar.lzma; \
	rm -rf build/gemhelp-bin-v$$VERSION; \
	mkdir -p build/gemhelp-$$VERSION; \
	cp go.mod go.sum Makefile Justfile build.sh package.sh LICENSE README.md AGENTS.md PKGBUILD gemhelp.1 build/gemhelp-$$VERSION/; \
	cp -r cmd packaging build/gemhelp-$$VERSION/; \
	tar -C build -czf build/gemhelp-$$VERSION.tar.gz gemhelp-$$VERSION; \
	tar -C build -cf - gemhelp-$$VERSION | xz -c > build/gemhelp-$$VERSION.tar.xz; \
	tar -C build -cf - gemhelp-$$VERSION | lzma -c > build/gemhelp-$$VERSION.tar.lzma; \
	rm -rf build/gemhelp-$$VERSION; \
	echo "Packaging Ms. Pac-Man (ms) packages..."; \
	GIT_REV_COUNT=$$(git rev-list --count HEAD 2>/dev/null || echo 0); \
	if [ "$$GIT_REV_COUNT" -ne 0 ]; then \
		VERSION_MS="r$$GIT_REV_COUNT.$$VERSION"; \
	else \
		VERSION_MS="$$VERSION"; \
	fi; \
	MS_GEMHELP_DIR="build/ms-gemhelp"; \
	MS_TLDR_DIR="build/ms-gemhelp-tldr"; \
	MS_WIKI_DIR="build/ms-gemhelp-wiki"; \
	MS_MAN_DIR="build/ms-gemhelp-man"; \
	rm -rf "$$MS_GEMHELP_DIR" "$$MS_TLDR_DIR" "$$MS_WIKI_DIR" "$$MS_MAN_DIR"; \
	mkdir -p "$$MS_GEMHELP_DIR/usr/bin" "$$MS_GEMHELP_DIR/usr/share/man/man1" "$$MS_TLDR_DIR" "$$MS_WIKI_DIR" "$$MS_MAN_DIR"; \
	cp build/gemhelp "$$MS_GEMHELP_DIR/usr/bin/gemhelp"; \
	cp gemhelp.1 "$$MS_GEMHELP_DIR/usr/share/man/man1/gemhelp.1"; \
	sed "s/@VERSION@/$$VERSION_MS/g" packaging/ms/gemhelp/metadata.yaml.in > "$$MS_GEMHELP_DIR/metadata.yaml"; \
	cp packaging/ms/gemhelp-tldr/metadata.yaml "$$MS_TLDR_DIR/metadata.yaml"; \
	cp packaging/ms/gemhelp-wiki/metadata.yaml "$$MS_WIKI_DIR/metadata.yaml"; \
	cp packaging/ms/gemhelp-man/metadata.yaml "$$MS_MAN_DIR/metadata.yaml"; \
	tar -C "$$MS_GEMHELP_DIR" -cf - metadata.yaml usr | lzma -c > build/gemhelp.ms.tar.lzma; \
	tar -C "$$MS_TLDR_DIR" -cf - metadata.yaml | lzma -c > build/gemhelp-tldr.ms.tar.lzma; \
	tar -C "$$MS_WIKI_DIR" -cf - metadata.yaml | lzma -c > build/gemhelp-wiki.ms.tar.lzma; \
	tar -C "$$MS_MAN_DIR" -cf - metadata.yaml | lzma -c > build/gemhelp-man.ms.tar.lzma; \
	rm -rf "$$MS_GEMHELP_DIR" "$$MS_TLDR_DIR" "$$MS_WIKI_DIR" "$$MS_MAN_DIR"; \
	echo "Packaging completed successfully."; \
	echo "Binary archives: build/gemhelp-bin-v$$VERSION.tar.{gz,xz,lzma}"; \
	echo "Source archives: build/gemhelp-$$VERSION.tar.{gz,xz,lzma}"; \
	echo "Ms. Pac-Man archives: build/gemhelp.ms.tar.lzma, build/gemhelp-tldr.ms.tar.lzma, build/gemhelp-wiki.ms.tar.lzma, build/gemhelp-man.ms.tar.lzma"

clean:
	rm -rf build
	rm -rf go
