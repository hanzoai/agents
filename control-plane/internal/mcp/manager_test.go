package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPManagerAndProcessManagerLogs(t *testing.T) {
	tempDir := t.TempDir()
	alias := "test-server"
	logDir := filepath.Join(tempDir, "packages", "mcp", alias)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	var tenLines []string
	for i := 1; i <= 10; i++ {
		tenLines = append(tenLines, fmt.Sprintf("line-%d", i))
	}
	fullLog := strings.Join(tenLines, "\n") + "\n"
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", alias))
	if err := os.WriteFile(logFile, []byte(fullLog), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	mgr := NewMCPManager(nil, tempDir, false)

	tests := []struct {
		name      string
		lines     int
		wantLines []string
	}{
		{
			name:      "lines=3 returns last 3 lines",
			lines:     3,
			wantLines: []string{"line-8", "line-9", "line-10"},
		},
		{
			name:      "lines=0 returns all 10 lines",
			lines:     0,
			wantLines: tenLines,
		},
		{
			name:      "lines=-1 returns all 10 lines",
			lines:     -1,
			wantLines: tenLines,
		},
		{
			name:      "lines=50 returns all 10 lines",
			lines:     50,
			wantLines: tenLines,
		},
	}

	for _, tc := range tests {
		t.Run("MCPManager_"+tc.name, func(t *testing.T) {
			rc, err := mgr.Logs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("Logs(%d) returned unexpected error: %v", tc.lines, err)
			}
			defer rc.Close()

			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}

			got := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
			if len(got) != len(tc.wantLines) {
				t.Fatalf("expected %d lines, got %d (%q)", len(tc.wantLines), len(got), string(b))
			}
			for i := range tc.wantLines {
				if got[i] != tc.wantLines[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tc.wantLines[i])
				}
			}
		})

		t.Run("ProcessManager_"+tc.name, func(t *testing.T) {
			rc, err := mgr.processManager.GetProcessLogs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("GetProcessLogs(%d) returned unexpected error: %v", tc.lines, err)
			}
			defer rc.Close()

			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}

			got := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
			if len(got) != len(tc.wantLines) {
				t.Fatalf("expected %d lines, got %d", len(tc.wantLines), len(got))
			}
		})
	}

	t.Run("missing log file returns expected error", func(t *testing.T) {
		_, err := mgr.Logs("nonexistent", false, 5)
		if err == nil {
			t.Fatal("expected error for missing log file, got nil")
		}
		wantErr := "log file not found for server nonexistent"
		if err.Error() != wantErr {
			t.Fatalf("got error %q, want %q", err.Error(), wantErr)
		}
	})
}
