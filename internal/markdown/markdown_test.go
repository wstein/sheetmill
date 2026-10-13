// SPDX-License-Identifier: EUPL-1.2
package markdown

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/wstein/sheetmill/internal/model"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update golden JSON snapshots")

func TestStructure(t *testing.T) {
	src := []byte("---\ntitle: Editor\n---\nIntro **bold**.\n\n## Editing\n### Selection {.col-span-2 #select}\n\n| Key | Meaning |\n| :-- | --: |\n| `x` | Select |\n\n- [x] Done\n- [ ] Next\n\n> ### Quoted heading\n\n````sh\n### Literal\n```\necho '$HOME'\n````\n\n#### Details\n[Guide](https://example.com) and ~~old~~.\n")
	d, ds := Parse(src)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if len(d.Preamble) != 1 || len(d.Sections) != 1 || len(d.Sections[0].Cards) != 1 {
		t.Fatalf("%+v", d)
	}
	c := d.Sections[0].Cards[0]
	if c.Title != "Selection" || c.ID != "select" || c.Layout.Columns != 2 || c.Span.Start.Line != 7 {
		t.Fatalf("%+v", c)
	}
	j, _ := json.Marshal(d)
	for _, s := range []string{"Table", "TaskCheckBox", "Blockquote", "Strikethrough", "FencedCodeBlock", "### Literal\\n```\\necho '$HOME'\\n"} {
		if !bytes.Contains(j, []byte(s)) {
			t.Errorf("missing %s", s)
		}
	}
}

func TestInvalid(t *testing.T) {
	for _, tt := range []struct{ heading, code string }{
		{"### Orphan", "MD001"}, {"## S\n### X {.navi} {.tip}", "ATTR001"}, {"## S\n### X {.col-span-0}", "ATTR001"}, {"## S\n### X {.col-span-2 .col-span-3}", "ATTR002"}, {"## S\n### X {.unknown}", "ATTR001"}, {"## S\n### X {.navi", "ATTR001"}, {"## S\n### X {#a}\n### Y {#a}", "ID001"}, {"## S {.navi}", "ATTR001"},
	} {
		t.Run(tt.heading, func(t *testing.T) {
			_, ds := Parse([]byte("---\ntitle: Test\n---\n" + tt.heading + "\n"))
			if len(ds) == 0 || ds[0].Code != tt.code {
				t.Fatalf("want %s got %v", tt.code, ds)
			}
		})
	}
}

func TestGolden(t *testing.T) {
	paths, err := filepath.Glob("../../testdata/reference-compatible/*.md")
	if err != nil || len(paths) == 0 {
		t.Fatal("missing fixtures", err)
	}
	for _, p := range paths {
		t.Run(filepath.Base(p), func(t *testing.T) {
			src, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			d, ds := Parse(src)
			if strings.HasSuffix(p, "invalid.md") {
				if len(ds) == 0 {
					t.Fatal("expected diagnostics")
				}
				return
			}
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			got, err := d.JSON()
			if err != nil {
				t.Fatal(err)
			}
			if *updateGolden {
				if err := os.WriteFile(p+".json", got, 0644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(p + ".json")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("golden mismatch: %s", p)
			}
			for range 3 {
				again, ds := Parse(src)
				b, _ := again.JSON()
				if len(ds) > 0 || !bytes.Equal(got, b) {
					t.Fatal("not deterministic")
				}
			}
		})
	}
}

func TestStableIDs(t *testing.T) {
	a, _ := Parse([]byte("---\ntitle: X\n---\n## S\n### A\nText\n### A\nOther\n"))
	b, _ := Parse([]byte("---\ntitle: X\n---\n\n## S\n### A\nText\n### A\nOther\n"))
	for i, c := range a.Sections[0].Cards {
		if c.ID != b.Sections[0].Cards[i].ID {
			t.Fatal("IDs depend on position")
		}
	}
	if a.Sections[0].Cards[0].ID == a.Sections[0].Cards[1].ID {
		t.Fatal("duplicate IDs")
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"---\ntitle: X\n---\n## S\n### C\n```\nx\n```", "---\ntitle: X\n---\n## {.navi}"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		d, ds := Parse(b)
		if len(ds) == 0 {
			a, e := d.JSON()
			if e != nil {
				t.Fatal(e)
			}
			again, _ := Parse(b)
			z, _ := again.JSON()
			if !bytes.Equal(a, z) {
				t.Fatal("unstable")
			}
		}
	})
}

func TestCRLFLiteral(t *testing.T) {
	src := []byte("---\r\ntitle: X\r\n---\r\n## S\r\n### C\r\n```sh\r\necho 'literal'\r\n```\r\n")
	d, ds := Parse(src)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	n := d.Sections[0].Cards[0].Content[0]
	if n.Text != "echo 'literal'\r\n" {
		t.Fatalf("literal changed: %q", n.Text)
	}
}
func TestNestedFenceSpan(t *testing.T) {
	src := []byte("---\ntitle: X\n---\n## S\n### C\n> ```sh\n> echo literal\n> ```\n\nAfter.\n")
	d, ds := Parse(src)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	quote := d.Sections[0].Cards[0].Content[0]
	code := quote.Children[0]
	if code.Text != "echo literal\n" || code.Span.End.Line > 10 {
		t.Fatalf("%+v", code)
	}
}

