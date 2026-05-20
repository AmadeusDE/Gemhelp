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

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <command> [question...]\n", progName)
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSubcommands / Multi-call Shortcuts:\n")
		fmt.Fprintf(os.Stderr, "  %s man <command>          Render parsed man page\n", progName)
		fmt.Fprintf(os.Stderr, "  %s tldr <command>         Render offline TLDR page\n", progName)
		fmt.Fprintf(os.Stderr, "  %s wiki <query>           Search/get Arch Wiki page\n", progName)
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  gemhelp ls\n")
		fmt.Fprintf(os.Stderr, "  gemhelp ls how to sort by size\n")
		fmt.Fprintf(os.Stderr, "  gemhelp --tldr tar\n")
	}

	flag.Parse()

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
