package caddy

import (
	"strings"
	"testing"
)

const editSample = `{
	email me@example.com
}

# ============================================================
# SNIPPETS
# ============================================================

# LAN only. Everyone else gets a 404.
(ip_restricted) {
	@not_allowed {
		not remote_ip 10.0.0.0/24
	}
	respond @not_allowed "404" 404
}

# Response headers.
# Import in every site.
(security_headers) {
	header {
		-Server
	}
}

(no_comment) {
	encode gzip
}

app.example.com {
	import ip_restricted
	import security_headers

	reverse_proxy 10.0.0.2:80
}
`

func TestSnippetCommentIsParsed(t *testing.T) {
	got := map[string]string{}
	for _, s := range ParseCaddyfileWithSnippets(editSample).Snippets {
		got[s.Name] = s.Comment
	}
	want := map[string]string{
		"ip_restricted":    "LAN only. Everyone else gets a 404.",
		"security_headers": "Response headers. Import in every site.",
		"no_comment":       "", // the banner above is a divider, not a comment
	}
	for name, c := range want {
		if got[name] != c {
			t.Errorf("%s comment = %q, want %q", name, got[name], c)
		}
	}
}

func TestReplaceSnippetBody(t *testing.T) {
	old, _ := FindSnippet(editSample, "no_comment")
	out, err := ReplaceSnippetBody(editSample, "no_comment", old.Content, "\tencode zstd gzip\n\tlog\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "(no_comment) {\n\tencode zstd gzip\n\tlog\n}\n") {
		t.Errorf("body not replaced:\n%s", out)
	}
	// Everything else is byte-identical.
	if strings.Replace(out, "\tencode zstd gzip\n\tlog\n", "\tencode gzip\n", 1) != editSample {
		t.Error("replacing one snippet changed something else")
	}
	// The entries still parse the same.
	if a, b := ParseCaddyfileWithSnippets(editSample).Entries, ParseCaddyfileWithSnippets(out).Entries; len(a) != len(b) || a[0].RawBlock != b[0].RawBlock {
		t.Error("domain entries changed")
	}
}

func TestReplaceSnippetBodyRefusesUnsafeEdits(t *testing.T) {
	old, _ := FindSnippet(editSample, "ip_restricted")
	tests := []struct {
		name, snippet, expected, body, want string
	}{
		{"unknown snippet", "nope", "", "\tx\n", "not found"},
		{"file changed since open", "ip_restricted", "\tstale body", "\tx\n", "changed"},
		{"unbalanced open", "ip_restricted", old.Content, "\theader {\n", "unbalanced"},
		{"unbalanced close", "ip_restricted", old.Content, "\t}\n", "unbalanced"},
	}
	for _, tt := range tests {
		_, err := ReplaceSnippetBody(editSample, tt.snippet, tt.expected, tt.body)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want it to mention %q", tt.name, err, tt.want)
		}
	}
	// A brace inside quotes is not structure.
	if _, err := ReplaceSnippetBody(editSample, "ip_restricted", old.Content, "\trespond \"{\" 200\n"); err != nil {
		t.Errorf("quoted brace rejected: %v", err)
	}
}

func TestRemoveSnippetBlock(t *testing.T) {
	out, err := RemoveSnippetBlock(editSample, "security_headers")
	if err != nil {
		t.Fatal(err)
	}
	for _, gone := range []string{"(security_headers)", "-Server", "# Response headers.", "# Import in every site."} {
		if strings.Contains(out, gone) {
			t.Errorf("%q left behind:\n%s", gone, out)
		}
	}
	for _, kept := range []string{"# SNIPPETS", "# ====", "(ip_restricted)", "(no_comment)", "app.example.com {", "# LAN only."} {
		if !strings.Contains(out, kept) {
			t.Errorf("%q was removed:\n%s", kept, out)
		}
	}
	if strings.Contains(out, "\n\n\n") {
		t.Errorf("double blank line left behind:\n%s", out)
	}
	if _, err := FindSnippet(out, "security_headers"); err == nil {
		t.Error("snippet still parses")
	}

	// The first snippet, directly under a banner, keeps the banner.
	out, err = RemoveSnippetBlock(editSample, "ip_restricted")
	if err != nil || !strings.Contains(out, "# SNIPPETS") || strings.Contains(out, "LAN only") {
		t.Errorf("first snippet removal wrong (err %v):\n%s", err, out)
	}
	if _, err := RemoveSnippetBlock(editSample, "nope"); err == nil {
		t.Error("removing an unknown snippet must fail")
	}
}
