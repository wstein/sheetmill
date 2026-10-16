// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

func TestHugoBuildNestedSelection(t *testing.T) {
	dir := tempDir(t)
	os.Mkdir(filepath.Join(dir, "format"), 0700)
	os.WriteFile(filepath.Join(dir, "cli.md"), []byte("---\ntitle: CLI\n---\n## S\n### C\n[Format](format/gfm-plus.md#ids) [Policy](../AGENTS.md)\n"), 0600)
	os.WriteFile(filepath.Join(dir, "format", "gfm-plus.md"), []byte("---\ntitle: Format\n---\n## S\n### IDs {#ids}\n[Back](../cli.md)\n"), 0600)
	site := filepath.Join(dir, "site")
	var out, errout bytes.Buffer
	code := Run([]string{"build", "hugo", dir, filepath.Join(dir, "format"), "--out", site, "--source-base-url", "https://github.com/o/r/blob/main/docs/"}, &out, &errout)
	if code != 0 || errout.Len() != 0 {
		t.Fatal(code, errout.String())
	}
	cli, _ := os.ReadFile(filepath.Join(site, "data", "sheets", "cli.json"))
	format, _ := os.ReadFile(filepath.Join(site, "data", "sheets", "gfm-plus.json"))
	for _, want := range []string{`href=\"/gfm-plus/#ref-ids\"`, `href=\"https://github.com/o/r/blob/main/AGENTS.md\"`} {
		if !bytes.Contains(cli, []byte(want)) {
			t.Error(want, string(cli))
		}
	}
	if !bytes.Contains(format, []byte(`href=\"/cli/\"`)) {
		t.Error(string(format))
	}
	for _, b := range [][]byte{cli, format} {
		if bytes.Contains(b, []byte(dir)) {
			t.Fatal("host path in output")
		}
	}
}

// The published manual must build cleanly from the listed pages.
func TestManualBuildsWithoutWarnings(t *testing.T) {
	list, err := os.ReadFile(filepath.Join("..", "..", "docs", "manual-pages.txt"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"build", "hugo"}
	for _, line := range strings.Split(string(list), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			args = append(args, filepath.Join("..", "..", filepath.FromSlash(line)))
		}
	}
	args = append(args, "--out", filepath.Join(tempDir(t), "manual"), "--base-path", "/sheetmill/", "--source-base-url", "https://github.com/wstein/sheetmill/blob/main/docs/", "--edit-base-url", "https://github.com/wstein/sheetmill/edit/main/docs/")
	var out, errout bytes.Buffer
	if code := Run(args, &out, &errout); code != 0 || errout.Len() != 0 {
		t.Fatal(code, errout.String())
	}
}

func TestHugoBuildDescription(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "sheet.md")
	os.WriteFile(input, []byte("---\ntitle: Sheet\n---\n## S\n### C\nText.\n"), 0600)
	var out, errout bytes.Buffer
	if code := Run([]string{"build", "hugo", input, "--out", filepath.Join(dir, "site"), "--description", "A tidy shelf."}, &out, &errout); code != 0 {
		t.Fatal(code, errout.String())
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "site", "hugo.json")); !bytes.Contains(b, []byte(`"description": "A tidy shelf."`)) {
		t.Fatal(string(b))
	}
}

func TestCheckStrict(t *testing.T) {
	dir := tempDir(t)
	good := filepath.Join(dir, "good.md")
	broken := filepath.Join(dir, "broken.md")
	os.WriteFile(good, []byte("---\ntitle: Good\n---\n## S\n### C {#c}\n[Self](#c) [Policy](../AGENTS.md)\n"), 0600)
	os.WriteFile(broken, []byte("---\ntitle: Broken\n---\n## S\n### C\n[Gone](good.md#missing)\n"), 0600)
	run := func(args ...string) (int, string, string) {
		var out, errout bytes.Buffer
		code := Run(args, &out, &errout)
		return code, out.String(), errout.String()
	}
	if code, out, e := run("check", dir); code != 0 || out != "" || e != "" {
		t.Fatal("plain check must stay structural", code, e)
	}
	code, out, e := run("check", "--strict", dir)
	if code != 1 || out != "" || !strings.Contains(e, broken+":6:1: REN103: link fragment matches no section or card: good.md#missing") || strings.Contains(e, "(warning)") {
		t.Fatal(code, e)
	}
	if !strings.Contains(e, good+":6:") {
		t.Fatal("the repository link must be reported without --source-base-url", e)
	}
	if code, _, e := run("check", "--strict", good, "--source-base-url", "https://github.com/o/r/blob/main/docs/"); code != 0 || e != "" {
		t.Fatal("repository links resolve with --source-base-url", code, e)
	}
	if code, _, e := run("check", good, "--source-base-url", "https://github.com/o/r/blob/main/docs/"); code != 2 || !strings.Contains(e, "--source-base-url requires --strict") {
		t.Fatal(code, e)
	}
	if _, err := os.Stat(filepath.Join(dir, "site")); !os.IsNotExist(err) {
		t.Fatal("strict check must not write a project")
	}
}

