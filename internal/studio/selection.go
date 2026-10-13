// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxFiles = 100
const maxSource = 8 << 20
const maxSources = 32 << 20

// Selection holds the explicitly granted directory, even if its path is renamed.
type Selection struct {
	root   *os.Root
	single string
	Scope  string
}
type Entry struct {
	Name  string
	Bytes []byte
	Error string
}
type Snapshot struct {
	Entries []Entry
	Error   string
}

func OpenSelection(path string) (*Selection, error) {
	if path == "" {
		path = "."
	}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == ".." {
			return nil, fmt.Errorf("selection traversal is unsupported")
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	current := string(filepath.Separator)
	var info os.FileInfo
	for _, part := range strings.Split(strings.TrimPrefix(abs, string(filepath.Separator)), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err = os.Lstat(current)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symlink selection is unsupported")
		}
	}
	dir, single := abs, ""
	if !info.IsDir() {
		if !info.Mode().IsRegular() || !strings.HasSuffix(abs, ".md") {
			return nil, fmt.Errorf("select a regular .md file or directory")
		}
		dir = filepath.Dir(abs)
		single = filepath.Base(abs)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return &Selection{root: root, single: single, Scope: abs}, nil
}
func (s *Selection) Close() error { return s.root.Close() }
func (s *Selection) Read() Snapshot {
	names := []string{}
	if s.single != "" {
		names = append(names, s.single)
	} else {
		f, e := s.root.Open(".")
		if e != nil {
			return Snapshot{Error: "selected directory is unavailable"}
		}
		entries, e := f.ReadDir(-1)
		f.Close()
		if e != nil {
			return Snapshot{Error: "cannot list selected directory"}
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".md") && !entry.IsDir() {
				names = append(names, entry.Name())
			}
		}
	}
	sort.Strings(names)
	if len(names) > maxFiles {
		return Snapshot{Error: "selection exceeds 100 files"}
	}
	snap := Snapshot{}
	total := 0
	for _, name := range names {
		before, e := s.root.Lstat(name)
		if os.IsNotExist(e) && s.single != "" {
			continue
		}
		entry := Entry{Name: name}
		if e != nil || !before.Mode().IsRegular() {
			entry.Error = "source is unavailable or not a regular file"
		} else if before.Size() > maxSource {
			entry.Error = "source exceeds 8 MiB"
		} else if total+int(before.Size()) > maxSources {
			return Snapshot{Error: "selection exceeds 32 MiB aggregate source bytes"}
		} else {
			f, e := s.root.Open(name)
			if e != nil {
				entry.Error = "cannot open selected source"
			} else {
				opened, e1 := f.Stat()
				b, e2 := readStableBytes(f, min(before.Size(), int64(maxSources-total)))
				_, seekErr := f.Seek(0, io.SeekStart)
				second, secondErr := readStableBytes(f, min(before.Size(), int64(maxSources-total)))
				final, statErr := f.Stat()
				f.Close()
				after, e3 := s.root.Lstat(name)
				if e1 != nil || e2 != nil || e3 != nil || seekErr != nil || secondErr != nil || statErr != nil || !bytes.Equal(b, second) || before.Size() != final.Size() || before.ModTime() != final.ModTime() || opened.Size() != final.Size() || opened.ModTime() != final.ModTime() || int64(len(b)) != final.Size() || !after.Mode().IsRegular() || !os.SameFile(before, opened) || !os.SameFile(opened, after) {
					entry.Error = "source changed during read; waiting for a stable regular file"
				} else if len(b) > maxSource {
					entry.Error = "source exceeds 8 MiB"
				} else {
					entry.Bytes = b
					total += len(b)
				}
			}
		}
		snap.Entries = append(snap.Entries, entry)
	}
	if total > maxSources {
		return Snapshot{Error: "selection exceeds 32 MiB aggregate source bytes"}
	}
	return snap
}
func (s Snapshot) hash() [32]byte { b, _ := json.Marshal(s); return sha256.Sum256(b) }

// The stat-sized buffer prevents ReadAll growth from multiplying the source cap.
func readStableBytes(f *os.File, size int64) ([]byte, error) {
	b := make([]byte, int(size)+1)
	n, err := io.ReadFull(f, b)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		err = nil
	}
	return b[:n], err
}
