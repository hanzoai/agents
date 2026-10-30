package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hanzoai/agents/control-plane/internal/config"
)

// setupTestLog creates a 10-line log file at packages/mcp/<alias>/<alias>.log within tempDir.
func setupTestLog(t *testing.T, tempDir, alias string, numLines int) string {
	t.Helper()
	logDir := filepath.Join(tempDir, "packages", "mcp", alias)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log directory: %v", err)
	}

	var sb strings.Builder
	for i := 1; i <= numLines; i++ {
		sb.WriteString(fmt.Sprintf("line %d\n", i))
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", alias))
	if err := os.WriteFile(logFile, []byte(sb.String()), 0644); err != nil {
		t.Fatalf("failed to write test log file: %v", err)
	}

	return logFile
}

func readAllAndClose(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read from log reader: %v", err)
	}
	return string(data)
}

func TestMCPManagerLogs_Lines(t *testing.T) {
	tempDir := t.TempDir()
	alias := "my-server"
	setupTestLog(t, tempDir, alias, 10)

	mgr := NewMCPManager(&config.Config{}, tempDir, false)

	expectedAll10 := "line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\nline 9\nline 10\n"
	expectedLast3 := "line 8\nline 9\nline 10\n"

	tests := []struct {
		name     string
		lines    int
		expected string
	}{
		{
			name:     "lines=3 returns exactly the last 3 lines",
			lines:    3,
			expected: expectedLast3,
		},
		{
			name:     "lines=0 returns all 10 lines",
			lines:    0,
			expected: expectedAll10,
		},
		{
			name:     "lines=-1 returns all 10 lines",
			lines:    -1,
			expected: expectedAll10,
		},
		{
			name:     "lines=50 returns all 10 lines",
			lines:    50,
			expected: expectedAll10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rc, err := mgr.Logs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
			got := readAllAndClose(t, rc)
			if got != tc.expected {
				t.Errorf("%s failed:\nexpected:\n%q\ngot:\n%q", tc.name, tc.expected, got)
			}
		})
	}
}

func TestMCPManagerLogs_MissingLog(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewMCPManager(&config.Config{}, tempDir, false)

	_, err := mgr.Logs("non-existent-server", false, 10)
	if err == nil {
		t.Fatal("expected error for non-existent server log, got nil")
	}

	expectedErr := "log file not found for server non-existent-server"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestProcessManagerGetProcessLogs_Lines(t *testing.T) {
	tempDir := t.TempDir()
	alias := "process-server"
	setupTestLog(t, tempDir, alias, 10)

	template := NewTemplateProcessor(tempDir, false)
	pm := NewProcessManager(tempDir, template, false)

	expectedAll10 := "line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\nline 9\nline 10\n"
	expectedLast3 := "line 8\nline 9\nline 10\n"

	tests := []struct {
		name     string
		lines    int
		expected string
	}{
		{
			name:     "lines=3 returns exactly the last 3 lines",
			lines:    3,
			expected: expectedLast3,
		},
		{
			name:     "lines=0 returns all 10 lines",
			lines:    0,
			expected: expectedAll10,
		},
		{
			name:     "lines=-1 returns all 10 lines",
			lines:    -1,
			expected: expectedAll10,
		},
		{
			name:     "lines=50 returns all 10 lines",
			lines:    50,
			expected: expectedAll10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rc, err := pm.GetProcessLogs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
			got := readAllAndClose(t, rc)
			if got != tc.expected {
				t.Errorf("%s failed:\nexpected:\n%q\ngot:\n%q", tc.name, tc.expected, got)
			}
		})
	}

	// Missing log check
	_, err := pm.GetProcessLogs("missing-alias", false, 5)
	if err == nil {
		t.Fatal("expected error for missing log, got nil")
	}
	expectedErr := "log file not found for server missing-alias"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestMCPManagerLogs_EmptyLog(t *testing.T) {
	tempDir := t.TempDir()
	alias := "empty-server"
	logDir := filepath.Join(tempDir, "packages", "mcp", alias)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log directory: %v", err)
	}
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", alias))
	if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty log file: %v", err)
	}

	mgr := NewMCPManager(&config.Config{}, tempDir, false)

	rc, err := mgr.Logs(alias, false, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := readAllAndClose(t, rc)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
