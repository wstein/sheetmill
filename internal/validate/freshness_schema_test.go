// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/wstein/sheetmill/internal/validate/freshcases"
)

// schemaNode evaluates the small JSON Schema subset the freshness properties use:
// type string, minLength, maxLength (in code points), pattern and allOf of patterns.
type schemaNode struct {
	Type      string       `json:"type"`
	MinLength *int         `json:"minLength"`
	MaxLength *int         `json:"maxLength"`
	Pattern   string       `json:"pattern"`
	AllOf     []schemaNode `json:"allOf"`
	Comment   string       `json:"$comment"`
}

var unicodeEscape = regexp.MustCompile(`\\u([0-9A-Fa-f]{4})`)

// re2 translates the ECMA-262 spelling used in the schema (\uXXXX) to RE2 (\x{XXXX}).
func re2(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	re, err := regexp.Compile(unicodeEscape.ReplaceAllString(pattern, `\x{$1}`))
	if err != nil {
		t.Fatalf("schema pattern %q is not RE2-compatible: %v", pattern, err)
	}
	return re
}

func (n schemaNode) accepts(t *testing.T, v any) bool {
	t.Helper()
	if n.Type == "string" {
		if _, ok := v.(string); !ok {
			return false
		}
	}
	s, isString := v.(string)
	if isString {
		if !utf8.ValidString(s) {
			t.Fatal("invalid UTF-8 cannot be written in a JSON document; the caller skips it")
		}
		count := utf8.RuneCountInString(s)
		if n.MinLength != nil && count < *n.MinLength || n.MaxLength != nil && count > *n.MaxLength {
			return false
		}
		if n.Pattern != "" && !re2(t, n.Pattern).MatchString(s) {
			return false
		}
	}
	for _, sub := range n.AllOf {
		if !sub.accepts(t, v) {
			return false
		}
	}
	return true
}

func freshnessSchema(t *testing.T, key string) schemaNode {
	t.Helper()
	b, err := os.ReadFile("../../schema/cheatsheet.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Properties struct {
			Cheatsheet struct {
				Properties map[string]schemaNode `json:"properties"`
			} `json:"cheatsheet"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	node, ok := doc.Properties.Cheatsheet.Properties[key]
	if !ok {
		t.Fatalf("schema lacks cheatsheet.%s", key)
	}
	return node
}

func TestSchemaMatchesFreshnessTables(t *testing.T) {
	// The schema is looser only where its $comment says so: calendar validity of a
	// well-formed date needs arithmetic a regular expression cannot express.
	calendarOnly := map[string]bool{"impossible february 30": true, "non-leap february 29": true, "century non-leap": true, "century non-leap in range": true, "april 31": true}
	for _, field := range []struct {
		key   string
		cases []freshcases.Case
	}{{"updated", freshcases.Dates}, {"tool_version", freshcases.Versions}} {
		node := freshnessSchema(t, field.key)
		if node.Type != "string" {
			t.Fatalf("%s must be a string", field.key)
		}
		for _, c := range field.cases {
			if s, ok := c.Value.(string); ok && !utf8.ValidString(s) {
				continue // JSON strings cannot carry invalid UTF-8
			}
			t.Run(field.key+"/"+c.Name, func(t *testing.T) {
				got := node.accepts(t, c.Value)
				switch {
				case got == c.Valid:
				case got && field.key == "updated" && calendarOnly[c.Name]:
					if !strings.Contains(node.Comment, "calendar") {
						t.Error("the schema is looser for calendar dates and must say so in its $comment")
					}
				default:
					t.Errorf("schema accepts=%v, contract valid=%v for %q", got, c.Valid, c.Value)
				}
			})
		}
	}
	if c := freshnessSchema(t, "tool_version").Comment; !strings.Contains(c, "U+3164") {
		t.Error("tool_version $comment must document the filler and source-only rules")
	}
}
