// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/markdown"
	"github.com/wstein/sheetmill/internal/model"
)

func sheet(t *testing.T, id, body string) Sheet {
	t.Helper()
	d, ds := markdown.Parse([]byte("---\ntitle: Test & Reference\ntags: [sample]\nsource: {license: EXAMPLE, attribution: Example contributors}\n---\n" + body))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	return Sheet{ID: id, SourceName: id + ".md", Document: d}
}
func body(t *testing.T, p Project, id string) string {
	t.Helper()
	var data struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(p.Files["data/sheets/"+id+".json"], &data); err != nil {
		t.Fatal(err)
	}
	return data.Body
}
func TestSemanticRendering(t *testing.T) {
	s := sheet(t, "test", "Intro **bold** &amp; \\*literal\\*.\n\n## Editing {#edit}\n### Select {.col-span-2 .note #select}\n\n| Key | Meaning |\n| :-- | --: |\n| `x` | ~~old~~ |\n\n- [x] Done\n- [ ] Next\n\n> Quoted *word*.\n\n```sh\nprintf '%s\\n' '<script>'\n```\n\n[Selection](#select) <a@example.com> [Other](other.md#target)\n")
	other := sheet(t, "other", "## S\n### Target {#target}\nOther content.\n")
	p, ds := Render([]Sheet{s, other}, Options{Title: "Library <safe>"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	b := body(t, p, "test")
	for _, want := range []string{"<strong>bold</strong>", "&amp; *literal*", `id="ref-select"`, `class="card col-span-2 row-span-1 note"`, "<table>", `class="align-right"`, "<del>old</del>", `type="checkbox" disabled checked`, `<blockquote>`, `<em>word</em>`, `&lt;script&gt;`, `href="#ref-select"`, `href="mailto:a@example.com"`, `href="/other/#ref-target"`, "EXAMPLE", "Example contributors"} {
		if !strings.Contains(b, want) {
			t.Errorf("missing %q in %s", want, b)
		}
	}
	for _, f := range []string{"hugo.json", "layouts/baseof.html", "layouts/home.html", "layouts/page.html", "static/assets/site.css", "static/assets/app.js", "static/search.json", "static/brand/favicon.svg", "static/brand/fonts/IBMPlexSans-Regular.woff2", "content/test/index.md", "data/ir/test.json"} {
		if len(p.Files[f]) == 0 {
			t.Errorf("missing %s", f)
		}
	}
}
func TestInertAndUnsafe(t *testing.T) {
	for _, tt := range []struct{ name, src, code string }{
		{"script", "## S\n### C\n<script>touch NEVER_EXECUTED</script>\n", ""},
		{"javascript", "## S\n### C\n[x](javascript:alert%281%29)\n", "REN002"},
		{"data", "## S\n### C\n[x](data:text/html,test)\n", "REN002"},
		{"protocol relative", "## S\n### C\n[x](//example.com)\n", "REN002"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, ds := Render([]Sheet{sheet(t, "test", tt.src)}, Options{})
			if tt.code != "" {
				if len(ds) == 0 || ds[0].Code != tt.code {
					t.Fatal(ds)
				}
				if len(p.Files) > 0 {
					t.Fatal("partial project returned")
				}
				return
			}
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			b := body(t, p, "test")
			if strings.Contains(b, "<script>") || !strings.Contains(b, "&lt;script&gt;") || len(p.Warnings) == 0 {
				t.Fatal("raw HTML not inert/reported", b)
			}
		})
	}
}
func TestUnknownIRAndDuplicateRoutes(t *testing.T) {
	s := sheet(t, "test", "Intro.")
	s.Document.Preamble = append(s.Document.Preamble, model.Node{Kind: "FutureUnknown"})
	if _, ds := Render([]Sheet{s}, Options{}); len(ds) == 0 || ds[0].Code != "REN001" {
		t.Fatal(ds)
	}
	s = sheet(t, "test", "Intro.")
	if _, ds := Render([]Sheet{s, s}, Options{}); len(ds) == 0 || ds[0].Code != "REN003" {
		t.Fatal(ds)
	}
}
func TestProjectDeterminism(t *testing.T) {
	a := sheet(t, "alpha", "## S\n### A\nContent.\n")
	b := sheet(t, "beta", "## S\n### B\nOther.\n")
	one, ds := Render([]Sheet{a, b}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	two, ds := Render([]Sheet{b, a}, Options{})
	if len(ds) > 0 || !reflect.DeepEqual(one.Files, two.Files) {
		t.Fatal("input order changed output")
	}
	for name, data := range one.Files {
		if bytes.Contains(data, []byte("/Users/")) || bytes.Contains(data, []byte("/private/tmp/")) {
			t.Fatal("host path in output", name)
		}
	}
}
func TestHugoIntegration(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("set SHEETMILL_HUGO to explicitly authorize installed tool")
	}
	v, err := exec.Command(binary, "version").CombinedOutput()
	if err != nil || !strings.Contains(string(v), "v"+Version+"+") {
		t.Fatalf("need Hugo %s: %s %v", Version, v, err)
	}
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### Literal {#literal}\n```sh\necho '<tag>'\n```\n\n{{< unsafe >}}\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var snapshots []map[string][]byte
	for range 2 {
		root := t.TempDir()
		for path, b := range p.Files {
			full := filepath.Join(root, path)
			if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, b, 0600); err != nil {
				t.Fatal(err)
			}
		}
		out := filepath.Join(root, "public")
		cmd := exec.Command(binary, "--source", root, "--destination", out, "--clock", "2000-01-01T00:00:00Z", "--noBuildLock", "--quiet")
		cmd.Env = append(os.Environ(), "HUGO_CACHEDIR="+filepath.Join(root, "cache"))
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("hugo: %v %s", err, b)
		}
		snapshot := map[string][]byte{}
		err := filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(out, path)
			b, e := os.ReadFile(path)
			snapshot[rel] = b
			return e
		})
		if err != nil {
			t.Fatal(err)
		}
		html := string(snapshot[filepath.Join("test", "index.html")])
		if !strings.Contains(html, `id="ref-literal"`) || !strings.Contains(html, "&lt;tag&gt;") || !strings.Contains(html, "{{&lt; unsafe &gt;}}") {
			t.Fatalf("missing built semantics: %s", html)
		}
		snapshots = append(snapshots, snapshot)
	}
	if !reflect.DeepEqual(snapshots[0], snapshots[1]) {
		t.Fatal("Hugo output differs across equivalent builds")
	}
}

