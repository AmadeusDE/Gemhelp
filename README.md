# Gemhelp

Gemhelp is an intelligent terminal helper utility written in Go. It integrates with the Gemini API using function calling to read system man pages, resolve offline TLDR examples, and search the Arch Wiki to answer natural language questions about terminal commands.

## Features
*   **Gemini Tool Calling Loop**: Intelligent resolution using custom system-integrated tools.
*   **Offline Man Page Parser**: Directly parses compressed (`.gz`) and raw man pages on Arch and other Linux distributions.
*   **Offline TLDR Database**: Downloads and extracts standard/language-specific TLDR zips, resolving them locally.
*   **Arch Wiki Client**: MediaWiki JSON client that fetches wikitext and strips template clutter.
*   **Multi-Call Binary Support**: Can act as a drop-in offline alternative to `man` or `tldr` when symlinked or executed as those names.
*   **Model Fallback & Backoff**: Automatically fallbacks from `gemini-3.5-flash` to `gemini-3.1-flash-lite` and `gemini-2.5-flash` with exponential backoffs on rate limits.
*   **Command Correction**: `thefuck`-style command correction using Gemini, with shell integration for bash, zsh, and mksh.
*   **Response Caching**: Instantaneous lookups using SHA-256 hashes of queries.

---

## Installation

### Standard Go Installation (Remote)
To install the latest release directly via Go's package manager:
```bash
go install github.com/AmadeusDE/Gemhelp/cmd/gemhelp@latest
```

### From Source (Local)
If you have cloned the repository:
```bash
go install ./cmd/gemhelp
```

### Pacman Installation (Arch Linux)
```bash
makepkg -si
```

# Build from Source
If Go is not installed on your system, you can use the bootstrap script:
## Clone the repository
git clone https://github.com/AmadeusDE/gemhelp.git
cd gemhelp

### Using the build script (downloads Go if not present in PATH)
`./build.sh`

### Using Makefile
`make`

### Using Justfile
`just`

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

### Command Correction (`fuck`)
Correct a failed command using Gemini, similar to [thefuck](https://github.com/nvbn/thefuck):
```bash
# Direct usage — prints corrected command
gemhelp fuck "pacman -s firefox"
# → pacman -S firefox

# With error output piped in
some_command 2>&1 | gemhelp fuck "some_command"
```

### Shell Integration
Add one of the following lines to your shell's rc file to get a `fuck()` function that helps you correct your last failed command:

#### Bash (`~/.bashrc`)
```bash
eval "$(gemhelp --init-shell bash)"
```

#### Zsh (`~/.zshrc`)
```zsh
eval "$(gemhelp --init-shell zsh)"
```

#### mksh (`~/.mkshrc`)
```sh
eval "$(gemhelp --init-shell mksh)"
```

After sourcing, simply type `fuck` after a failed command. For safety, the tool will **not** execute the command automatically:
*   **Zsh**: The corrected command is placed directly into your input buffer.
*   **Bash**: The corrected command is added to your history; press `[Up]` to review and run it.

```
$ pacman -s firefox
error: ...
$ fuck
Fixing: pacman -s firefox
# [Zsh: command buffer is now populated with 'pacman -S firefox']
# [Bash: 'pacman -S firefox' added to history]
```

---

## Configuration & Storage Paths
*   **Configuration**: `~/.config/gemhelp/config.json` (permissions `0600`).
*   **TLDR Archive Cache**: `~/.local/share/gemhelp/tldr/`
*   **Response Cache**: `~/.cache/gemhelp/responses/`
*   **Man Page Directory Override**: By default, man pages are searched under `/usr/share/man`. You can override this search path by setting the `GEMHELP_MAN_DIR` environment variable.

---

## License & Code of Conduct
This project is licensed under the [Open Software License (OSL) v. 3.0](LICENSE).
Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).