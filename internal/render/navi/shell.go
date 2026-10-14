// SPDX-License-Identifier: EUPL-1.2
package navi

import (
	"fmt"
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// quoteProblem uses only the syntax package: no interpreter, shell process,
// environment expansion or filesystem access. The pinned parser understands
// case-pattern parentheses and nested substitutions instead of guessing their
// quote context. This profile reports quote errors only, not full shell validity.
func quoteProblem(command string, profile Profile) string {
	if profile.Name == "bash" || profile.Name == "zsh" {
		variant := syntax.LangBash
		if profile.Name == "zsh" {
			variant = syntax.LangZsh
		}
		// Navi placeholders are templates, not shell redirections. Validation-only
		// sentinels avoid syntax errors from e.g. unquoted <path>; emitted bytes never
		// use this copy. Placeholder declarations/escapes are checked separately.
		masked := maskPlaceholders(command)
		_, err := syntax.NewParser(syntax.Variant(variant)).Parse(strings.NewReader(masked), "")
		// v3.14.1 quoteErr emits this distinct diagnostic. Other parser errors are
		// deliberately not converted into quote failures or a syntax-validity claim.
		if parseErr, ok := err.(syntax.ParseError); ok && strings.Contains(parseErr.Text, "without closing quote") {
			return "the command has an unterminated quote: " + parseErr.Text
		}
		return ""
	}
	// Fish remains an experimental heuristic; Nushell performs no quote check.
	if profile.quotes && !strings.Contains(command, "<<") {
		if open := unterminatedQuote(command, profile); open != 0 {
			return fmt.Sprintf("the command has an unterminated %c quote", open)
		}
	}
	return ""
}

var sentinelPrefixes = regexp.MustCompile(`__sheetmill_input_[0-9]+_`)

// maskPlaceholders preserves identity between repeated tokens without merging
// different heredoc delimiters or matching literal source text. Collect occupied
// prefixes in one scan so choosing an unused prefix stays bounded by input size.
func maskPlaceholders(command string) string {
	occupied := map[string]bool{}
	for _, prefix := range sentinelPrefixes.FindAllString(command, -1) {
		occupied[prefix] = true
	}
	prefix := ""
	for index := 0; ; index++ {
		prefix = fmt.Sprintf("__sheetmill_input_%d_", index)
		if !occupied[prefix] {
			break
		}
	}
	tokens := map[string]string{}
	return placeholder.ReplaceAllStringFunc(command, func(token string) string {
		if sentinel, ok := tokens[token]; ok {
			return sentinel
		}
		sentinel := fmt.Sprintf("%s%d__", prefix, len(tokens))
		tokens[token] = sentinel
		return sentinel
	})
}
