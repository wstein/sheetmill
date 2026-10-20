// SPDX-License-Identifier: EUPL-1.2

// Package varcases holds the accept/reject tables of the variable-source contract
// (specs/020-navi-variable-sources). It is imported only by tests of the validator
// and the JSON Schema so both share one table.
package varcases

import "strings"

// Case is one parameter declaration and whether the contract accepts it. Looser
// names the rule the JSON Schema leaves to source validation; it is empty when the
// schema must agree with the validator.
type Case struct {
	Name   string
	Fields map[string]any
	Valid  bool
	Looser string
}

// Looser rule names; the schema $comment must mention each one it relies on.
const (
	RuleCharacters = "characters"
	RuleIntChoices = "int choices"
	RuleDependency = "dependencies"
	RuleBothFlags  = "multi and prevent_extra"
	RuleFloats     = "floats"
)

func decl(typ string, kv ...any) map[string]any {
	m := map[string]any{"type": typ}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

func list(s ...string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

// n returns count distinct short choices.
func n(count int) []any {
	out := make([]any, count)
	for i := range out {
		out[i] = "c" + string(rune('A'+i/26)) + string(rune('a'+i%26))
	}
	return out
}

func ok(name string, f map[string]any) Case  { return Case{Name: name, Fields: f, Valid: true} }
func bad(name string, f map[string]any) Case { return Case{Name: name, Fields: f} }
func badIn(rule, name string, f map[string]any) Case {
	return Case{Name: name, Fields: f, Looser: rule}
}

// Choices are the shape cases for the choices field.
var Choices = []Case{
	ok("one choice", decl("string", "choices", list("main"))),
	ok("authored order and punctuation", decl("string", "choices", list("main", "it's", "a\\b", "$HOME", "%s", "-x", "süd", "<x>", "a---b", "`c`"))),
	ok("64 choices", decl("string", "choices", n(64))),
	ok("128 code points", decl("string", "choices", list(strings.Repeat("ü", 128)))),
	ok("path type", decl("path", "choices", list("a"))),
	ok("typed string list", decl("string", "choices", []string{"a", "b"})),
	bad("empty list", decl("string", "choices", []any{})),
	bad("65 choices", decl("string", "choices", n(65))),
	bad("129 code points", decl("string", "choices", list(strings.Repeat("ü", 129)))),
	bad("duplicate", decl("string", "choices", list("a", "b", "a"))),
	bad("number", decl("string", "choices", []any{int64(1), int64(2)})),
	bad("boolean", decl("string", "choices", []any{true})),
	bad("null item", decl("string", "choices", []any{nil})),
	bad("not a list", decl("string", "choices", "a")),
	bad("mapping", decl("string", "choices", map[string]any{"a": "b"})),
	bad("null", decl("string", "choices", nil)),
	bad("empty string", decl("string", "choices", list(""))),
	badIn(RuleCharacters, "leading space", decl("string", "choices", list(" a"))),
	badIn(RuleCharacters, "trailing space", decl("string", "choices", list("a "))),
	badIn(RuleCharacters, "no-break space edge", decl("string", "choices", list("\u00a0a"))),
	badIn(RuleCharacters, "newline", decl("string", "choices", list("a\nb"))),
	badIn(RuleCharacters, "tab", decl("string", "choices", list("a\tb"))),
	badIn(RuleCharacters, "nul", decl("string", "choices", list("a\x00b"))),
	badIn(RuleCharacters, "format character", decl("string", "choices", list("a\u200bb"))),
	badIn(RuleCharacters, "line separator", decl("string", "choices", list("a\u2028b"))),
	badIn(RuleCharacters, "paragraph separator", decl("string", "choices", list("a\u2029b"))),
	badIn(RuleCharacters, "braille blank", decl("string", "choices", list("a\u2800b"))),
}

// IntChoices are the cases for choices of an int parameter.
var IntChoices = []Case{
	ok("integers", decl("int", "choices", list("3", "-1", "0", "007"))),
	ok("string type keeps text", decl("string", "choices", list("x", "1.5"))),
	badIn(RuleIntChoices, "word", decl("int", "choices", list("x"))),
	badIn(RuleIntChoices, "decimal", decl("int", "choices", list("1.5"))),
	badIn(RuleIntChoices, "plus sign", decl("int", "choices", list("+1"))),
	badIn(RuleIntChoices, "fullwidth digit", decl("int", "choices", list("\uff11"))),
	badIn(RuleIntChoices, "bare minus", decl("int", "choices", list("-"))),
}

// Sources are the shape cases for the source field.
var Sources = []Case{
	ok("command", decl("string", "source", "git remote")),
	ok("quotes and format", decl("string", "source", "git branch --format='%(refname:short)'")),
	ok("navi separator is a navi concern", decl("string", "source", "echo a --- b")),
	ok("unicode", decl("string", "source", "echo süd")),
	bad("number", decl("string", "source", int64(5))),
	bad("boolean", decl("string", "source", true)),
	bad("list", decl("string", "source", list("a"))),
	bad("null", decl("string", "source", nil)),
	bad("empty", decl("string", "source", "")),
	badIn(RuleCharacters, "leading space", decl("string", "source", " ls")),
	badIn(RuleCharacters, "trailing space", decl("string", "source", "ls ")),
	badIn(RuleCharacters, "newline", decl("string", "source", "a\nb")),
	badIn(RuleCharacters, "tab", decl("string", "source", "a\tb")),
	badIn(RuleCharacters, "nul", decl("string", "source", "a\x00b")),
	badIn(RuleCharacters, "format character", decl("string", "source", "a\u200bb")),
	badIn(RuleCharacters, "line separator", decl("string", "source", "a\u2028b")),
	badIn(RuleCharacters, "paragraph separator", decl("string", "source", "a\u2029b")),
	badIn(RuleCharacters, "braille blank", decl("string", "source", "a\u2800b")),
}

// Exclusive are the cases for choices together with source.
var Exclusive = []Case{
	ok("choices only", decl("string", "choices", list("a"))),
	ok("source only", decl("string", "source", "ls")),
	ok("neither", decl("string")),
	bad("both", decl("string", "choices", list("a"), "source", "ls")),
}

func opt(kv ...any) map[string]any { return decl("string", "source", "ls", "options", decl0(kv...)) }

func decl0(kv ...any) map[string]any {
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

func optChoices(kv ...any) map[string]any {
	return decl("string", "choices", list("a"), "options", decl0(kv...))
}

// Options are the shape cases for the options mapping.
var Options = []Case{
	ok("column low", opt("column", int64(1))),
	ok("column high", opt("column", int64(255))),
	ok("header_lines zero", opt("header_lines", int64(0))),
	ok("header_lines high", opt("header_lines", int64(255))),
	ok("delimiter", opt("delimiter", "\\s\\s+")),
	ok("multi", opt("multi", true)),
	ok("multi false", opt("multi", false)),
	ok("prevent_extra", opt("prevent_extra", true)),
	ok("multi with prevent_extra false", opt("multi", true, "prevent_extra", false)),
	ok("all keys", opt("column", int64(2), "delimiter", ",", "header_lines", int64(1), "multi", true)),
	ok("multi with choices", optChoices("multi", true)),
	ok("prevent_extra with choices", optChoices("prevent_extra", true)),
	bad("column zero", opt("column", int64(0))),
	bad("column 256", opt("column", int64(256))),
	bad("column string", opt("column", "2")),
	bad("column boolean", opt("column", true)),
	badIn(RuleFloats, "column float", opt("column", 2.0)),
	bad("header_lines negative", opt("header_lines", int64(-1))),
	bad("header_lines 256", opt("header_lines", int64(256))),
	bad("multi string", opt("multi", "true")),
	bad("multi number", opt("multi", int64(1))),
	bad("prevent_extra string", opt("prevent_extra", "yes")),
	bad("delimiter number", opt("delimiter", int64(1))),
	bad("delimiter empty", opt("delimiter", "")),
	badIn(RuleCharacters, "delimiter padded", opt("delimiter", " ,")),
	badIn(RuleCharacters, "delimiter newline", opt("delimiter", "a\nb")),
	bad("unknown key", opt("foo", int64(1))),
	bad("map is out of scope", opt("map", "x")),
	bad("not a mapping", decl("string", "source", "ls", "options", list("multi"))),
	bad("null options", decl("string", "source", "ls", "options", nil)),
	badIn(RuleDependency, "options without source or choices", decl("string", "options", decl0("multi", true))),
	badIn(RuleDependency, "empty options without source or choices", decl("string", "options", decl0())),
	badIn(RuleDependency, "column with choices", optChoices("column", int64(1))),
	badIn(RuleDependency, "delimiter with choices", optChoices("delimiter", ",")),
	badIn(RuleDependency, "header_lines with choices", optChoices("header_lines", int64(1))),
	badIn(RuleBothFlags, "multi with prevent_extra", opt("multi", true, "prevent_extra", true)),
	badIn(RuleBothFlags, "multi with prevent_extra and choices", optChoices("multi", true, "prevent_extra", true)),
}
