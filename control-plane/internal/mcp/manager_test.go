package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPManager_Logs(t *testing.T) {
	tmpDir := t.TempDir()
	alias := "testserver"
	logDir := filepath.Join(tmpDir, "packages", "mcp", alias)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	var tenLines []string
	for i := 1; i <= 10; i++ {
		tenLines = append(tenLines, fmt.Sprintf("line %d", i))
	}
	content := strings.Join(tenLines, "\n") + "\n"
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", alias))
	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	mgr := &MCPManager{projectDir: tmpDir}
	pm := &ProcessManager{projectDir: tmpDir}

	readAll := func(rc io.ReadCloser) string {
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("failed to read: %v", err)
		}
		return string(b)
	}

	tests := []struct {
		name     string
		lines    int
		expected string
	}{
		{
			name:     "lines=3 returns last 3 lines",
			lines:    3,
			expected: "line 8\nline 9\nline 10\n",
		},
		{
			name:     "lines=0 returns all 10 lines",
			lines:    0,
			expected: content,
		},
		{
			name:     "lines=-1 returns all 10 lines",
			lines:    -1,
			expected: content,
		},
		{
			name:     "lines=50 returns all 10 lines",
			lines:    50,
			expected: content,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test MCPManager.Logs
			rc1, err := mgr.Logs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("MCPManager.Logs failed: %v", err)
			}
			out1 := readAll(rc1)
			if out1 != tc.expected {
				t.Errorf("MCPManager.Logs: got %q, want %q", out1, tc.expected)
			}

			// Test ProcessManager.GetProcessLogs
			rc2, err := pm.GetProcessLogs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("ProcessManager.GetProcessLogs failed: %v", err)
			}
			out2 := readAll(rc2)
			if out2 != tc.expected {
				t.Errorf("ProcessManager.GetProcessLogs: got %q, want %q", out2, tc.expected)
			}
		})
	}

	t.Run("missing log returns error", func(t *testing.T) {
		_, err := mgr.Logs("nonexistent", false, 0)
		if err == nil {
			t.Fatal("expected error for nonexistent server log, got nil")
		}
		expectedErr := "log file not found for server nonexistent"
		if err.Error() != expectedErr {
			t.Errorf("got error %q, want %q", err.Error(), expectedErr)
		}

		_, err = pm.GetProcessLogs("nonexistent", false, 0)
		if err == nil {
			t.Fatal("expected error for nonexistent server log, got nil")
		}
		if err.Error() != expectedErr {
			t.Errorf("got error %q, want %q", err.Error(), expectedErr)
		}
	})
}
