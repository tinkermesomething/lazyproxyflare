package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/cmdlog"
)

func key(s string) tea.KeyMsg {
	if s == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// seedCommandLog replaces the process-wide log with n numbered API entries.
func seedCommandLog(t *testing.T, n int) {
	t.Helper()
	cmdlog.Default.Clear()
	t.Cleanup(cmdlog.Default.Clear)
	for i := 0; i < n; i++ {
		e := cmdlog.Entry{
			Time:     time.Date(2026, 1, 1, 10, 0, i%60, 0, time.Local),
			Kind:     cmdlog.API,
			Summary:  fmt.Sprintf("GET /zones/z/dns_records?page=%d", i),
			Status:   "200",
			Duration: 120 * time.Millisecond,
		}
		if i%4 == 3 {
			e.Status, e.Err = "403", "Invalid access token"
		}
		cmdlog.Default.Add(e)
	}
}

func renderCommandLog(m Model) string { return stripANSI(m.renderCommandLogView()) }

func TestCommandLogOpensAndClosesWithKeys(t *testing.T) {
	seedCommandLog(t, 3)
	m := createTestModel()
	m.width, m.height = 120, 30

	m, _ = m.handleKeyMsg(key("@"))
	if m.currentView != ViewCommandLog {
		t.Fatalf("@ did not open the command log, view = %v", m.currentView)
	}
	m, _ = m.handleKeyMsg(key("@"))
	if m.currentView != ViewList {
		t.Errorf("@ again should close it, view = %v", m.currentView)
	}
	m, _ = m.handleKeyMsg(key("@"))
	m, _ = m.handleKeyMsg(key("esc"))
	if m.currentView != ViewList {
		t.Errorf("esc should close it, view = %v", m.currentView)
	}
}

func TestCommandLogShowsNewestFirstWithStatusAndError(t *testing.T) {
	seedCommandLog(t, 4)
	m := createTestModel()
	m.width, m.height = 179, 49
	m.currentView = ViewCommandLog

	out := renderCommandLog(m)
	for _, want := range []string{"Command Log", "4 commands", "page=0", "200", "403", "Invalid access token", "120ms"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Index(out, "page=3") > strings.Index(out, "page=0") {
		t.Errorf("newest entry should be listed first:\n%s", out)
	}
}

func TestCommandLogEmpty(t *testing.T) {
	cmdlog.Default.Clear()
	m := createTestModel()
	m.width, m.height = 120, 30
	m.currentView = ViewCommandLog
	if out := renderCommandLog(m); !strings.Contains(out, "Nothing has run yet") {
		t.Errorf("empty state missing:\n%s", out)
	}
}

func TestCommandLogScrollReachesOldestAndKeepsHint(t *testing.T) {
	const n = 200
	for _, sz := range screenSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			seedCommandLog(t, n)
			m := createTestModel()
			m.width, m.height = sz.w, sz.h
			m.currentView = ViewCommandLog

			for i := 0; i < 400; i++ {
				m, _ = m.handleNavigateDown()
				if i < 60 {
					if out := renderCommandLog(m); !strings.Contains(out, "esc: close") {
						t.Fatalf("hint pushed out at scroll %d:\n%s", m.cmdLogScroll, out)
					}
				}
			}
			if m.cmdLogScroll != m.commandLogMaxScroll() {
				t.Errorf("scroll = %d, want max %d", m.cmdLogScroll, m.commandLogMaxScroll())
			}
			out := renderCommandLog(m)
			if !strings.Contains(out, "page=0") || !strings.Contains(out, "esc: close") {
				t.Errorf("oldest entry or hint not visible at the end:\n%s", out)
			}

			m, _ = m.handleGoToTop()
			if m.cmdLogScroll != 0 {
				t.Errorf("g: scroll = %d, want 0", m.cmdLogScroll)
			}
			m, _ = m.handleGoToBottom()
			if m.cmdLogScroll != m.commandLogMaxScroll() {
				t.Errorf("G: scroll = %d, want %d", m.cmdLogScroll, m.commandLogMaxScroll())
			}
			for i := 0; i < 400; i++ {
				m, _ = m.handleNavigateUp()
			}
			if m.cmdLogScroll != 0 {
				t.Errorf("scroll up: scroll = %d, want 0", m.cmdLogScroll)
			}
		})
	}
}

func TestCommandLogDoesNotOpenOverOtherViews(t *testing.T) {
	m := createTestModel()
	m.currentView = ViewBackupManager
	m, _ = m.handleKeyMsg(key("@"))
	if m.currentView != ViewBackupManager {
		t.Errorf("@ must not switch away from other views, view = %v", m.currentView)
	}
}
