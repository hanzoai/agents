package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeLog creates <projectDir>/packages/mcp/<alias>/<alias>.log with the
// given lines (one per line, each terminated by a newline).
func writeLog(t *testing.T, projectDir, alias string, lines []string) {
	t.Helper()
	dir := filepath.Join(projectDir, "packages", "mcp", alias)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	if err := os.WriteFile(filepath.Join(dir, alias+".log"), []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
}

func tenLines() []string {
	lines := make([]string, 0, 10)
	for i := 1; i <= 10; i++ {
		lines = append(lines, fmt.Sprintf("line%d", i))
	}
	return lines
}

// drain reads rc to completion; err is passed through unchanged so callers
// can forward a two-value method result directly.
func drain(rc io.ReadCloser, err error) (string, error) {
	if err != nil {
		return "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func TestManagerLogsHonorsLines(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "srv", tenLines())
	m := &MCPManager{projectDir: dir, processes: map[string]*MCPProcess{}}

	all := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"

	got, err := drain(m.Logs("srv", false, 3))
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if want := "line8\nline9\nline10\n"; got != want {
		t.Errorf("lines=3: got %q, want %q", got, want)
	}

	for _, n := range []int{0, -1, 50} {
		got, err := drain(m.Logs("srv", false, n))
		if err != nil {
			t.Fatalf("Logs(lines=%d): %v", n, err)
		}
		if got != all {
			t.Errorf("lines=%d: got %q, want the whole file", n, got)
		}
	}
}

func TestProcessLogsHonorsLines(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "srv", tenLines())
	pm := &ProcessManager{projectDir: dir}

	got, err := drain(pm.GetProcessLogs("srv", false, 3))
	if err != nil {
		t.Fatalf("GetProcessLogs: %v", err)
	}
	if want := "line8\nline9\nline10\n"; got != want {
		t.Errorf("lines=3: got %q, want %q", got, want)
	}

	got, err = drain(pm.GetProcessLogs("srv", false, 0))
	if err != nil {
		t.Fatalf("GetProcessLogs(lines=0): %v", err)
	}
	if want := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"; got != want {
		t.Errorf("lines=0: got %q, want whole file", got)
	}
}

func TestLogsMissingFile(t *testing.T) {
	dir := t.TempDir()
	m := &MCPManager{projectDir: dir, processes: map[string]*MCPProcess{}}
	if _, err := m.Logs("ghost", false, 3); err == nil ||
		!strings.Contains(err.Error(), "log file not found for server ghost") {
		t.Errorf("missing log: got %v, want the log-not-found error", err)
	}

	pm := &ProcessManager{projectDir: dir}
	if _, err := pm.GetProcessLogs("ghost", false, 3); err == nil ||
		!strings.Contains(err.Error(), "log file not found for server ghost") {
		t.Errorf("missing log: got %v, want the log-not-found error", err)
	}
}

func TestMethodsShareOneImplementation(t *testing.T) {
	// Both paths must go through readServerLog: identical alias directory
	// layout and identical `lines` semantics for the same fixture.
	dir := t.TempDir()
	writeLog(t, dir, "shared", tenLines())
	m := &MCPManager{projectDir: dir, processes: map[string]*MCPProcess{}}
	pm := &ProcessManager{projectDir: dir}

	a, err := drain(m.Logs("shared", false, 4))
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	b, err := drain(pm.GetProcessLogs("shared", false, 4))
	if err != nil {
		t.Fatalf("GetProcessLogs: %v", err)
	}
	if a != b {
		t.Errorf("implementations diverged: manager=%q process=%q", a, b)
	}
	if want := "line7\nline8\nline9\nline10\n"; a != want {
		t.Errorf("lines=4: got %q, want %q", a, want)
	}
}
