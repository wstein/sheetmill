// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"slices"
	"strings"

	"github.com/wstein/sheetmill/internal/diagnostic"
)

// MaxNavigationOrder is the inclusive upper bound of navigation.order.
const MaxNavigationOrder = 1000000

// NavigationSettings are runtime values derived from the optional navigation
// metadata; omitted fields stay zero/nil and are never written back to metadata.
type NavigationSettings struct {
	Group string
	Order *int64
}

// Navigation validates the closed, optional navigation mapping. Orders must be
// canonical int64 values: strings, Booleans, nulls and floats are never coerced.
// Locations may be nil (direct IR); diagnostics then sit at the document start.
func Navigation(m map[string]any, locations map[string]diagnostic.Span) (NavigationSettings, []diagnostic.Diagnostic) {
	var settings NavigationSettings
	raw, present := m["navigation"]
	if !present {
		return settings, nil
	}
	var ds []diagnostic.Diagnostic
	add := func(msg string, keys ...string) {
		span := diagnostic.Span{Start: diagnostic.Position{Line: 1, Column: 1}, End: diagnostic.Position{Line: 1, Column: 1}}
		for _, key := range append(keys, "navigation", "") {
			if s, ok := locations[key]; ok {
				span = s
				break
			}
		}
		ds = append(ds, diagnostic.Diagnostic{Code: "VAL003", Message: msg, Span: span})
	}
	nav, ok := raw.(map[string]any)
	if !ok {
		add("navigation must be a mapping")
		return NavigationSettings{}, ds
	}
	keys := make([]string, 0, len(nav))
	for key := range nav {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if key != "group" && key != "order" {
			add("navigation."+key+" is not supported; only group and order are allowed", "navigation."+key)
		}
	}
	if v, ok := nav["group"]; ok {
		if s, ok := v.(string); !ok || strings.TrimSpace(s) == "" {
			add("navigation.group must be a string with non-whitespace content", "navigation.group")
		} else {
			settings.Group = s
		}
	}
	if v, ok := nav["order"]; ok {
		if n, ok := v.(int64); !ok || n < 0 || n > MaxNavigationOrder {
			add("navigation.order must be an integer from 0 to 1000000", "navigation.order")
		} else {
			settings.Order = &n
		}
	}
	if len(ds) > 0 {
		return NavigationSettings{}, ds
	}
	return settings, nil
}
