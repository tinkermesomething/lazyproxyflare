package caddy

import (
	"strings"
	"testing"
)

func parseOne(t *testing.T, src string) *CaddyEntry {
	t.Helper()
	entries, err := ParseCaddyfile(src)
	if err != nil || len(entries) != 1 {
		t.Fatalf("parse: %v, %d entries", err, len(entries))
	}
	return &entries[0]
}

// Whatever the generator writes must count as managed, or the app could never
// edit its own entries.
func TestGeneratedBlocksAreManaged(t *testing.T) {
	for _, snippets := range [][]string{nil, {"ip_restricted"}, {"log_access", "performance", "frame_embedding"}} {
		for _, lan := range []bool{false, true} {
			for _, oauth := range []bool{false, true} {
				for _, ws := range []bool{false, true} {
					for _, port := range []int{80, 443, 8080} {
						block := GenerateCaddyBlock(GenerateBlockInput{
							FQDN: "app.example.com", Target: "10.0.0.5", Port: port, SSL: port == 443,
							LANOnly: lan, OAuth: oauth, WebSocket: ws,
							LANSubnet: "10.0.0.0/8", AllowedExtIP: "127.0.0.1/32", SelectedSnippets: snippets,
						})
						e := parseOne(t, block)
						if ok, why := IsManagedBlock(e); !ok {
							t.Errorf("generated block reported as custom (%q):\n%s", why, block)
						}
					}
				}
			}
		}
	}
}

func TestCustomBlocksAreNotManaged(t *testing.T) {
	tests := []struct {
		name, src, reasonHas string
	}{
		{"extra header_up", `mail.example.com {
	import log_access
	reverse_proxy 10.0.0.9:80 {
		header_up X-Real-IP {remote_host}
		import performance
	}
}`, "X-Real-IP"},
		{"request_body and transport", `git.example.com {
	import log_access
	request_body {
		max_size 512MB
	}
	reverse_proxy 10.0.0.10:3080 {
		transport http {
			read_timeout 120s
		}
	}
}`, "request_body"},
		{"several domains", `a.example.com, b.example.com {
	reverse_proxy 10.0.0.9:80
}`, "2 domains"},
		{"comment inside block", `a.example.com {
	# keep this
	reverse_proxy 10.0.0.9:80
}`, "# keep this"},
	}
	for _, tt := range tests {
		ok, why := IsManagedBlock(parseOne(t, tt.src))
		if ok {
			t.Errorf("%s: reported as managed", tt.name)
		} else if !strings.Contains(why, tt.reasonHas) {
			t.Errorf("%s: reason %q does not mention %q", tt.name, why, tt.reasonHas)
		}
	}
}

// Blocks written by hand in the same shape as generated ones are fine: the
// scheme prefix, import order and comments outside the block do not matter.
func TestHandWrittenPlainBlockIsManaged(t *testing.T) {
	src := `# a comment above the block
app.example.com {
	import ip_restricted
	import log_access

	reverse_proxy 10.0.0.3:4080 {
		import performance
		import frame_embedding
	}
}`
	if ok, why := IsManagedBlock(parseOne(t, src)); !ok {
		t.Errorf("plain block reported as custom: %q", why)
	}
	if ok, _ := IsManagedBlock(nil); !ok {
		t.Error("nil entry must be managed")
	}
}
