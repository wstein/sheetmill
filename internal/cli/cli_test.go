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

func TestReferenceImport(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "source.md")
	output := filepath.Join(dir, "converted.md")
	source := []byte("---\ntitle: Reference\nplugins: [runCode]\n---\n## Section\n### Card\n```sh\ntouch NEVER_EXECUTED\n```\n")
	marker := filepath.Join(dir, "executed")
	source = bytes.ReplaceAll(source, []byte("NEVER_EXECUTED"), []byte(marker))
	if err := os.WriteFile(input, source, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"import", "reference", input, "--repository", "https://github.com/example/reference", "--source-path", "source/source.md", "--revision", strings.Repeat("a", 40), "--license", "GPL-3.0", "--attribution", "Example contributors", "--out", output, "--report", "json"}
	var out, errout bytes.Buffer
	if code := Run(append(append([]string{}, args...), "--dry-run"), &out, &errout); code != 0 || !strings.Contains(out.String(), "IMP101") {
		t.Fatalf("dry run: %d %s %s", code, out.String(), errout.String())
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote output", err)
	}
	out.Reset()
	errout.Reset()
	if code := Run(args, &out, &errout); code != 0 || !strings.Contains(out.String(), "source_sha256") {
		t.Fatalf("import: %d %s %s", code, out.String(), errout.String())
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("source command executed", err)
	}
	original, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errout.Reset()
	if code := Run(args, &out, &errout); code != 2 || !strings.Contains(errout.String(), "IO003") {
		t.Fatalf("overwrite: %d %s %s", code, out.String(), errout.String())
	}
	after, _ := os.ReadFile(output)
	if !bytes.Equal(original, after) {
		t.Fatal("overwrote output")
	}
	out.Reset()
	errout.Reset()
	if code := Run([]string{"check", output}, &out, &errout); code != 0 {
		t.Fatal("converted source invalid", errout.String())
	}
	os.Remove(output)
	os.Symlink(input, output)
	out.Reset()
	errout.Reset()
	if Run(args, &out, &errout) != 2 {
		t.Fatal("output symlink accepted")
	}
	after, _ = os.ReadFile(input)
	if !bytes.Equal(source, after) {
		t.Fatal("input modified")
	}
}
func TestImportUsage(t *testing.T) {
	for _, args := range [][]string{{"import"}, {"import", "auto", "x.md"}, {"import", "reference", "x.md"}, {"import", "reference", "x.md", "--dry-run", "--report", "xml"}} {
		var out, errout bytes.Buffer
		if Run(args, &out, &errout) != 2 || out.Len() != 0 {
			t.Fatal(args, out.String(), errout.String())
		}
	}
}

func TestImportErrorJSON(t *testing.T) {
	dir := tempDir(t)
	p := filepath.Join(dir, "invalid.md")
	if err := os.WriteFile(p, []byte("---\ntitle: X\n---\n## S\n### C {.unsupported}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errout bytes.Buffer
	args := []string{"import", "reference", p, "--repository", "https://github.com/example/reference", "--source-path", "source/example.md", "--revision", strings.Repeat("a", 40), "--license", "GPL-3.0", "--attribution", "Examples", "--dry-run", "--report", "json"}
	if code := Run(args, &out, &errout); code != 1 || !strings.Contains(out.String(), "\"severity\": \"error\"") || !strings.Contains(out.String(), "ATTR001") || !strings.Contains(errout.String(), "ATTR001") {
		t.Fatalf("%d %s %s", code, out.String(), errout.String())
	}
}

func TestCobraHelpAndFlags(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"check", "--help"}, {"import", "reference", "--help"}} {
		var out, errout bytes.Buffer
		if code := Run(args, &out, &errout); code != 0 || !strings.Contains(out.String(), "Usage:") || errout.Len() != 0 {
			t.Fatalf("%v: %d %s %s", args, code, out.String(), errout.String())
		}
	}
	dir := tempDir(t)
	input := filepath.Join(dir, "reference.md")
	if err := os.WriteFile(input, []byte("---\ntitle: Example\n---\n## Section\n### Card\nText.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"import", "reference", "--dry-run", input},
		{"import", "reference", input, "--dry-run"},
	} {
		args = append(args, "--repository", "https://github.com/example/reference", "--source-path", "source/example.md", "--revision", strings.Repeat("a", 40), "--license", "EXAMPLE", "--attribution", "Example")
		var out, errout bytes.Buffer
		if code := Run(args, &out, &errout); code != 0 {
			t.Fatalf("%v: %d %s", args, code, errout.String())
		}
	}
	for _, args := range [][]string{{"version", "extra"}, {"check", "--unknown"}, {"import", "reference", input, "--unknown"}} {
		if code := Run(args, &bytes.Buffer{}, &bytes.Buffer{}); code != 2 {
			t.Fatalf("%v: %d", args, code)
		}
	}
}

func TestHelpWriteFailure(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"help"}, {"check", "--help"}, {"import", "reference", "--help"}} {
		var errout bytes.Buffer
		if code := Run(args, brokenWriter{}, &errout); code != 2 || !strings.Contains(errout.String(), "IO002") {
			t.Fatalf("%v: %d %s", args, code, errout.String())
		}
	}
}

func TestNilArgsMeansNoArguments(t *testing.T) {
	// Cobra falls back to the host process arguments for nil; embedders must not inherit them.
	previous := os.Args
	os.Args = []string{"host", "version"}
	defer func() { os.Args = previous }()
	var out, errout bytes.Buffer
	if code := Run(nil, &out, &errout); code != 2 || !strings.Contains(errout.String(), "a command is required") {
		t.Fatal(code, errout.String())
	}
}

func TestImportUsageOnParseErrors(t *testing.T) {
	for _, args := range [][]string{{"import"}, {"import", "reference"}, {"import", "reference", "x.md", "--unknown"}, {"import", "reference", "a.md", "b.md"}} {
		var out, errout bytes.Buffer
		if code := Run(args, &out, &errout); code != 2 || !strings.Contains(errout.String(), importUsage) {
			t.Fatal(args, code, errout.String())
		}
	}
}
