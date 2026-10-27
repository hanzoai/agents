package mcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// readServerLog opens the log file for the MCP server identified by alias
// under <projectDir>/packages/mcp/<alias>/<alias>.log.
//
// When lines > 0 only the last lines lines are returned; when lines <= 0 the
// whole file is returned, as before. follow is not supported yet: callers
// should treat it as a no-op until log tailing is implemented.
func readServerLog(projectDir string, alias string, lines int) (io.ReadCloser, error) {
	logFile := filepath.Join(projectDir, "packages", "mcp", alias, fmt.Sprintf("%s.log", alias))

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("log file not found for server %s", alias)
	}

	if lines <= 0 {
		file, err := os.Open(logFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		return file, nil
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// TODO: support follow (stream the file as it grows).
	content := string(data)
	endsWithNewline := strings.HasSuffix(content, "\n")
	split := strings.Split(content, "\n")
	if endsWithNewline {
		// Drop the empty trailing element produced by the final newline.
		split = split[:len(split)-1]
	}
	if len(split) > lines {
		split = split[len(split)-lines:]
	}

	out := strings.Join(split, "\n")
	if endsWithNewline && len(split) > 0 {
		out += "\n"
	}

	return io.NopCloser(strings.NewReader(out)), nil
}
