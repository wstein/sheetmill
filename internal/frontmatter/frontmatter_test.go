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
	for _, s := range []string{"", "---\ntitle: X\n---\n", "---\na: &a [*a]\n---\n"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, b []byte) { Parse(b) })
}
