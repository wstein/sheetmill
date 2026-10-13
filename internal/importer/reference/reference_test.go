// SPDX-License-Identifier: EUPL-1.2
package reference

import (
	"bytes"
	"flag"
	"github.com/wstein/sheetmill/internal/markdown"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var sample = []byte("---\ntitle: Sample\nplugins: [copyCode, runCode]\nbackground: bg-blue\ntags: [cli]\ncustom: {keep: true}\n---\n## Start\n### Literal {.col-span-2 .row-span-3}\n````python\nprint('### literal')\n```\n````\n\n| Key | Meaning |\n| --- | --- |\n| x | sample |\n")

func provenance() Provenance {
	return Provenance{Repository: "https://github.com/example/reference", Path: "source/_posts/sample.md", Revision: strings.Repeat("a", 40), License: "GPL-3.0", Attribution: "Example contributors"}
}
func TestConvert(t *testing.T) {
	p := provenance()
	a, ds := Convert(sample, p)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	b, ds := Convert(sample, p)
	if len(ds) > 0 || !bytes.Equal(a.Markdown, b.Markdown) {
		t.Fatal("unstable conversion")
	}
	aj, _ := a.Report.JSON()
	bj, _ := b.Report.JSON()
	if !bytes.Equal(aj, bj) {
		t.Fatal("unstable report")
	}
	d, ds := markdown.Parse(a.Markdown)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	body := sample[bytes.Index(sample, []byte("## Start")):]
	if d.Source != string(body) {
		t.Fatal("body changed")
	}
	if _, ok := d.Metadata["plugins"]; ok {
		t.Fatal("plugin declarations remain active metadata")
	}
	src := d.Metadata["source"].(map[string]any)
	if src["revision"] != p.Revision || src["license"] != p.License || src["format"] != "fechin-reference" {
		t.Fatal(src)
	}
	if len(a.Report.Findings) != 2 || a.Report.Findings[0].Code != "IMP101" || a.Report.Findings[1].Code != "IMP102" {
		t.Fatal(a.Report)
	}
	if a.Report.SourceSHA256 == "" || a.Report.ConvertedSHA256 == "" {
		t.Fatal("missing hashes")
	}
	if _, ds = Convert(a.Markdown, p); len(ds) == 0 || ds[0].Code != "IMP002" {
		t.Fatal("reimported canonical provenance", ds)
	}
}
func TestInvalidProvenance(t *testing.T) {
	for _, name := range []string{"repository", "path", "revision", "license", "attribution"} {
		t.Run(name, func(t *testing.T) {
			p := provenance()
			switch name {
			case "repository":
				p.Repository = "file:///tmp/x"
			case "path":
				p.Path = "../escape.md"
			case "revision":
				p.Revision = "main"
			case "license":
				p.License = ""
			case "attribution":
				p.Attribution = ""
			}
			if _, ds := Convert(sample, p); len(ds) == 0 || ds[0].Code != "IMP001" {
				t.Fatal(ds)
			}
		})
	}
}

// Unsupported classes are archived (TestLegacyMapping); malformed or
// conflicting groups that remain still fail before anything is written.
func TestUnsupportedFails(t *testing.T) {
	for _, tt := range []struct{ from, to, code string }{
		{".col-span-2", ".col-span-2 .col-span-3", "ATTR002"},
		{"### Literal {.col-span-2 .row-span-3}", "### Literal {.col-span-2 .row-span-3", "ATTR001"},
		{"### Literal", "### Literal {.tip} {.note}", "ATTR001"},
	} {
		b := bytes.Replace(sample, []byte(tt.from), []byte(tt.to), 1)
		if _, ds := Convert(b, provenance()); len(ds) == 0 || ds[0].Code != tt.code {
			t.Errorf("%s: want %s, got %v", tt.to, tt.code, ds)
		}
	}
}
func TestLegacyMetadataPreserved(t *testing.T) {
	a, ds := Convert(sample, provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	d, ds := markdown.Parse(a.Markdown)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	src := d.Metadata["source"].(map[string]any)
	legacy := src["original_metadata"].(map[string]any)
	if len(legacy["plugins"].([]any)) != 2 || legacy["background"] != "bg-blue" || d.Metadata["custom"].(map[string]any)["keep"] != true {
		t.Fatal(d.Metadata)
	}
}
func TestPinnedPython(t *testing.T) {
	p := os.Getenv("SHEETMILL_REFERENCE_PYTHON")
	if p == "" {
		t.Skip("optional pinned upstream benchmark; no network in tests")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	prov := Provenance{Repository: "https://github.com/Fechin/reference", Path: "source/_posts/python.md", Revision: "82cbc0929f2d7a44e0bb5fc77e0a8442d0230b59", License: "GPL-3.0", Attribution: "Fechin/reference contributors (benchmark, no redistribution)"}
	if hash(b) != "e56491b4bbe22d6e58b259b02cd62bfea136cbe79f2704a1631d3b919db8d1d5" {
		t.Fatal("benchmark bytes do not match pinned source")
	}
	a, ds := Convert(b, prov)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	d, ds := markdown.Parse(a.Markdown)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	cards := 0
	for _, s := range d.Sections {
		cards += len(s.Cards)
	}
	if cards < 50 || len(d.Sections) < 5 {
		t.Fatalf("unexpected benchmark structure %d sections %d cards", len(d.Sections), cards)
	}
	t.Logf("revision=%s source_sha256=%s sections=%d cards=%d findings=%d", prov.Revision, a.Report.SourceSHA256, len(d.Sections), cards, len(a.Report.Findings))
}
func FuzzConvert(f *testing.F) {
	f.Add(sample)
	f.Add([]byte("---\ntitle: X\n---\n"))
	f.Fuzz(func(t *testing.T, b []byte) {
		a, ds := Convert(b, provenance())
		if len(ds) == 0 {
			if _, ds := markdown.Parse(a.Markdown); len(ds) > 0 {
				t.Fatal(ds)
			}
			again, _ := Convert(b, provenance())
			if !bytes.Equal(a.Markdown, again.Markdown) {
				t.Fatal("unstable")
			}
		}
	})
}

var updateGolden = flag.Bool("update", false, "update importer golden fixtures")

func TestGolden(t *testing.T) {
	paths, err := filepath.Glob("../../../testdata/import/reference/*.md")
	if err != nil || len(paths) == 0 {
		t.Fatal("missing import fixtures", err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			result, ds := Convert(b, provenance())
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			report, err := result.Report.JSON()
			if err != nil {
				t.Fatal(err)
			}
			for suffix, got := range map[string][]byte{".canonical": result.Markdown, ".report.json": report} {
				if *updateGolden {
					if err := os.WriteFile(path+suffix, got, 0644); err != nil {
						t.Fatal(err)
					}
				}
				want, err := os.ReadFile(path + suffix)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("golden mismatch: %s%s", path, suffix)
				}
			}
		})
	}
}
func TestLossReports(t *testing.T) {
	b := []byte("---\ntitle: X\n---\n## S\n### C\n{% unsupported example %}\n\n<script>alert('inert')</script>\n\n```text\n{% literal %}\n```\n")
	r, ds := Convert(b, provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if len(r.Report.Findings) != 2 || r.Report.Findings[0].Code != "IMP104" || r.Report.Findings[1].Code != "IMP103" {
		t.Fatal(r.Report.Findings)
	}
}
func TestErrorReport(t *testing.T) {
	r, ds := Convert([]byte("---\ntitle: X\n---\n## S\n### C {.tip\n"), provenance())
	if len(ds) == 0 || len(r.Report.Findings) != 1 || r.Report.Findings[0].Severity != "error" || r.Report.Findings[0].Code != "ATTR001" || len(r.Markdown) != 0 {
		t.Fatal(r, ds)
	}
}

func TestHeadingLossReports(t *testing.T) {
	b := []byte("---\ntitle: X\n---\n## <script>x</script>\n### {% plugin x %}\n\nBody.\n\n### `{% literal %}`\n")
	r, ds := Convert(b, provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	html, directives := 0, 0
	lastOffset := -1
	for _, f := range r.Report.Findings {
		if f.Span.Start.Offset < lastOffset {
			t.Fatal("findings out of source order", r.Report.Findings)
		}
		lastOffset = f.Span.Start.Offset
		switch f.Code {
		case "IMP103":
			html++
			if f.Span.Start.Line != 4 {
				t.Fatal(f)
			}
		case "IMP104":
			directives++
			if f.Span.Start.Line != 5 {
				t.Fatal(f)
			}
		}
	}
	if html == 0 || directives != 1 {
		t.Fatalf("missing heading findings or code literal scanned: %+v", r.Report.Findings)
	}
}

// Independently written fixture using Reference conventions that SheetMill
// does not support natively.
var legacy = []byte("---\ntitle: Tool\nintro: A short intro.\ncategories: [Linux Command, Other]\ntags: [cli, null, \"\"]\n---\n## Basics {.cols-2}\n### Keys {.col-span-2 .secondary}\n| Key | Action |\n| --- | --- |\n| `x` | cut |\n{.shortcuts}\n\n- one\n{.marker-none}\n\n#### Sub {.left-text}\n\n```text\n{.keep-in-code}\n```\n### Plain {.row-span-2 .tip}\nText.\n")

func TestLegacyMapping(t *testing.T) {
	r, ds := Convert(legacy, provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	d, ds := markdown.Parse(r.Markdown)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if d.Metadata["description"] != "A short intro." || d.Metadata["category"] != "Linux Command" {
		t.Fatal(d.Metadata)
	}
	if tags := d.Metadata["tags"].([]any); len(tags) != 1 || tags[0] != "cli" {
		t.Fatal("invalid tags must be dropped", tags)
	}
	original := d.Metadata["source"].(map[string]any)["original_metadata"].(map[string]any)
	if original["intro"] != "A short intro." || len(original["categories"].([]any)) != 2 || len(original["tags"].([]any)) != 3 {
		t.Fatal("original values must be archived", original)
	}
	removed := original["removed_attributes"].([]any)
	if len(removed) != 4 {
		t.Fatal(removed)
	}
	if first := removed[0].(map[string]any); first["line"] != int64(8) || first["text"] != ".secondary" {
		t.Fatal(first)
	}
	body := string(r.Markdown[bytes.Index(r.Markdown, []byte("## Basics")):])
	for _, want := range []string{"## Basics {.cols-2}\n", "### Keys {.col-span-2}\n", "#### Sub\n", "{.keep-in-code}", "### Plain {.row-span-2 .tip}\n"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in %s", want, body)
		}
	}
	for _, gone := range []string{".secondary", "{.shortcuts}", "{.marker-none}", ".left-text"} {
		if strings.Contains(body, gone) {
			t.Errorf("%s survived in %s", gone, body)
		}
	}
	codes := map[string]int{}
	for _, f := range r.Report.Findings {
		codes[f.Code]++
		if f.Severity != "warning" {
			t.Error(f)
		}
	}
	if codes["IMP105"] != 1 || codes["IMP106"] != 1 || codes["IMP107"] != 1 || codes["IMP108"] != 2 || codes["IMP109"] != 2 {
		t.Fatal(codes, r.Report.Findings)
	}
	for _, f := range r.Report.Findings {
		if f.Code == "IMP109" && f.Span.Start.Line != 12 && f.Span.Start.Line != 15 {
			t.Error("findings must point at original lines", f)
		}
	}
}

func TestLegacyMappingKeepsExplicitFields(t *testing.T) {
	r, ds := Convert([]byte("---\ntitle: T\ndescription: Kept.\nintro: Other.\ncategory: kept\ncategories: [x]\ntags: solo\n---\n## S\n"), provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	d, _ := markdown.Parse(r.Markdown)
	if d.Metadata["description"] != "Kept." || d.Metadata["category"] != "kept" {
		t.Fatal(d.Metadata)
	}
	if _, ok := d.Metadata["tags"]; ok {
		t.Fatal("a non-array tags value must be archived, not kept", d.Metadata)
	}
}
