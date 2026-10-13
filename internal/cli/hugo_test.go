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
