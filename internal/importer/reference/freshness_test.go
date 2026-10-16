// SPDX-License-Identifier: EUPL-1.2
package reference

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/frontmatter"
	"github.com/wstein/sheetmill/internal/markdown"
)

// Upstream updated/tool_version have unknown semantics: archive them inertly and
// never activate, infer or fail on them (spec 017 FR-010).
func TestUpstreamFreshnessIsArchivedInert(t *testing.T) {
	for _, tt := range []struct {
		name    string
		fields  []string // archived keys in finding order
		cards   bool
		codes   string
		lines   []int // source line of each IMP112 finding
		message string
	}{
		{"valid", []string{"updated", "tool_version"}, false, "IMP112,IMP112", []int{5, 6}, "updated"},
		{"wrong-type", []string{"updated", "tool_version"}, false, "IMP112,IMP112", []int{4, 5}, "updated"},
		{"with-cards", []string{"updated", "tool_version"}, true, "IMP111,IMP112,IMP112", []int{9, 4}, "cards"},
		{"invalid-date", []string{"updated"}, false, "IMP112", []int{4}, "updated"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join("../../../testdata/import/reference/freshness", tt.name+".md"))
			if err != nil {
				t.Fatal(err)
			}
			decoded, dds := frontmatter.Decode(src)
			if len(dds) > 0 {
				t.Fatal(dds)
			}
			upstream := decoded.Metadata["cheatsheet"].(map[string]any)

			r, ds := Convert(src, provenance())
			if len(ds) > 0 {
				t.Fatalf("freshness of any shape must never fail an import: %+v", ds)
			}
			d, pds := markdown.Parse(r.Markdown)
			if len(pds) > 0 {
				t.Fatalf("the archived output must validate: %+v", pds)
			}
			if active, _ := d.Metadata["cheatsheet"].(map[string]any); active != nil {
				for _, key := range []string{"updated", "tool_version"} {
					if _, ok := active[key]; ok {
						t.Fatalf("%s remains active metadata: %v", key, active)
					}
				}
			}
			original := d.Metadata["source"].(map[string]any)["original_metadata"].(map[string]any)
			archive, _ := original["cheatsheet"].(map[string]any)
			for _, key := range tt.fields {
				if !reflect.DeepEqual(archive[key], upstream[key]) {
					t.Fatalf("%s must be archived verbatim: got %#v want %#v", key, archive[key], upstream[key])
				}
			}
			if tt.cards != (archive["cards"] != nil) || (tt.cards && !reflect.DeepEqual(archive["cards"], upstream["cards"])) {
				t.Fatalf("the cards archive must stay intact beside freshness: %#v", archive)
			}
			var codes []string
			var lines []int
			for _, f := range r.Report.Findings {
				codes = append(codes, f.Code)
				if f.Code == "IMP112" {
					lines = append(lines, f.Span.Start.Line)
					if f.Span.Start.Column != 3 || !strings.Contains(f.Message, "cheatsheet.") {
						t.Errorf("IMP112 must name and sit at its key: %+v", f)
					}
				}
			}
			if strings.Join(codes, ",") != tt.codes {
				t.Fatalf("finding order: %v", r.Report.Findings)
			}
			if tt.name != "with-cards" && !reflect.DeepEqual(lines, tt.lines) {
				t.Fatalf("IMP112 lines %v want %v", lines, tt.lines)
			}
			// Repeating the conversion yields identical bytes: no inferred date.
			again, _ := Convert(src, provenance())
			if string(again.Markdown) != string(r.Markdown) {
				t.Fatal("conversion is not deterministic")
			}
		})
	}
}

func TestReferenceWithoutFreshnessEmitsNoIMP112(t *testing.T) {
	r, ds := Convert([]byte("---\ntitle: X\ncheatsheet:\n  tool: git\n---\n## S\n### C\nText.\n"), provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	for _, f := range r.Report.Findings {
		if f.Code == "IMP112" {
			t.Fatalf("unexpected %+v", f)
		}
	}
}

// When updated/tool_version were the only cheatsheet keys, archiving them must not
// leave an empty "cheatsheet: {}" behind as active metadata.
func TestEmptiedCheatsheetMappingIsRemoved(t *testing.T) {
	r, ds := Convert([]byte("---\ntitle: X\ncheatsheet:\n  updated: 2026-10-14\n  tool_version: \"1.0\"\n---\n## S\n### C\nText.\n"), provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	d, pds := markdown.Parse(r.Markdown)
	if len(pds) > 0 {
		t.Fatal(pds)
	}
	if v, ok := d.Metadata["cheatsheet"]; ok {
		t.Fatalf("emptied cheatsheet mapping must be removed, got %#v", v)
	}
	original := d.Metadata["source"].(map[string]any)["original_metadata"].(map[string]any)
	if original["cheatsheet"].(map[string]any)["updated"] == nil {
		t.Fatalf("the archive must keep the values: %#v", original)
	}
}

func TestAuthoredEmptyCheatsheetMappingIsKept(t *testing.T) {
	r, ds := Convert([]byte("---\ntitle: X\ncheatsheet: {}\n---\n## S\n### C\nText.\n"), provenance())
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	d, pds := markdown.Parse(r.Markdown)
	if len(pds) > 0 {
		t.Fatal(pds)
	}
	if _, ok := d.Metadata["cheatsheet"]; !ok {
		t.Fatal("an authored empty mapping is not ours to remove")
	}
}
