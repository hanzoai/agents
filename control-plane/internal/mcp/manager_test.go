package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTenLineLog(t *testing.T, dir, alias string) string {
	t.Helper()
	logDir := filepath.Join(dir, "packages", "mcp", alias)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}
	var b strings.Builder
	for i := 1; i <= 10; i++ {
		fmt.Fprintf(&b, "line-%d\n", i)
	}
	full := b.String()
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", alias))
	if err := os.WriteFile(logPath, []byte(full), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}
	return full
}

func readAllLog(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read log stream: %v", err)
	}
	return string(data)
}

func TestMCPManagerAndProcessManagerLogs(t *testing.T) {
	dir := t.TempDir()
	alias := "demo-server"
	full := writeTenLineLog(t, dir, alias)
	lastThree := "line-8\nline-9\nline-10\n"

	mgr := NewMCPManager(nil, dir, false)
	pm := NewProcessManager(dir, nil, false)

	cases := []struct {
		name  string
		lines int
		want  string
	}{
		{name: "last_3_lines", lines: 3, want: lastThree},
		{name: "zero_returns_all", lines: 0, want: full},
		{name: "negative_returns_all", lines: -1, want: full},
		{name: "exceeding_total_returns_all", lines: 50, want: full},
	}

	for _, tc := range cases {
		t.Run("MCPManager_"+tc.name, func(t *testing.T) {
			rc, err := mgr.Logs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("MCPManager.Logs(%d) unexpected error: %v", tc.lines, err)
			}
			if got := readAllLog(t, rc); got != tc.want {
				t.Errorf("MCPManager.Logs(%d) = %q, want %q", tc.lines, got, tc.want)
			}
		})

		t.Run("ProcessManager_"+tc.name, func(t *testing.T) {
			rc, err := pm.GetProcessLogs(alias, false, tc.lines)
			if err != nil {
				t.Fatalf("ProcessManager.GetProcessLogs(%d) unexpected error: %v", tc.lines, err)
			}
			if got := readAllLog(t, rc); got != tc.want {
				t.Errorf("ProcessManager.GetProcessLogs(%d) = %q, want %q", tc.lines, got, tc.want)
			}
		})
	}
}

func TestLogsMissingServerLogReturnsError(t *testing.T) {
	dir := t.TempDir()
	mgr := NewMCPManager(nil, dir, false)
	pm := NewProcessManager(dir, nil, false)

	wantErr := "log file not found for server missing-srv"

	if _, err := mgr.Logs("missing-srv", false, 5); err == nil || err.Error() != wantErr {
		t.Fatalf("MCPManager.Logs missing error = %v, want %q", err, wantErr)
	}
	if _, err := pm.GetProcessLogs("missing-srv", false, 5); err == nil || err.Error() != wantErr {
		t.Fatalf("ProcessManager.GetProcessLogs missing error = %v, want %q", err, wantErr)
	}
}
