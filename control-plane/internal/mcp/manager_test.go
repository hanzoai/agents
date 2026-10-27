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
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	var tenLines []string
	for i := 1; i <= 10; i++ {
		tenLines = append(tenLines, fmt.Sprintf("line %d", i))
	}
	fullContent := strings.Join(tenLines, "\n") + "\n"
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", alias))
	if err := os.WriteFile(logPath, []byte(fullContent), 0o644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	mgr := NewMCPManager(nil, tempDir, false)
	pm := NewProcessManager(tempDir, nil, false)

	readAll := func(t *testing.T, rc io.ReadCloser) string {
		t.Helper()
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("failed reading log: %v", err)
		}
		return string(b)
	}

	wantLast3 := "line 8\nline 9\nline 10\n"

	for _, tc := range []struct {
		name  string
		lines int
		want  string
	}{
		{name: "lines=3 returns last 3 lines", lines: 3, want: wantLast3},
		{name: "lines=0 returns all 10 lines", lines: 0, want: fullContent},
		{name: "lines=-1 returns all 10 lines", lines: -1, want: fullContent},
		{name: "lines=50 returns all 10 lines", lines: 50, want: fullContent},
	} {
		t.Run("MCPManager_"+tc.name, func(t *testing.T) {
			rc, err := mgr.Logs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := readAll(t, rc)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})

		t.Run("ProcessManager_"+tc.name, func(t *testing.T) {
			rc, err := pm.GetProcessLogs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := readAll(t, rc)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("missing log returns expected error", func(t *testing.T) {
		_, err := mgr.Logs("missing-alias", false, 5)
		if err == nil || err.Error() != "log file not found for server missing-alias" {
			t.Fatalf("got err %v, want 'log file not found for server missing-alias'", err)
		}

		_, err = pm.GetProcessLogs("missing-alias", false, 5)
		if err == nil || err.Error() != "log file not found for server missing-alias" {
			t.Fatalf("got err %v, want 'log file not found for server missing-alias'", err)
		}
	})
}
