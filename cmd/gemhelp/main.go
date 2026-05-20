package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	ctx := context.Background()

	// Multi-call binary detection based on the base name of the executable
	progName := filepath.Base(os.Args[0])
	if progName == "tldr" {
		runAsTldr()
		return
	} else if progName == "man" {
		runAsMan()
		return
	} else if progName == "wiki" {
		runAsWiki()
		return
	}

	// Standard Flag Parsing
	noCache := flag.Bool("no-cache", false, "Bypass response cache")
	flag.BoolVar(noCache, "n", false, "Bypass response cache (shorthand)")

	clearCacheFlag := flag.Bool("clear-cache", false, "Clear the response cache")

	tldrSub := flag.Bool("tldr", false, "Render the offline TLDR page directly")
	flag.BoolVar(tldrSub, "t", false, "Render the offline TLDR page directly (shorthand)")

	manSub := flag.Bool("man", false, "Render the offline parsed man page directly")
	flag.BoolVar(manSub, "m", false, "Render the offline parsed man page directly (shorthand)")

	wikiSub := flag.Bool("wiki", false, "Search the Arch Wiki directly")
	flag.BoolVar(wikiSub, "w", false, "Search the Arch Wiki directly (shorthand)")

	fuckSub := flag.Bool("fuck", false, "Correct a failed command using Gemini")
	flag.BoolVar(fuckSub, "f", false, "Correct a failed command using Gemini (shorthand)")

	initShell := flag.String("init-shell", "", "Print shell integration script (bash, zsh, mksh)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <command> [question...]\n", progName)
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSubcommands / Multi-call Shortcuts:\n")
		fmt.Fprintf(os.Stderr, "  %s man <command>          Render parsed man page\n", progName)
		fmt.Fprintf(os.Stderr, "  %s tldr <command>         Render offline TLDR page\n", progName)
		fmt.Fprintf(os.Stderr, "  %s wiki <query>           Search/get Arch Wiki page\n", progName)
		fmt.Fprintf(os.Stderr, "  %s fuck <command>         Correct a failed command\n", progName)
		fmt.Fprintf(os.Stderr, "\nShell Integration:\n")
		fmt.Fprintf(os.Stderr, "  eval \"$(%s --init-shell bash)\"   Add to ~/.bashrc\n", progName)
		fmt.Fprintf(os.Stderr, "  eval \"$(%s --init-shell zsh)\"    Add to ~/.zshrc\n", progName)
		fmt.Fprintf(os.Stderr, "  eval \"$(%s --init-shell mksh)\"   Add to ~/.mkshrc\n", progName)
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  gemhelp ls\n")
		fmt.Fprintf(os.Stderr, "  gemhelp ls how to sort by size\n")
		fmt.Fprintf(os.Stderr, "  gemhelp --tldr tar\n")
		fmt.Fprintf(os.Stderr, "  gemhelp fuck \"pacman -s firefox\"\n")
	}

	flag.Parse()

	// Shell integration script output
	if *initShell != "" {
		printShellIntegration(*initShell)
		return
	}

	// Clear cache and exit if requested
	if *clearCacheFlag {
		if err := ClearCache(); err != nil {
			fmt.Printf("Error clearing cache: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Response cache cleared successfully.")
		return
	}

	args := flag.Args()

	// Check if the user specified subcommands as the first argument
	if len(args) > 0 {
		sub := strings.ToLower(args[0])
		switch sub {
		case "tldr":
			if len(args) < 2 {
				fmt.Println("Error: Missing command name. Usage: gemhelp tldr <command>")
				os.Exit(1)
			}
			runOfflineTldr(args[1])
			return
		case "man":
			if len(args) < 2 {
				fmt.Println("Error: Missing command name. Usage: gemhelp man <command>")
				os.Exit(1)
			}
			runOfflineMan(args[1])
			return
		case "wiki":
			if len(args) < 2 {
				fmt.Println("Error: Missing query. Usage: gemhelp wiki <query>")
				os.Exit(1)
			}
			runDirectWiki(strings.Join(args[1:], " "))
			return
		case "fuck":
			if len(args) < 2 {
				fmt.Println("Error: Missing command. Usage: gemhelp fuck \"<failed command>\"")
				os.Exit(1)
			}
			runFuckMode(ctx, strings.Join(args[1:], " "))
			return
		}
	}

	// Flag-driven subcommands
	if *tldrSub {
		if len(args) == 0 {
			fmt.Println("Error: Missing command name for --tldr flag.")
			os.Exit(1)
		}
		runOfflineTldr(args[0])
		return
	}

	if *manSub {
		if len(args) == 0 {
			fmt.Println("Error: Missing command name for --man flag.")
			os.Exit(1)
		}
		runOfflineMan(args[0])
		return
	}

	if *wikiSub {
		if len(args) == 0 {
			fmt.Println("Error: Missing query for --wiki flag.")
			os.Exit(1)
		}
		runDirectWiki(strings.Join(args, " "))
		return
	}

	if *fuckSub {
		if len(args) == 0 {
			fmt.Println("Error: Missing command for --fuck flag.")
			os.Exit(1)
		}
		runFuckMode(ctx, strings.Join(args, " "))
		return
	}

	// Default help usage if no arguments provided
	if len(args) == 0 {
		flag.Usage()
		return
	}

	// Load or prompt configuration
	cfg, err := LoadConfig()
	if err != nil || cfg.APIKey == "" {
		cfg, err = PromptConfig(ctx)
		if err != nil {
			fmt.Printf("Setup failed: %v\n", err)
			os.Exit(1)
		}
	}

	command := args[0]
	question := ""
	if len(args) > 1 {
		question = strings.Join(args[1:], " ")
	}

	// Construct cache query string
	cacheQuery := command
	if question != "" {
		cacheQuery = command + " " + question
	}

	// Response Cache Check
	if !*noCache {
		if cachedResponse, found := GetCachedResponse(cacheQuery); found {
			fmt.Println(FormatMarkdown(cachedResponse))
			return
		}
	}

	// Invoke Gemini Session
	fmt.Printf("Querying Gemini for help with '%s'...\n", command)
	response, err := RunGeminiConversation(ctx, cfg.APIKey, cfg.Language, command, question)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print response and save to cache
	fmt.Println(FormatMarkdown(response))

	if err := SaveCachedResponse(cacheQuery, response); err != nil {
		fmt.Printf("Warning: Failed to save response to cache: %v\n", err)
	}
}

