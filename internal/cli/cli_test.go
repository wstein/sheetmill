// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	if err := os.WriteFile(p, []byte("---\ntitle: X\n---\n## S\n### C {.tip\n"), 0600); err != nil {
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

func TestRootHelpPointsAgentsToGuides(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"help"}} {
		var out, errout bytes.Buffer
		if code := Run(args, &out, &errout); code != 0 {
			t.Fatal(args, code, errout.String())
		}
		for _, want := range []string{"Are you an AI? Use these resources ONLY IF your task specifically asks you to:", agentGuideURL, llmsURL} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("%v help lacks %q", args, want)
			}
		}
	}
	var out bytes.Buffer
	if Run([]string{"check", "--help"}, &out, &bytes.Buffer{}); strings.Contains(out.String(), "Are you an AI?") {
		t.Fatal("subcommand help repeats the agent notice")
	}
}

// The help URLs name files on main; they must exist in this tree.
func TestAgentResourcesExist(t *testing.T) {
	const prefix = "https://raw.githubusercontent.com/wstein/sheetmill/main/"
	llms, err := os.ReadFile(filepath.Join("..", "..", "llms.txt"))
	if err != nil {
		t.Fatal(err)
	}
	// The Pages workflow copies these two files to the site root.
	published := map[string]string{agentGuideURL: "docs/agent-guide.md", llmsURL: "llms.txt"}
	for u, file := range published {
		if _, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(file))); err != nil {
			t.Errorf("%s: %v", u, err)
		}
	}
	var urls []string
	for _, field := range strings.FieldsFunc(string(llms), func(r rune) bool { return r == '(' || r == ')' || r == ' ' || r == '\n' }) {
		if strings.HasPrefix(field, "https://") {
			urls = append(urls, field)
		}
	}
	for _, u := range urls {
		if !strings.HasPrefix(u, prefix) {
			if strings.HasPrefix(u, "https://github.com/wstein/sheetmill") || strings.HasPrefix(u, "https://wstein.github.io/sheetmill/") {
				continue
			}
			t.Errorf("unexpected external link %s", u)
			continue
		}
		if _, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(strings.TrimPrefix(u, prefix)))); err != nil {
			t.Errorf("%s: %v", u, err)
		}
	}
}

func TestVersionCommandPrintsTheStampedVersion(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })
	Version = "v0.1.0"
	var out, errOut bytes.Buffer
	if code := Run([]string{"version"}, &out, &errOut); code != 0 || out.String() != "smill v0.1.0\n" || errOut.Len() != 0 {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out.String(), errOut.String())
	}
}

func TestLinkerFlagStampsTheVersion(t *testing.T) {
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go tool not available")
	}
	name := "smill"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	build := exec.Command(goTool, "build", "-trimpath", "-o", binary,
		"-ldflags", "-X github.com/wstein/sheetmill/internal/cli.Version=v0.1.0", "../../cmd/smill")
	// The caller's GOFLAGS could add its own -ldflags and change the stamp under test.
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GOFLAGS=") {
			build.Env = append(build.Env, entry)
		}
	}
	build.Env = append(build.Env, "CGO_ENABLED=0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}
	output, err := exec.Command(binary, "version").Output()
	if err != nil || string(output) != "smill v0.1.0\n" {
		t.Fatalf("stamped binary printed %q (%v)", output, err)
	}
}
