// SPDX-License-Identifier: EUPL-1.2

package pagesworkflow

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/cli"
)

// demoSpec mirrors the three demo exports of .github/workflows/pages.yml.
type demoSpec struct {
	dir   string
	theme string
	title string
	links []string
}

var demoSpecs = []demoSpec{
	{"demo", "", "SheetMill Demo", []string{"Compact layout=/sheetmill/demo-compact/", "Cards layout=/sheetmill/demo-cards/"}},
	{"demo-compact", "classic-compact", "SheetMill Demo (compact layout)", []string{"Editorial layout=/sheetmill/demo/", "Cards layout=/sheetmill/demo-cards/"}},
	{"demo-cards", "reference-cards", "SheetMill Demo (cards layout)", []string{"Editorial layout=/sheetmill/demo/", "Compact layout=/sheetmill/demo-compact/"}},
}

// smill runs the real CLI in-process, as the workflow's `go run ./cmd/smill` does.
func smill(t *testing.T, args ...string) {
	t.Helper()
	var out, errout bytes.Buffer
	if code := cli.Run(args, &out, &errout); code != 0 {
		t.Fatalf("smill %s: exit %d\n%s%s", strings.Join(args, " "), code, out.String(), errout.String())
	}
}

// hugoRun mirrors the workflow's hugo() helper: same flags, a private cache.
func hugoRun(t *testing.T, binary, cache, source, destination string) {
	t.Helper()
	cmd := exec.Command(binary, "--source", source, "--destination", destination, "--clock", "2000-01-01T00:00:00Z", "--noBuildLock", "--quiet")
	cmd.Env = append(os.Environ(), "HUGO_CACHEDIR="+cache)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("hugo %s: %v\n%s", source, err, b)
	}
}

// buildPages assembles the whole Pages tree below work the way the workflow does:
// the real manual from docs/manual-pages.txt, the three real demos, the guards,
// and llms.txt plus the agent guide.
func buildPages(t *testing.T, binary, work string) string {
	t.Helper()
	repo := repoRoot(t)
	listing, err := os.ReadFile(filepath.Join(repo, "docs", "manual-pages.txt"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"build", "hugo"}
	for _, line := range strings.Split(string(listing), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			args = append(args, filepath.Join(repo, filepath.FromSlash(line)))
		}
	}
	const site = "https://wstein.github.io"
	smill(t, append(args, "--out", filepath.Join(work, "manual"),
		"--header-link", "Example library=/sheetmill/demo/", "--title", "SheetMill Manual",
		"--description", "How to write, check and publish reference sheets with SheetMill.",
		"--base-path", "/sheetmill/", "--site-url", site+"/sheetmill/",
		"--source-base-url", "https://github.com/wstein/sheetmill/blob/main/docs/",
		"--edit-base-url", "https://github.com/wstein/sheetmill/edit/main/docs/")...)
	for _, d := range demoSpecs {
		args := []string{"build", "hugo", filepath.Join(repo, "examples", "sheets"), "--out", filepath.Join(work, d.dir)}
		if d.theme != "" {
			args = append(args, "--theme", d.theme)
		}
		args = append(args, "--title", d.title,
			"--description", "Example reference sheets, independently authored and compiled with SheetMill.",
			"--base-path", "/sheetmill/"+d.dir+"/", "--site-url", site+"/sheetmill/"+d.dir+"/",
			"--edit-base-url", "https://github.com/wstein/sheetmill/edit/main/examples/sheets/")
		for _, l := range d.links {
			args = append(args, "--header-link", l)
		}
		smill(t, args...)
	}
	cache := filepath.Join(work, "hugo-cache")
	public := filepath.Join(work, "public")
	hugoRun(t, binary, cache, filepath.Join(work, "manual"), public)
	for _, d := range demoSpecs {
		if _, err := os.Stat(filepath.Join(public, d.dir)); err == nil {
			t.Fatalf("the manual already owns %s", d.dir)
		}
		hugoRun(t, binary, cache, filepath.Join(work, d.dir), filepath.Join(public, d.dir))
	}
	for src, dst := range map[string]string{"llms.txt": "llms.txt", "docs/agent-guide.md": "agent-guide.md"} {
		data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(src)))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(public, dst), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return public
}

func hugoBinary(t *testing.T) string {
	t.Helper()
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("SKIPPED, NOT PASSED: set SHEETMILL_HUGO to an installed pinned Hugo to build the Pages tree")
	}
	return binary
}

