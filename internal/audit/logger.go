package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// OperationType represents the type of operation performed
type OperationType string

const (
	OperationCreate      OperationType = "create"
	OperationUpdate      OperationType = "update"
	OperationDelete      OperationType = "delete"
	OperationSync        OperationType = "sync"
	OperationBatchDelete OperationType = "batch_delete"
	OperationBatchSync   OperationType = "batch_sync"
	OperationRestore     OperationType = "restore"
)

// EntityType represents what entity was affected
type EntityType string

const (
	EntityDNS   EntityType = "dns"
	EntityCaddy EntityType = "caddy"
	EntityBoth  EntityType = "both"
)

// Result represents the outcome of an operation
type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
)

// LogEntry represents a single audit log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Operation   OperationType          `json:"operation"`
	EntityType  EntityType             `json:"entity_type"`
	Domain      string                 `json:"domain"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Result      Result                 `json:"result"`
	Error       string                 `json:"error,omitempty"`
	BatchCount  int                    `json:"batch_count,omitempty"` // For batch operations
}

// Logger handles audit logging operations
type Logger struct {
	logPath string
}

// NewLogger creates a new audit logger
func NewLogger(configDir string) (*Logger, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	logPath := filepath.Join(configDir, "audit.log")
	return &Logger{logPath: logPath}, nil
}

// Log appends a new entry to the audit log
func (l *Logger) Log(entry LogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Open log file in append mode
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer file.Close()

	// Marshal entry to JSON
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Write JSON line
	if _, err := file.Write(append(jsonBytes, '\n')); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// LoadLogs reads all log entries from the audit log file
func (l *Logger) LoadLogs() ([]LogEntry, error) {
	// Check if log file exists
	if _, err := os.Stat(l.logPath); os.IsNotExist(err) {
		// No log file yet - return empty slice
		return []LogEntry{}, nil
	}

	// Read entire file
	content, err := os.ReadFile(l.logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit log: %w", err)
	}

	// Parse JSON lines
	var entries []LogEntry
	lines := splitLines(string(content))

	for i, line := range lines {
		if line == "" {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Log parsing error but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to parse log line %d: %v\n", i+1, err)
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// RotateLogs keeps only the last maxEntries in the log file
func (l *Logger) RotateLogs(maxEntries int) error {
	entries, err := l.LoadLogs()
	if err != nil {
		return fmt.Errorf("failed to load logs for rotation: %w", err)
	}

	// If under limit, nothing to do
	if len(entries) <= maxEntries {
		return nil
	}

	// Keep only the last maxEntries
	entries = entries[len(entries)-maxEntries:]

	// Rewrite log file with truncated entries
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log for rotation: %w", err)
	}
	defer file.Close()

	for _, entry := range entries {
		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry during rotation: %w", err)
		}

		if _, err := file.Write(append(jsonBytes, '\n')); err != nil {
			return fmt.Errorf("failed to write log entry during rotation: %w", err)
		}
	}

	return nil
}

// splitLines splits a string by newlines, handling both \n and \r\n
func splitLines(s string) []string {
	var lines []string
	var line string

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, line)
			line = ""
		} else if s[i] != '\r' {
			line += string(s[i])
		}
	}

	// Add last line if not empty
	if line != "" {
		lines = append(lines, line)
	}

	return lines
}
