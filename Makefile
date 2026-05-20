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
	mkdir -p build/gemhelp-bin-v$$VERSION; \
	cp build/gemhelp build/gemhelp-bin-v$$VERSION/; \
	cd build/gemhelp-bin-v$$VERSION && ln -sf gemhelp man && ln -sf gemhelp tldr && ln -sf gemhelp wiki; \
	cd ../..; \
	tar -C build -czf build/gemhelp-bin-v$$VERSION.tar.gz gemhelp-bin-v$$VERSION; \
	tar -C build -cf - gemhelp-bin-v$$VERSION | xz -c > build/gemhelp-bin-v$$VERSION.tar.xz; \
	tar -C build -cf - gemhelp-bin-v$$VERSION | lzma -c > build/gemhelp-bin-v$$VERSION.tar.lzma; \
	rm -rf build/gemhelp-bin-v$$VERSION; \
	mkdir -p build/gemhelp-$$VERSION; \
	cp go.mod go.sum Makefile Justfile build.sh package.sh LICENSE README.md AGENTS.md PKGBUILD build/gemhelp-$$VERSION/; \
	cp -r src build/gemhelp-$$VERSION/; \
	tar -C build -czf build/gemhelp-$$VERSION.tar.gz gemhelp-$$VERSION; \
	tar -C build -cf - gemhelp-$$VERSION | xz -c > build/gemhelp-$$VERSION.tar.xz; \
	tar -C build -cf - gemhelp-$$VERSION | lzma -c > build/gemhelp-$$VERSION.tar.lzma; \
	rm -rf build/gemhelp-$$VERSION
	@echo "Packaging completed successfully."

clean:
	rm -rf build
	rm -rf go
