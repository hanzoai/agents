package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogsHonorsLines verifies that MCPManager.Logs returns only the last N
// lines of the log file (the core bug being fixed: `lines` was ignored).
func TestLogsHonorsLines(t *testing.T) {
	dir := t.TempDir()
	alias := "srv"
	logDir := filepath.Join(dir, "packages", "mcp", alias)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&sb, "line %d\n", i)
	}
	if err := os.WriteFile(filepath.Join(logDir, alias+".log"), []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &MCPManager{projectDir: dir}
	r, err := m.Logs(alias, false, 10)
	if err != nil {
		t.Fatalf("Logs error: %v", err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	got := strings.TrimRight(string(b), "\n")
	lines := strings.Split(got, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d:\n%s", len(lines), got)
	}
	if lines[0] != "line 91" || lines[len(lines)-1] != "line 100" {
		t.Fatalf("wrong tail window: first=%q last=%q", lines[0], lines[len(lines)-1])
	}
}

// TestLogsAllLinesWhenZero verifies lines<=0 returns the whole file.
func TestLogsAllLinesWhenZero(t *testing.T) {
	dir := t.TempDir()
	alias := "srv"
	logDir := filepath.Join(dir, "packages", "mcp", alias)
	os.MkdirAll(logDir, 0o755)
	os.WriteFile(filepath.Join(logDir, alias+".log"), []byte("a\nb\nc\n"), 0o644)

	m := &MCPManager{projectDir: dir}
	r, err := m.Logs(alias, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	if string(b) != "a\nb\nc\n" {
		t.Fatalf("expected full file, got %q", string(b))
	}
}
