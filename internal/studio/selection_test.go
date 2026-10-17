// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
)

func fixture(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

const validSource = "---\ntitle: Example\n---\n## Section\n### Card {#card}\nhello\n"

func TestSelectionRetainsScopeAndRejectsReplacement(t *testing.T) {
	path := fixture(t, "one.md", validSource)
	os.WriteFile(filepath.Join(filepath.Dir(path), "private.md"), []byte("SECRET"), 0600)
	selection, err := OpenSelection(path)
	if err != nil {
		t.Fatal(err)
	}
	defer selection.Close()
	snap := selection.Read()
	if len(snap.Entries) != 1 || snap.Entries[0].Name != "one.md" {
		t.Fatalf("scope: %+v", snap)
	}
	replacement := path + ".new"
	os.WriteFile(replacement, []byte(validSource+"new"), 0600)
	os.Rename(replacement, path)
	if got := selection.Read().Entries[0]; got.Error != "" || string(got.Bytes) != validSource+"new" {
		t.Fatalf("atomic replacement: %+v", got)
	}
	os.Remove(path)
	os.Symlink("private.md", path)
	if got := selection.Read().Entries[0]; got.Error == "" || len(got.Bytes) != 0 {
		t.Fatal("symlink disclosed source")
	}
}
func TestSelectionDirectoryAndBounds(t *testing.T) {
	path := fixture(t, "one.md", validSource)
	dir := filepath.Dir(path)
	s, err := OpenSelection(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	os.Mkdir(filepath.Join(dir, "nested"), 0700)
	os.WriteFile(filepath.Join(dir, "nested", "hidden.md"), []byte(validSource), 0600)
	if len(s.Read().Entries) != 1 {
		t.Fatal("selection recursed")
	}
	f, _ := os.Create(path)
	f.Truncate(8<<20 + 1)
	f.Close()
	if s.Read().Entries[0].Error == "" {
		t.Fatal("oversized file accepted")
	}
	if _, err := OpenSelection(dir + "/nested/../one.md"); err == nil {
		t.Fatal("traversal accepted")
	}
}

func TestSelectionFileAndAggregateCeilings(t *testing.T) {
	path := fixture(t, "one.md", validSource)
	dir := filepath.Dir(path)
	s, e := OpenSelection(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	for i := 0; i < 5; i++ {
		f, _ := os.Create(filepath.Join(dir, fmt.Sprintf("large-%d.md", i)))
		f.Truncate(maxSource)
		f.Close()
	}
	if s.Read().Error == "" {
		t.Fatal("aggregate bytes truncated")
	}
	for i := 0; i < 101; i++ {
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("item-%d.md", i)), nil, 0600)
	}
	if s.Read().Error == "" {
		t.Fatal("file ceiling truncated")
	}
}

func TestAggregateRejectedBeforeReadingAllFiles(t *testing.T) {
	path := fixture(t, "one.md", validSource)
	dir := filepath.Dir(path)
	for i := 0; i < 99; i++ {
		f, _ := os.Create(filepath.Join(dir, fmt.Sprintf("large-%d.md", i)))
		f.Truncate(maxSource)
		f.Close()
	}
	s, e := OpenSelection(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	snapshot := s.Read()
	runtime.ReadMemStats(&after)
	if snapshot.Error == "" {
		t.Fatal("aggregate ceiling absent")
	}
	if after.TotalAlloc-before.TotalAlloc > 160<<20 {
		t.Fatal("allocated across the full oversized selection")
	}
}
func TestInPlaceWritesNeverAcceptMixedSnapshot(t *testing.T) {
	a, b := bytes.Repeat([]byte{'a'}, maxSource), bytes.Repeat([]byte{'b'}, maxSource)
	path := fixture(t, "one.md", string(a))
	s, e := OpenSelection(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	f, _ := os.OpenFile(path, os.O_WRONLY, 0600)
	defer f.Close()
	stop := make(chan struct{})
	done := make(chan struct{})
	// gen counts started and finished writes. A writer that is descheduled in the middle of
	// a write leaves a file that is torn at rest; no reader can tell that from a stable file,
	// so only reads during which the writer made progress (gen changed) can prove a bug.
	var gen atomic.Int64
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
				gen.Add(1)
				f.WriteAt(a, 0)
				gen.Add(1)
				gen.Add(1)
				f.WriteAt(b, 0)
				gen.Add(1)
			}
		}
	}()
	defer func() { close(stop); <-done }()
	for i := 0; i < 20; i++ {
		before := gen.Load()
		entry := s.Read().Entries[0]
		if gen.Load() == before {
			continue // writer idle or stalled mid-write: a torn file at rest is indistinguishable
		}
		if entry.Error == "" && !bytes.Equal(entry.Bytes, a) && !bytes.Equal(entry.Bytes, b) {
			t.Fatal("accepted torn in-place read while the writer was making progress")
		}
	}
}

func TestDefaultSelectionAndRetainedRootReplacement(t *testing.T) {
	base := t.TempDir()
	base, _ = filepath.EvalSymlinks(base)
	dir := filepath.Join(base, "selected")
	os.Mkdir(dir, 0700)
	os.WriteFile(filepath.Join(dir, "one.md"), []byte(validSource), 0600)
	t.Chdir(dir)
	s, e := OpenSelection("")
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	os.Chdir(base)
	original := filepath.Join(base, "original")
	os.Rename(dir, original)
	outside := filepath.Join(base, "outside")
	os.Mkdir(outside, 0700)
	os.WriteFile(filepath.Join(outside, "one.md"), []byte("UNSELECTED"), 0600)
	os.Symlink(outside, dir)
	snap := s.Read()
	if len(snap.Entries) != 1 || string(snap.Entries[0].Bytes) != validSource {
		t.Fatal("retained root followed replaced directory path")
	}
}
