// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"fmt"
	"os"
	"path/filepath"
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