// runAsTldr executes when the binary base name is 'tldr'
func runAsTldr() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tldr <command>")
		os.Exit(1)
	}
	runOfflineTldr(os.Args[1])
}

// runAsMan executes when the binary base name is 'man'
func runAsMan() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: man <command>")
		os.Exit(1)
	}
	runOfflineMan(os.Args[1])
}

// runAsWiki executes when the binary base name is 'wiki'
func runAsWiki() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: wiki <query>")
		os.Exit(1)
	}
	runDirectWiki(strings.Join(os.Args[1:], " "))
}

func runOfflineTldr(command string) {
	cfg, err := LoadConfig()
	lang := "en"
	if err == nil && cfg.Language != "" {
		lang = cfg.Language
	}

	page, err := GetTldrPage(command, lang)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(FormatMarkdown(page))
}

func runOfflineMan(command string) {
	path, exists := FindManPagePath(command)
	if !exists {
		fmt.Printf("No manual entry for %s\n", command)
		os.Exit(1)
	}

	page, err := ParseManPage(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(FormatMarkdown(page))
}

func runDirectWiki(query string) {
	fmt.Printf("Searching Arch Wiki for '%s'...\n", query)
	titles, err := SearchArchWiki(query)
	if err != nil {
		fmt.Printf("Error searching Arch Wiki: %v\n", err)
		os.Exit(1)
	}

	if len(titles) == 0 {
		fmt.Println("No pages found on Arch Wiki matching query.")
		return
	}

	// If there's an exact match or we only got one, fetch it directly
	var targetPage string
	for _, t := range titles {
		if strings.EqualFold(t, query) {
			targetPage = t
			break
		}
	}
	if targetPage == "" {
		targetPage = titles[0]
	}

	fmt.Printf("Retrieving page '%s'...\n\n", targetPage)
	page, err := GetArchWikiPage(targetPage)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(FormatMarkdown(page))
}

func runFuckMode(ctx context.Context, failedCommand string) {
	cfg, err := LoadConfig()
	if err != nil || cfg.APIKey == "" {
		cfg, err = PromptConfig(ctx)
		if err != nil {
			fmt.Printf("Setup failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Read error output from stdin if piped
	errorOutput := readStdinIfPiped()

	fmt.Fprintf(os.Stderr, "Analyzing: %s\n", failedCommand)
	corrected, explanation, err := RunFuckCommand(ctx, cfg.APIKey, cfg.Language, failedCommand, errorOutput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Corrected command goes to stdout (for shell capture)
	fmt.Print(corrected)

	// Explanation goes to stderr (visible but doesn't interfere with capture)
	if explanation != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", FormatMarkdown(explanation))
	}
}

func printShellIntegration(shell string) {
	switch strings.ToLower(shell) {
	case "bash":
		fmt.Print(`fuck() {
    local last_cmd
    last_cmd=$(fc -ln -1 | sed 's/^\s*//')
    if [ -z "$last_cmd" ]; then
        echo "No previous command found." >&2
        return 1
    fi
    echo "Fixing: $last_cmd" >&2
    local corrected
    corrected=$(gemhelp fuck "$last_cmd")
    if [ -n "$corrected" ]; then
        # In bash, we print it and put it in history.
        # Ideally we'd put it in the buffer, but that's complex without bind -x.
        echo -e "\nProposed fix: $corrected" >&2
        history -s "$corrected"
        echo "Command added to history. Press [Up] to review/run." >&2
    fi
}
`)
	case "zsh":
		fmt.Print(`fuck() {
    local last_cmd
    last_cmd=$(fc -ln -1 | sed 's/^\s*//')
    if [[ -z "$last_cmd" ]]; then
        echo "No previous command found." >&2
        return 1
    fi
    echo "Fixing: $last_cmd" >&2
    local corrected
    corrected=$(gemhelp fuck "$last_cmd")
    if [[ -n "$corrected" ]]; then
        # Put it in the editing buffer
        print -z "$corrected"
    fi
}
`)
	case "mksh":
		fmt.Print(`fuck() {
    typeset last_cmd
    last_cmd=$(fc -ln -1 | sed 's/^[[:space:]]*//')
    if [ -z "$last_cmd" ]; then
        echo "No previous command found." >&2
        return 1
    fi
    echo "Fixing: $last_cmd" >&2
    typeset corrected
    corrected=$(gemhelp fuck "$last_cmd")
    if [ -n "$corrected" ]; then
        echo -e "\nProposed fix: $corrected" >&2
    fi
}
`)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported shell: %s. Supported: bash, zsh, mksh\n", shell)
		os.Exit(1)
	}
}
