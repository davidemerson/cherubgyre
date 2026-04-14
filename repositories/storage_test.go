package repositories

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// newTempStore returns a fileStore pointed at a fresh path under
// t.TempDir(), so each test gets isolated state and nothing leaks
// into the repo.
func newTempStore(t *testing.T) *fileStore {
	t.Helper()
	return newFileStore(filepath.Join(t.TempDir(), "store.json"))
}

func TestFileStoreLoadMissingFileIsNil(t *testing.T) {
	s := newTempStore(t)
	var out []string
	if err := s.load(&out); err != nil {
		t.Fatalf("load on missing file should be nil: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil slice on missing file, got %v", out)
	}
}

func TestFileStoreSaveThenLoadRoundTrip(t *testing.T) {
	s := newTempStore(t)
	in := []string{"alpha", "beta", "gamma"}
	if err := func() error {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.saveLocked(in)
	}(); err != nil {
		t.Fatalf("save: %v", err)
	}
	var out []string
	if err := s.load(&out); err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(out) != len(in) {
		t.Fatalf("length mismatch: got %d want %d", len(out), len(in))
	}
	for i := range in {
		if out[i] != in[i] {
			t.Errorf("out[%d]=%q want %q", i, out[i], in[i])
		}
	}
}

func TestFileStoreAtomicReplaceLeavesNoTempFiles(t *testing.T) {
	s := newTempStore(t)
	for i := 0; i < 5; i++ {
		s.mu.Lock()
		err := s.saveLocked([]int{i, i + 1, i + 2})
		s.mu.Unlock()
		if err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
	}
	dir := filepath.Dir(s.path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	// After saves complete, only the final file should remain.
	// Any surviving .tmp-XXXX entries indicate that temp-file cleanup
	// leaked, which would be a disk-space bug.
	for _, e := range entries {
		name := e.Name()
		if name == filepath.Base(s.path) {
			continue
		}
		t.Errorf("unexpected leftover file in store dir: %q", name)
	}
}

func TestFileStoreConcurrentWritersNoCorruption(t *testing.T) {
	s := newTempStore(t)

	// Seed with an empty slice.
	s.mu.Lock()
	if err := s.saveLocked([]int{}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	s.mu.Unlock()

	const writers = 30
	const perWriter = 5
	var appended int64
	var wg sync.WaitGroup
	wg.Add(writers)

	for w := 0; w < writers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				s.mu.Lock()
				var cur []int
				if err := s.loadLocked(&cur); err != nil {
					t.Errorf("writer %d load: %v", id, err)
					s.mu.Unlock()
					return
				}
				cur = append(cur, id*1000+i)
				if err := s.saveLocked(cur); err != nil {
					t.Errorf("writer %d save: %v", id, err)
					s.mu.Unlock()
					return
				}
				s.mu.Unlock()
				atomic.AddInt64(&appended, 1)
			}
		}(w)
	}
	wg.Wait()

	if got := atomic.LoadInt64(&appended); got != writers*perWriter {
		t.Fatalf("appended %d, want %d", got, writers*perWriter)
	}

	// Final file must decode cleanly AND contain exactly writers*perWriter
	// entries — any fewer means a race corrupted the sequence.
	var final []int
	if err := s.load(&final); err != nil {
		t.Fatalf("final load: %v", err)
	}
	if len(final) != writers*perWriter {
		t.Errorf("final file has %d entries, want %d", len(final), writers*perWriter)
	}
}

func TestFileStoreRejectsInvalidJSONGracefully(t *testing.T) {
	s := newTempStore(t)
	// Write garbage directly to the file on disk.
	if err := os.WriteFile(s.path, []byte("not json at all"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var out []string
	err := s.load(&out)
	if err == nil {
		t.Error("expected load to fail on invalid JSON, got nil")
	}
}

func TestFileStoreEmptyFileIsNil(t *testing.T) {
	s := newTempStore(t)
	if err := os.WriteFile(s.path, []byte{}, 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var out []string
	if err := s.load(&out); err != nil {
		t.Fatalf("load on empty file should be nil: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil on empty file, got %v", out)
	}
}

func TestHealthCheckWritesAndCleansUp(t *testing.T) {
	// HealthCheck uses "." so redirect CWD to t.TempDir() for the
	// duration of the test.
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	if err := HealthCheck(); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		t.Errorf("HealthCheck left probe file behind: %q", e.Name())
	}
}

// touchTestJSONGoldenShape keeps the decode/encode paths honest
// across Go version bumps by round-tripping a realistic-looking
// record.
func TestFileStoreEncodesNestedStructure(t *testing.T) {
	s := newTempStore(t)
	type entry struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
		Meta struct {
			Count int `json:"count"`
		} `json:"meta"`
	}
	in := []entry{{Name: "a", Tags: []string{"x", "y"}}, {Name: "b"}}
	in[0].Meta.Count = 7
	s.mu.Lock()
	if err := s.saveLocked(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	s.mu.Unlock()
	raw, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatalf("readfile: %v", err)
	}
	var out []entry
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out) != 2 || out[0].Meta.Count != 7 {
		t.Errorf("round-trip lost data: %+v", out)
	}
}
