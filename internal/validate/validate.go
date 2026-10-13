// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"github.com/wstein/sheetmill/internal/diagnostic"
	"strings"
)

// Metadata validates known fields; unknown JSON-safe fields remain intact.
func Metadata(m map[string]any, locations map[string]diagnostic.Span) []diagnostic.Diagnostic {
	var ds []diagnostic.Diagnostic
	add := func(code, msg, key string) {
		span, ok := locations[key]
		if !ok {
			span = locations[""]
		}
		ds = append(ds, diagnostic.Diagnostic{Code: code, Message: msg, Span: span})
	}
	title, ok := m["title"].(string)
	if !ok || strings.TrimSpace(title) == "" {
		add("VAL001", "title must be a nonempty string", "title")
	}
	for _, key := range []string{"category", "description"} {
		if v, ok := m[key]; ok {
			if _, ok := v.(string); !ok {
				add("VAL002", key+" must be a string", key)
			}
		}
	}
	if v, ok := m["tags"]; ok {
		tags, ok := v.([]any)
		if !ok {
			add("VAL002", "tags must be an array of strings", "tags")
		} else {
			for _, tag := range tags {
				if s, ok := tag.(string); !ok || strings.TrimSpace(s) == "" {
					add("VAL002", "tags must contain nonempty strings", "tags")
					break
				}
			}
		}
	}
	if v, ok := m["cheatsheet"]; ok {
		cs, ok := v.(map[string]any)
		if !ok {
			add("VAL003", "cheatsheet must be a mapping", "cheatsheet")
		} else {
			if v, ok := cs["schema"]; ok && v != int64(1) {
				add("VAL003", "cheatsheet.schema must be integer 1", "cheatsheet.schema")
			}
			if v, ok := cs["tool"]; ok {
				if s, ok := v.(string); !ok || strings.TrimSpace(s) == "" {
					add("VAL003", "cheatsheet.tool must be a nonempty string", "cheatsheet.tool")
				}
			}
		}
	}
	return ds
}
