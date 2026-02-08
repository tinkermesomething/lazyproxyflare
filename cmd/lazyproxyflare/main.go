package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
	"lazyproxyflare/internal/ui"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Show version and exit")
	profileFlag := flag.String("profile", "", "Load a specific profile by name")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "LazyProxyFlare - Cloudflare DNS + Caddy reverse proxy manager\n\n")
		fmt.Fprintf(os.Stderr, "Usage: lazyproxyflare [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nWith no flags, launches the interactive TUI.\n")
		fmt.Fprintf(os.Stderr, "Profiles are stored in ~/.config/lazyproxyflare/profiles/\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("lazyproxyflare %s\n", Version)
		os.Exit(0)
	}

	// If --profile specified, load that directly
	if *profileFlag != "" {
		autoLoadProfile(*profileFlag)
		return
	}

	// Discover available profiles
	profiles, err := config.ListProfiles()
	if err != nil {
		log.Fatalf("Failed to discover profiles: %v", err)
	}

	// Determine startup mode based on profile count
	switch len(profiles) {
	case 0:
		// No profiles exist - launch wizard
		launchWizard()

	case 1:
		// One profile exists - auto-load it
		autoLoadProfile(profiles[0])

	default:
		// Multiple profiles exist - show profile selector
		launchProfileSelector(profiles)
	}
}

// launchWizard starts the TUI in wizard mode (no data loaded)
func launchWizard() {
	// Launch TUI with wizard view (no config, no data)
	p := tea.NewProgram(
		ui.NewModelWithWizard(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// autoLoadProfile loads a single profile and launches TUI with data
func autoLoadProfile(profileName string) {
	// Load profile
	profileConfig, err := config.LoadProfile(profileName)
	if err != nil {
		log.Fatalf("Failed to load profile '%s': %v", profileName, err)
	}

	// Convert to legacy config format
	cfg := config.ProfileToLegacyConfig(profileConfig)

	// Load data (Caddyfile + DNS records)
	entries, snippets := loadData(cfg)

	// Set as last used profile
	config.SetLastUsedProfile(profileName)

	// Launch TUI with loaded data
	p := tea.NewProgram(
		ui.NewModelWithProfile(entries, snippets, cfg, profileName, ""), // No password
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// launchProfileSelector starts the TUI in profile selector mode
func launchProfileSelector(profiles []string) {
	// Get last used profile (if any)
	lastUsed, _ := config.GetLastUsedProfile()

	// Launch TUI with profile selector view (no config, no data)
	p := tea.NewProgram(
		ui.NewModelWithProfileSelector(profiles, lastUsed),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// loadData loads Caddyfile and DNS records for a given config
func loadData(cfg *config.Config) ([]diff.SyncedEntry, []caddy.Snippet) {
	// Parse Caddyfile
	caddyContent, err := os.ReadFile(cfg.Caddy.CaddyfilePath)
	if err != nil {
		log.Printf("Warning: Failed to read Caddyfile: %v", err)
		// Return empty data instead of crashing
		return []diff.SyncedEntry{}, []caddy.Snippet{}
	}

	// Parse Caddyfile with snippets
	parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))

	// Get API token
	apiToken, err := cfg.GetAPIToken()
	if err != nil {
		log.Fatalf("Failed to get API token: %v", err)
	}

	// Fetch DNS records from Cloudflare
	cfClient := cloudflare.NewClient(apiToken)

	cnameRecords, err := cfClient.ListDNSRecords(cfg.Cloudflare.ZoneID, "CNAME")
	if err != nil {
		log.Printf("Warning: Failed to fetch CNAME records: %v", err)
		cnameRecords = []cloudflare.DNSRecord{}
	}

	aRecords, err := cfClient.ListDNSRecords(cfg.Cloudflare.ZoneID, "A")
	if err != nil {
		log.Printf("Warning: Failed to fetch A records: %v", err)
		aRecords = []cloudflare.DNSRecord{}
	}

	// Combine all DNS records
	allDNS := append(cnameRecords, aRecords...)

	// Run diff engine
	syncedEntries := diff.Compare(allDNS, parsed.Entries)

	return syncedEntries, parsed.Snippets
}
