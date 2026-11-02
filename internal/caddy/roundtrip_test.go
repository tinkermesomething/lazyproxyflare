package caddy

import (
	"os"
	"path/filepath"
	"testing"
)

const rtBlock = "# === new.example.com ===\nnew.example.com {\n\timport ip_restricted\n\n\treverse_proxy http://localhost:80\n}\n"

func rtRun(t *testing.T, original string) (string, string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "Caddyfile")
	if err := os.WriteFile(p, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AppendEntry(p, rtBlock); err != nil {
		t.Fatal(err)
	}
	added, _ := os.ReadFile(p)
	if err := RemoveEntry(p, "new.example.com"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	return string(added), string(got)
}

// Adding an entry and deleting it again must give back the exact original file.
func TestAppendThenRemoveRestoresTheFile(t *testing.T) {
	cases := map[string]string{
		"ends with newline":    "a.example.com {\n\treverse_proxy 1.1.1.1\n}\n",
		"no trailing newline":  "a.example.com {\n\treverse_proxy 1.1.1.1\n}",
		"two blank lines":      "a.example.com {\n}\n\n",
		"global block + snips": "{\n\temail x@y.z\n}\n\n(snip) {\n\tencode gzip\n}\n\na.example.com {\n\timport snip\n}\n",
	}
	for name, orig := range cases {
		added, got := rtRun(t, orig)
		if added == orig {
			t.Errorf("%s: append did nothing", name)
		}
		// A missing final newline is added by Append, so compare ignoring that.
		want := orig
		if want[len(want)-1] != '\n' {
			want += "\n"
		}
		if got != want {
			t.Errorf("%s: round trip changed the file\n--- want %q\n--- got  %q", name, want, got)
		}
	}
}

// Deleting a block from the middle must keep exactly one blank line between
// its neighbours, and must not touch anything else.
func TestRemoveMiddleEntryKeepsOneSeparator(t *testing.T) {
	src := "a.example.com {\n}\n\n" + rtBlock + "\nb.example.com {\n}\n"
	p := filepath.Join(t.TempDir(), "Caddyfile")
	os.WriteFile(p, []byte(src), 0o644)
	if err := RemoveEntry(p, "new.example.com"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	if want := "a.example.com {\n}\n\nb.example.com {\n}\n"; string(got) != want {
		t.Errorf("got %q want %q", got, want)
	}

	tight := "a.example.com {\n}\n" + rtBlock + "b.example.com {\n}\n"
	os.WriteFile(p, []byte(tight), 0o644)
	RemoveEntry(p, "new.example.com")
	got, _ = os.ReadFile(p)
	if want := "a.example.com {\n}\nb.example.com {\n}\n"; string(got) != want {
		t.Errorf("tight neighbours: got %q want %q", got, want)
	}
}

// Same round trip with what the app really appends, and an edit (remove +
// append) in between: the file must end up with the new block and nothing else.
func TestRoundTripWithGeneratedBlocks(t *testing.T) {
	orig := "(ip_restricted) {\n\trespond \"404\" 404\n}\n\na.example.com {\n\timport ip_restricted\n\treverse_proxy http://10.0.0.2:80\n}\n"
	p := filepath.Join(t.TempDir(), "Caddyfile")
	os.WriteFile(p, []byte(orig), 0o644)

	gen := func(port int) string {
		return GenerateCaddyBlock(GenerateBlockInput{
			FQDN: "new.example.com", Target: "10.0.0.9", Port: port,
			AvailableSnippets: []string{"ip_restricted"}, SelectedSnippets: []string{"ip_restricted"},
		})
	}
	if err := AppendEntry(p, gen(80)); err != nil {
		t.Fatal(err)
	}
	// An edit is remove + append.
	if err := RemoveEntry(p, "new.example.com"); err != nil {
		t.Fatal(err)
	}
	if err := AppendEntry(p, gen(8099)); err != nil {
		t.Fatal(err)
	}
	if err := RemoveEntry(p, "new.example.com"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	if string(got) != orig {
		t.Errorf("add, edit, delete changed the file\n--- want %q\n--- got  %q", orig, got)
	}
}
