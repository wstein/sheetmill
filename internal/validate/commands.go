// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/model"
)

var commandID = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)
var commandParam = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

func commandKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func commandSpan(locations map[string]diagnostic.Span, key string) diagnostic.Span {
	for {
		if span, ok := locations[key]; ok && span.Start.Line > 0 {
			return span
		}
		i := strings.LastIndexByte(key, '.')
		if i < 0 {
			break
		}
		key = key[:i]
	}
	if span, ok := locations[""]; ok && span.Start.Line > 0 {
		return span
	}
	return diagnostic.Span{Start: diagnostic.Position{Line: 1, Column: 1}, End: diagnostic.Position{Line: 1, Column: 1}}
}

// CommandMetadata validates shared declarations without mutating authored metadata.
// Legacy cheatsheet.navi remains the exporter's compatibility contract.
func CommandMetadata(m map[string]any, locations map[string]diagnostic.Span) []diagnostic.Diagnostic {
	cs, _ := m["cheatsheet"].(map[string]any)
	raw, present := cs["cards"]
	if !present {
		return nil
	}
	var ds []diagnostic.Diagnostic
	add := func(key, msg string) {
		ds = append(ds, diagnostic.Diagnostic{Code: "CMD001", Message: msg, Span: commandSpan(locations, key)})
	}
	cards, ok := raw.(map[string]any)
	if !ok {
		add("cheatsheet.cards", "cheatsheet.cards must be a mapping")
		return ds
	}
	for _, id := range commandKeys(cards) {
		where := "cheatsheet.cards." + id
		if !commandID.MatchString(id) {
			add(where, "shared card keys must be explicit ASCII card IDs")
		}
		card, ok := cards[id].(map[string]any)
		if !ok {
			add(where, where+" must be a mapping")
			continue
		}
		for _, key := range commandKeys(card) {
			if key != "params" {
				add(where+"."+key, where+": unknown field "+key)
			}
		}
		raw, has := card["params"]
		if !has {
			continue
		}
		params, ok := raw.(map[string]any)
		if !ok {
			add(where+".params", where+".params must be a mapping")
			continue
		}
		for _, name := range commandKeys(params) {
			p := where + ".params." + name
			if !commandParam.MatchString(name) {
				add(p, "parameter names match [a-z][a-z0-9_]{0,31}")
			}
			fields, ok := params[name].(map[string]any)
			if !ok {
				add(p, p+" must be a mapping")
				continue
			}
			if _, has := fields["type"]; !has {
				add(p, p+".type is required")
			}
			for _, key := range commandKeys(fields) {
				v, isString := fields[key].(string)
				switch key {
				case "type":
					if !isString || (v != "string" && v != "int" && v != "path") {
						add(p+".type", p+".type must be string, int or path")
					}
				case "description":
					if !isString || !commandDescription(v) {
						add(p+".description", p+".description must be a nonempty single-line string without surrounding whitespace or invisible controls")
					}
				default:
					add(p+"."+key, p+": unknown field "+key)
				}
			}
		}
	}
	return ds
}

func commandDescription(s string) bool {
	if s == "" || strings.TrimSpace(s) != s {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == '\u2028' || r == '\u2029' {
			return false
		}
	}
	return true
}

// IsCommand recognizes the semantic marker and its legacy selection alias.
func IsCommand(card model.Card) bool {
	return slices.Contains(card.Classes, "command") || slices.Contains(card.Classes, "navi")
}

// Commands protects parser and direct-IR consumers with identical shared rules.
// Legacy-only cards keep their existing exporter-owned missing-ID diagnostics.
func Commands(d *model.Document, locations map[string]diagnostic.Span) []diagnostic.Diagnostic {
	if d == nil {
		return nil
	}
	ds := CommandMetadata(d.Metadata, locations)
	if len(ds) > 0 {
		return ds
	}
	cards := map[string][]model.Card{}
	for _, section := range d.Sections {
		for _, card := range section.Cards {
			cards[card.ID] = append(cards[card.ID], card)
			if slices.Contains(card.Classes, "command") && (!card.IDExplicit || !commandID.MatchString(card.ID)) {
				span := card.Span
				if span.Start.Line == 0 {
					span = commandSpan(locations, "")
				}
				ds = append(ds, diagnostic.Diagnostic{Code: "CMD002", Message: ".command cards require an explicit ASCII card ID", Span: span})
			}
		}
	}
	cs, _ := d.Metadata["cheatsheet"].(map[string]any)
	shared, _ := cs["cards"].(map[string]any)
	for _, id := range commandKeys(shared) {
		matches := cards[id]
		if len(matches) != 1 || !matches[0].IDExplicit || !IsCommand(matches[0]) {
			ds = append(ds, diagnostic.Diagnostic{Code: "CMD002", Message: "shared declaration must reference one explicitly identified command card: " + id, Span: commandSpan(locations, "cheatsheet.cards."+id)})
		}
	}
	return ds
}
