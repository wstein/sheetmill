// SPDX-License-Identifier: EUPL-1.2
package frontmatter

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, tt := range []struct{ name, src, code string }{
		{"valid", "---\ntitle: Editor\ntags: [editor, cli]\ncustom: {enabled: true, count: 3}\n---\n## Editing\n", ""},
		{"missing", "title: Editor\n", "FM001"},
		{"unclosed", "---\ntitle: Editor\n", "FM002"},
		{"duplicate", "---\ntitle: One\ntitle: Two\n---\n", "FM004"},
		{"nested duplicate", "---\ntitle: One\ncustom: {a: 1, a: 2}\n---\n", "FM004"},
		{"title required", "---\ntags: [x]\n---\n", "VAL001"},
		{"title type", "---\ntitle: 123\n---\n", "VAL001"},
		{"tags type", "---\ntitle: One\ntags: x\n---\n", "VAL002"},
		{"schema", "---\ntitle: One\ncheatsheet: {schema: 2}\n---\n", "VAL003"},
		{"root anchor", "---\n&a {title: One}\n---\n", "FM005"},
		{"alias", "---\ntitle: One\ncustom: &a [x]\nother: *a\n---\n", "FM005"},
		{"complex key", "---\ntitle: One\ncustom: {1: x}\n---\n", "FM005"},
		{"tagged mapping", "---\ntitle: One\ncustom: !application {a: 1}\n---\n", "FM005"},
		{"tagged sequence", "---\ntitle: One\ncustom: !application [a, b]\n---\n", "FM005"},
		{"tagged set", "---\ntitle: One\ncustom: !!set {a: null}\n---\n", "FM005"},
		{"tagged root mapping", "---\n!application {title: One}\n---\n", "FM005"},
		{"tagged root set", "---\n!!set {title: null}\n---\n", "FM005"},
		{"navigation valid", "---\ntitle: Editor\nnavigation: {group: SheetMill, order: 20}\n---\n## Editing\n", ""},
		{"navigation empty", "---\ntitle: Editor\nnavigation: {}\n---\n## Editing\n", ""},
		{"navigation float order", "---\ntitle: Editor\nnavigation: {order: 1.0}\n---\n", "VAL003"},
		{"navigation string order", "---\ntitle: Editor\nnavigation: {order: \"1\"}\n---\n", "VAL003"},
		{"navigation unknown key", "---\ntitle: Editor\nnavigation: {ordr: 1}\n---\n", "VAL003"},
		{"navigation title deferred", "---\ntitle: Editor\nnavigation: {title: Short}\n---\n", "VAL003"},
		{"navigation scalar", "---\ntitle: Editor\nnavigation: top\n---\n", "VAL003"},
		{"navigation null", "---\ntitle: Editor\nnavigation:\n---\n", "VAL003"},
		{"navigation blank group", "---\ntitle: Editor\nnavigation: {group: \" \"}\n---\n", "VAL003"},
		{"syntax", "---\ntitle: [\n---\n", "FM003"},
		{"utf8", "---\ntitle: \xff\n---\n", "SRC001"},
		{"crlf", "---\r\ntitle: Editor\r\n---\r\n## Editing\r\n", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, ds := Parse([]byte(tt.src))
			if tt.code == "" {
				if len(ds) > 0 {
					t.Fatalf("%v", ds)
				}
				if r.Metadata["title"] != "Editor" {
					t.Fatal(r.Metadata)
				}
				if !strings.HasPrefix(string(r.Body), "## Editing") {
					t.Fatal(string(r.Body))
				}
				return
			}
			if len(ds) == 0 || ds[0].Code != tt.code {
				t.Fatalf("want %s, got %v", tt.code, ds)
			}
			if ds[0].Span.Start.Line < 1 || ds[0].Span.Start.Column < 1 {
				t.Fatal(ds)
			}
		})
	}
}