// Every diagnostic code the sources can emit must be documented, and every
// documented NAV, CMD or STU code must still be emitted unless it is retired.
func TestDiagnosticsCatalogueIsComplete(t *testing.T) {
	code := regexp.MustCompile(`"((?:SRC|FM|VAL|MD|ATTR|ID|IO|REN|IMP|NAV|CMD|STU)[0-9]{3})"`)
	// NAV000 is printed from a format string rather than a quoted code literal.
	printed := regexp.MustCompile(`smill:1:1: ((?:NAV)[0-9]{3}):`)
	documented := ""
	for _, doc := range []string{"diagnostics.md", "imports.md"} {
		b, err := os.ReadFile(filepath.Join("..", "..", "docs", doc))
		if err != nil {
			t.Fatal(err)
		}
		documented += string(b)
	}
	// A range such as `NAV101`–`NAV106` documents every code between its ends.
	rng := regexp.MustCompile("`([A-Z]+)([0-9]{3})`–`[A-Z]+([0-9]{3})`")
	for _, m := range rng.FindAllStringSubmatch(documented, -1) {
		lo, _ := strconv.Atoi(m[2])
		hi, _ := strconv.Atoi(m[3])
		for n := lo; n <= hi; n++ {
			documented += fmt.Sprintf(" `%s%03d`", m[1], n)
		}
	}
	emitted := map[string]bool{}
	found := 0
	err := filepath.WalkDir(filepath.Join("..", ".."), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		b, err := os.ReadFile(path)
		for _, m := range printed.FindAllStringSubmatch(string(b), -1) {
			emitted[m[1]] = true
		}
		for _, m := range code.FindAllStringSubmatch(string(b), -1) {
			found++
			emitted[m[1]] = true
			if !strings.Contains(documented, "`"+m[1]+"`") {
				t.Errorf("%s emits undocumented %s", path, m[1])
			}
		}
		return err
	})
	if err != nil || found == 0 {
		t.Fatal(err, found)
	}
	retired := map[string]bool{"NAV007": true, "NAV107": true}
	live := regexp.MustCompile("`((?:NAV|CMD|STU)[0-9]{3})`")
	for _, m := range live.FindAllStringSubmatch(documented, -1) {
		if !emitted[m[1]] && !retired[m[1]] {
			t.Errorf("%s is documented but no source emits it; retire it explicitly in this test and the catalogue", m[1])
		}
	}
}

func TestHugoBuildSiteURL(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "sheet.md")
	os.WriteFile(input, []byte("---\ntitle: Sheet\n---\n## S\n### C\nText.\n"), 0600)
	var out, errout bytes.Buffer
	if code := Run([]string{"build", "hugo", input, "--out", filepath.Join(dir, "site"), "--base-path", "/docs/", "--site-url", "https://example.org/docs/"}, &out, &errout); code != 0 {
		t.Fatal(code, errout.String())
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "site", "static", "sitemap.xml")); !bytes.Contains(b, []byte("<loc>https://example.org/docs/sheet/</loc>")) {
		t.Fatal(string(b))
	}
	errout.Reset()
	if code := Run([]string{"build", "hugo", input, "--out", filepath.Join(dir, "bad"), "--site-url", "https://example.org/docs/"}, &out, &errout); code != 1 || !strings.Contains(errout.String(), "REN003: site URL must be https://host/") {
		t.Fatal(code, errout.String())
	}
}

func TestHeaderLinkFlag(t *testing.T) {
	dir := tempDir(t)
	input := filepath.Join(dir, "a.md")
	if err := os.WriteFile(input, []byte("---\ntitle: A\n---\n## S\n### C\nText.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errout bytes.Buffer
	good := filepath.Join(dir, "good")
	if code := Run([]string{"build", "hugo", input, "--out", good, "--header-link", "Example library=/sheetmill/demo/", "--header-link", "Query=https://example.org/?a=b"}, &out, &errout); code != 0 {
		t.Fatalf("%d %s", code, errout.String())
	}
	config, _ := os.ReadFile(filepath.Join(good, "hugo.json"))
	if !bytes.Contains(config, []byte(`"url": "https://example.org/?a=b"`)) || !bytes.Contains(config, []byte(`"label": "Example library"`)) {
		t.Fatalf("%s", config)
	}
	for _, bad := range []string{"no-separator", "=/x/", "Demo=javascript:alert(1)", "Demo=//evil.example/", "Demo=/a\x01b"} {
		errout.Reset()
		target := filepath.Join(dir, "bad")
		if code := Run([]string{"build", "hugo", input, "--out", target, "--header-link", bad}, &out, &errout); code != 1 || !strings.Contains(errout.String(), "REN003: invalid header link") {
			t.Errorf("%q: %d %s", bad, code, errout.String())
		}
		if _, err := os.Stat(target); err == nil {
			t.Errorf("%q created output", bad)
		}
	}
}
