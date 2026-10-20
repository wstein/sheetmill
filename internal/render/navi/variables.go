// SPDX-License-Identifier: EUPL-1.2
package navi

import (
	"strconv"
	"strings"
)

// quoteChoices builds the suggestion command for a fixed list of literals, in
// authored order, with the quoting of the selected profile. The caller has
// already rejected controls and the text navi's line grammar cannot carry.
func quoteChoices(profile Profile, choices []string) string {
	quoted := make([]string, len(choices))
	for i, c := range choices {
		switch profile.Name {
		case "nu":
			quoted[i] = `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(c) + `"`
		case "fish":
			quoted[i] = `'` + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(c) + `'`
		default:
			quoted[i] = `'` + strings.ReplaceAll(c, `'`, `'\''`) + `'`
		}
	}
	if profile.Name == "nu" {
		return "[ " + strings.Join(quoted, " ") + " ] | to text"
	}
	return `printf '%s\n' ` + strings.Join(quoted, " ")
}

// options are the closed navi argument options of one variable. The booleans
// hasColumn and hasHeaderLines tell an absent integer from a written zero.
type options struct {
	Column, HeaderLines int
	Delimiter           string
	Multi, PreventExtra bool

	hasColumn, hasHeaderLines bool
}

// variableFields reads already validated choices, source and options.
func variableFields(fields map[string]any) (choices []string, source string, o options) {
	switch v := fields["choices"].(type) {
	case []any:
		for _, item := range v {
			choices = append(choices, item.(string))
		}
	case []string:
		choices = append(choices, v...)
	}
	source, _ = fields["source"].(string)
	raw, _ := fields["options"].(map[string]any)
	if n, ok := raw["column"].(int64); ok {
		o.Column, o.hasColumn = int(n), true
	}
	if n, ok := raw["header_lines"].(int64); ok {
		o.HeaderLines, o.hasHeaderLines = int(n), true
	}
	o.Delimiter, _ = raw["delimiter"].(string)
	o.Multi, _ = raw["multi"].(bool)
	o.PreventExtra, _ = raw["prevent_extra"].(bool)
	return choices, source, o
}

// words lists the option words in the fixed order of the data model.
func (o options) words() []string {
	var w []string
	if o.hasColumn {
		w = append(w, "--column", strconv.Itoa(o.Column))
	}
	if o.Delimiter != "" {
		w = append(w, "--delimiter", "'"+o.Delimiter+"'")
	}
	if o.hasHeaderLines {
		w = append(w, "--header-lines", strconv.Itoa(o.HeaderLines))
	}
	if o.Multi {
		w = append(w, "--multi")
	}
	if o.PreventExtra {
		w = append(w, "--prevent-extra")
	}
	return w
}

// variableLine writes one navi variable line. It does not check the text; the
// caller rejects what the line grammar cannot carry.
func variableLine(name, command string, o options) string {
	line := "$ " + name + ": " + command
	if w := o.words(); len(w) > 0 {
		line += " --- " + strings.Join(w, " ")
	}
	return line
}
