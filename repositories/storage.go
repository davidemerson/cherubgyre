package repositories

import (
	"encoding/json"
	"errors"
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

func newFileStore(path string) *fileStore {
	return &fileStore{path: path}
}

// load decodes the file contents into out. If the file does not exist or
// is empty, out is left untouched and nil is returned.
func (s *fileStore) load(out interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadLocked(out)
}

func (s *fileStore) loadLocked(out interface{}) error {
	f, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

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

// save serializes in to JSON and atomically replaces the target file.
// The write goes to a sibling temp file which is then os.Rename'd over
// the target — on POSIX this is atomic, so readers either see the old
// contents or the new contents, never a half-written file.
func (s *fileStore) save(in interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(in)
}

func (s *fileStore) saveLocked(in interface{}) error {
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
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// mutate atomically runs fn under the write lock, letting the caller
// load-mutate-save without dropping the lock in between. fn receives a
// fresh decode of the current file contents (or an empty value if the
// file doesn't exist yet), mutates it, and returns the value to persist.
// If fn returns an error, no write happens.
func (s *fileStore) mutate(initial interface{}, fn func() (interface{}, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.loadLocked(initial); err != nil {
		return err
	}
	out, err := fn()
	if err != nil {
		return err
	}
	return s.saveLocked(out)
}