// digest maps every file of the tree to its SHA-256.
func digest(t *testing.T, root string) map[string]string {
	t.Helper()
	sums := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		rel, _ := filepath.Rel(root, path)
		sums[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return sums
}

var (
	scriptTag = regexp.MustCompile(`<script[^>]*>`)
	srcAttr   = regexp.MustCompile(` src="([^"]*)"`)
	styleAttr = regexp.MustCompile(`<[a-zA-Z][^>]* style=`)
	srcTag    = regexp.MustCompile(`<[a-zA-Z][^>]* src="([^"]*)"`)
	linkHref  = regexp.MustCompile(`<link [^>]*href="([^"]*)"`)
)

// checkPage asserts the CSP and request promises of FR-11 / SC-005 for one page.
func checkPage(t *testing.T, name, base string, html []byte) {
	t.Helper()
	scripts := scriptTag.FindAll(html, -1)
	if len(scripts) != 1 {
		t.Errorf("%s: want exactly one <script>, got %d", name, len(scripts))
	} else if m := srcAttr.FindSubmatch(scripts[0]); m == nil {
		t.Errorf("%s: inline script %s", name, scripts[0])
	} else if ok, _ := regexp.MatchString("^"+regexp.QuoteMeta(base)+`assets/app\.[0-9a-f]{64}\.js$`, string(m[1])); !ok {
		t.Errorf("%s: script %q is not %sassets/app.<digest>.js", name, m[1], base)
	}
	if styleAttr.Match(html) {
		t.Errorf("%s: inline style attribute", name)
	}
	var urls [][]byte
	for _, m := range srcTag.FindAllSubmatch(html, -1) {
		urls = append(urls, m[1])
	}
	for _, m := range linkHref.FindAllSubmatch(html, -1) {
		urls = append(urls, m[1])
	}
	for _, u := range urls {
		s := string(u)
		if !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "https://wstein.github.io/") && !strings.HasPrefix(s, "#") {
			t.Errorf("%s: absolute subresource or link %q", name, s)
		}
		if strings.HasPrefix(s, "//") {
			t.Errorf("%s: protocol-relative %q", name, s)
		}
	}
}

func TestPagesTreeWithRealHugo(t *testing.T) {
	binary := hugoBinary(t)
	public := buildPages(t, binary, t.TempDir())

	// The verifier accepts the real manual-plus-demos tree, including the
	// http://127.0.0.1:1313/ prose link of docs/hugo.md (rule 8 regression).
	code, out := runScript(t, public)
	if code != 0 {
		t.Fatalf("verify-pages.sh rejected the real tree (exit %d):\n%s", code, out)
	}
	hugoDoc, err := os.ReadFile(filepath.Join(public, "hugo", "index.html"))
	if err != nil || !bytes.Contains(hugoDoc, []byte(`href="http://127.0.0.1:1313/`)) {
		t.Errorf("the real manual no longer contains the loopback prose link the regression relies on: %v", err)
	}

	pages := 0
	err = filepath.WalkDir(public, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") {
			return err
		}
		rel, _ := filepath.Rel(public, path)
		rel = filepath.ToSlash(rel)
		base := "/sheetmill/"
		for _, d := range demoSpecs {
			if strings.HasPrefix(rel, d.dir+"/") {
				base = "/sheetmill/" + d.dir + "/"
			}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		pages++
		checkPage(t, rel, base, data)
		return nil
	})
	if err != nil || pages == 0 {
		t.Fatalf("pages checked: %d, error %v", pages, err)
	}

	// Each demo links to the other two layouts on its home page and not to itself.
	for _, d := range demoSpecs {
		home, err := os.ReadFile(filepath.Join(public, d.dir, "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range d.links {
			_, url, _ := strings.Cut(l, "=")
			if !bytes.Contains(home, []byte(`href="`+url+`"`)) {
				t.Errorf("%s home lacks a link to %s", d.dir, url)
			}
		}
	}
}

func TestPagesTreeIsDeterministic(t *testing.T) {
	binary := hugoBinary(t)
	first := digest(t, buildPages(t, binary, t.TempDir()))
	second := digest(t, buildPages(t, binary, t.TempDir()))
	if len(first) == 0 || len(first) != len(second) {
		t.Fatalf("file counts differ: %d vs %d", len(first), len(second))
	}
	for name, sum := range first {
		if second[name] != sum {
			t.Errorf("%s differs between two builds", name)
		}
	}
}
