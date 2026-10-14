// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"github.com/wstein/sheetmill/internal/markdown"
	"strings"
	"testing"
)

func TestDirectIRSharedCommandValidation(t *testing.T) {
	s := sheet(t, "commands", "Text.")
	s.Document.Metadata["cheatsheet"] = map[string]any{"cards": map[string]any{"missing": map[string]any{}}}
	p, ds := Render([]Sheet{s}, Options{})
	if len(p.Files) != 0 || len(ds) != 1 || ds[0].Code != "CMD002" || ds[0].Span.Start.Line != 1 {
		t.Fatalf("invalid IR rendered: %+v", ds)
	}
	s.Document.Metadata["cheatsheet"] = map[string]any{"cards": nil}
	p, ds = Render([]Sheet{s}, Options{})
	if len(p.Files) != 0 || len(ds) != 1 || ds[0].Code != "CMD001" {
		t.Fatalf("malformed IR rendered: %+v", ds)
	}
}

func TestSharedCommandRendersLiteralReference(t *testing.T) {
	doc, ds := markdown.Parse([]byte("---\ntitle: Commands\ncheatsheet: {cards: {check: {params: {sheet: {type: path}}}}}\n---\n## Validation\n### Check {#check .command}\n```bash\nsmill check \"<sheet>\"\n```\n"))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	p, issues := Render([]Sheet{{ID: "commands", SourceName: "commands.md", Document: doc}}, Options{})
	if len(issues) > 0 {
		t.Fatal(issues)
	}
	data := string(p.Files["data/sheets/commands.json"])
	if !strings.Contains(data, `smill check`) || strings.Contains(data, `.command`) || !strings.Contains(data, `ref-check`) {
		t.Fatal(data)
	}
}
