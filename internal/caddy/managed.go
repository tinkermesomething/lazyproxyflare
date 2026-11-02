package caddy

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	importLine       = regexp.MustCompile(`^import \S+$`)
	reverseProxyLine = regexp.MustCompile(`^reverse_proxy \S+( \{)?$`)
	notRemoteIPLine  = regexp.MustCompile(`^not remote_ip \S+( \S+)*$`)
)

// generatedLines are the fixed lines GenerateCaddyBlock can emit besides
// imports and the reverse_proxy line.
var generatedLines = map[string]bool{
	"}":                     true,
	"@external {":           true,
	"respond @external 404": true,
	"header_up Upgrade {http.request.header.Upgrade}":                                               true,
	"header_up Connection {http.request.header.Connection}":                                         true,
	"header_up X-Forwarded-User {http.request.header.X-Forwarded-User}":                             true,
	"header_up X-Forwarded-Groups {http.request.header.X-Forwarded-Groups}":                         true,
	"header_up X-Forwarded-Email {http.request.header.X-Forwarded-Email}":                           true,
	"header_up X-Forwarded-Preferred-Username {http.request.header.X-Forwarded-Preferred-Username}": true,
}

// IsManagedBlock reports whether the entry's block is exactly what
// GenerateCaddyBlock can produce for the entry's own settings. Editing an
// entry removes the block and writes a freshly generated one, so anything else
// in the block (custom directives, inner comments, extra domains) would be
// lost. When it returns false, reason names the first thing that would be lost.
func IsManagedBlock(e *CaddyEntry) (managed bool, reason string) {
	if e == nil {
		return true, ""
	}
	if len(e.Domains) > 1 {
		return false, fmt.Sprintf("block serves %d domains", len(e.Domains))
	}
	lines := strings.Split(e.RawBlock, "\n")
	for i, raw := range lines {
		if i == 0 { // the "domain {" header, already parsed into Domains
			continue
		}
		line := strings.TrimSpace(raw)
		switch {
		case line == "":
		case generatedLines[line], importLine.MatchString(line),
			reverseProxyLine.MatchString(line), notRemoteIPLine.MatchString(line):
		default:
			if len(line) > 50 {
				line = line[:47] + "..."
			}
			return false, line
		}
	}
	return true, ""
}
