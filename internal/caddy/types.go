package caddy

// CaddyEntry represents a parsed Caddy configuration entry
type CaddyEntry struct {
	Domain       string   // Primary domain (e.g., "plex.angelsomething.com")
	Domains      []string // All domains if multi-domain block
	Target       string   // Target IP or hostname from reverse_proxy
	Port         int      // Port number from reverse_proxy
	SSL          bool     // true if https://, false if http://
	IPRestricted bool     // true if has IP restriction (import or inline)
	OAuthHeaders bool     // true if has OAuth/OIDC headers
	WebSocket    bool     // true if has WebSocket headers
	Imports      []string // List of imported snippets
	RawBlock     string   // Original block text
	LineStart    int      // Line number where block starts (1-indexed)
	LineEnd      int      // Line number where block ends
	HasMarker    bool     // true if has # === domain === marker
}
