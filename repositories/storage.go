package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// fileStore wraps a single JSON-on-disk file with a mutex so concurrent
// readers and writers cannot corrupt it, and with an atomic temp-file +
// rename pattern so a crash mid-write cannot leave the file empty.
//
// This is the single source of truth for every JSON persistence path in
// the repositories package. Before it existed each repository re-
// implemented the open/decode/seek/truncate/encode dance with no locking,
// racing on every write. Keep all new persistence going through a
// fileStore.
type fileStore struct {
	mu   sync.RWMutex
	path string
}

// HealthCheck writes a short probe file in the current working
// directory (the same place the JSON stores live), syncs it, re-reads
// it, and deletes it. Any failure returns an error so the caller can
// convert it to an HTTP 503. Used by the /ready readiness endpoint
// to tell a container orchestrator that the app has actually lost
// disk access, not just that the process is still running.
func HealthCheck() error {
	dir := "."
	f, err := os.CreateTemp(dir, ".health-probe-*")
	if err != nil {
		return fmt.Errorf("create probe: %w", err)
	}
	name := f.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := f.Write([]byte("ok")); err != nil {
		_ = f.Close()
		return fmt.Errorf("write probe: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("sync probe: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close probe: %w", err)
	}
	if _, err := os.ReadFile(name); err != nil {
		return fmt.Errorf("read probe: %w", err)
	}
	return nil
}

func newFileStore(path string) *fileStore {
	return &fileStore{path: path}
}

// load decodes the file contents into out. If the file does not exist or
// is empty, out is left untouched and nil is returned.
func (s *fileStore) load(out any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadLocked(out)
}

func (s *fileStore) loadLocked(out any) error {
	f, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return nil
	}

	if err := json.NewDecoder(f).Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

// saveLocked serializes in to JSON and atomically replaces the target
// file. The write goes to a sibling temp file which is then os.Rename'd
// over the target — on POSIX this is atomic, so readers either see the
// old contents or the new contents, never a half-written file. Errors
// during cleanup are intentionally dropped: we already have a primary
// error to return, and logging a second error would only confuse the
// caller.
func (s *fileStore) saveLocked(in any) error {
	dir := filepath.Dir(s.path)
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(s.path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	enc := json.NewEncoder(tmp)
	if err := enc.Encode(in); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}
