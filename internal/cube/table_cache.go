package cube

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func tableCachePath(filename string) string {
	path := coordinateCachePath()
	if path == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(path), filename)
}

// Check the target before opening it: opening a FIFO for reading can block
// before File.Stat or any deadline check. Stat also follows symlink targets.
// Check the opened descriptor again to validate its type and bounded size.
func openTableCache(path string, minSize, maxSize int64, deadline time.Time) (*os.File, int64, error) {
	if tableDeadlineExceeded(deadline) {
		return nil, 0, context.DeadlineExceeded
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}
	if !info.Mode().IsRegular() || info.Size() < minSize || info.Size() > maxSize {
		return nil, 0, fmt.Errorf("invalid table cache: %s", path)
	}
	if tableDeadlineExceeded(deadline) {
		return nil, 0, context.DeadlineExceeded
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	info, err = f.Stat()
	if err == nil && (!info.Mode().IsRegular() || info.Size() < minSize || info.Size() > maxSize) {
		err = fmt.Errorf("invalid table cache: %s", path)
	}
	if err == nil && tableDeadlineExceeded(deadline) {
		err = context.DeadlineExceeded
	}
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

func createTableCache(path, pattern string, deadline time.Time) (*os.File, error) {
	if tableDeadlineExceeded(deadline) {
		return nil, context.DeadlineExceeded
	}
	if path == "" {
		return nil, fmt.Errorf("table cache directory is unavailable")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	if tableDeadlineExceeded(deadline) {
		return nil, context.DeadlineExceeded
	}
	return os.CreateTemp(filepath.Dir(path), pattern)
}

// Publish only a complete, closed file. Persistence callers receive errors
// from directory creation, writing, closing and rename; solvers can explicitly
// ignore them when the cache is optional.
func writeTableCache(path, pattern string, deadline time.Time, write func(*os.File) error) error {
	f, err := createTableCache(path, pattern, deadline)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	writeErr := write(f)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if tableDeadlineExceeded(deadline) {
		return context.DeadlineExceeded
	}
	return os.Rename(f.Name(), path)
}

type tableDeadlineWriter struct {
	w        io.Writer
	deadline time.Time
}

func (w tableDeadlineWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		if tableDeadlineExceeded(w.deadline) {
			return written, context.DeadlineExceeded
		}
		chunk := p[:min(len(p), 64<<10)]
		n, err := w.w.Write(chunk)
		written += n
		if err != nil {
			return written, err
		}
		if n != len(chunk) {
			return written, io.ErrShortWrite
		}
		p = p[n:]
	}
	if tableDeadlineExceeded(w.deadline) {
		return written, context.DeadlineExceeded
	}
	return written, nil
}
