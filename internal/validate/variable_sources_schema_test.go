// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/wstein/sheetmill/internal/validate/varcases"
)

// accepts evaluates the small JSON Schema subset the parameter declaration uses.
// An unknown keyword fails the test, so the schema cannot grow past the evaluator
// unnoticed.
func accepts(t *testing.T, schema map[string]any, v any) bool {
	t.Helper()
	if strs, ok := v.([]string); ok {
		list := make([]any, len(strs))
		for i, s := range strs {
			list[i] = s
		}
		v = list
	}
	number := func(key string) (float64, bool) { n, ok := schema[key].(float64); return n, ok }
	for _, key := range commandKeys(schema) {
		switch key {
		case "type":
			if !hasJSONType(v, schema[key].(string)) {
				return false
			}
		case "enum":
			found := false
			for _, e := range schema[key].([]any) {
				found = found || e == v
			}
			if !found {
				return false
			}
		case "required":
			m, isMap := v.(map[string]any)
			for _, name := range schema[key].([]any) {
				if _, has := m[name.(string)]; isMap && !has {
					return false
				}
			}
		case "properties", "additionalProperties":
			m, isMap := v.(map[string]any)
			if !isMap {
				continue
			}
			props, _ := schema["properties"].(map[string]any)
			for _, name := range commandKeys(m) {
				sub, known := props[name].(map[string]any)
				if !known && schema["additionalProperties"] == false {
					return false
				}
				if known && !accepts(t, sub, m[name]) {
					return false
				}
			}
		case "items":
			if list, isList := v.([]any); isList {
				for _, item := range list {
					if !accepts(t, schema[key].(map[string]any), item) {
						return false
					}
				}
			}
		case "uniqueItems":
			seen := map[any]bool{}
			if list, isList := v.([]any); isList && schema[key] == true {
				for _, item := range list {
					if seen[item] {
						return false
					}
					seen[item] = true
				}
			}
		case "minItems", "maxItems":
			if list, isList := v.([]any); isList {
				n, _ := number(key)
				if key == "minItems" && float64(len(list)) < n || key == "maxItems" && float64(len(list)) > n {
					return false
				}
			}
		case "minLength", "maxLength":
			if s, isString := v.(string); isString {
				n, _ := number(key)
				count := float64(utf8.RuneCountInString(s))
				if key == "minLength" && count < n || key == "maxLength" && count > n {
					return false
				}
			}
		case "minimum", "maximum":
			if f, isNumber := asFloat(v); isNumber {
				n, _ := number(key)
				if key == "minimum" && f < n || key == "maximum" && f > n {
					return false
				}
			}
		case "pattern":
			if s, isString := v.(string); isString && !re2(t, schema[key].(string)).MatchString(s) {
				return false
			}
		case "not":
			if accepts(t, schema[key].(map[string]any), v) {
				return false
			}
		case "$comment", "description":
		default:
			t.Fatalf("schema keyword %q is not supported by the test evaluator", key)
		}
	}
	return true
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

func hasJSONType(v any, name string) bool {
	switch name {
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "integer":
		f, ok := asFloat(v)
		return ok && f == math.Trunc(f)
	}
	return false
}

func paramSchema(t *testing.T) map[string]any {
	t.Helper()
	b, err := os.ReadFile("../../schema/cheatsheet.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var node any
	if err := json.Unmarshal(b, &node); err != nil {
		t.Fatal(err)
	}
	for _, step := range []string{"properties", "cheatsheet", "properties", "cards", "additionalProperties", "properties", "params", "additionalProperties"} {
		m, ok := node.(map[string]any)
		if !ok {
			t.Fatalf("schema path breaks at %q", step)
		}
		if node, ok = m[step]; !ok {
			t.Fatalf("schema lacks %q on the way to the parameter declaration", step)
		}
	}
	return node.(map[string]any)
}

func TestSchemaMatchesVariableSourceRules(t *testing.T) {
	schema := paramSchema(t)
	comment, _ := schema["$comment"].(string)
	for _, table := range []struct {
		name  string
		cases []varcases.Case
	}{{"choices", varcases.Choices}, {"int choices", varcases.IntChoices}, {"source", varcases.Sources}, {"exclusive", varcases.Exclusive}, {"options", varcases.Options}} {
		for _, c := range table.cases {
			t.Run(table.name+"/"+c.Name, func(t *testing.T) {
				got := accepts(t, schema, c.Fields)
				switch {
				case got == c.Valid:
				case got && c.Looser != "":
					// The schema is looser only where its $comment says so.
					if !strings.Contains(comment, c.Looser) {
						t.Errorf("the schema accepts %v and must name %q in its $comment", c.Fields, c.Looser)
					}
				default:
					t.Errorf("schema accepts=%v, contract valid=%v for %v", got, c.Valid, c.Fields)
				}
			})
		}
	}
	for _, rule := range []string{varcases.RuleCharacters, varcases.RuleIntChoices, varcases.RuleDependency, varcases.RuleBothFlags, varcases.RuleFloats} {
		if !strings.Contains(comment, rule) {
			t.Errorf("the parameter $comment must document the %q rule", rule)
		}
	}
}
