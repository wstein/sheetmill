// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

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
	for _, f := range []string{"hugo.json", "layouts/baseof.html", "layouts/home.html", "layouts/page.html", "static/" + assetPath(t, p, "stylesheet"), "static/" + assetPath(t, p, "script"), "static/search.json", "static/brand/favicon.svg", "static/brand/fonts/IBMPlexSans-Regular.woff2", "content/test/index.md", "data/ir/test.json"} {
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
	if err != nil || !pinnedVersion(string(v)) {
		t.Fatalf("need Hugo %s: %s %v", Version, v, err)
	}
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### Literal {#literal}\n```sh\necho '<tag>'\n```\n\n{{< unsafe >}}\n\n[gone](#missing) ![](diagram.png) [](https://example.com) [top](#)\n")}, Options{})
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
		built := []string{`id="ref-literal"`, "&lt;tag&gt;", "{{&lt; unsafe &gt;}}", `href="#missing"`, `>diagram.png</a>`, `>https://example.com</a>`, `href="#"`}
		if missing := slices.DeleteFunc(slices.Clone(built), func(want string) bool { return strings.Contains(html, want) }); len(missing) > 0 {
			t.Fatalf("missing built semantics %q: %s", missing, html)
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
	css := string(p.Files["static/"+assetPath(t, p, "stylesheet")])
	for _, selector := range []string{".row-span-4", ".row-span-12"} {
		if !strings.Contains(css, selector) {
			t.Error("missing responsive hint", selector)
		}
	}
	// Column spans beyond the default three-column grid are clamped, never dropped.
	if got := body(t, p, "test"); strings.Count(got, `class="card col-span-3 `) != 2 {
		t.Error("wide spans must clamp to the section grid", got)
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

func TestImageAltLinksAndTrailingURLSpace(t *testing.T) {
	s := sheet(t, "test", "## S\n### C\n![an [inner](https://inner.example) image](https://img.example/a.png)\n\n[Space](<https://example.com/a >)\n")
	p, ds := Render([]Sheet{s}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "test")
	if strings.Contains(got, `href="https://inner.example"`) || !strings.Contains(got, `href="https://example.com/a%20"`) {
		t.Fatal(got)
	}
}

// Official standard builds and Homebrew extended builds have different suffixes.
func pinnedVersion(output string) bool {
	fields := strings.Fields(output)
	if len(fields) < 2 || fields[0] != "hugo" {
		return false
	}
	version := fields[1]
	return version == "v"+Version || strings.HasPrefix(version, "v"+Version+"+") || strings.HasPrefix(version, "v"+Version+"-")
}
func TestPinnedVersion(t *testing.T) {
	for _, s := range []string{"hugo v0.167.0 linux/amd64", "hugo v0.167.0-abcdef linux/amd64", "hugo v0.167.0+extended darwin/arm64"} {
		if !pinnedVersion(s) {
			t.Fatal(s)
		}
	}
	for _, s := range []string{"hugo v0.167.01 linux/amd64", "hugo v0.168.0 linux/amd64", "different v0.167.0"} {
		if pinnedVersion(s) {
			t.Fatal(s)
		}
	}
}

func TestDuplicateRouteReportsEverySource(t *testing.T) {
	a := sheet(t, "same", "Intro.")
	b := sheet(t, "same", "Intro.")
	a.SourceName, b.SourceName = "Same.md", "same.md"
	_, ds := Render([]Sheet{a, b}, Options{})
	sources := map[string]bool{}
	for _, d := range ds {
		if d.Code == "REN003" {
			sources[d.Source] = true
		}
	}
	if !sources["Same.md"] || !sources["same.md"] {
		t.Fatal(ds)
	}
}

type failingFS struct{}

func (failingFS) Open(string) (fs.File, error) { return nil, fs.ErrPermission }

func TestEmbeddedAssetFailureIsFatal(t *testing.T) {
	for name, set := range map[string]*fs.FS{"theme": &themeFiles, "brand": &brandFiles} {
		t.Run(name, func(t *testing.T) {
			previous := *set
			*set = failingFS{}
			defer func() { *set = previous }()
			p, ds := Render([]Sheet{sheet(t, "test", "Intro.")}, Options{})
			if len(ds) == 0 || ds[0].Code != "REN001" || len(p.Files) > 0 {
				t.Fatal(ds, len(p.Files))
			}
		})
	}
}

func TestSearchIndexKeepsCodeVerbatim(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### C\n`\\*` and \\*x\\* &amp;\n\n```\nC:\\\\Temp a &amp;amp; b\n```\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var index []entry
	if err := json.Unmarshal(p.Files["static/search.json"], &index); err != nil {
		t.Fatal(err)
	}
	text := index[0].Text
	for _, want := range []string{`\*`, `and *x* &`, `C:\\Temp`, `a &amp;amp; b`} {
		if !strings.Contains(text, want) {
			t.Errorf("search text %q lacks %q", text, want)
		}
	}
}

func TestEmptyLinkLabels(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### C\n![](diagram.png) [](https://example.com) [![](https://img.example/a.png)](https://example.com/b) [](<>)\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "test")
	for _, want := range []string{`<a href="diagram.png" title="">diagram.png</a>`, `<a href="https://example.com" title="">https://example.com</a>`, `<span class="image-reference">image: https://img.example/a.png</span>`, `<a href="" title="">link</a>`} {
		if !strings.Contains(got, want) {
			t.Error(want, got)
		}
	}
}

func TestUnresolvedInternalLinks(t *testing.T) {
	other := sheet(t, "other", "## S\n### Target {#target}\nText.\n")
	for _, tt := range []struct {
		name, link, href string
		warn             bool
	}{
		{"local card", "[x](#select)", `href="#ref-select"`, false},
		{"local section", "[x](#s)", `href="#ref-s"`, false},
		{"cross sheet card", "[x](other.md#target)", `href="/other/#ref-target"`, false},
		{"cross sheet page", "[x](other.md)", `href="/other/"`, false},
		{"external", "[x](https://example.com/#usage)", `href="https://example.com/#usage"`, false},
		{"empty fragment", "[x](#)", `href="#"`, false},
		{"page with empty fragment", "[x](other.md#)", `href="/other/#"`, false},
		{"missing local fragment", "[x](#usage)", `href="#usage"`, true},
		{"fragment of other sheet", "[x](#target)", `href="#target"`, true},
		{"missing cross sheet fragment", "[x](other.md#install-notes)", `href="/other/#install-notes"`, true},
		{"unselected relative sheet", "[x](../net/dns.md)", `href="../net/dns.md"`, true},
		{"site path", "[x](/elsewhere/)", `href="/elsewhere/"`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := sheet(t, "test", "## S {#s}\n### Select {#select}\n"+tt.link+"\n")
			p, ds := Render([]Sheet{s, other}, Options{})
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			if got := body(t, p, "test"); !strings.Contains(got, tt.href) {
				t.Fatal(tt.href, got)
			}
			found := false
			for _, w := range p.Warnings {
				if w.Code == "REN103" {
					found = true
					if w.Source != "test.md" || w.Span.Start.Line != 8 {
						t.Fatal("wrong origin", w)
					}
				}
			}
			if found != tt.warn {
				t.Fatal("REN103 reported:", found, p.Warnings)
			}
		})
	}
}

func TestSystemDarkModeWithoutScript(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "Intro.")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	css := string(p.Files["static/"+assetPath(t, p, "stylesheet")])
	explicit := css[strings.Index(css, ":root[data-theme=dark]{"):]
	explicit = explicit[len(":root[data-theme=dark]") : strings.Index(explicit, "}")+1]
	for _, want := range []string{"@media(prefers-color-scheme:dark){:root:not([data-theme=light])" + explicit, ":root:not([data-theme=light]) .logo-light{display:none!important}", ":root:not([data-theme=light]) .logo-dark{display:block!important}"} {
		if !strings.Contains(css, want) {
			t.Error("missing", want)
		}
	}
}

func TestUnknownMetadataReachesIR(t *testing.T) {
	d, ds := markdown.Parse([]byte("---\ntitle: T\nx-review:\n  owner: docs\n  levels: [1, 2]\nfuture_flag: true\n---\n## S\n### C\nText.\n"))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	want, err := d.JSON()
	if err != nil {
		t.Fatal(err)
	}
	p, issues := Render([]Sheet{{ID: "test", SourceName: "test.md", Document: d}}, Options{})
	if len(issues) > 0 {
		t.Fatal(issues)
	}
	got := p.Files["data/ir/test.json"]
	if !bytes.Equal(got, want) {
		t.Fatal("IR bytes differ from the parser's canonical JSON")
	}
	var ir struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.Unmarshal(got, &ir); err != nil {
		t.Fatal(err)
	}
	review, _ := ir.Metadata["x-review"].(map[string]any)
	if review["owner"] != "docs" || len(review["levels"].([]any)) != 2 || ir.Metadata["future_flag"] != true {
		t.Fatal("unknown metadata lost", ir.Metadata)
	}
}

func TestBasePath(t *testing.T) {
	a := sheet(t, "a", "## S\n### C\n[Other](b.md#target) [Top](#ref) [Ext](https://example.com/x)\n### Ref {#ref}\nText.\n")
	b := sheet(t, "b", "## S\n### Target {#target}\nText.\n")
	p, ds := Render([]Sheet{a, b}, Options{BasePath: "/sheetmill/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "a")
	for _, want := range []string{`href="/sheetmill/b/#ref-target"`, `href="#ref-ref"`, `href="https://example.com/x"`} {
		if !strings.Contains(got, want) {
			t.Error(want, got)
		}
	}
	for file, want := range map[string]string{
		"data/library.json":  `"url": "/sheetmill/a/"`,
		"static/search.json": `"url": "/sheetmill/a/#ref-c"`,
		"hugo.json":          `"base": "/sheetmill/"`,
		"content/a/index.md": `"url":"/a/"`,
		"data/sheets/a.json": `"url": "/sheetmill/a/"`,
	} {
		if !strings.Contains(string(p.Files[file]), want) {
			t.Errorf("%s lacks %s", file, want)
		}
	}
	// Default output keeps the origin root.
	root, ds := Render([]Sheet{a, b}, Options{})
	if len(ds) > 0 || !strings.Contains(string(root.Files["hugo.json"]), `"base": "/"`) || !strings.Contains(body(t, root, "a"), `href="/b/#ref-target"`) {
		t.Fatal("default base path changed", ds)
	}
}

func TestThemeHasNoRootRelativeURLs(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "Intro.")}, Options{BasePath: "/sheetmill/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	for name, b := range p.Files {
		if !strings.HasPrefix(name, "layouts/") && !strings.HasPrefix(name, "static/assets/") {
			continue
		}
		for _, bad := range []string{`href="/`, `src="/`, `url('/`, `url("/`, `fetch('/`, `fetch("/`} {
			if bytes.Contains(b, []byte(bad)) {
				t.Errorf("%s contains root-relative %s", name, bad)
			}
		}
	}
}

func TestInvalidBasePath(t *testing.T) {
	for _, base := range []string{"sheetmill", "sheetmill/", "/sheetmill", "//host/", "/a/../", "/./", "/a b/", "/a?x/", "/a#x/", "https://x/", `/a\b/`, "/a//b/", "/%2e%2e/", "/a\x00/"} {
		if _, ds := Render([]Sheet{sheet(t, "test", "Intro.")}, Options{BasePath: base}); len(ds) == 0 || ds[0].Code != "REN003" {
			t.Errorf("%q accepted: %v", base, ds)
		}
	}
	for _, base := range []string{"", "/", "/sheetmill/", "/a/b-c_d.e~/"} {
		if _, ds := Render([]Sheet{sheet(t, "test", "Intro.")}, Options{BasePath: base}); len(ds) > 0 {
			t.Errorf("%q rejected: %v", base, ds)
		}
	}
}

func TestHugoIntegrationBasePath(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("set SHEETMILL_HUGO to explicitly authorize installed tool")
	}
	a := sheet(t, "a", "## S\n### C\n[Other](b.md#target)\n")
	b := sheet(t, "b", "## S\n### Target {#target}\nText.\n")
	p, ds := Render([]Sheet{a, b}, Options{BasePath: "/sheetmill/", EditBaseURL: "https://github.com/wstein/sheetmill/edit/main/examples/sheets/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	root := t.TempDir()
	for path, data := range p.Files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(root, "public")
	cmd := exec.Command(binary, "--source", root, "--destination", out, "--clock", "2000-01-01T00:00:00Z", "--noBuildLock", "--quiet")
	cmd.Env = append(os.Environ(), "HUGO_CACHEDIR="+filepath.Join(root, "cache"))
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("hugo: %v %s", err, b)
	}
	// Files stay at the publish root; only URLs carry the base path.
	for _, name := range []string{"index.html", "a/index.html", "b/index.html", "search.json", assetPath(t, p, "stylesheet"), assetPath(t, p, "script"), "brand/favicon.svg"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"index.html", "a/index.html"} {
		html, _ := os.ReadFile(filepath.Join(out, name))
		for _, want := range []string{`data-base="/sheetmill/"`, `href="/sheetmill/` + assetPath(t, p, "stylesheet") + `"`, `src="/sheetmill/` + assetPath(t, p, "script") + `"`, `href="/sheetmill/a/"`} {
			if !bytes.Contains(html, []byte(want)) {
				t.Errorf("%s lacks %s", name, want)
			}
		}
		if name == "a/index.html" && !bytes.Contains(html, []byte(`href="https://github.com/wstein/sheetmill/edit/main/examples/sheets/a.md"`)) {
			t.Error("built sheet lacks GitHub edit link")
		}
		if name == "a/index.html" && !bytes.Contains(html, []byte(`aria-label="Edit on GitHub — this cheatsheet"`)) {
			t.Error("edit action accessible name must start with its visible label")
		}
		if name == "index.html" && bytes.Contains(html, []byte(`class="edit-link"`)) {
			t.Error("home page guessed a sheet edit destination")
		}
		for _, bad := range []string{`href="/assets`, `src="/brand`, `href="/a/"`, `href="/"`} {
			if bytes.Contains(html, []byte(bad)) {
				t.Errorf("%s contains root-relative %s", name, bad)
			}
		}
	}
	if html, _ := os.ReadFile(filepath.Join(out, "a", "index.html")); !bytes.Contains(html, []byte(`href="/sheetmill/b/#ref-target"`)) {
		t.Error("cross-sheet link lacks base path")
	}
}

func TestGitHubEditLinksAreSeparateFromProvenance(t *testing.T) {
	s := sheet(t, "bdock", "## CLI\n### Help\n`bdock --help`\n")
	s.SourceName = "Bundle Dock.md"
	s.Document.Metadata["source"] = map[string]any{"repository": "https://github.com/wstein/bundledock", "path": "cmd/bdock/tools.go", "revision": "81c2688ea8dcb782f34b51f96c3c2f18bdd00d85"}
	p, ds := Render([]Sheet{s}, Options{EditBaseURL: "https://github.com/wstein/sheetmill/edit/main/examples/sheets/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var page map[string]any
	json.Unmarshal(p.Files["data/sheets/bdock.json"], &page)
	if page["edit_url"] != "https://github.com/wstein/sheetmill/edit/main/examples/sheets/Bundle%20Dock.md" {
		t.Fatal(page["edit_url"])
	}
	if !strings.Contains(body(t, p, "bdock"), "cmd/bdock/tools.go") || !strings.Contains(body(t, p, "bdock"), "wstein/bundledock") {
		t.Fatal("provenance was replaced by edit destination")
	}
	p, ds = Render([]Sheet{s}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	page = map[string]any{}
	json.Unmarshal(p.Files["data/sheets/bdock.json"], &page)
	if _, ok := page["edit_url"]; ok {
		t.Fatal("generic export guessed a GitHub destination")
	}
	for _, base := range []string{"javascript:alert(1)", "http://github.com/o/r/edit/main/", "https://evil.invalid/o/r/edit/main/", "https://user@github.com/o/r/edit/main/", "https://github.com/o/r/blob/main/", "https://github.com/o/r/edit/main/?x=1", "https://github.com/o/r/edit/main/#fragment", "https://github.com/o/r/edit/main/../"} {
		if p, ds := Render([]Sheet{s}, Options{EditBaseURL: base}); len(ds) == 0 || len(p.Files) > 0 {
			t.Fatal("invalid edit base accepted", base)
		}
	}
}

func TestCompactProvenance(t *testing.T) {
	s := sheet(t, "test", "Content.")
	s.Document.Metadata["source"] = map[string]any{"license": "EUPL-1.2", "attribution": "Contributors <safe>", "repository": "https://github.com/example/project", "custom": map[string]any{"retained": true}}
	p, ds := Render([]Sheet{s}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	b := body(t, p, "test")
	for _, want := range []string{`<details class="source-details"><summary><span class="source-facts">Contributors &lt;safe&gt;</span>`, `<span class="source-facts">EUPL-1.2</span>`, `Source details</summary>`, `<dt>Attribution:</dt><dd>Contributors &lt;safe&gt;</dd>`, `<dt>License:</dt><dd>EUPL-1.2</dd>`, `retained: true`} {
		if !strings.Contains(b, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(b, `<details class="source-details" open`) {
		t.Fatal("metadata must start collapsed")
	}
}

// Asset URLs must identify their exact bytes, independently of export location.
func assetPath(t *testing.T, p Project, key string) string {
	t.Helper()
	var config struct {
		Params map[string]string `json:"params"`
	}
	if err := json.Unmarshal(p.Files["hugo.json"], &config); err != nil {
		t.Fatal(err)
	}
	name := config.Params[key]
	if name == "" {
		t.Fatalf("missing fingerprinted %s URL", key)
	}
	data, ok := p.Files["static/"+name]
	if !ok {
		t.Fatalf("missing asset %s", name)
	}
	stem, ext := "site", "css"
	if key == "script" {
		stem, ext = "app", "js"
	}
	want := fmt.Sprintf("assets/%s.%x.%s", stem, sha256.Sum256(data), ext)
	if name != want {
		t.Fatalf("asset digest mismatch: got %s want %s", name, want)
	}
	if _, ok := p.Files["static/assets/"+stem+"."+ext]; ok {
		t.Fatal("unversioned asset alias remains")
	}
	return name
}

func TestAssetFingerprintsTrackContent(t *testing.T) {
	original := themeFiles
	defer func() { themeFiles = original }()
	files := fstest.MapFS{}
	if err := fs.WalkDir(original, "theme", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(original, name)
		files[name] = &fstest.MapFile{Data: data}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	themeFiles = files
	input := []Sheet{sheet(t, "test", "Text.")}
	before, ds := Render(input, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	css, js := assetPath(t, before, "stylesheet"), assetPath(t, before, "script")
	for _, key := range []string{"stylesheet", "script"} {
		file := "theme/static/assets/site.css"
		if key == "script" {
			file = "theme/static/assets/app.js"
		}
		saved := files[file]
		files[file] = &fstest.MapFile{Data: append(append([]byte{}, saved.Data...), []byte("\n/* asset change */\n")...)}
		after, ds := Render(input, Options{BasePath: "/sheetmill/"})
		if len(ds) > 0 {
			t.Fatal(ds)
		}
		newCSS, newJS := assetPath(t, after, "stylesheet"), assetPath(t, after, "script")
		if (newCSS != css) != (key == "stylesheet") || (newJS != js) != (key == "script") {
			t.Fatal("only changed asset should get a new URL")
		}
		files[file] = saved
	}
	delete(files, "theme/static/assets/site.css")
	if p, ds := Render(input, Options{}); len(ds) == 0 || ds[0].Code != "REN001" || len(p.Files) != 0 {
		t.Fatal("missing fingerprint input must fail without a partial project", ds)
	}
}

func TestSourceDetailsPreserveValuesAndEscapeMarkup(t *testing.T) {
	value := map[string]any{"attribution": "<script>unsafe</script>", "license": "EXAMPLE", "note": "Line one\nLine two", "revision": "0000123", "future<key>": []any{"<img>", true, int64(9007199254740993)}, "nil": nil}
	got, err := sourceDetails(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<dt>Attribution:</dt><dd>&lt;script&gt;unsafe&lt;/script&gt;</dd>`, `<dt>future&lt;key&gt;:</dt>`, `&lt;img&gt;`, `9007199254740993`, `<dt>note:</dt><dd>Line one` + "\n" + `Line two</dd>`, `<dt>revision:</dt><dd>0000123</dd>`, `<dt>nil:</dt><dd>null</dd>`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "<script>") || strings.Contains(got, "<pre>") {
		t.Fatal("unsafe or raw JSON presentation", got)
	}
	if strings.Index(got, "<dt>Attribution:") > strings.Index(got, "<dt>License:") {
		t.Fatal("unstable field order")
	}
	for _, root := range []any{nil, []any{"first", true}, "plain"} {
		got, err := sourceDetails(root)
		if err != nil || !strings.Contains(got, "<dt>source:</dt>") {
			t.Fatal("non-map metadata lost", root, err, got)
		}
	}
	if _, err := sourceDetails(map[string]any{"bad": func() {}}); err == nil {
		t.Fatal("unsupported metadata accepted")
	}
}

func TestSourceDetailsEscapeControlCharacters(t *testing.T) {
	got, err := sourceDetails(map[string]any{"note": "a\x00b\rc", "key\x00suffix": "value", "key\nline": "tab\tand\nnewline"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<dt>&#34;key\0suffix&#34;:</dt>`, `<dd>&#34;a\0b\rc&#34;</dd>`, `<dt>&#34;key\nline&#34;:</dt>`, "<dd>tab\tand\nnewline</dd>"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
	if strings.ContainsAny(got, "\x00\r") {
		t.Fatalf("HTML-normalized controls leaked: %q", got)
	}
}

func nested(t *testing.T, id, sourcePath, body string) Sheet {
	t.Helper()
	s := sheet(t, id, body)
	s.SourceName, s.Path = path.Base(sourcePath), sourcePath
	return s
}

func TestNestedSheetLinks(t *testing.T) {
	cli := nested(t, "cli", "cli.md", "## S\n### Usage {#usage}\n[Format](format/gfm-plus.md#ids) [Wrong](gfm-plus.md)\n")
	format := nested(t, "gfm-plus", "format/gfm-plus.md", "## S\n### IDs {#ids}\n[Back](../cli.md#usage) [Same dir](compatibility.md)\n")
	compat := nested(t, "compatibility", "format/compatibility.md", "## S\n### C\nText.\n")
	p, ds := Render([]Sheet{cli, format, compat}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	for id, want := range map[string]string{"cli": `href="/gfm-plus/#ref-ids"`, "gfm-plus": `href="/cli/#ref-usage"`} {
		if got := body(t, p, id); !strings.Contains(got, want) {
			t.Error(id, want, got)
		}
	}
	if got := body(t, p, "gfm-plus"); !strings.Contains(got, `href="/compatibility/"`) {
		t.Error("same-directory link", got)
	}
	var unresolved []string
	for _, w := range p.Warnings {
		if w.Code == "REN103" {
			unresolved = append(unresolved, w.Message)
		}
	}
	if len(unresolved) != 1 || !strings.Contains(unresolved[0], "gfm-plus.md") {
		t.Fatal("only the root-relative basename link must stay unresolved", unresolved)
	}
}

func TestSourceBaseURL(t *testing.T) {
	cli := nested(t, "cli", "cli.md", "## S\n### C\n[Policy](../AGENTS.md#scope) [Notes](e2e-notes.md) [Site](/elsewhere/) [Escape](../../../x.md)\n")
	format := nested(t, "gfm-plus", "format/gfm-plus.md", "## S\n### C\n[Schema](../../schema/cheatsheet.schema.json)\n")
	p, ds := Render([]Sheet{cli, format}, Options{SourceBaseURL: "https://github.com/wstein/sheetmill/blob/main/docs/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	for id, want := range map[string]string{
		"cli":      `href="https://github.com/wstein/sheetmill/blob/main/AGENTS.md#scope"`,
		"gfm-plus": `href="https://github.com/wstein/sheetmill/blob/main/schema/cheatsheet.schema.json"`,
	} {
		if got := body(t, p, id); !strings.Contains(got, want) {
			t.Error(id, want, got)
		}
	}
	if got := body(t, p, "cli"); !strings.Contains(got, `href="https://github.com/wstein/sheetmill/blob/main/docs/e2e-notes.md"`) {
		t.Error("same-directory repository link", got)
	}
	// Site-absolute paths and links leaving the repository stay as written and are reported.
	var unresolved []string
	for _, w := range p.Warnings {
		if w.Code == "REN103" {
			unresolved = append(unresolved, w.Message)
		}
	}
	if len(unresolved) != 2 || !strings.Contains(unresolved[0]+unresolved[1], "/elsewhere/") || !strings.Contains(unresolved[0]+unresolved[1], "../../../x.md") {
		t.Fatal(unresolved)
	}
	for _, bad := range []string{"http://github.com/o/r/blob/main/", "https://example.com/o/r/blob/main/", "https://github.com/o/r/edit/main/", "https://github.com/o/r/blob/main/docs/?q", "https://github.com/o/r/blob/main/../x/"} {
		if _, ds := Render([]Sheet{cli}, Options{SourceBaseURL: bad}); len(ds) == 0 || ds[0].Code != "REN003" {
			t.Error(bad, ds)
		}
	}
}

func TestNestedEditLinkAndInvalidPath(t *testing.T) {
	format := nested(t, "gfm-plus", "format/gfm-plus.md", "## S\n### C\nText.\n")
	p, ds := Render([]Sheet{format}, Options{EditBaseURL: "https://github.com/wstein/sheetmill/edit/main/docs/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if !strings.Contains(string(p.Files["data/sheets/gfm-plus.json"]), `https://github.com/wstein/sheetmill/edit/main/docs/format/gfm-plus.md`) {
		t.Fatal(string(p.Files["data/sheets/gfm-plus.json"]))
	}
	for _, bad := range []string{"/abs/gfm-plus.md", "../gfm-plus.md", "format/../gfm-plus.md", "format/other.md", `format\gfm-plus.md`, "format//gfm-plus.md"} {
		s := nested(t, "gfm-plus", "format/gfm-plus.md", "Intro.")
		s.Path = bad
		if _, ds := Render([]Sheet{s}, Options{}); len(ds) == 0 || ds[0].Code != "REN003" {
			t.Error(bad, ds)
		}
	}
}

func TestSectionColumnsClampCardSpans(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## Two {.cols-2}\n### Wide {.col-span-3}\nA.\n## Default\n### Huge {.col-span-12}\nB.\n## Six {.cols-6}\n### Five {.col-span-5}\nC.\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "test")
	for _, want := range []string{`<div class="cards cols-2"><article class="card col-span-2 `, `<div class="cards cols-3"><article class="card col-span-3 `, `<div class="cards cols-6"><article class="card col-span-5 `} {
		if !strings.Contains(got, want) {
			t.Error(want, got)
		}
	}
	css := string(p.Files["static/"+assetPath(t, p, "stylesheet")])
	for n := 1; n <= 6; n++ {
		for _, want := range []string{fmt.Sprintf(".cards.cols-%d{", n), fmt.Sprintf(".col-span-%d{grid-column:span %d}", n, n)} {
			if n == 1 && strings.HasPrefix(want, ".col-span") {
				continue
			}
			if !strings.Contains(css, want) {
				t.Error("missing", want)
			}
		}
	}
}

func described(t *testing.T, id, frontmatter, body string) Sheet {
	t.Helper()
	d, ds := markdown.Parse([]byte("---\n" + frontmatter + "---\n" + body))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	return Sheet{ID: id, SourceName: id + ".md", Document: d}
}

func TestSheetMetadataAndShelves(t *testing.T) {
	sheets := []Sheet{
		described(t, "git", "title: Git\ndescription: Track changes.\ncategory: version-control\ntags: [vcs, cli]\n", "## S\n### C\nText.\n"),
		described(t, "tar", "title: tar\ncategory: files\n", "## S\n### C\nText.\n"),
		described(t, "misc", "title: Misc\n", "Intro.\n"),
		described(t, "find", "title: find\ncategory: files\n", "Intro.\n"),
	}
	p, ds := Render(sheets, Options{Title: "Library", Description: "Ten tools, well arranged."})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	git := string(p.Files["data/sheets/git.json"])
	for _, want := range []string{`"description": "Track changes."`, `"category": "version-control"`, `"vcs"`} {
		if !strings.Contains(git, want) {
			t.Error(want, git)
		}
	}
	var shelves []struct {
		Category string `json:"category"`
		Sheets   []struct {
			Title string `json:"title"`
		} `json:"sheets"`
	}
	if err := json.Unmarshal(p.Files["data/shelves.json"], &shelves); err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, s := range shelves {
		var titles []string
		for _, sheet := range s.Sheets {
			titles = append(titles, sheet.Title)
		}
		order = append(order, s.Category+"="+strings.Join(titles, ","))
	}
	// Categories sort by name; sheets without a category come last.
	if got := strings.Join(order, ";"); got != "files=find,tar;version-control=Git;=Misc" {
		t.Fatal(got)
	}
	if !strings.Contains(string(p.Files["hugo.json"]), `"description": "Ten tools, well arranged."`) {
		t.Fatal(string(p.Files["hugo.json"]))
	}
}

func TestHugoIntegrationMetadata(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("set SHEETMILL_HUGO to explicitly authorize installed tool")
	}
	p, ds := Render([]Sheet{
		described(t, "git", "title: Git\ndescription: Track <changes>.\ncategory: version-control\ntags: [vcs, cli]\n", "## S\n### C\nText.\n"),
		described(t, "misc", "title: Misc\n", "Intro.\n"),
	}, Options{Title: "My Library", Description: "Ten tools, well arranged."})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	out := hugoBuild(t, binary, p)
	tagIndex, err := os.ReadFile(filepath.Join(out, "tags", "index.html"))
	if err != nil || !bytes.Contains(tagIndex, []byte("<h1>Tags</h1>")) {
		t.Fatalf("tag index: %v %s", err, tagIndex)
	}
	home, _ := os.ReadFile(filepath.Join(out, "index.html"))
	page, _ := os.ReadFile(filepath.Join(out, "git", "index.html"))
	for _, want := range []string{"<h1>My Library</h1>", "Ten tools, well arranged.", `<h2 class="shelf-title">version-control</h2>`, `<h2 class="shelf-title">Other references</h2>`, "Track &lt;changes&gt;."} {
		if !bytes.Contains(home, []byte(want)) {
			t.Errorf("home lacks %s", want)
		}
	}
	for _, want := range []string{`<p class="sheet-description">Track &lt;changes&gt;.</p>`, `<li><a href="/tags/vcs-a8d1cfa1fc63/">vcs</a></li>`, `version-control`} {
		if !bytes.Contains(page, []byte(want)) {
			t.Errorf("page lacks %s", want)
		}
	}
	if misc, _ := os.ReadFile(filepath.Join(out, "misc", "index.html")); bytes.Contains(misc, []byte("sheet-description")) || bytes.Contains(misc, []byte("Keep the essentials")) {
		t.Error("a sheet without description must not show a placeholder tagline")
	}
}

// hugoBuild writes p to a temporary source tree and builds it with the pinned binary.
func hugoBuild(t *testing.T, binary string, p Project) string {
	t.Helper()
	root := t.TempDir()
	for name, data := range p.Files {
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(root, "public")
	cmd := exec.Command(binary, "--source", root, "--destination", out, "--clock", "2000-01-01T00:00:00Z", "--noBuildLock", "--quiet")
	cmd.Env = append(os.Environ(), "HUGO_CACHEDIR="+filepath.Join(root, "cache"))
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("hugo: %v %s", err, b)
	}
	return out
}

func TestCardKindLabels(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### A {.tip}\nA.\n### B {.warning}\nB.\n### C {.note}\nC.\n### D\nD.\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "test")
	for _, want := range []string{`<p class="card-kind">Tip</p>`, `<p class="card-kind">Warning</p>`, `<p class="card-kind">Note</p>`} {
		if !strings.Contains(got, want) {
			t.Error(want, got)
		}
	}
	if strings.Count(got, "card-kind") != 3 {
		t.Error("plain cards must not get a label", got)
	}
	css := string(p.Files["static/"+assetPath(t, p, "stylesheet")])
	for _, want := range []string{".card.tip{", ".card.warning{", ".card.note{", ".card-kind{"} {
		if !strings.Contains(css, want) {
			t.Error("missing", want)
		}
	}
}

func TestSearchEntriesCarryTagsAndCategory(t *testing.T) {
	p, ds := Render([]Sheet{described(t, "git", "title: Git\ncategory: version-control\ntags: [vcs]\n", "## S\n### Undo\nText.\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var index []struct {
		Title    string   `json:"title"`
		Tags     []string `json:"tags"`
		Category string   `json:"category"`
	}
	if err := json.Unmarshal(p.Files["static/search.json"], &index); err != nil {
		t.Fatal(err)
	}
	for _, e := range index {
		if e.Category != "version-control" || len(e.Tags) != 1 || e.Tags[0] != "vcs" {
			t.Fatal(e)
		}
	}
}

func TestLicenseInSearchAndLibrary(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### C\nText.\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if !strings.Contains(string(p.Files["static/search.json"]), `"license": "EXAMPLE"`) {
		t.Fatal(string(p.Files["static/search.json"]))
	}
	if !strings.Contains(string(p.Files["data/library.json"]), `"license": "EXAMPLE"`) {
		t.Fatal(string(p.Files["data/library.json"]))
	}
	if !strings.Contains(string(p.Files["layouts/home.html"]), "library-license") {
		t.Fatal("home cards must show the license")
	}
}

func TestKeyTables(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "test", "## S\n### Keys {.keys}\n| Key | Action |\n| --- | --- |\n| `Ctrl+C` | Copy `text` |\n| `+` | Zoom |\n| `<Esc>` | Leave |\n\nUse `Ctrl+V` too.\n### Plain\n| Key | Action |\n| --- | --- |\n| `Ctrl+C` | Copy |\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "test")
	for _, want := range []string{
		`<td><kbd><kbd>Ctrl</kbd>+<kbd>C</kbd></kbd></td>`,
		`<td>Copy <code>text</code></td>`,
		`<td><kbd>+</kbd></td>`,
		`<td><kbd>&lt;Esc&gt;</kbd></td>`,
		`Use <code>Ctrl+V</code> too.`,
		`<th>Key</th>`,
	} {
		if !strings.Contains(got, want) {
			t.Error(want, got)
		}
	}
	if strings.Count(got, "<kbd>") != 5 {
		t.Error("only first-column code in .keys cards becomes keys", got)
	}
}

func TestSectionJumpList(t *testing.T) {
	p, ds := Render([]Sheet{
		sheet(t, "many", "Intro.\n\n## First *steps* {#first}\n### C\nA.\n## Second\n### D\nB.\n"),
		sheet(t, "one", "## Only\n### C\nA.\n"),
	}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := body(t, p, "many")
	want := `<nav class="sheet-toc" aria-label="On this page"><ul><li><a href="#ref-first">First steps</a></li><li><a href="#ref-second">Second</a></li></ul></nav>`
	if !strings.Contains(got, want) {
		t.Fatal(want, got)
	}
	if strings.Index(got, want) > strings.Index(got, `<section`) || strings.Index(got, "Intro.") > strings.Index(got, want) {
		t.Fatal("the jump list belongs between the preamble and the first section", got)
	}
	if strings.Contains(body(t, p, "one"), "sheet-toc") {
		t.Fatal("a single section needs no jump list")
	}
}

// Performance budget: the theme ships at most 32 KiB of CSS and JavaScript, and
// the search index stays within twice the size of the sources it indexes.
func TestPerformanceBudget(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "..", "examples", "sheets", "*.md"))
	if err != nil || len(files) == 0 {
		t.Fatal(err, files)
	}
	var sheets []Sheet
	sources := 0
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		sources += len(b)
		d, ds := markdown.Parse(b)
		if len(ds) > 0 {
			t.Fatal(f, ds)
		}
		name := filepath.Base(f)
		sheets = append(sheets, Sheet{ID: strings.TrimSuffix(name, ".md"), SourceName: name, Document: d})
	}
	p, ds := Render(sheets, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	assets := len(p.Files["static/"+assetPath(t, p, "stylesheet")]) + len(p.Files["static/"+assetPath(t, p, "script")])
	if assets > 32<<10 {
		t.Errorf("CSS and JavaScript use %d bytes; budget is 32 KiB", assets)
	}
	if index := len(p.Files["static/search.json"]); index > 2*sources {
		t.Errorf("search index is %d bytes for %d source bytes; budget is 2x", index, sources)
	}
}

func TestSiteURLSitemapAndValidation(t *testing.T) {
	sheets := []Sheet{sheet(t, "a&b", "Intro."), sheet(t, "c", "Intro.")}
	sheets[0].ID = "ab"
	p, ds := Render(sheets, Options{BasePath: "/sheetmill/", SiteURL: "https://wstein.github.io/sheetmill/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	want := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n<url><loc>https://wstein.github.io/sheetmill/</loc></url>\n<url><loc>https://wstein.github.io/sheetmill/ab/</loc></url>\n<url><loc>https://wstein.github.io/sheetmill/c/</loc></url>\n<url><loc>https://wstein.github.io/sheetmill/tags/</loc></url>\n<url><loc>https://wstein.github.io/sheetmill/tags/sample-af2bdbe1aa9b/</loc></url>\n</urlset>\n"
	if got := string(p.Files["static/sitemap.xml"]); got != want {
		t.Fatalf("%q", got)
	}
	if !strings.Contains(string(p.Files["hugo.json"]), `"origin": "https://wstein.github.io"`) || len(p.Files["static/brand/social-preview.png"]) == 0 {
		t.Fatal("missing origin param or social image")
	}
	if p, _ := Render(sheets, Options{}); p.Files["static/sitemap.xml"] != nil {
		t.Fatal("no sitemap without a canonical origin")
	}
	for _, bad := range []string{"http://wstein.github.io/sheetmill/", "https://wstein.github.io/other/", "https://wstein.github.io/sheetmill/?x", "https://user@wstein.github.io/sheetmill/", "https:///sheetmill/", "wstein.github.io/sheetmill/"} {
		if _, ds := Render(sheets, Options{BasePath: "/sheetmill/", SiteURL: bad}); len(ds) == 0 || ds[0].Code != "REN003" {
			t.Error(bad, ds)
		}
	}
}

func TestHugoIntegrationSocialMetadata(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("set SHEETMILL_HUGO to explicitly authorize installed tool")
	}
	p, ds := Render([]Sheet{
		described(t, "git", "title: Git\ndescription: Track \"changes\" & <more>.\n", "## S\n### C\nText.\n"),
		described(t, "misc", "title: Misc\n", "Intro.\n"),
	}, Options{Title: "Library", Description: "Site intro.", BasePath: "/sheetmill/", SiteURL: "https://wstein.github.io/sheetmill/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	out := hugoBuild(t, binary, p)
	page, _ := os.ReadFile(filepath.Join(out, "git", "index.html"))
	for _, want := range []string{
		`<meta name="description" content="Track &#34;changes&#34; &amp; &lt;more&gt;.">`,
		`<meta property="og:title" content="Git · Library">`,
		`<link rel="canonical" href="https://wstein.github.io/sheetmill/git/">`,
		`<meta property="og:url" content="https://wstein.github.io/sheetmill/git/">`,
		`<meta property="og:image" content="https://wstein.github.io/sheetmill/brand/social-preview.png">`,
	} {
		if !bytes.Contains(page, []byte(want)) {
			t.Errorf("page lacks %s", want)
		}
	}
	misc, _ := os.ReadFile(filepath.Join(out, "misc", "index.html"))
	home, _ := os.ReadFile(filepath.Join(out, "index.html"))
	if !bytes.Contains(misc, []byte(`<meta name="description" content="Site intro.">`)) || !bytes.Contains(home, []byte(`<link rel="canonical" href="https://wstein.github.io/sheetmill/">`)) {
		t.Error("fallback description or home canonical missing")
	}
	if _, err := os.Stat(filepath.Join(out, "sitemap.xml")); err != nil {
		t.Error(err)
	}
}

func TestRelatedSheets(t *testing.T) {
	p, ds := Render([]Sheet{
		described(t, "git", "title: Git\ncategory: vcs\ntags: [cli, history]\n", "Intro.\n"),
		described(t, "hg", "title: Mercurial\ncategory: vcs\ntags: [history]\n", "Intro.\n"),
		described(t, "svn", "title: Subversion\ncategory: vcs\n", "Intro.\n"),
		described(t, "tar", "title: tar\ncategory: files\ntags: [cli]\n", "Intro.\n"),
		described(t, "zip", "title: zip\ncategory: files\n", "Intro.\n"),
		described(t, "alone", "title: Alone\n", "Intro.\n"),
	}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	related := func(id string) string {
		var data struct {
			Related []struct {
				Title string `json:"title"`
				URL   string `json:"url"`
			} `json:"related"`
		}
		if err := json.Unmarshal(p.Files["data/sheets/"+id+".json"], &data); err != nil {
			t.Fatal(err)
		}
		var titles []string
		for _, r := range data.Related {
			titles = append(titles, r.Title)
		}
		return strings.Join(titles, ",")
	}
	// hg: shared tag + category = 3; svn: category = 1; tar: shared tag = 2.
	if got := related("git"); got != "Mercurial,tar,Subversion" {
		t.Fatal(got)
	}
	if got := related("alone"); got != "" {
		t.Fatal("unrelated sheets get no list", got)
	}
	if !strings.Contains(string(p.Files["layouts/page.html"]), "related") {
		t.Fatal("page template must show related sheets")
	}
}
