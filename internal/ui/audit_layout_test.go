package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"lazyproxyflare/internal/audit"
)

// newAuditModel builds n log entries of mixed height (1 to 4 lines), oldest first.
func newAuditModel(n, width, height int) Model {
	logs := make([]audit.LogEntry, n)
	for i := range logs {
		e := audit.LogEntry{
			Timestamp:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local).Add(time.Duration(i) * time.Minute),
			Operation:  audit.OperationUpdate,
			EntityType: audit.EntityBoth,
			Domain:     fmt.Sprintf("d%03d.example.com", i),
			Result:     audit.ResultSuccess,
		}
		if i%3 == 0 {
			e.Details = map[string]interface{}{"dns_type": "CNAME", "target": "t.example.com"}
		}
		if i%5 == 0 {
			e.Result = audit.ResultFailure
			e.Error = "boom"
		}
		if i%7 == 0 {
			e.BatchCount = 3
		}
		logs[i] = e
	}
	m := createTestModel()
	m.width, m.height = width, height
	m.currentView = ViewAuditLog
	m.audit.Logs = logs
	return m
}

func auditDomain(i int) string { return fmt.Sprintf("d%03d.example.com", i) }

// Newest entries are shown first, so display index k is log index n-1-k.
func TestAuditLogScrollKeepsHintAndReachesOldestEntry(t *testing.T) {
	const n = 40
	for _, sz := range screenSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := newAuditModel(n, sz.w, sz.h)

			check := func(step string) {
				out := stripANSI(m.renderAuditLogView())
				if !strings.Contains(out, "ESC: close") {
					t.Fatalf("%s (scroll %d): hint line was pushed out of the modal:\n%s", step, m.audit.Scroll, out)
				}
				if want := auditDomain(n - 1 - m.audit.Scroll); !strings.Contains(out, want) {
					t.Fatalf("%s (scroll %d): first entry %s not shown:\n%s", step, m.audit.Scroll, want, out)
				}
			}

			check("start")
			for i := 0; i < 500; i++ {
				m, _ = m.handleNavigateDown()
				if i < 100 {
					check("down")
				}
			}
			check("end")
			if got, want := m.audit.Scroll, m.auditMaxScroll(); got != want {
				t.Errorf("Scroll after scrolling to the end = %d, want %d", got, want)
			}
			if out := stripANSI(m.renderAuditLogView()); !strings.Contains(out, auditDomain(0)) {
				t.Errorf("oldest entry not reachable at max scroll:\n%s", out)
			}

			for m.audit.Scroll > 0 {
				m, _ = m.handleNavigateUp()
				check("up")
			}
		})
	}
}

func TestAuditLogShortListDoesNotScroll(t *testing.T) {
	m := newAuditModel(2, 179, 49)
	if got := m.auditMaxScroll(); got != 0 {
		t.Errorf("auditMaxScroll() = %d for a list that fits, want 0", got)
	}
	m, _ = m.handleNavigateDown()
	if m.audit.Scroll != 0 {
		t.Errorf("scrolled a list that fits: Scroll = %d", m.audit.Scroll)
	}
}
