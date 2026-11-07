package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestLog(t *testing.T, projectDir, alias string, lines []string) {
	t.Helper()
	dir := filepath.Join(projectDir, "packages", "mcp", alias)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir log dir: %v", err)
	}
	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, alias+".log"), []byte(content), 0o644); err != nil {
		t.Fatalf("write log file: %v", err)
	}
}

func readAll(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	return string(data)
}

func tenLines() []string {
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%d", i+1)
	}
	return lines
}

func TestMCPManager_Logs_Lines(t *testing.T) {
	dir := t.TempDir()
	writeTestLog(t, dir, "test-server", tenLines())
	m := &MCPManager{projectDir: dir}

	cases := []struct {
		name  string
		lines int
		want  string
	}{
		{"last 3 lines", 3, "line8\nline9\nline10\n"},
		{"zero returns all", 0, "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"},
		{"negative returns all", -1, "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"},
		{"more than file returns all", 50, "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rc, err := m.Logs("test-server", false, tc.lines)
			if err != nil {
				t.Fatalf("Logs() error: %v", err)
			}
			if got := readAll(t, rc); got != tc.want {
				t.Errorf("Logs(lines=%d) = %q, want %q", tc.lines, got, tc.want)
			}
		})
	}
}

func TestMCPManager_Logs_MissingFile(t *testing.T) {
	m := &MCPManager{projectDir: t.TempDir()}
	_, err := m.Logs("missing-server", false, 0)
	if err == nil {
		t.Fatal("Logs() expected error for missing log file, got nil")
	}
	if !strings.Contains(err.Error(), "log file not found for server missing-server") {
		t.Errorf("Logs() error = %q, want it to contain %q", err.Error(), "log file not found for server missing-server")
	}
}

func TestProcessManager_GetProcessLogs_Lines(t *testing.T) {
	dir := t.TempDir()
	writeTestLog(t, dir, "test-server", tenLines())
	pm := &ProcessManager{projectDir: dir}

	rc, err := pm.GetProcessLogs("test-server", false, 3)
	if err != nil {
		t.Fatalf("GetProcessLogs() error: %v", err)
	}
	if got := readAll(t, rc); got != "line8\nline9\nline10\n" {
		t.Errorf("GetProcessLogs(lines=3) = %q, want %q", got, "line8\nline9\nline10\n")
	}

	_, err = pm.GetProcessLogs("missing-server", false, 0)
	if err == nil || !strings.Contains(err.Error(), "log file not found for server missing-server") {
		t.Errorf("GetProcessLogs() missing-file error = %v, want log-file-not-found error", err)
	}
}
