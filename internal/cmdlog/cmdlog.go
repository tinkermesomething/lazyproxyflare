// Package cmdlog keeps an in-memory record of the external actions the app
// performs (Cloudflare API calls and shell commands), so the user can see
// exactly what ran. It never stores request bodies, headers or tokens.
package cmdlog

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Kind is the type of recorded action.
type Kind string

const (
	API  Kind = "api"
	Exec Kind = "exec"
)

// maxErrLen caps the stored error text; command output can be long.
const maxErrLen = 200

// Entry is one recorded action.
type Entry struct {
	Time     time.Time
	Kind     Kind
	Summary  string // "GET /zones/.../dns_records" or the command line
	Status   string // HTTP status, "ok", "exit N" or "error"
	Duration time.Duration
	Err      string // one line, empty on success
}

// Log is a fixed-size, concurrency-safe record, newest entries kept.
type Log struct {
	mu      sync.Mutex
	max     int
	entries []Entry
}

// Default is the process-wide log the app writes to.
var Default = New(500)

// New creates a Log that keeps at most max entries.
func New(max int) *Log {
	return &Log{max: max}
}

// Add records an entry, stamping the time if unset.
func (l *Log) Add(e Entry) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	e.Err = oneLine(e.Err)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, e)
	if over := len(l.entries) - l.max; over > 0 {
		l.entries = append([]Entry(nil), l.entries[over:]...)
	}
}

// Entries returns a copy of the record, oldest first.
func (l *Log) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Entry(nil), l.entries...)
}

// Clear removes every entry.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}

// Run executes cmd, records it, and returns its combined output.
func (l *Log) Run(cmd *exec.Cmd) ([]byte, error) {
	start := time.Now()
	out, err := cmd.CombinedOutput()

	e := Entry{
		Time:     start,
		Kind:     Exec,
		Summary:  strings.Join(cmd.Args, " "),
		Status:   "ok",
		Duration: time.Since(start),
	}
	if err != nil {
		e.Status = "error"
		if exitErr, ok := err.(*exec.ExitError); ok {
			e.Status = fmt.Sprintf("exit %d", exitErr.ExitCode())
		}
		// Prefer the command's own output: it says why it failed.
		e.Err = strings.TrimSpace(string(out))
		if e.Err == "" {
			e.Err = err.Error()
		}
	}
	l.Add(e)
	return out, err
}

// oneLine collapses s to a single, length-limited line.
func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxErrLen {
		// Cutting by bytes can split a multibyte character; drop the broken tail.
		s = strings.ToValidUTF8(s[:maxErrLen], "") + "…"
	}
	return s
}
