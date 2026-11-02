package textdiff

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func split(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "")
}

// apply rebuilds the new text from the old text and the ops, checking the ops
// are consistent with both inputs.
func apply(t *testing.T, old []string, ops []Op) []string {
	t.Helper()
	var out []string
	oi := 0
	for _, op := range ops {
		switch op.Kind {
		case Equal:
			if oi >= len(old) || old[oi] != op.Text {
				t.Fatalf("equal op %q does not match old[%d]", op.Text, oi)
			}
			out = append(out, op.Text)
			oi++
		case Delete:
			if oi >= len(old) || old[oi] != op.Text {
				t.Fatalf("delete op %q does not match old[%d]", op.Text, oi)
			}
			oi++
		case Insert:
			out = append(out, op.Text)
		}
	}
	if oi != len(old) {
		t.Fatalf("ops consumed %d of %d old lines", oi, len(old))
	}
	return out
}

// lcsLen is a slow reference implementation.
func lcsLen(a, b []string) int {
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	return dp[len(a)][len(b)]
}

func TestLinesCases(t *testing.T) {
	tests := []struct{ name, old, new string }{
		{"both empty", "", ""},
		{"old empty", "", "abc"},
		{"new empty", "abc", ""},
		{"identical", "abc", "abc"},
		{"append", "ab", "abc"},
		{"prepend", "bc", "abc"},
		{"replace middle", "abc", "axc"},
		{"duplicated line", "abc", "acbc"},
		{"moved line", "abc", "acb"},
		{"repeated removed", "aab", "ab"},
		{"all different", "abc", "xyz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old, nw := split(tt.old), split(tt.new)
			ops := Lines(old, nw)
			if got := apply(t, old, ops); !reflect.DeepEqual(got, nw) && !(len(got) == 0 && len(nw) == 0) {
				t.Errorf("rebuilt %v, want %v", got, nw)
			}
		})
	}
}

func TestLinesIsMinimal(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	letters := []string{"a", "b", "c"}
	gen := func() []string {
		n := rng.Intn(14)
		s := make([]string, n)
		for i := range s {
			s[i] = letters[rng.Intn(len(letters))]
		}
		return s
	}
	for i := 0; i < 2000; i++ {
		old, nw := gen(), gen()
		ops := Lines(old, nw)
		rebuilt := apply(t, old, ops)
		if len(rebuilt) != len(nw) || (len(nw) > 0 && !reflect.DeepEqual(rebuilt, nw)) {
			t.Fatalf("old=%v new=%v rebuilt=%v", old, nw, rebuilt)
		}
		equal := 0
		for _, op := range ops {
			if op.Kind != Insert && (op.OldLine < 1 || op.OldLine > len(old) || old[op.OldLine-1] != op.Text) {
				t.Fatalf("old=%v new=%v: bad OldLine in %+v", old, nw, op)
			}
			if op.Kind != Delete && (op.NewLine < 1 || op.NewLine > len(nw) || nw[op.NewLine-1] != op.Text) {
				t.Fatalf("old=%v new=%v: bad NewLine in %+v", old, nw, op)
			}
			if op.Kind == Equal {
				equal++
			}
		}
		if want := lcsLen(old, nw); equal != want {
			t.Fatalf("old=%v new=%v: %d equal ops, LCS is %d (diff not minimal)", old, nw, equal, want)
		}
	}
}

func TestLineNumbers(t *testing.T) {
	ops := Lines(split("abc"), split("axc"))
	want := []Op{
		{Equal, "a", 1, 1},
		{Delete, "b", 2, 0},
		{Insert, "x", 0, 2},
		{Equal, "c", 3, 3},
	}
	if !reflect.DeepEqual(ops, want) {
		t.Errorf("got  %+v\nwant %+v", ops, want)
	}
}

func TestHunksGroupAndContext(t *testing.T) {
	old := split("abcdefghijklmnopqrstuvwxyz")
	nw := append([]string{}, old...)
	nw[2] = "X"  // change c
	nw[20] = "Y" // change u
	hunks := Hunks(Lines(old, nw), 2)
	if len(hunks) != 2 {
		t.Fatalf("got %d hunks, want 2 (changes are far apart)", len(hunks))
	}
	h := hunks[0]
	// context 2 before (a,b), the change, 2 after (d,e)
	if h.OldStart != 1 || h.OldLen != 5 || h.NewStart != 1 || h.NewLen != 5 {
		t.Errorf("hunk 0 header = -%d,%d +%d,%d", h.OldStart, h.OldLen, h.NewStart, h.NewLen)
	}

	// Close changes merge into one hunk.
	nw2 := append([]string{}, old...)
	nw2[2] = "X"
	nw2[5] = "Y"
	if got := len(Hunks(Lines(old, nw2), 3)); got != 1 {
		t.Errorf("close changes: got %d hunks, want 1", got)
	}
}

func TestHunksIdenticalIsEmpty(t *testing.T) {
	x := split("abc")
	if h := Hunks(Lines(x, x), 3); len(h) != 0 {
		t.Errorf("identical input produced hunks: %+v", h)
	}
}

func TestHunksPureInsertAndDeleteHeaders(t *testing.T) {
	// Insert into empty old: old range is "0,0" in unified-diff convention.
	h := Hunks(Lines(nil, split("ab")), 3)
	if len(h) != 1 || h[0].OldStart != 0 || h[0].OldLen != 0 || h[0].NewStart != 1 || h[0].NewLen != 2 {
		t.Errorf("insert-into-empty header wrong: %+v", h)
	}
	h = Hunks(Lines(split("ab"), nil), 3)
	if len(h) != 1 || h[0].OldStart != 1 || h[0].OldLen != 2 || h[0].NewStart != 0 || h[0].NewLen != 0 {
		t.Errorf("delete-to-empty header wrong: %+v", h)
	}
}
