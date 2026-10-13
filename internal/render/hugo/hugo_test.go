// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
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
	css := string(p.Files["static/assets/site.css"])
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
	p, ds := Render([]Sheet{a, b}, Options{BasePath: "/sheetmill/"})
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
	for _, name := range []string{"index.html", "a/index.html", "b/index.html", "search.json", "assets/site.css", "brand/favicon.svg"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"index.html", "a/index.html"} {
		html, _ := os.ReadFile(filepath.Join(out, name))
		for _, want := range []string{`data-base="/sheetmill/"`, `href="/sheetmill/assets/site.css"`, `src="/sheetmill/assets/app.js"`, `href="/sheetmill/a/"`} {
			if !bytes.Contains(html, []byte(want)) {
				t.Errorf("%s lacks %s", name, want)
			}
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
