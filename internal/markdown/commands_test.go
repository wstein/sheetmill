// SPDX-License-Identifier: EUPL-1.2
package markdown

import (
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
