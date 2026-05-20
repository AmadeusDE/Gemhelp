# Gemhelp

Gemhelp is an intelligent terminal helper utility written in Go. It integrates with the Gemini API using function calling to read system man pages, resolve offline TLDR examples, and search the Arch Wiki to answer natural language questions about terminal commands.

## Features
*   **Gemini Tool Calling Loop**: Intelligent resolution using custom system-integrated tools.
*   **Offline Man Page Parser**: Directly parses compressed (`.gz`) and raw man pages on Arch and other Linux distributions.
*   **Offline TLDR Database**: Downloads and extracts standard/language-specific TLDR zips, resolving them locally.
*   **Arch Wiki Client**: MediaWiki JSON client that fetches wikitext and strips template clutter.
*   **Multi-Call Binary Support**: Can act as a drop-in offline alternative to `man` or `tldr` when symlinked or executed as those names.
*   **Model Fallback & Backoff**: Automatically fallbacks from `gemini-3.5-flash` to `gemini-3.1-flash-lite` and `gemini-2.5-flash` with exponential backoffs on rate limits.
*   **Response Caching**: Instantaneous lookups using SHA-256 hashes of queries.

---

## Installation

### Standard Go Installation
To install the latest release directly via Go's package manager:
```bash
go install github.com/AmadeusDE/gemhelp@latest
```

### Pacman Installation
```bash
makepkg -si
```

# Build from Source
If Go is installed on your system, or to bootstrap it automatically:
## Clone the repository
git clone https://github.com/AmadeusDE/gemhelp.git
cd gemhelp

### Using go build
`go build -o build/gemhelp ./src`

## Or other options (downloads Go if not present in PATH)
### Using Makefile
`make`

### Using Justfile
`just`

### Or just run the build script
`./build.sh`

---

## Usage

### Natural Language Help
Query Gemini with a command name and your question:
```bash
gemhelp ls how to show hidden files and sort by size
```

### Standalone Subcommands
You can call the offline readers directly via flags or subcommands:
```bash
# View offline man page
gemhelp man pacman
gemhelp --man pacman

# View offline TLDR cheatsheet
gemhelp tldr pacman
gemhelp --tldr pacman

# Search/View Arch Wiki page
gemhelp wiki pacman
gemhelp --wiki pacman
```

### Multi-Call Binary Setup
If you symlink the compiled binary to `tldr` or `man` inside your `$PATH` (e.g., in `~/.local/bin`), it behaves directly as those subcommands:
```bash
# Behave as offline manpage reader
ln -s /usr/bin/gemhelp ~/.local/bin/man
man pacman

# Behave as offline TLDR cheatsheet reader
ln -s /usr/bin/gemhelp ~/.local/bin/tldr
tldr pacman

# Behave as offline Arch Wiki reader
ln -s /usr/bin/gemhelp ~/.local/bin/wiki
wiki pacman
```

---

## Configuration & Storage Paths
*   **Configuration**: `~/.config/gemhelp/config.json` (permissions `0600`).
*   **TLDR Archive Cache**: `~/.local/share/gemhelp/tldr/`
*   **Response Cache**: `~/.cache/gemhelp/responses/`

---

## License & Code of Conduct
This project is licensed under the [Open Software License (OSL) v. 3.0](LICENSE).
Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).