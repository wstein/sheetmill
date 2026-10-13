// SPDX-License-Identifier: EUPL-1.2
package markdown

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
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
