package mcp

import (
	"bufio"
	"io"
	"os"
	"time"
)

// tailReader returns a reader over the last `lines` lines of the already-open
// file. If lines <= 0 it returns the file from the beginning. The returned
// ReadCloser, when closed, closes the underlying file.
//
// Implementation: scan the file to count offsets of line starts, keep only the
// last N offsets, then seek to the start of the earliest retained line.
func tailReader(file *os.File, lines int) (io.ReadCloser, error) {
	if lines <= 0 {
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		return file, nil
	}

	// Record byte offset at the start of each line.
	var offsets []int64
	offsets = append(offsets, 0)
	r := bufio.NewReaderSize(file, 64*1024)
	var pos int64
	for {
		chunk, err := r.ReadBytes('\n')
		pos += int64(len(chunk))
		if err != nil {
			break
		}
		offsets = append(offsets, pos)
		if len(offsets) > lines*4 { // keep memory bounded but always keep tail
			offsets = offsets[len(offsets)-lines:]
			// adjust: offsets are absolute, so trimming the head is fine
		}
	}

	start := int64(0)
	if len(offsets) > lines {
		start = offsets[len(offsets)-lines]
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}
	return file, nil
}

// followReader wraps a file and, after EOF, blocks and re-reads as new bytes are
// appended (like `tail -f`). Close stops the follow loop and closes the file.
type followReader struct {
	f    *os.File
	done chan struct{}

}

func newFollowReader(f *os.File) *followReader {
	fr := &followReader{f: f, done: make(chan struct{})}
	return fr
}

func (fr *followReader) Read(p []byte) (int, error) {
	for {
		n, err := fr.f.Read(p)
		if n > 0 {
			return n, nil
		}
		if err == io.EOF {
			select {
			case <-fr.done:
				return 0, io.EOF
			case <-time.After(250 * time.Millisecond):
				// new data may have been appended; retry
				continue
			}
		}
		if err != nil {
			return n, err
		}
	}
}

func (fr *followReader) Close() error {
	select {
	case <-fr.done:
	default:
		close(fr.done)
	}
	return fr.f.Close()
}
