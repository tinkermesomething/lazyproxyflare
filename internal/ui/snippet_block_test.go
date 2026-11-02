package ui

import (
	"strings"
	"testing"

	"lazyproxyflare/internal/caddy"
	sw "lazyproxyflare/internal/ui/snippet_wizard"
)

func TestSnippetBlockWrapsOnlyBodies(t *testing.T) {
	if got, want := snippetBlock("x", "header A 1"), "(x) {\n\theader A 1\n}"; got != want {
		t.Errorf("flush-left body: got %q want %q", got, want)
	}
	if got, want := snippetBlock("x", "header {\n\t\tA 1\n\t}"), "(x) {\nheader {\n\t\tA 1\n\t}\n}"; got == want {
		t.Log("(already-indented body is kept as-is)")
	}
	complete := "(x) {\n\theader A 1\n}"
	if got := snippetBlock("x", complete); got != complete {
		t.Errorf("a complete block must not be wrapped again: %q", got)
	}
}

// Every wizard mode must produce a Caddyfile the parser reads back with the
// same snippet names, exactly one level of wrapper, and balanced braces.
func TestEveryWizardModeWritesParseableSnippets(t *testing.T) {
	templated := sw.SnippetWizardData{Mode: sw.ModeTemplated, SelectedTemplates: map[string]bool{}, SnippetConfigs: map[string]sw.SnippetConfig{}}
	for key := range sw.GetAvailableTemplates() {
		templated.SelectedTemplates[key] = true
	}
	cases := map[string]sw.SnippetWizardData{
		"custom":    {Mode: sw.ModeCustom, CustomSnippetName: "mine", CustomSnippetContent: "header X-A 1\nencode gzip"},
		"guided":    {Mode: sw.ModeGuided, CreateIPRestriction: true, LANSubnet: "10.0.0.0/8", CreateSecurityHeaders: true, SecurityPreset: "basic", CreatePerformance: true},
		"templated": templated,
	}
	for name, data := range cases {
		data := data
		w := sw.Wizard{State: &sw.WizardState{Data: &data}}
		gen := w.GenerateSnippets()
		if len(gen) == 0 {
			t.Errorf("%s: nothing generated", name)
			continue
		}
		var src strings.Builder
		for _, g := range gen {
			src.WriteString(snippetBlock(g.Name, g.Content))
			src.WriteString("\n\n")
		}
		text := src.String()
		if strings.Count(text, "{") != strings.Count(text, "}") {
			t.Errorf("%s: unbalanced braces:\n%s", name, text)
		}
		parsed := caddy.ParseCaddyfileWithSnippets(text)
		if len(parsed.Snippets) != len(gen) {
			t.Errorf("%s: wrote %d snippets, parser found %d:\n%s", name, len(gen), len(parsed.Snippets), text)
		}
		for _, g := range gen {
			if strings.Count(text, "("+g.Name+")") != 1 {
				t.Errorf("%s: %s is declared %d times:\n%s", name, g.Name, strings.Count(text, "("+g.Name+")"), text)
			}
		}
	}
}
