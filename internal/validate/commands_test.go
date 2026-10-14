// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"encoding/json"
	"testing"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/model"
)

func TestCommandDescription(t *testing.T) {
	for _, s := range []string{"", " leading", "trailing ", "two\nlines", "tab\there", "null\x00here", "hidden\u200bhere", "line\u2028here", "line\u2029here", "\u00a0space"} {
		t.Run(s, func(t *testing.T) {
			m := map[string]any{"cheatsheet": map[string]any{"cards": map[string]any{"run": map[string]any{"params": map[string]any{"file": map[string]any{"type": "path", "description": s}}}}}}
			if len(CommandMetadata(m, nil)) == 0 {
				t.Fatalf("accepted %q", s)
			}
		})
	}
}

func TestCommandsDirectIR(t *testing.T) {
	for _, tt := range []struct {
		name     string
		classes  []string
		explicit bool
		id       string
		code     string
	}{
		{"semantic", []string{"command"}, true, "run", ""},
		{"alias", []string{"navi"}, true, "run", ""},
		{"dual", []string{"command", "navi"}, true, "run", ""},
		{"unmarked", nil, true, "run", "CMD002"},
		{"generated", []string{"command"}, false, "run", "CMD002"},
		{"bad direct ID", []string{"command"}, true, "1run", "CMD002"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := &model.Document{Schema: 1, Metadata: map[string]any{"cheatsheet": map[string]any{"cards": map[string]any{"run": map[string]any{}}}}, Sections: []model.Section{{Cards: []model.Card{{ID: tt.id, IDExplicit: tt.explicit, Classes: tt.classes}}}}}
			before, _ := json.Marshal(d)
			ds := Commands(d, nil)
			after, _ := json.Marshal(d)
			if string(before) != string(after) {
				t.Fatal("validation mutated authored IR")
			}
			if tt.code == "" {
				if len(ds) > 0 {
					t.Fatal(ds)
				}
				return
			}
			if len(ds) == 0 || ds[0].Code != tt.code || ds[0].Span.Start.Line != 1 || ds[0].Span.Start.Column != 1 {
				t.Fatalf("bad fallback: %+v", ds)
			}
		})
	}
}

func TestCommandMetadataDeterministicPositions(t *testing.T) {
	m := map[string]any{"cheatsheet": map[string]any{"cards": map[string]any{"z": map[string]any{"wrong": true}, "a": map[string]any{"params": map[string]any{"arg": map[string]any{"type": "wrong"}}}}}}
	pos := diagnostic.Span{Start: diagnostic.Position{Line: 8, Column: 5}, End: diagnostic.Position{Line: 8, Column: 9}}
	ds := CommandMetadata(m, map[string]diagnostic.Span{"cheatsheet.cards.a.params.arg.type": pos})
	if len(ds) != 2 || ds[0].Span != pos {
		t.Fatalf("unsorted or unpositioned: %+v", ds)
	}
	for range 20 {
		next, _ := json.Marshal(CommandMetadata(m, map[string]diagnostic.Span{"cheatsheet.cards.a.params.arg.type": pos}))
		first, _ := json.Marshal(ds)
		if string(next) != string(first) {
			t.Fatal("map-order diagnostics")
		}
	}
}
