// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/markdown"
	"github.com/wstein/sheetmill/internal/model"
	"github.com/wstein/sheetmill/internal/render/hugo"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const maxGenerated = 64 << 20
const maxLog = 1 << 20

type retained struct {
	Document  *model.Document
	Published uint64
}
type Build struct {
	Files    map[string][]byte
	Sheets   []SheetState
	Retained map[string]retained
	Error    string
}

func routeID(name string) string {
	name = strings.TrimSuffix(name, ".md")
	var b strings.Builder
	separator := false
	for _, c := range strings.ToLower(name) {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			if separator && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(c)
			separator = false
		} else {
			separator = true
		}
	}
	if b.Len() == 0 {
		sum := sha256.Sum256([]byte(name))
		return fmt.Sprintf("sheet-%x", sum[:6])
	}
	return b.String()
}

var hugoVersion = regexp.MustCompile(`\bhugo v0\.167\.0(?:[-+ ]|$)`)

// toolEnv never inherits HUGO_*, HOME, proxy settings or source-defined config.
func toolEnv(root string) []string {
	return []string{"HOME=" + root, "TMPDIR=" + root, "XDG_CONFIG_HOME=" + root, "XDG_CACHE_HOME=" + root, "HUGO_CACHEDIR=" + filepath.Join(root, "cache"), "LANG=C", "TZ=UTC"}
}

type boundedLog struct {
	mu       sync.Mutex
	size     int
	overflow bool
	text     strings.Builder
}