func TestEscapedHeadingAttributes(t *testing.T) {
	for _, tt := range []struct {
		heading, title string
		classes        int
	}{
		{`x \{.unknown}`, `x {.unknown}`, 0},
		{`x \{.note}`, `x {.note}`, 0},
		{`x \\{.note}`, `x \`, 1},
		{`x \\\{.note}`, `x \{.note}`, 0},
		{`x \{.note} {.tip}`, `x {.note}`, 1},
	} {
		t.Run(tt.heading, func(t *testing.T) {
			d, ds := Parse([]byte("---\ntitle: X\n---\n## S\n### " + tt.heading + "\n"))
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			c := d.Sections[0].Cards[0]
			if c.Title != tt.title || len(c.Classes) != tt.classes {
				t.Fatalf("title=%q classes=%v", c.Title, c.Classes)
			}
		})
	}
}

func TestImplicitFenceContainerBoundary(t *testing.T) {
	for name, body := range map[string]string{
		"quote":       "## S\n### C\n> ```sh\n> literal\n\nOutside.\n\n### D\nend\n",
		"list":        "## S\n### C\n- ```sh\n  literal\n\nOutside.\n\n### D\nend\n",
		"empty quote": "## S\n### C\n> ```sh\n\nOutside.\n\n### D\nend\n",
		"later fence": "## S\n### C\n> ```sh\n> literal\n\nOutside.\n\n```\nunrelated\n```\n\n### D\nend\n",
	} {
		t.Run(name, func(t *testing.T) {
			src := []byte("---\ntitle: X\n---\n" + body)
			d, ds := Parse(src)
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			if len(d.Sections[0].Cards) != 2 {
				t.Fatal("lost card boundary")
			}
			first := d.Sections[0].Cards[0].Content[0]
			boundary := bytes.Index(src, []byte("Outside."))
			if first.Span.End.Offset > boundary {
				t.Fatalf("container overlaps outside content: %+v", first.Span)
			}
			found := false
			var walk func(n model.Node)
			walk = func(n model.Node) {
				if n.Kind == "FencedCodeBlock" {
					found = true
					if n.Span.End.Offset > boundary || n.Span.End.Offset <= n.Span.Start.Offset {
						t.Fatalf("invalid fence span: %+v", n.Span)
					}
				}
				for _, c := range n.Children {
					walk(c)
				}
			}
			walk(first)
			if !found {
				t.Fatal("missing code node")
			}
		})
	}
}

func TestEmailAutolinkDestinations(t *testing.T) {
	d, ds := Parse([]byte("---\ntitle: X\n---\n<a@example.com> user@example.com <mailto:other@example.com> <https://example.com>\n"))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var got []string
	var walk func(model.Node)
	walk = func(n model.Node) {
		if n.Kind == "AutoLink" {
			got = append(got, n.Properties["destination"].(string))
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range d.Preamble {
		walk(n)
	}
	want := []string{"mailto:a@example.com", "mailto:user@example.com", "mailto:other@example.com", "https://example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("destinations %v; want %v", got, want)
	}
}

// Attribute groups apply only to root ## and ### headings. Anywhere else they
// would silently become visible text, so they are diagnosed.
func TestStrayAttributeGroups(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		line       int
	}{
		{"paragraph line", "## S\n### C\nText.\n\n{.marker-none}\n", 8},
		{"after list item", "## S\n### C\n- one\n{.marker-none}\n", 7},
		{"after table", "## S\n### C\n| A | B |\n| - | - |\n| a | b |\n{.show-header}\n", 9},
		{"preamble", "{.cols-2}\n", 4},
		{"inside paragraph", "## S\nFirst line\n{#anchor}\nlast line\n", 6},
		{"content heading", "## S\n### C\n#### Sub {.secondary}\n", 6},
		{"level one heading", "# Title {#top}\n", 4},
		{"nested heading", "## S\n### C\n> ### Quoted {.tip}\n", 6},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, ds := Parse([]byte("---\ntitle: Test\n---\n" + tt.body))
			if len(ds) != 1 || ds[0].Code != "ATTR003" || ds[0].Span.Start.Line != tt.line {
				t.Fatalf("want one ATTR003 on line %d, got %v", tt.line, ds)
			}
		})
	}
	for _, body := range []string{
		"## S\n### C\n\\{.marker-none}\n",
		"## S\n### C\n`{.marker-none}`\n",
		"## S\n### C\n```text\n{.marker-none}\n```\n",
		"## S\n### C\nUse {braces} and {.class} inside prose.\n",
		"## S\n### C\n#### Sub \\{.secondary}\n",
		"## S {#s}\n### C {.tip}\n",
		"## S\n### C\n    {.indented-code}\n",
	} {
		if _, ds := Parse([]byte("---\ntitle: Test\n---\n" + body)); len(ds) > 0 {
			t.Errorf("%q: %v", body, ds)
		}
	}
}

func TestSectionColumns(t *testing.T) {
	d, ds := Parse([]byte("---\ntitle: Test\n---\n## Two {.cols-2 #two}\n### C\n## Default\n"))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if d.Sections[0].Columns != 2 || d.Sections[0].ID != "two" || d.Sections[1].Columns != 0 {
		t.Fatal(d.Sections)
	}
	b, _ := d.JSON()
	if !bytes.Contains(b, []byte(`"columns": 2`)) || bytes.Count(b, []byte(`"columns"`)) != 1 {
		t.Fatal("section columns must be present only when set", string(b))
	}
	for body, code := range map[string]string{
		"## S {.cols-0}":             "ATTR001",
		"## S {.cols-7}":             "ATTR001",
		"## S {.cols-02}":            "ATTR001",
		"## S {.cols-2 .cols-3}":     "ATTR002",
		"## S\n### C {.cols-2}":      "ATTR001",
		"## S {.cols-2 .col-span-2}": "ATTR001",
	} {
		if _, ds := Parse([]byte("---\ntitle: Test\n---\n" + body + "\n")); len(ds) == 0 || ds[0].Code != code {
			t.Errorf("%q: want %s, got %v", body, code, ds)
		}
	}
}