func TestStructuralHeadingPresentation(t *testing.T) {
	s := sheet(t, "test", "## **Editing** {#edit}\n### Use `code` {.col-span-2 #code}\nText.\n### Literal \\{#literal}\nMore.\n")
	p, ds := Render([]Sheet{s}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	b := body(t, p, "test")
	for _, want := range []string{"<strong>Editing</strong>", "Use <code>code</code>", "Literal {#literal}"} {
		if !strings.Contains(b, want) {
			t.Error(want, b)
		}
	}
	if strings.Contains(b, "{#edit}") || strings.Contains(b, "{.col-span-2 #code}") {
		t.Fatal("rendered parsed attributes", b)
	}
}

func TestLinkedCardHeadingAndDotRoute(t *testing.T) {
	a := sheet(t, "a", "## S\n### [Guide](https://example.com) {#guide}\n[Other](./b.md#target)\n")
	b := sheet(t, "b", "## S\n### Target {#target}\nText.\n")
	p, ds := Render([]Sheet{a, b}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "a")
	if strings.Contains(got, `<h3><a href="#ref-guide"><a`) || !strings.Contains(got, `href="/b/#ref-target"`) {
		t.Fatal(got)
	}
	if !strings.Contains(got, `class="permalink"`) {
		t.Fatal("missing separate permalink", got)
	}
}
func TestLargeLayoutHints(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### Four {.col-span-4 .row-span-4}\nText.\n### Twelve {.col-span-12 .row-span-12}\nText.\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	css := string(p.Files["static/assets/site.css"])
	for _, selector := range []string{".col-span-4", ".col-span-12", ".row-span-4", ".row-span-12"} {
		if !strings.Contains(css, selector) {
			t.Error("missing responsive hint", selector)
		}
	}
}

func TestLinkedImage(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### C\n[![Alt](https://img.example/a.png)](https://example.com)\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	b := body(t, p, "test")
	if !strings.Contains(b, `href="https://example.com"`) || strings.Contains(b, `href="https://img.example/a.png"`) || !strings.Contains(b, "Alt (image: https://img.example/a.png)") {
		t.Fatal(b)
	}
	if len(p.Warnings) == 0 {
		t.Fatal("missing image warning")
	}
}

func TestURLSpacesAndUnsafeSchemes(t *testing.T) {
	a := sheet(t, "a", "## S\n### C\n[Space](<https://example.com/a b?q=a b#c d>) [Selected](<./Sheet One.md#target>)\n")
	b := sheet(t, "sheet-one", "## S\n### Target {#target}\nText.\n")
	b.SourceName = "Sheet One.md"
	p, ds := Render([]Sheet{a, b}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "a")
	for _, want := range []string{`href="https://example.com/a%20b?q=a%20b#c%20d"`, `href="/sheet-one/#ref-target"`} {
		if !strings.Contains(got, want) {
			t.Error(want, got)
		}
	}
	for _, destination := range []string{"javascript&colon;alert(1)", "file:///private/secret", "https://exam ple.com", "mailto:a b@example.com"} {
		s := sheet(t, "test", "## S\n### C\n[x](<"+destination+">)\n")
		if _, ds := Render([]Sheet{s}, Options{}); len(ds) == 0 || ds[0].Code != "REN002" {
			t.Fatal(destination, ds)
		}
	}
}
