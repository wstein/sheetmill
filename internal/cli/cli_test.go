// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {
	dir := tempDir(t)
	good := filepath.Join(dir, "good.md")
	bad := filepath.Join(dir, "bad.md")
	os.WriteFile(good, []byte("---\ntitle: Test\n---\n## S\n### C\n```sh\ntouch NEVER_EXECUTED\n```\n"), 0600)
	os.WriteFile(bad, []byte("---\ntitle: 12\n---\n"), 0600)
	for _, tt := range []struct {
		args     []string
		code     int
		out, err string
	}{
		{[]string{"version"}, 0, "smill dev", ""}, {[]string{"check", good}, 0, "", ""}, {[]string{"inspect", good}, 0, "\"schema\": 1", ""}, {[]string{"check", bad}, 1, "", "VAL001"}, {[]string{"inspect", bad}, 1, "", "VAL001"}, {[]string{"build"}, 2, "", "usage:"}, {[]string{"check"}, 2, "", "usage:"}, {[]string{"inspect", dir}, 2, "", "IO001"}, {[]string{"check", filepath.Join(dir, "missing.md")}, 2, "", "IO001"},
	} {
		var out, err bytes.Buffer
		code := Run(tt.args, &out, &err)
		if code != tt.code || !strings.Contains(out.String(), tt.out) || !strings.Contains(err.String(), tt.err) {
			t.Fatalf("%v: %d %q %q", tt.args, code, out.String(), err.String())
		}
		if tt.out == "" && out.Len() != 0 {
			t.Fatal("unexpected stdout", out.String())
		}
		if tt.err == "" && err.Len() != 0 {
			t.Fatal("unexpected stderr", err.String())
		}
	}
}
func TestDirectoryBoundaries(t *testing.T) {
	dir := tempDir(t)
	outside := tempDir(t)
	os.WriteFile(filepath.Join(outside, "secret.md"), []byte("---\ntitle: Secret\n---\n"), 0600)
	os.Symlink(filepath.Join(outside, "secret.md"), filepath.Join(dir, "link.md"))
	var out, err bytes.Buffer
	if Run([]string{"check", dir}, &out, &err) != 2 || !strings.Contains(err.String(), "symlink") {
		t.Fatal(out.String(), err.String())
	}
	os.Remove(filepath.Join(dir, "link.md"))
	os.Mkdir(filepath.Join(dir, "nested"), 0700)
	os.WriteFile(filepath.Join(dir, "nested", "invalid.md"), []byte("invalid"), 0600)
	out.Reset()
	err.Reset()
	if Run([]string{"check", dir}, &out, &err) != 2 || !strings.Contains(err.String(), "no .md files") {
		t.Fatal(out.String(), err.String())
	}
}
func TestAncestorSymlink(t *testing.T) {
	dir := tempDir(t)
	target := tempDir(t)
	os.WriteFile(filepath.Join(target, "x.md"), []byte("---\ntitle: X\n---\n"), 0600)
	os.Symlink(target, filepath.Join(dir, "link"))
	var out, err bytes.Buffer
	if Run([]string{"check", filepath.Join(dir, "link", "x.md")}, &out, &err) != 2 {
		t.Fatal("ancestor symlink accepted")
	}
}
func TestWriteFailure(t *testing.T) {
	if Run([]string{"version"}, brokenWriter{}, &bytes.Buffer{}) != 2 {
		t.Fatal("write failure ignored")
	}
}

type brokenWriter struct{}

func (brokenWriter) Write(p []byte) (int, error) { return 0, os.ErrClosed }

func tempDir(t *testing.T) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return p
}