func TestDuplicatePosition(t *testing.T) {
	_, ds := Parse([]byte("---\ntitle: One\ntitle: Two\n---\n"))
	if len(ds) != 1 || ds[0].Span.Start.Line != 3 || ds[0].Span.Start.Column != 1 {
		t.Fatal(ds)
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"", "---\ntitle: X\n---\n", "---\na: &a [*a]\n---\n",
		"---\ntitle: X\ncheatsheet:\n  updated: 2026-10-14\n  tool_version: \"2.56.0\"\n---\n",
		"---\ntitle: X\ncheatsheet: {updated: 2026-13-45, tool_version: \"\\u200b\"}\n---\n"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, b []byte) { Parse(b) })
}

func TestNavigationExactPositionsAndNoPartialResult(t *testing.T) {
	for _, key := range []string{"ordr", "order", "group"} {
		value := "bad"
		if key == "group" {
			value = "''"
		}
		src := []byte("---\ntitle: T\nnavigation:\n  " + key + ": " + value + "\n---\n")
		result, ds := Parse(src)
		if len(ds) != 1 || ds[0].Code != "VAL003" || ds[0].Span.Start.Line != 4 || ds[0].Span.Start.Column != 3 {
			t.Fatalf("%s: %+v", key, ds)
		}
		if result.Metadata != nil || result.Body != nil {
			t.Fatal("partial result returned")
		}
	}
}

func TestFreshnessYAMLForms(t *testing.T) {
	for _, tt := range []struct {
		name, fields string
		code         string
		line, column int
	}{
		{"unquoted date", "  updated: 2026-10-14\n", "", 0, 0},
		{"quoted date", "  updated: \"2026-10-14\"\n", "", 0, 0},
		{"single quoted date", "  updated: '2026-10-14'\n", "", 0, 0},
		{"both fields", "  tool_version: \"2.56.0\"\n  updated: 2026-10-14\n", "", 0, 0},
		{"timestamp with time", "  updated: 2026-10-14T09:00:00Z\n", "VAL004", 4, 3},
		{"unpadded date", "  updated: 2026-1-5\n", "VAL004", 4, 3},
		{"impossible date", "  updated: 2026-02-30\n", "VAL004", 4, 3},
		{"number", "  updated: 20261014\n", "VAL004", 4, 3},
		{"bool version", "  tool: git\n  tool_version: true\n", "VAL004", 5, 3},
		{"empty version", "  tool_version: ''\n", "VAL004", 4, 3},
		{"newline version", "  tool_version: \"1\\n2\"\n", "VAL004", 4, 3},
		{"zero width version", "  tool_version: \"1.\\u200b0\"\n", "VAL004", 4, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result, ds := Parse([]byte("---\ntitle: T\ncheatsheet:\n" + tt.fields + "---\nBody\n"))
			if tt.code == "" {
				if len(ds) != 0 {
					t.Fatalf("want accepted, got %+v", ds)
				}
				return
			}
			if len(ds) != 1 || ds[0].Code != tt.code || ds[0].Span.Start.Line != tt.line || ds[0].Span.Start.Column != tt.column {
				t.Fatalf("want one %s at %d:%d, got %+v", tt.code, tt.line, tt.column, ds)
			}
			if result.Metadata != nil || result.Body != nil {
				t.Fatal("partial result returned")
			}
		})
	}
}

func TestFreshnessPreservesAuthoredStrings(t *testing.T) {
	result, ds := Parse([]byte("---\ntitle: T\ncheatsheet:\n  updated: 2026-10-14\n  tool_version: 10.3p1\n  owner: docs-team\n---\n"))
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	cs, _ := result.Metadata["cheatsheet"].(map[string]any)
	if cs["updated"] != "2026-10-14" || cs["tool_version"] != "10.3p1" || cs["owner"] != "docs-team" {
		t.Fatalf("authored strings and unknown keys must survive verbatim: %#v", cs)
	}
}
