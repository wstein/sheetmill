// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHugoBuildBoundaries(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "Sheet One.md")
	destination := filepath.Join(dir, "project")
	os.WriteFile(input, []byte("---\ntitle: Sheet One\n---\n## S\n### C\nText.\n"), 0600)
	run := func(args ...string) (int, string) {
		var out, errout bytes.Buffer
		code := Run(args, &out, &errout)
		return code, errout.String()
	}
	if code, e := run("build", "hugo", input, "--out", destination); code != 0 {
		t.Fatal(code, e)
	}
	before, err := os.ReadFile(filepath.Join(destination, "hugo.json"))
	if err != nil {
		t.Fatal(err)
	}
	if code, e := run("build", "hugo", input, "--out", destination); code != 2 || !strings.Contains(e, "IO003") {
		t.Fatal(code, e)
	}
	after, _ := os.ReadFile(filepath.Join(destination, "hugo.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("overwrote project")
	}
	bad := filepath.Join(dir, "bad.md")
	os.WriteFile(bad, []byte("---\ntitle: 12\n---\n"), 0600)
	absent := filepath.Join(dir, "absent")
	if code, e := run("build", "hugo", input, bad, "--out", absent); code != 1 {
		t.Fatal(code, e)
	}
	if _, err := os.Stat(absent); !os.IsNotExist(err) {
		t.Fatal("invalid input created output", err)
	}
	alias := filepath.Join(dir, "alias")
	os.Symlink(destination, alias)
	if code, e := run("build", "hugo", input, "--out", alias); code != 2 {
		t.Fatal(code, e)
	}
	if code, e := run("build", "hugo", input); code != 2 {
		t.Fatal(code, e)
	}
	second := filepath.Join(dir, "sheet-one.md")
	os.WriteFile(second, []byte("---\ntitle: Other\n---\n"), 0600)
	if code, e := run("build", "hugo", input, second, "--out", absent); code != 1 || !strings.Contains(e, "REN003") {
		t.Fatal(code, e)
	}
}

func TestHugoDiagnosticOrigins(t *testing.T) {
	dir := tempDir(t)
	good := filepath.Join(dir, "good.md")
	bad := filepath.Join(dir, "bad.md")
	os.WriteFile(good, []byte("---\ntitle: Good\n---\n## S\n### C\nGood.\n"), 0600)
	for _, tt := range []struct {
		body, code string
		status     int
	}{{"[Bad](javascript:alert%281%29)", "REN002", 1}, {"<script>inert</script>", "REN101", 0}} {
		os.WriteFile(bad, []byte("---\ntitle: Bad\n---\n## S\n### C\n"+tt.body+"\n"), 0600)
		var out, errout bytes.Buffer
		if code := Run([]string{"build", "hugo", good, bad, "--out", filepath.Join(dir, tt.code)}, &out, &errout); code != tt.status || !strings.Contains(errout.String(), bad+":6:1: "+tt.code) {
			t.Fatal(code, errout.String())
		}
	}
}

func TestHugoAmbiguousBasenameNamesEveryPath(t *testing.T) {
	dir := tempDir(t)
	var paths []string
	for _, sub := range []string{"a", "b"} {
		os.Mkdir(filepath.Join(dir, sub), 0700)
		path := filepath.Join(dir, sub, "x.md")
		os.WriteFile(path, []byte("---\ntitle: X\n---\n"), 0600)
		paths = append(paths, path)
	}
	var out, errout bytes.Buffer
	code := Run([]string{"build", "hugo", paths[0], paths[1], "--out", filepath.Join(dir, "site")}, &out, &errout)
	lines := strings.Split(strings.TrimSpace(errout.String()), "\n")
	if code != 1 || len(lines) != 2 || !strings.HasPrefix(lines[0], paths[0]+":") || !strings.HasPrefix(lines[1], paths[1]+":") {
		t.Fatal(code, errout.String())
	}
}

