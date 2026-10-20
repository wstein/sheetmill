// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"encoding/json"
	"testing"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/model"
	"github.com/wstein/sheetmill/internal/validate/varcases"
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

func paramMetadata(fields map[string]any) map[string]any {
	return map[string]any{"cheatsheet": map[string]any{"cards": map[string]any{"run": map[string]any{"params": map[string]any{"arg": fields}}}}}
}

// checkParamCases runs a shared case table: valid declarations produce nothing,
// invalid ones only CMD001, and authored metadata is never mutated.
func checkParamCases(t *testing.T, cases []varcases.Case) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			m := paramMetadata(c.Fields)
			before, _ := json.Marshal(m)
			ds := CommandMetadata(m, nil)
			after, _ := json.Marshal(m)
			if string(before) != string(after) {
				t.Fatal("validation mutated authored metadata")
			}
			if c.Valid && len(ds) > 0 {
				t.Fatalf("rejected: %+v", ds)
			}
			if !c.Valid && len(ds) == 0 {
				t.Fatal("accepted")
			}
			for _, d := range ds {
				if d.Code != "CMD001" {
					t.Fatalf("want CMD001, got %+v", d)
				}
			}
		})
	}
}

func TestCommandParamChoicesShape(t *testing.T) { checkParamCases(t, varcases.Choices) }

func TestCommandParamIntChoicesAreIntegerStrings(t *testing.T) {
	checkParamCases(t, varcases.IntChoices)
}

func TestCommandParamSourceShape(t *testing.T) { checkParamCases(t, varcases.Sources) }

func TestCommandParamChoicesAndSourceExclusive(t *testing.T) {
	checkParamCases(t, varcases.Exclusive)
}

func TestCommandParamOptionsShape(t *testing.T) { checkParamCases(t, varcases.Options) }

func TestCommandParamNewFieldsPositions(t *testing.T) {
	at := func(line, col int) diagnostic.Span {
		return diagnostic.Span{Start: diagnostic.Position{Line: line, Column: col}, End: diagnostic.Position{Line: line, Column: col + 1}}
	}
	const p = "cheatsheet.cards.run.params.arg"
	locations := map[string]diagnostic.Span{
		p: at(5, 9), p + ".choices": at(7, 11), p + ".source": at(8, 11), p + ".options": at(9, 11),
		p + ".options.column": at(10, 13), p + ".options.multi": at(11, 13), p + ".options.prevent_extra": at(12, 13),
	}
	for _, tt := range []struct {
		name   string
		fields map[string]any
		want   diagnostic.Span
	}{
		{"empty choices", map[string]any{"type": "string", "choices": []any{}}, locations[p+".choices"]},
		{"source newline", map[string]any{"type": "string", "source": "a\nb"}, locations[p+".source"]},
		{"exclusive", map[string]any{"type": "string", "choices": []any{"a"}, "source": "ls"}, locations[p+".source"]},
		{"nested option", map[string]any{"type": "string", "source": "ls", "options": map[string]any{"column": int64(0)}}, locations[p+".options.column"]},
		{"option dependency", map[string]any{"type": "string", "choices": []any{"a"}, "options": map[string]any{"column": int64(1)}}, locations[p+".options.column"]},
		{"flags", map[string]any{"type": "string", "source": "ls", "options": map[string]any{"multi": true, "prevent_extra": true}}, locations[p+".options.prevent_extra"]},
		{"options alone", map[string]any{"type": "string", "options": map[string]any{}}, locations[p+".options"]},
		{"nearest parent", map[string]any{"type": "string", "source": "ls", "options": map[string]any{"delimiter": ""}}, locations[p+".options"]},
	} {
		t.Run(tt.name, func(t *testing.T) {
			loc := locations
			if tt.name == "nearest parent" {
				loc = map[string]diagnostic.Span{p + ".options": locations[p+".options"]}
			}
			ds := CommandMetadata(paramMetadata(tt.fields), loc)
			if len(ds) != 1 || ds[0].Code != "CMD001" || ds[0].Span != tt.want {
				t.Fatalf("want one CMD001 at %+v, got %+v", tt.want, ds)
			}
			// Direct IR carries no locations: every diagnostic sits at 1:1.
			direct := CommandMetadata(paramMetadata(tt.fields), nil)
			if len(direct) != 1 || direct[0].Span.Start != (diagnostic.Position{Line: 1, Column: 1}) {
				t.Fatalf("direct IR must anchor at 1:1: %+v", direct)
			}
		})
	}
}
