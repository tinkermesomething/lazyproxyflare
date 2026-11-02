package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"lazyproxyflare/internal/caddy"
)

// With many snippets the custom-directives field used to fall below the modal.
func TestFormScrollsToFocusedField(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {140, 40}} {
		m := mgrModel(t)
		for i := 0; i < 30; i++ {
			m.snippets = append(m.snippets, caddy.Snippet{Name: fmt.Sprintf("snip_%02d", i), Description: "d"})
		}
		m.width, m.height = size[0], size[1]
		m.currentView = ViewAdd
		m.addForm.CustomCaddyConfig = "header X-A 1\nheader X-B 2"
		rows := m.formRows()
		last := 8 + len(m.snippets) // custom directives field

		for f := 0; f <= last; f++ {
			m.addForm.FocusedField = f
			out := ansi.Strip(m.renderAddFormContent())
			if n := len(strings.Split(out, "\n")); n > rows {
				t.Fatalf("%v field %d: %d lines, modal holds %d", size, f, n, rows)
			}
			if f == last && !strings.Contains(out, "header X-A 1") {
				t.Errorf("%v: custom directives not visible when focused:\n%s", size, out)
			}
			if f >= 8 && f < last && !strings.Contains(out, m.snippets[f-8].Name) {
				t.Errorf("%v: focused snippet %q not visible", size, m.snippets[f-8].Name)
			}
			if !strings.Contains(out, "Cancel: esc") {
				t.Errorf("%v field %d: key hints scrolled away", size, f)
			}
		}
	}
}
