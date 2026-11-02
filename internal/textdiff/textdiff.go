// Package textdiff computes minimal line diffs (Myers) and groups them into
// unified-diff style hunks. It exists so backup previews show real differences:
// order and repeated lines matter in a Caddyfile.
package textdiff

// Kind is the type of one diff operation.
type Kind int

const (
	Equal Kind = iota
	Delete
	Insert
)

// Op is one line in the edit script. OldLine and NewLine are 1-based and are
// 0 where the line does not exist on that side (NewLine for Delete, OldLine for Insert).
type Op struct {
	Kind    Kind
	Text    string
	OldLine int
	NewLine int
}

// Lines returns a minimal edit script turning old into new.
func Lines(old, new []string) []Op {
	// Trim the common prefix and suffix: Caddyfile diffs are usually small
	// and this keeps the search below cheap.
	prefix := 0
	for prefix < len(old) && prefix < len(new) && old[prefix] == new[prefix] {
		prefix++
	}
	oldEnd, newEnd := len(old), len(new)
	for oldEnd > prefix && newEnd > prefix && old[oldEnd-1] == new[newEnd-1] {
		oldEnd--
		newEnd--
	}

	ops := make([]Op, 0, len(old)+len(new))
	for i := 0; i < prefix; i++ {
		ops = append(ops, Op{Equal, old[i], i + 1, i + 1})
	}
	for _, e := range myers(old[prefix:oldEnd], new[prefix:newEnd]) {
		switch e.kind {
		case Equal:
			ops = append(ops, Op{Equal, e.text, prefix + e.a + 1, prefix + e.b + 1})
		case Delete:
			ops = append(ops, Op{Delete, e.text, prefix + e.a + 1, 0})
		case Insert:
			ops = append(ops, Op{Insert, e.text, 0, prefix + e.b + 1})
		}
	}
	for i := oldEnd; i < len(old); i++ {
		ops = append(ops, Op{Equal, old[i], i + 1, newEnd + (i - oldEnd) + 1})
	}
	return ops
}

// edit is an op with 0-based indexes into the slices given to myers.
type edit struct {
	kind Kind
	text string
	a, b int
}

// myers implements the greedy O(ND) algorithm from Myers, "An O(ND) Difference
// Algorithm and Its Variations" (1986).
func myers(a, b []string) []edit {
	n, m := len(a), len(b)
	if n == 0 && m == 0 {
		return nil
	}
	maxD := n + m
	off := maxD + 1
	v := make([]int, 2*maxD+3)

	// trace[d] is the furthest-reaching x per diagonal k before round d,
	// stored only for k in [-(d+1), d+1].
	trace := make([][]int, 0, 16)

	for d := 0; d <= maxD; d++ {
		snap := make([]int, 2*d+3)
		copy(snap, v[off-d-1:off+d+2])
		trace = append(trace, snap)

		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[off+k-1] < v[off+k+1]) {
				x = v[off+k+1] // step down: insertion
			} else {
				x = v[off+k-1] + 1 // step right: deletion
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[off+k] = x
			if x >= n && y >= m {
				return backtrack(a, b, trace)
			}
		}
	}
	return nil // unreachable: d == n+m always reaches the end
}

func backtrack(a, b []string, trace [][]int) []edit {
	x, y := len(a), len(b)
	var rev []edit
	for d := len(trace) - 1; d >= 0; d-- {
		snap := trace[d]
		at := func(k int) int { return snap[k+d+1] }

		k := x - y
		var prevK int
		if k == -d || (k != d && at(k-1) < at(k+1)) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := at(prevK)
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			x--
			y--
			rev = append(rev, edit{Equal, a[x], x, y})
		}
		if d > 0 {
			if x == prevX {
				y--
				rev = append(rev, edit{Insert, b[y], x, y})
			} else {
				x--
				rev = append(rev, edit{Delete, a[x], x, y})
			}
		}
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

// Hunk is a group of changes with surrounding context, as in a unified diff.
// Start values follow unified-diff convention: 1-based, and one less than the
// first changed line when the side has no lines in the hunk.
type Hunk struct {
	OldStart, OldLen int
	NewStart, NewLen int
	Ops              []Op
}

// Hunks groups ops into hunks with the given number of context lines.
// It returns nil when old and new are identical.
func Hunks(ops []Op, context int) []Hunk {
	type span struct{ start, end int } // [start, end)
	var spans []span
	for i, op := range ops {
		if op.Kind == Equal {
			continue
		}
		s, e := max(0, i-context), min(len(ops), i+1+context)
		if last := len(spans) - 1; last >= 0 && s <= spans[last].end {
			spans[last].end = max(spans[last].end, e)
		} else {
			spans = append(spans, span{s, e})
		}
	}

	hunks := make([]Hunk, 0, len(spans))
	for _, sp := range spans {
		h := Hunk{Ops: ops[sp.start:sp.end]}
		oldBefore, newBefore := 0, 0
		for _, op := range ops[:sp.start] {
			if op.Kind != Insert {
				oldBefore++
			}
			if op.Kind != Delete {
				newBefore++
			}
		}
		for _, op := range h.Ops {
			if op.Kind != Insert {
				h.OldLen++
			}
			if op.Kind != Delete {
				h.NewLen++
			}
		}
		h.OldStart, h.NewStart = oldBefore+1, newBefore+1
		if h.OldLen == 0 {
			h.OldStart = oldBefore
		}
		if h.NewLen == 0 {
			h.NewStart = newBefore
		}
		hunks = append(hunks, h)
	}
	if len(hunks) == 0 {
		return nil
	}
	return hunks
}
