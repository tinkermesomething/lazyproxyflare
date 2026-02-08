package cloudflare

// DNSRecord represents a Cloudflare DNS record
type DNSRecord struct {
	ID       string `json:"id"`
	Type     string `json:"type"`     // A, AAAA, CNAME, etc.
	Name     string `json:"name"`     // Full FQDN (e.g., "plex.example.com")
	Content  string `json:"content"`  // Target (IP for A record, domain for CNAME)
	Proxied  bool   `json:"proxied"`  // Cloudflare proxy (orange cloud)
	TTL      int    `json:"ttl"`      // Time to live in seconds (1 = auto)
	ZoneID   string `json:"zone_id"`
	ZoneName string `json:"zone_name"`
}

// APIResponse is the standard Cloudflare API response wrapper
type APIResponse struct {
	Success    bool          `json:"success"`
	Errors     []APIError    `json:"errors"`
	Result     []DNSRecord   `json:"result"`
	ResultInfo *ResultInfo   `json:"result_info,omitempty"`
}

// ResultInfo contains pagination metadata
type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// APIError represents a Cloudflare API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