func (b *boundedLog) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.size += len(p)
	if b.size > maxLog {
		b.overflow = true
		return 0, fmt.Errorf("publisher output exceeds 1 MiB")
	}
	b.text.Write(p)
	return len(p), nil
}
func runTool(ctx context.Context, tool string, args []string, root string, version bool, budget int64) error {
	deadline := 30 * time.Second
	if version {
		deadline = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Dir = root
	cmd.Env = toolEnv(root)
	stopEscalation := ownProcess(cmd)
	defer stopEscalation()
	if version {
		output := &boundedLog{}
		cmd.Stdout = output
		cmd.Stderr = output
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cannot run Hugo version check")
		}
		if !hugoVersion.MatchString(output.text.String()) {
			return fmt.Errorf("installed Hugo 0.167.0 is required")
		}
		return nil
	}
	log := &boundedLog{}
	cmd.Stdout = log
	cmd.Stderr = log
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot start installed Hugo")
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("Hugo generation failed or exceeded its deadline")
			}
			if log.overflow {
				return fmt.Errorf("publisher log ceiling exceeded")
			}
			return nil
		case <-ticker.C:
			if _, err := treeSize(root, budget); err != nil {
				cancel()
				<-done
				return fmt.Errorf("publisher output exceeded its budget or became unreadable")
			}
		}
	}
}
func treeSize(root string, limit int64) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("generated output must contain regular files")
		}
		total += info.Size()
		if total > limit {
			return fmt.Errorf("generated data exceeds 64 MiB budget")
		}
		return nil
	})
	return total, err
}
func compile(ctx context.Context, tool, ownerRoot string, seq uint64, snap Snapshot, old map[string]retained, oldBytes int) Build {
	b := Build{Files: map[string][]byte{}, Retained: map[string]retained{}}
	if snap.Error != "" {
		b.Error = snap.Error
		return b
	}
	ids := map[string]int{}
	for _, e := range snap.Entries {
		ids[routeID(e.Name)]++
	}
	var sheets []hugo.Sheet
	for _, e := range snap.Entries {
		id := routeID(e.Name)
		state := SheetState{Name: e.Name, ID: id, Status: "current", Published: seq, Diagnostics: []diagnostic.Diagnostic{}}
		var d *model.Document
		var ds []diagnostic.Diagnostic
		if e.Error != "" {
			ds = []diagnostic.Diagnostic{{Code: "STU001", Message: e.Error}}
		} else {
			d, ds = markdown.Parse(e.Bytes)
		}
		if ids[id] > 1 || id == "studio" || id == "assets" || id == "brand" {
			ds = append(ds, diagnostic.Diagnostic{Code: "REN003", Message: "duplicate or reserved sheet route: " + id})
		}
		if len(ds) > 0 {
			state.Status = "invalid"
			state.Diagnostics = ds
			state.Published = 0
			d = nil
			if previous, ok := old[e.Name]; ok && ids[id] == 1 && id != "studio" && id != "assets" && id != "brand" {
				d = previous.Document
				state.Stale = true
				state.Published = previous.Published
				b.Retained[e.Name] = previous
			}
		}
		if d != nil {
			state.Title, _ = d.Metadata["title"].(string)
			sheets = append(sheets, hugo.Sheet{ID: id, SourceName: e.Name, Document: d})
			if state.Status == "current" {
				b.Retained[e.Name] = retained{d, seq}
			}
		}
		b.Sheets = append(b.Sheets, state)
	}
	// Even an entirely invalid/removed selection gets an authenticated status page.
	if len(sheets) == 0 {
		b.Files["index.html"] = []byte(`<!doctype html><html><head><meta charset="utf-8"><title>SheetMill Studio</title></head><body><main><h1>SheetMill Studio</h1><p>No current preview. Inspect the source diagnostics below.</p></main></body></html>`)
		return b
	}
	project, issues := hugo.Render(sheets, hugo.Options{Title: "SheetMill References"})
	if len(issues) > 0 {
		b.Error = "renderer refused the selected references"
		return b
	}
	for _, warning := range project.Warnings {
		for i := range b.Sheets {
			if b.Sheets[i].Name == warning.Source {
				b.Sheets[i].Diagnostics = append(b.Sheets[i].Diagnostics, warning.Diagnostic)
			}
		}
	}
	work, err := os.MkdirTemp(ownerRoot, "publish-")
	if err != nil {
		b.Error = "cannot prepare owned publisher"
		return b
	}
	defer os.RemoveAll(work)
	budget := maxGenerated - oldBytes
	for _, entry := range b.Retained {
		data, e := entry.Document.JSON()
		if e != nil {
			b.Error = "cannot retain validated IR"
			return b
		}
		budget -= len(data)
	}
	size := 0
	for name, data := range project.Files {
		size += len(data)
		if size > budget {
			b.Error = "generated data exceeds 64 MiB budget"
			return b
		}
		if !filepath.IsLocal(name) {
			b.Error = "invalid renderer output path"
			return b
		}
		path := filepath.Join(work, name)
		if err = os.MkdirAll(filepath.Dir(path), 0700); err == nil {
			err = os.WriteFile(path, data, 0600)
		}
		if err != nil {
			b.Error = "cannot write owned publisher input"
			return b
		}
	}
	public := filepath.Join(work, "public")
	args := []string{"--source", work, "--destination", public, "--clock", "2000-01-01T00:00:00Z", "--noBuildLock", "--quiet"}
	if err = runTool(ctx, tool, args, work, false, int64(budget)); err != nil {
		b.Error = err.Error()
		return b
	}
	disk, err := treeSize(work, int64(budget))
	if err != nil {
		b.Error = "generated data exceeds 64 MiB budget or has an invalid file"
		return b
	}
	err = filepath.WalkDir(public, func(path string, entry fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if entry.IsDir() {
			return nil
		}
		info, e := entry.Info()
		if e != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("nonregular output")
		}
		if disk+info.Size() > int64(budget) {
			return fmt.Errorf("snapshot exceeds 64 MiB budget")
		}
		disk += info.Size()
		f, e := os.Open(path)
		if e != nil {
			return e
		}
		data, e := io.ReadAll(io.LimitReader(f, int64(budget)+1))
		f.Close()
		if e != nil {
			return e
		}
		name, _ := filepath.Rel(public, path)
		b.Files[filepath.ToSlash(name)] = data
		return nil
	})
	if err != nil {
		b.Error = "cannot load generated snapshot within 64 MiB budget"
	}
	return b
}
