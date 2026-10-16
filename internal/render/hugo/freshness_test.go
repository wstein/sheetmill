// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"testing"

	"github.com/wstein/sheetmill/internal/validate/freshcases"
)

func freshnessSheet(t *testing.T, id string, cs map[string]any) Sheet {
	t.Helper()
	s := sheet(t, id, "## S\n### C\nText.\n")
	s.Document.Metadata["cheatsheet"] = cs
	return s
}

func TestFreshnessDirectIRValidation(t *testing.T) {
	for _, field := range []struct {
		key   string
		cases []freshcases.Case
	}{{"updated", freshcases.Dates}, {"tool_version", freshcases.Versions}} {
		for _, c := range field.cases {
			t.Run(field.key+"/"+c.Name, func(t *testing.T) {
				p, ds := Render([]Sheet{freshnessSheet(t, "a", map[string]any{field.key: c.Value})}, Options{})
				if c.Valid {
					if len(ds) != 0 || len(p.Files) == 0 {
						t.Fatalf("want accepted, got %+v", ds)
					}
					return
				}
				if len(ds) != 1 || ds[0].Code != "VAL004" || ds[0].Source != "a.md" || len(p.Files) != 0 {
					t.Fatalf("want one VAL004 and an empty project, got %+v with %d files", ds, len(p.Files))
				}
				if ds[0].Span.Start.Line != 1 || ds[0].Span.Start.Column != 1 {
					t.Fatalf("direct IR has no source positions, want document start: %+v", ds[0])
				}
			})
		}
	}
}

func TestFreshnessDirectIRIgnoresNonMappingCheatsheet(t *testing.T) {
	s := sheet(t, "a", "## S\n### C\nText.\n")
	s.Document.Metadata["cheatsheet"] = "2026-10-14"
	if p, ds := Render([]Sheet{s}, Options{}); len(ds) != 0 || len(p.Files) == 0 {
		t.Fatalf("a non-mapping cheatsheet is ignored as for cheatsheet.tool: %+v", ds)
	}
}
