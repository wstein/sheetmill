// SPDX-License-Identifier: EUPL-1.2
package reference

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/frontmatter"
	"github.com/wstein/sheetmill/internal/markdown"
)

func TestUpstreamSharedCardsArchivedBeforeValidation(t *testing.T) {
	for _, value := range []string{"null", "false", "[broken]", "{bad: null}", "{}", "{run: {params: {file: {type: path, description: Input file}}}}"} {
		t.Run(value, func(t *testing.T) {
			src := []byte("---\ntitle: T\ncheatsheet:\n  tool: smill\n  custom: {keep: true}\n  cards: " + value + "\n  navi: {shell: bash}\n---\n## Commands\n### Example {#run .command .navi}\n```bash\nprintf '%s' literal\n```\n")
			decoded, ds := frontmatter.Decode(src)
			if len(ds) != 0 {
				t.Fatal(ds)
			}
			result, ds := Convert(src, provenance())
			if len(ds) != 0 {
				t.Fatal(ds)
			}
			doc, ds := markdown.Parse(result.Markdown)
			if len(ds) != 0 {
				t.Fatal(ds)
			}
			sheet := doc.Metadata["cheatsheet"].(map[string]any)
			if _, active := sheet["cards"]; active {
				t.Fatal("upstream shared cards remain active")
			}
			original := doc.Metadata["source"].(map[string]any)["original_metadata"].(map[string]any)
			archived := original["cheatsheet"].(map[string]any)
			expected := decoded.Metadata["cheatsheet"].(map[string]any)
			if got, ok := archived["cards"]; !ok || !reflect.DeepEqual(got, expected["cards"]) {
				t.Fatal("shared cards not preserved exactly", archived)
			}
			if sheet["tool"] != "smill" || !reflect.DeepEqual(sheet["custom"], expected["custom"]) || !reflect.DeepEqual(sheet["navi"], expected["navi"]) {
				t.Fatal("unrelated cheatsheet metadata changed", sheet)
			}
			if !reflect.DeepEqual(doc.Sections[0].Cards[0].Classes, []string{"navi"}) {
				t.Fatal("marker handling changed", doc.Sections[0].Cards[0])
			}
			var metadataFinding, markerFinding bool
			for _, finding := range result.Report.Findings {
				if finding.Code != "IMP111" {
					continue
				}
				if strings.Contains(finding.Message, "cheatsheet.cards") {
					metadataFinding = finding.Span.Start.Line == 6 && finding.Span.Start.Column == 3
				}
				if strings.Contains(finding.Message, ".command") {
					markerFinding = finding.Span.Start.Line == 10 && finding.Span.Start.Column > 1
				}
			}
			if !metadataFinding || !markerFinding {
				t.Fatal("missing original-position command findings", result.Report.Findings)
			}
			attrs := original["removed_attributes"].([]any)
			if len(attrs) != 1 || attrs[0].(map[string]any)["text"] != ".command" {
				t.Fatal("original marker not preserved", attrs)
			}
		})
	}
}

func TestUpstreamCommandMarkersAreInertAndLiteralCodeUntouched(t *testing.T) {
	src := []byte("---\ntitle: T\n---\n## Commands\n### New {.command}\n```text\n### Literal {.command}\n{.command}\n```\n### Old {#old .navi}\n```bash\nprintf old\n```\n")
	result, ds := Convert(src, provenance())
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	doc, ds := markdown.Parse(result.Markdown)
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	cards := doc.Sections[0].Cards
	if len(cards[0].Classes) != 0 || !reflect.DeepEqual(cards[1].Classes, []string{"navi"}) {
		t.Fatal(cards)
	}
	if !strings.Contains(string(result.Markdown), "### Literal {.command}\n{.command}") {
		t.Fatal("literal code changed", string(result.Markdown))
	}
	if len(result.Report.Findings) != 1 || result.Report.Findings[0].Code != "IMP111" || result.Report.Findings[0].Span.Start.Line != 5 {
		t.Fatal(result.Report.Findings)
	}
}

func TestImportKeepsUpstreamVariableSourcesInert(t *testing.T) {
	src := []byte("---\ntitle: T\ncheatsheet:\n  tool: smill\n  cards:\n    run:\n      params:\n        file:\n          type: string\n          choices: [a, b]\n          source: ls\n          options: {multi: true}\n---\n## Commands\n### Example {#run .navi}\n```bash\nprintf '%s' literal\n```\n")
	decoded, ds := frontmatter.Decode(src)
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	result, ds := Convert(src, provenance())
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	doc, ds := markdown.Parse(result.Markdown)
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	if _, active := doc.Metadata["cheatsheet"].(map[string]any)["cards"]; active {
		t.Fatal("upstream variable sources became active")
	}
	archived := doc.Metadata["source"].(map[string]any)["original_metadata"].(map[string]any)["cheatsheet"].(map[string]any)["cards"]
	if !reflect.DeepEqual(archived, decoded.Metadata["cheatsheet"].(map[string]any)["cards"]) {
		t.Fatal("variable sources not archived exactly", archived)
	}
	var found bool
	for _, finding := range result.Report.Findings {
		if finding.Code == "IMP111" && strings.Contains(finding.Message, "cheatsheet.cards") && finding.Span.Start.Line == 5 {
			found = true
		}
	}
	if !found {
		t.Fatal("missing IMP111", result.Report.Findings)
	}
}