func TestWriteProjectMidWriteFailure(t *testing.T) {
	dir := tempDir(t)
	sibling := filepath.Join(dir, "keep.md")
	os.WriteFile(sibling, []byte("unchanged\n"), 0600)
	destination := filepath.Join(dir, "project")
	// "a" is written as a file first, so creating directory "a" for "a/b" fails.
	err := writeProject(destination, map[string][]byte{"a": []byte("first"), "a/b": []byte("second"), "z": []byte("never")})
	if err == nil || !strings.Contains(err.Error(), "partial project may remain") {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(sibling); string(b) != "unchanged\n" {
		t.Fatal("pre-existing path modified")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Fatal("wrote outside the new project", entries)
	}
	if b, _ := os.ReadFile(filepath.Join(destination, "a")); string(b) != "first" {
		t.Fatal("expected the partial project to remain for inspection")
	}
	if _, err := os.Stat(filepath.Join(destination, "z")); !os.IsNotExist(err) {
		t.Fatal("writing continued after the failure")
	}
}

func TestHugoBuildReportsMidWriteFailure(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "sheet.md")
	os.WriteFile(input, []byte("---\ntitle: Sheet\n---\n## S\n### C\nText.\n"), 0600)
	destination := filepath.Join(dir, "project")
	defer func(previous func(string, map[string][]byte) error) { writeFiles = previous }(writeFiles)
	writeFiles = func(destination string, files map[string][]byte) error {
		files["a"], files["a/b"] = []byte("x"), []byte("y")
		return writeProject(destination, files)
	}
	var out, errout bytes.Buffer
	code := Run([]string{"build", "hugo", input, "--out", destination}, &out, &errout)
	if code != 2 || !strings.Contains(errout.String(), destination+":1:1: IO003: partial project may remain") || out.Len() != 0 {
		t.Fatal(code, errout.String(), out.String())
	}
	if b, _ := os.ReadFile(input); !strings.HasPrefix(string(b), "---\ntitle: Sheet") {
		t.Fatal("input modified")
	}
}

func TestHugoBuildBasePath(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "sheet.md")
	os.WriteFile(input, []byte("---\ntitle: Sheet\n---\n## S\n### C\nText.\n"), 0600)
	var out, errout bytes.Buffer
	if code := Run([]string{"build", "hugo", input, "--out", filepath.Join(dir, "site"), "--base-path", "/sheetmill/"}, &out, &errout); code != 0 {
		t.Fatal(code, errout.String())
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "site", "hugo.json")); !bytes.Contains(b, []byte(`"base": "/sheetmill/"`)) {
		t.Fatal(string(b))
	}
	errout.Reset()
	if code := Run([]string{"build", "hugo", input, "--out", filepath.Join(dir, "bad"), "--base-path", "sheetmill"}, &out, &errout); code != 1 || !strings.Contains(errout.String(), "REN003: invalid base path") {
		t.Fatal(code, errout.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "bad")); !os.IsNotExist(err) {
		t.Fatal("invalid base path created output")
	}
}

func TestHugoEditBaseFlag(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "bdock.md")
	os.WriteFile(input, []byte("---\ntitle: BundleDock CLI\n---\n## CLI\n### Help\nContent.\n"), 0600)
	destination := filepath.Join(dir, "site")
	var out, errout bytes.Buffer
	if status := Run([]string{"build", "hugo", input, "--out", destination, "--edit-base-url", "https://github.com/wstein/sheetmill/edit/main/examples/sheets/"}, &out, &errout); status != 0 {
		t.Fatal(status, errout.String())
	}
	page, _ := os.ReadFile(filepath.Join(destination, "data/sheets/bdock.json"))
	if !bytes.Contains(page, []byte("https://github.com/wstein/sheetmill/edit/main/examples/sheets/bdock.md")) {
		t.Fatal("CLI omitted authored sheet edit link")
	}
}
