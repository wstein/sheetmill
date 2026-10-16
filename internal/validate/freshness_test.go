// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/validate/freshcases"
)

func span(line, col int) diagnostic.Span {
	return diagnostic.Span{Start: diagnostic.Position{Line: line, Column: col}, End: diagnostic.Position{Line: line, Column: col + 5}}
}

func TestFreshnessTableSizes(t *testing.T) {
	if len(freshcases.Dates) < 40 || len(freshcases.Versions) < 25 {
		t.Fatalf("SC-001 needs at least 40 date and 25 version cases, have %d and %d", len(freshcases.Dates), len(freshcases.Versions))
	}
}

func TestFreshnessDates(t *testing.T) {
	where := span(4, 3)
	for _, c := range freshcases.Dates {
		t.Run(c.Name, func(t *testing.T) {
			m := map[string]any{"cheatsheet": map[string]any{"updated": c.Value}}
			got, ds := Freshness(m, map[string]diagnostic.Span{"cheatsheet.updated": where, "": span(1, 1)})
			if c.Valid {
				if len(ds) != 0 || got.Updated != c.Value.(string) || got.ToolVersion != "" {
					t.Fatalf("want accepted verbatim, got %+v %+v", got, ds)
				}
				return
			}
			if len(ds) != 1 || ds[0].Code != "VAL004" || ds[0].Span != where || !strings.Contains(ds[0].Message, "cheatsheet.updated") {
				t.Fatalf("want one positioned VAL004, got %+v", ds)
			}
			if got != (FreshnessSettings{}) {
				t.Fatalf("invalid input must yield empty settings, got %+v", got)
			}
		})
	}
}

func TestFreshnessVersions(t *testing.T) {
	where := span(5, 3)
	for _, c := range freshcases.Versions {
		t.Run(c.Name, func(t *testing.T) {
			m := map[string]any{"cheatsheet": map[string]any{"tool_version": c.Value}}
			got, ds := Freshness(m, map[string]diagnostic.Span{"cheatsheet.tool_version": where, "": span(1, 1)})
			if c.Valid {
				if len(ds) != 0 || got.ToolVersion != c.Value.(string) || got.Updated != "" {
					t.Fatalf("want accepted verbatim, got %+v %+v", got, ds)
				}
				return
			}
			if len(ds) != 1 || ds[0].Code != "VAL004" || ds[0].Span != where || !strings.Contains(ds[0].Message, "cheatsheet.tool_version") {
				t.Fatalf("want one positioned VAL004, got %+v", ds)
			}
			if got != (FreshnessSettings{}) {
				t.Fatalf("invalid input must yield empty settings, got %+v", got)
			}
		})
	}
}

func TestFreshnessAbsentAndShapes(t *testing.T) {
	for name, m := range map[string]map[string]any{
		"nil metadata":           nil,
		"no cheatsheet":          {"title": "T"},
		"empty cheatsheet":       {"cheatsheet": map[string]any{}},
		"tool only":              {"cheatsheet": map[string]any{"tool": "git"}},
		"cheatsheet is a list":   {"cheatsheet": []any{"updated"}},
		"cheatsheet is a scalar": {"cheatsheet": "2026-10-14"},
		"cheatsheet is null":     {"cheatsheet": nil},
	} {
		t.Run(name, func(t *testing.T) {
			got, ds := Freshness(m, nil)
			if len(ds) != 0 || got != (FreshnessSettings{}) {
				t.Fatalf("absent or non-mapping cheatsheet must be silent (VAL003 owns it), got %+v %+v", got, ds)
			}
		})
	}
}

func TestFreshnessBothValidAndIndependent(t *testing.T) {
	m := map[string]any{"cheatsheet": map[string]any{"updated": "2026-10-14", "tool_version": "2.56.0", "tool": "git"}}
	got, ds := Freshness(m, nil)
	if len(ds) != 0 || got != (FreshnessSettings{Updated: "2026-10-14", ToolVersion: "2.56.0"}) {
		t.Fatalf("got %+v %+v", got, ds)
	}
}

func TestFreshnessNilLocationsUseDocumentStart(t *testing.T) {
	m := map[string]any{"cheatsheet": map[string]any{"updated": "nope", "tool_version": ""}}
	_, ds := Freshness(m, nil)
	if len(ds) != 2 {
		t.Fatalf("want two diagnostics, got %+v", ds)
	}
	for _, d := range ds {
		if d.Code != "VAL004" || d.Span.Start.Line != 1 || d.Span.Start.Column != 1 {
			t.Fatalf("want document start, got %+v", d)
		}
	}
}

func TestFreshnessDeterministicOrderAndPositions(t *testing.T) {
	m := map[string]any{"cheatsheet": map[string]any{"tool_version": "\n", "updated": 5}}
	locs := map[string]diagnostic.Span{"cheatsheet.updated": span(9, 1), "cheatsheet.tool_version": span(3, 1)}
	first, _ := Freshness(m, locs)
	_ = first
	var want []diagnostic.Diagnostic
	for i := 0; i < 50; i++ {
		_, ds := Freshness(m, locs)
		if i == 0 {
			want = ds
			if len(ds) != 2 || !strings.Contains(ds[0].Message, "updated") || !strings.Contains(ds[1].Message, "tool_version") {
				t.Fatalf("updated is reported before tool_version, got %+v", ds)
			}
			if ds[0].Span != span(9, 1) || ds[1].Span != span(3, 1) {
				t.Fatalf("each diagnostic sits at its own key, got %+v", ds)
			}
		}
		if !reflect.DeepEqual(ds, want) {
			t.Fatal("diagnostics must be deterministic")
		}
	}
}

func TestFreshnessKeepsUnknownCheatsheetKeys(t *testing.T) {
	m := map[string]any{"cheatsheet": map[string]any{"updated": "2026-10-14", "owner": "docs-team", "extra": map[string]any{"a": int64(1)}}}
	before := map[string]any{"cheatsheet": map[string]any{"updated": "2026-10-14", "owner": "docs-team", "extra": map[string]any{"a": int64(1)}}}
	got, ds := Freshness(m, nil)
	if len(ds) != 0 || got.Updated != "2026-10-14" {
		t.Fatalf("an unknown key next to valid freshness is not an error: %+v %+v", got, ds)
	}
	if !reflect.DeepEqual(m, before) {
		t.Fatal("Freshness must not modify or drop metadata")
	}
	if ds := Metadata(map[string]any{"title": "T", "cheatsheet": m["cheatsheet"]}, nil); len(ds) != 0 {
		t.Fatalf("Metadata must accept unknown cheatsheet keys beside freshness: %+v", ds)
	}
}

func TestFreshnessNonStringAdvisesQuoting(t *testing.T) {
	for _, key := range []string{"updated", "tool_version"} {
		for _, v := range []any{3.81, 3.10, int64(3), true, nil} {
			_, ds := Freshness(map[string]any{"cheatsheet": map[string]any{key: v}}, nil)
			if len(ds) != 1 || !strings.Contains(ds[0].Message, "must be a quoted YAML string") || !strings.Contains(ds[0].Message, "cheatsheet."+key) {
				t.Errorf("%s=%#v: want a quote-it message, got %+v", key, v, ds)
			}
		}
	}
	// A string that is merely invalid keeps the rule text, not the quoting advice.
	_, ds := Freshness(map[string]any{"cheatsheet": map[string]any{"tool_version": ""}}, nil)
	if len(ds) != 1 || strings.Contains(ds[0].Message, "quoted YAML string") || !strings.Contains(ds[0].Message, "1 to 64") {
		t.Errorf("invalid string: want the rule message, got %+v", ds)
	}
}
