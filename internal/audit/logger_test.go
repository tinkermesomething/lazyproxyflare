package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	t.Run("creates directory if missing", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nested", "audit")
		logger, err := NewLogger(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if logger == nil {
			t.Fatal("expected non-nil logger")
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("expected directory to be created")
		}
	})

	t.Run("works with existing directory", func(t *testing.T) {
		dir := t.TempDir()
		logger, err := NewLogger(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if logger == nil {
			t.Fatal("expected non-nil logger")
		}
	})
}

func TestLog(t *testing.T) {
	dir := t.TempDir()
	logger, _ := NewLogger(dir)

	entry := LogEntry{
		Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Operation:  OperationCreate,
		EntityType: EntityDNS,
		Domain:     "test.example.com",
		Result:     ResultSuccess,
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created and has content
	content, err := os.ReadFile(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if len(content) == 0 {
		t.Error("expected non-empty log file")
	}
}

func TestLogAutoTimestamp(t *testing.T) {
	dir := t.TempDir()
	logger, _ := NewLogger(dir)

	before := time.Now()
	entry := LogEntry{
		Operation:  OperationDelete,
		EntityType: EntityCaddy,
		Domain:     "test.example.com",
		Result:     ResultSuccess,
	}
	if err := logger.Log(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logs, _ := logger.LoadLogs()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(logs))
	}
	if logs[0].Timestamp.Before(before) {
		t.Error("auto-timestamp should be set to current time")
	}
}

func TestLoadLogs(t *testing.T) {
	t.Run("no log file returns empty", func(t *testing.T) {
		dir := t.TempDir()
		logger, _ := NewLogger(dir)

		logs, err := logger.LoadLogs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 0 {
			t.Errorf("expected 0 entries, got %d", len(logs))
		}
	})

	t.Run("parses multiple entries", func(t *testing.T) {
		dir := t.TempDir()
		logger, _ := NewLogger(dir)

		for i := 0; i < 3; i++ {
			logger.Log(LogEntry{
				Operation:  OperationCreate,
				EntityType: EntityDNS,
				Domain:     "test.example.com",
				Result:     ResultSuccess,
			})
		}

		logs, err := logger.LoadLogs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 3 {
			t.Errorf("expected 3 entries, got %d", len(logs))
		}
	})

	t.Run("handles malformed lines gracefully", func(t *testing.T) {
		dir := t.TempDir()
		logFile := filepath.Join(dir, "audit.log")
		content := `{"operation":"create","entity_type":"dns","domain":"good.com","result":"success","timestamp":"2024-01-01T00:00:00Z"}
not valid json
{"operation":"delete","entity_type":"caddy","domain":"also-good.com","result":"success","timestamp":"2024-01-01T00:00:00Z"}
`
		os.WriteFile(logFile, []byte(content), 0644)

		logger, _ := NewLogger(dir)
		logs, err := logger.LoadLogs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 2 {
			t.Errorf("expected 2 valid entries (skipping malformed), got %d", len(logs))
		}
	})
}

func TestRotateLogs(t *testing.T) {
	t.Run("keeps last N entries", func(t *testing.T) {
		dir := t.TempDir()
		logger, _ := NewLogger(dir)

		for i := 0; i < 10; i++ {
			logger.Log(LogEntry{
				Operation:  OperationCreate,
				EntityType: EntityDNS,
				Domain:     "test.example.com",
				Result:     ResultSuccess,
				BatchCount: i,
			})
		}

		if err := logger.RotateLogs(3); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		logs, _ := logger.LoadLogs()
		if len(logs) != 3 {
			t.Errorf("expected 3 entries after rotation, got %d", len(logs))
		}
		// Should keep the last 3 (BatchCount 7, 8, 9)
		if logs[0].BatchCount != 7 {
			t.Errorf("expected first remaining entry BatchCount=7, got %d", logs[0].BatchCount)
		}
	})

	t.Run("no-op when under limit", func(t *testing.T) {
		dir := t.TempDir()
		logger, _ := NewLogger(dir)

		logger.Log(LogEntry{
			Operation:  OperationCreate,
			EntityType: EntityDNS,
			Domain:     "test.example.com",
			Result:     ResultSuccess,
		})

		if err := logger.RotateLogs(100); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		logs, _ := logger.LoadLogs()
		if len(logs) != 1 {
			t.Errorf("expected 1 entry (unchanged), got %d", len(logs))
		}
	})
}
