// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"testing"

	"github.com/wstein/sheetmill/internal/diagnostic"
)

func TestNavigation(t *testing.T) {
	zero, max, twenty := int64(0), int64(1000000), int64(20)
	for _, tc := range []struct {
		name  string
		meta  map[string]any
		group string
		order *int64
		bad   bool
	}{
		{"absent", map[string]any{"title": "T"}, "", nil, false},
		{"empty mapping", map[string]any{"navigation": map[string]any{}}, "", nil, false},
		{"group only keeps exact text", map[string]any{"navigation": map[string]any{"group": " Sheet Mill "}}, " Sheet Mill ", nil, false},
		{"order only", map[string]any{"navigation": map[string]any{"order": int64(20)}}, "", &twenty, false},
		{"explicit zero", map[string]any{"navigation": map[string]any{"order": int64(0)}}, "", &zero, false},
		{"upper bound", map[string]any{"navigation": map[string]any{"group": "G", "order": int64(1000000)}}, "G", &max, false},
		{"null", map[string]any{"navigation": nil}, "", nil, true},
		{"scalar", map[string]any{"navigation": "x"}, "", nil, true},
		{"array", map[string]any{"navigation": []any{}}, "", nil, true},
		{"unknown key", map[string]any{"navigation": map[string]any{"ordr": int64(1)}}, "", nil, true},
		{"deferred title", map[string]any{"navigation": map[string]any{"title": "x"}}, "", nil, true},
		{"empty group", map[string]any{"navigation": map[string]any{"group": ""}}, "", nil, true},
		{"blank group", map[string]any{"navigation": map[string]any{"group": " \t"}}, "", nil, true},
		{"group type", map[string]any{"navigation": map[string]any{"group": int64(1)}}, "", nil, true},
		{"group null", map[string]any{"navigation": map[string]any{"group": nil}}, "", nil, true},
		{"negative", map[string]any{"navigation": map[string]any{"order": int64(-1)}}, "", nil, true},
		{"too large", map[string]any{"navigation": map[string]any{"order": int64(1000001)}}, "", nil, true},
		{"float", map[string]any{"navigation": map[string]any{"order": 1.0}}, "", nil, true},
		{"string", map[string]any{"navigation": map[string]any{"order": "1"}}, "", nil, true},
		{"bool", map[string]any{"navigation": map[string]any{"order": true}}, "", nil, true},
		{"order null", map[string]any{"navigation": map[string]any{"order": nil}}, "", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ds := Navigation(tc.meta, nil)
			if tc.bad {
				if len(ds) == 0 || ds[0].Code != "VAL003" || ds[0].Span.Start.Line != 1 || ds[0].Span.Start.Column != 1 {
					t.Fatalf("want positioned VAL003, got %+v", ds)
				}
				return
			}
			if len(ds) > 0 || got.Group != tc.group || (got.Order == nil) != (tc.order == nil) || (tc.order != nil && *got.Order != *tc.order) {
				t.Fatalf("%+v %+v", got, ds)
			}
		})
	}
}

func TestNavigationSourcePosition(t *testing.T) {
	span := diagnostic.Span{Start: diagnostic.Position{Offset: 9, Line: 3, Column: 5}}
	meta := map[string]any{"title": "T", "navigation": map[string]any{"order": 1.5}}
	_, ds := Navigation(meta, map[string]diagnostic.Span{"": {}, "navigation.order": span})
	if len(ds) != 1 || ds[0].Span != span {
		t.Fatalf("%+v", ds)
	}
	if ds := Metadata(meta, map[string]diagnostic.Span{"navigation.order": span}); len(ds) != 1 || ds[0].Code != "VAL003" {
		t.Fatalf("Metadata must apply the shared rules: %+v", ds)
	}
}
