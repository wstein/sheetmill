// SPDX-License-Identifier: EUPL-1.2
package navi

import "strings"

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
