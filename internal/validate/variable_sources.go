// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Variable source bounds (specs/020-navi-variable-sources).
const (
	maxChoices       = 64
	maxChoicePoints  = 128
	maxOptionInteger = 255
)

var integerChoice = regexp.MustCompile(`^-?[0-9]+$`)

// variableText reports whether s is a nonempty single line without surrounding
// whitespace, controls, format characters, U+2028, U+2029 or U+2800.
func variableText(s string) bool {
	if s == "" || strings.TrimSpace(s) != s {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == '\u2028' || r == '\u2029' || r == '\u2800' {
			return false
		}
	}
	return true
}

const variableTextRule = "must be a nonempty single line without surrounding whitespace, controls, format characters, U+2028, U+2029 or U+2800"

// variableSources checks the optional choices, source and options of one
// parameter declaration. It only reads fields; add reports CMD001 at a nested key.
func variableSources(p string, fields map[string]any, add func(key, msg string)) {
	choices, hasChoices := fields["choices"]
	source, hasSource := fields["source"]
	if hasChoices {
		variableChoices(p, fields["type"] == "int", choices, add)
	}
	if hasSource {
		if s, ok := source.(string); !ok || !variableText(s) {
			add(p+".source", p+".source "+variableTextRule)
		}
	}
	if hasChoices && hasSource {
		add(p+".source", p+": source and choices are mutually exclusive")
	}
	if options, has := fields["options"]; has {
		variableOptions(p, options, hasSource, hasChoices, add)
	}
}

// VariableSourceProblems applies the shared rules to one parameter declaration
// of another namespace (the legacy navi alias) and returns one message per problem.
func VariableSourceProblems(path string, fields map[string]any) []string {
	var problems []string
	variableSources(path, fields, func(_, msg string) { problems = append(problems, msg) })
	return problems
}

func variableChoices(p string, isInt bool, raw any, add func(key, msg string)) {
	var items []any
	switch v := raw.(type) {
	case []any:
		items = v
	case []string:
		for _, s := range v {
			items = append(items, s)
		}
	default:
		add(p+".choices", p+".choices must be a list of strings")
		return
	}
	if len(items) == 0 || len(items) > maxChoices {
		add(p+".choices", p+".choices must hold 1 to 64 items")
		return
	}
	seen := map[string]bool{}
	for _, item := range items {
		s, ok := item.(string)
		switch {
		case !ok:
			add(p+".choices", p+".choices items must be strings; numbers, booleans and null are never coerced")
		case !variableText(s) || utf8.RuneCountInString(s) > maxChoicePoints:
			add(p+".choices", p+".choices items "+variableTextRule+" and have at most 128 code points")
		case isInt && !integerChoice.MatchString(s):
			add(p+".choices", p+".choices items of an int parameter must be decimal integer strings such as \"3\" or \"-1\"")
		case seen[s]:
			add(p+".choices", p+".choices must not repeat an item")
		default:
			seen[s] = true
			continue
		}
		return
	}
}

func variableOptions(p string, raw any, hasSource, hasChoices bool, add func(key, msg string)) {
	opts, ok := raw.(map[string]any)
	if !ok {
		add(p+".options", p+".options must be a mapping")
		return
	}
	if !hasSource && !hasChoices {
		add(p+".options", p+".options requires source or choices")
	}
	integer := func(key string, lo int64) {
		if n, ok := opts[key].(int64); !ok || n < lo || n > maxOptionInteger {
			add(p+".options."+key, p+".options."+key+" must be an integer from "+strconv.FormatInt(lo, 10)+" to 255")
		}
	}
	for _, key := range commandKeys(opts) {
		where := p + ".options." + key
		switch key {
		case "column", "header_lines", "delimiter":
			if !hasSource && hasChoices {
				add(where, where+" requires source")
			}
			switch key {
			case "column":
				integer(key, 1)
			case "header_lines":
				integer(key, 0)
			default:
				if s, ok := opts[key].(string); !ok || !variableText(s) {
					add(where, where+" "+variableTextRule)
				}
			}
		case "multi", "prevent_extra":
			if _, ok := opts[key].(bool); !ok {
				add(where, where+" must be a boolean")
			}
		default:
			add(where, p+".options: unknown option "+key)
		}
	}
	if opts["multi"] == true && opts["prevent_extra"] == true {
		add(p+".options.prevent_extra", p+".options: multi and prevent_extra must not both be true")
	}
}
