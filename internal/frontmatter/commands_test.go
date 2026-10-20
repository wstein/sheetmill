// SPDX-License-Identifier: EUPL-1.2
package frontmatter

import (
	"reflect"
	"testing"
)

func TestSharedCommandMetadataPositions(t *testing.T) {
	src := []byte("---\ntitle: Commands\ncheatsheet:\n  cards:\n    run:\n      params:\n        file:\n          type: invalid\n---\n## Commands\n")
	r, ds := Parse(src)
	if len(ds) != 1 || ds[0].Code != "CMD001" || ds[0].Span.Start.Line != 8 || ds[0].Span.Start.Column != 11 || r.Metadata != nil {
		t.Fatalf("bad metadata diagnostic: %+v result=%+v", ds, r)
	}
}

const variableSourceHead = "---\ntitle: Git\ncheatsheet:\n  tool: git\n  cards:\n    run:\n      params:\n        arg:\n"

func TestFrontmatterAcceptsVariableSources(t *testing.T) {
	src := "---\ntitle: Git\ncheatsheet:\n  tool: git\n  cards:\n    fetch-branch:\n      params:\n" +
		"        remote:\n          type: string\n          source: git remote\n          options: {column: 1, delimiter: ',', multi: true}\n" +
		"        branch:\n          type: string\n          choices: [main, \"it's\", süd]\n" +
		"---\n## Branches\n"
	r, ds := Parse([]byte(src))
	if len(ds) > 0 {
		t.Fatalf("rejected: %+v", ds)
	}
	params := r.Metadata["cheatsheet"].(map[string]any)["cards"].(map[string]any)["fetch-branch"].(map[string]any)["params"]
	want := map[string]any{
		"remote": map[string]any{"type": "string", "source": "git remote", "options": map[string]any{"column": int64(1), "delimiter": ",", "multi": true}},
		"branch": map[string]any{"type": "string", "choices": []any{"main", "it's", "süd"}},
	}
	if !reflect.DeepEqual(params, want) {
		t.Fatalf("metadata not preserved verbatim: %#v", params)
	}
}

func TestFrontmatterRejectsVariableSourceErrors(t *testing.T) {
	for _, tt := range []struct {
		name, typ, lines string
		line, column     int
	}{
		{"empty choices", "string", "choices: []\n", 10, 11},
		{"duplicate choices", "string", "choices: [a, a]\n", 10, 11},
		{"numeric choices", "string", "choices: [1, 2]\n", 10, 11},
		{"newline in choice", "string", "choices: [\"a\\nb\"]\n", 10, 11},
		{"non-integer int choice", "int", "choices: [x]\n", 10, 11},
		{"empty source", "string", "source: \"\"\n", 10, 11},
		{"source with choices", "string", "choices: [a]\n          source: ls\n", 11, 11},
		{"options without source", "string", "options: {multi: true}\n", 10, 11},
		{"nested option range", "string", "source: ls\n          options:\n            column: 0\n", 12, 13},
		{"unknown option map", "string", "source: ls\n          options:\n            map: x\n", 12, 13},
		{"multi with prevent_extra", "string", "source: ls\n          options:\n            multi: true\n            prevent_extra: true\n", 13, 13},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, ds := Parse([]byte(variableSourceHead + "          type: " + tt.typ + "\n          " + tt.lines + "---\n"))
			if len(ds) != 1 || ds[0].Code != "CMD001" || ds[0].Span.Start.Line != tt.line || ds[0].Span.Start.Column != tt.column || r.Metadata != nil {
				t.Fatalf("want one CMD001 at %d:%d, got %+v", tt.line, tt.column, ds)
			}
		})
	}
}
