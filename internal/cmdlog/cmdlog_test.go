package cmdlog

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"
)

func TestKeepsOnlyNewestEntries(t *testing.T) {
	l := New(3)
	for i := 0; i < 5; i++ {
		l.Add(Entry{Summary: fmt.Sprintf("e%d", i)})
	}
	got := l.Entries()
	if len(got) != 3 || got[0].Summary != "e2" || got[2].Summary != "e4" {
		t.Errorf("got %+v, want e2..e4 oldest first", got)
	}
}

func TestEntriesReturnsACopy(t *testing.T) {
	l := New(5)
	l.Add(Entry{Summary: "a"})
	l.Entries()[0].Summary = "changed"
	if got := l.Entries()[0].Summary; got != "a" {
		t.Errorf("caller mutated the log: %q", got)
	}
}

func TestAddSetsTimeAndShortensError(t *testing.T) {
	l := New(5)
	before := time.Now()
	l.Add(Entry{Summary: "x", Err: strings.Repeat("e", 1000) + "\nsecond line"})
	e := l.Entries()[0]
	if e.Time.Before(before) {
		t.Errorf("time not set: %v", e.Time)
	}
	if strings.Contains(e.Err, "\n") {
		t.Errorf("error must be one line: %q", e.Err)
	}
	if len(e.Err) > maxErrLen+len("…") {
		t.Errorf("error not shortened: %d bytes", len(e.Err))
	}
}

func TestRunRecordsSuccess(t *testing.T) {
	l := New(5)
	out, err := l.Run(exec.Command("echo", "hello"))
	if err != nil || strings.TrimSpace(string(out)) != "hello" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	e := l.Entries()[0]
	if e.Kind != Exec || e.Summary != "echo hello" || e.Status != "ok" || e.Err != "" {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestRunRecordsFailureAndKeepsOutput(t *testing.T) {
	l := New(5)
	out, err := l.Run(exec.Command("sh", "-c", "echo oops >&2; exit 3"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(string(out), "oops") {
		t.Errorf("output lost: %q", out)
	}
	e := l.Entries()[0]
	if e.Status != "exit 3" || !strings.Contains(e.Err, "oops") {
		t.Errorf("failure not recorded with output: %+v", e)
	}
}

func TestRunRecordsCommandThatCannotStart(t *testing.T) {
	l := New(5)
	if _, err := l.Run(exec.Command("definitely-not-a-real-binary-xyz")); err == nil {
		t.Fatal("expected error")
	}
	e := l.Entries()[0]
	if e.Status != "error" || e.Err == "" {
		t.Errorf("start failure not recorded: %+v", e)
	}
}

func TestConcurrentUse(t *testing.T) {
	l := New(50)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				l.Add(Entry{Summary: "x"})
				l.Entries()
			}
		}()
	}
	wg.Wait()
	if n := len(l.Entries()); n != 50 {
		t.Errorf("len = %d, want 50", n)
	}
}

func TestClear(t *testing.T) {
	l := New(5)
	l.Add(Entry{Summary: "a"})
	l.Clear()
	if n := len(l.Entries()); n != 0 {
		t.Errorf("len = %d after Clear", n)
	}
}

func TestShortenedErrorStaysValidUTF8(t *testing.T) {
	l := New(5)
	// 3-byte characters: a byte cut at maxErrLen lands inside one.
	l.Add(Entry{Summary: "x", Err: strings.Repeat("€", 300)})
	if e := l.Entries()[0]; !utf8.ValidString(e.Err) {
		t.Errorf("error is not valid UTF-8: %q", e.Err)
	}
}
