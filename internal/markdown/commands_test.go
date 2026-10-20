// SPDX-License-Identifier: EUPL-1.2
package markdown

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCommandCards(t *testing.T) {
	for _, tt := range []struct{ name, settings, heading, code string }{
		{"shared", "cards: {run: {params: {file: {type: path, description: Input file}}}}", "Run {#run .command}", ""},
		{"no params", "", "Run {#run .command}", ""},
		{"empty mappings", "cards: {}", "Run {#run .command}", ""},
		{"legacy shared ref", "cards: {run: {}}", "Run {#run .navi}", ""},
		{"dual markers", "cards: {run: {}}", "Run {#run .navi .command}", ""},
		{"missing explicit", "", "Run {.command}", "CMD002"},
		{"dangling", "cards: {other: {}}", "Run {#run .command}", "CMD002"},
		{"unmarked", "cards: {run: {}}", "Run {#run}", "CMD002"},
		{"generated reference", "cards: {run: {}}", "Run {.navi}", "CMD002"},
		{"legacy generated unchanged", "", "Run {.navi}", ""},
		{"cards null", "cards: null", "Run {#run .command}", "CMD001"},
		{"settings null", "cards: {run: null}", "Run {#run .command}", "CMD001"},
		{"unknown settings", "cards: {run: {shell: bash}}", "Run {#run .command}", "CMD001"},
		{"params list", "cards: {run: {params: []}}", "Run {#run .command}", "CMD001"},
		{"invalid name", "cards: {run: {params: {File: {type: path}}}}", "Run {#run .command}", "CMD001"},
		{"invalid id", "cards: {'1run': {}}", "Run {#run .command}", "CMD001"},
		{"missing type", "cards: {run: {params: {file: {}}}}", "Run {#run .command}", "CMD001"},
		{"bad type", "cards: {run: {params: {file: {type: choice}}}}", "Run {#run .command}", "CMD001"},
		{"coercion", "cards: {run: {params: {file: {type: 1}}}}", "Run {#run .command}", "CMD001"},
		{"unknown param", "cards: {run: {params: {file: {type: path, default: a}}}}", "Run {#run .command}", "CMD001"},
		{"variable sources", "cards: {run: {params: {file: {type: string, choices: [a, b]}}}}", "Run {#run .command}", ""},
		{"trimmed description", "cards: {run: {params: {file: {type: path, description: ' input'}}}}", "Run {#run .command}", "CMD001"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			src := "---\ntitle: Commands\ncheatsheet: {" + tt.settings + "}\n---\n## Work\n### " + tt.heading + "\n```bash\necho '<file>'\n```\n"
			d, ds := Parse([]byte(src))
			if tt.code != "" {
				if len(ds) == 0 || ds[0].Code != tt.code || d != nil {
					t.Fatalf("want %s, got %#v %v", tt.code, d, ds)
				}
				if ds[0].Span.Start.Line < 3 {
					t.Fatalf("missing source span: %+v", ds)
				}
				return
			}
			if len(ds) > 0 || d == nil {
				t.Fatalf("parse: %v", ds)
			}
			if !strings.Contains(d.Source, tt.heading) {
				t.Fatal("source rewritten")
			}
			if strings.Contains(tt.heading, ".command") && !strings.Contains(strings.Join(d.Sections[0].Cards[0].Classes, ","), "command") {
				t.Fatal("marker lost")
			}
		})
	}
}

// TestGoldenCommandCardsUnchanged: the new parameter fields live in metadata only;
// the parsed sections of a command sheet are identical with and without them.
func TestGoldenCommandCardsUnchanged(t *testing.T) {
	const body = "---\ntitle: Git\ncheatsheet:\n  tool: git\n  cards:\n    fetch:\n      params:\n        branch:\n          type: string\n%s---\n## Branches\n### Fetch {#fetch .command}\n```bash\ngit fetch <branch>\n```\n"
	// The plain sheet carries a description of equal length so the spans match.
	declared := "          choices: [main, \"it's\"]\n"
	filler := "          description: " + strings.Repeat("x", len(declared)-len("          description: \n")) + "\n"
	plain, ds := Parse([]byte(strings.Replace(body, "%s", filler, 1)))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	sourced, ds := Parse([]byte(strings.Replace(body, "%s", declared, 1)))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	a, _ := json.Marshal(plain.Sections)
	b, _ := json.Marshal(sourced.Sections)
	if string(a) != string(b) {
		t.Fatalf("sections differ:\n%s\n%s", a, b)
	}
	params := sourced.Metadata["cheatsheet"].(map[string]any)["cards"].(map[string]any)["fetch"].(map[string]any)["params"]
	want := map[string]any{"branch": map[string]any{"type": "string", "choices": []any{"main", "it's"}}}
	if !reflect.DeepEqual(params, want) {
		t.Fatalf("metadata not preserved verbatim: %#v", params)
	}
}
