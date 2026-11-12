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

func TestMCPManagerLogsAndGetProcessLogs_Lines(t *testing.T) {
	projectDir := t.TempDir()
	alias := "test-server"
	logDir := filepath.Join(projectDir, "packages", "mcp", alias)
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

	mgr := NewMCPManager(&config.Config{}, projectDir, false)
	lastThree := strings.Join(tenLines[7:], "\n") + "\n"

	cases := []struct {
		name string
		n    int
		want string
	}{
		{"last_3_lines", 3, lastThree},
		{"zero_lines_returns_all", 0, fullLog},
		{"negative_lines_returns_all", -1, fullLog},
		{"more_than_total_returns_all", 50, fullLog},
	}

	for _, tc := range cases {
		t.Run("MCPManager.Logs_"+tc.name, func(t *testing.T) {
			rc, err := mgr.Logs(alias, false, tc.n)
			if err != nil {
				t.Fatalf("Logs(%d) returned error: %v", tc.n, err)
			}
			defer rc.Close()
			got, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("Logs(%d) = %q, want %q", tc.n, string(got), tc.want)
			}
		})

		t.Run("ProcessManager.GetProcessLogs_"+tc.name, func(t *testing.T) {
			rc, err := mgr.processManager.GetProcessLogs(alias, false, tc.n)
			if err != nil {
				t.Fatalf("GetProcessLogs(%d) returned error: %v", tc.n, err)
			}
			defer rc.Close()
			got, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("GetProcessLogs(%d) = %q, want %q", tc.n, string(got), tc.want)
			}
		})
	}
}
