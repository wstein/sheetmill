// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNaviBuild(t *testing.T) {
	dir := tempDir(t)
	good := filepath.Join(dir, "Git Tips.md")
	os.WriteFile(good, []byte("---\ntitle: Git tips\nlicense: EUPL-1.2\ncheatsheet:\n  tool: git\n---\n## Basics\n\n### Status {.navi #status}\n\n```bash\ngit status\n```\n\n### Broken {.navi #broken}\n\n```bash\necho 'oops\n```\n"), 0600)
	run := func(args ...string) (int, string, string) {
		var out, errout bytes.Buffer
		code := Run(args, &out, &errout)
		return code, out.String(), errout.String()
	}
	destination := filepath.Join(dir, "cheats")
	code, out, e := run("build", "navi", good, "--out", destination)
	if code != 0 || !strings.Contains(out, "navi was not run") || !strings.Contains(e, "NAV102") || !strings.Contains(e, "(warning)") || !strings.Contains(e, good+":") {
		t.Fatal(code, out, e)
	}
	b, err := os.ReadFile(filepath.Join(destination, "git-tips.cheat"))
	if err != nil || !strings.Contains(string(b), "# Status\ngit status\n") || strings.Contains(string(b), "oops") {
		t.Fatalf("%v %q", err, b)
	}
	// An existing destination is never overwritten.
	if code, _, e := run("build", "navi", good, "--out", destination); code != 2 || !strings.Contains(e, "IO003") {
		t.Fatal(code, e)
	}
	// Unknown profile is an error and creates nothing.
	absent := filepath.Join(dir, "absent")
	if code, _, e := run("build", "navi", good, "--out", absent, "--shell", "powershell"); code != 1 || !strings.Contains(e, "NAV001") {
		t.Fatal(code, e)
	}
	if _, err := os.Stat(absent); !os.IsNotExist(err) {
		t.Fatal("failed build created output", err)
	}
	// A selection with nothing to export fails instead of writing an empty directory.
	plain := filepath.Join(dir, "plain.md")
	os.WriteFile(plain, []byte("---\ntitle: Plain\nlicense: x\n---\n## S\n### C\nText.\n"), 0600)
	if code, _, e := run("build", "navi", plain, "--out", absent); code != 1 || !strings.Contains(e, "NAV106") {
		t.Fatal(code, e)
	}
	if _, err := os.Stat(absent); !os.IsNotExist(err) {
		t.Fatal("empty export created output", err)
	}
	if code, _, e := run("build", "navi", good); code != 2 {
		t.Fatal(code, e)
	}
}

func TestNaviRootNamesCardsRelatively(t *testing.T) {
	dir := tempDir(t)
	sub := filepath.Join(dir, "sheets")
	os.Mkdir(sub, 0700)
	input := filepath.Join(sub, "a.md")
	os.WriteFile(input, []byte("---\ntitle: A\nlicense: x\n---\n## S\n\n### C {.navi #c}\n\n```bash\necho hi\n```\n"), 0600)
	var out, errout bytes.Buffer
	destination := filepath.Join(dir, "cheats")
	if code := Run([]string{"build", "navi", input, "--root", dir, "--out", destination}, &out, &errout); code != 0 {
		t.Fatal(code, errout.String())
	}
	b, _ := os.ReadFile(filepath.Join(destination, "a.cheat"))
	if !strings.Contains(string(b), "% sheetmill:sheets/a.md#c!\n") || strings.Contains(string(b), dir) {
		t.Fatalf("%q", b)
	}
	other := tempDir(t)
	if code := Run([]string{"build", "navi", input, "--root", other, "--out", filepath.Join(dir, "x")}, &out, &errout); code != 2 || !strings.Contains(errout.String(), "outside --root") {
		t.Fatal(code, errout.String())
	}
}

func TestNaviStrictFailsOnWarnings(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "a.md")
	os.WriteFile(input, []byte("---\ntitle: A\nlicense: x\n---\n## S\n\n### Good {.navi #g}\n\n```bash\necho hi\n```\n\n### Bad {.navi #b}\n\n```bash\necho 'oops\n```\n"), 0600)
	var out, errout bytes.Buffer
	strictOut := filepath.Join(dir, "strict")
	if code := Run([]string{"build", "navi", input, "--strict", "--out", strictOut}, &out, &errout); code != 1 || !strings.Contains(errout.String(), "--strict") {
		t.Fatal(code, errout.String())
	}
	if _, err := os.Stat(strictOut); !os.IsNotExist(err) {
		t.Fatal("strict failure created output", err)
	}
	if code := Run([]string{"build", "navi", input, "--out", filepath.Join(dir, "lenient")}, &out, &errout); code != 0 {
		t.Fatal(code, errout.String())
	}
}

func TestNaviStrictRejectsOmittedParameterMetadata(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "a.md")
	source := "---\ntitle: A\nlicense: x\ncheatsheet:\n  navi:\n    cards:\n      c: {params: {sheet: {type: path, description: Markdown input}}}\n---\n## S\n### C {.navi #c}\n```bash\necho <sheet>\n```\n"
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errout bytes.Buffer
	destination := filepath.Join(dir, "output")
	code := Run([]string{"build", "navi", input, "--strict", "--out", destination}, &out, &errout)
	if code != 1 || !strings.Contains(errout.String(), "NAV107") || !strings.Contains(errout.String(), "NAV000") {
		t.Fatalf("code=%d stderr=%s", code, &errout)
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("failed strict export wrote output: %v", err)
	}
}

func TestNaviRootMustBeDirectory(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "sheet.md")
	if err := os.WriteFile(input, []byte("---\ntitle: T\nlicense: x\n---\n## S\n### C {.navi #c}\n```bash\necho ok\n```\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errout bytes.Buffer
	destination := filepath.Join(dir, "output")
	code := Run([]string{"build", "navi", input, "--root", input, "--out", destination}, &out, &errout)
	if code != 2 || !strings.Contains(errout.String(), "directory") {
		t.Fatalf("code=%d stderr=%s", code, &errout)
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("invalid root wrote output: %v", err)
	}
}
