// SPDX-License-Identifier: EUPL-1.2
package validate

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/wstein/sheetmill/internal/diagnostic"
)

// Freshness bounds (specs/017-freshness).
const (
	minFreshnessYear     = 1970
	maxToolVersionPoints = 64
)

// FreshnessSettings are the validated optional sheet freshness strings; an empty
// field is absent. They are the authored text, never parsed into a time for output.
type FreshnessSettings struct {
	Updated     string
	ToolVersion string
}

// Freshness validates the optional cheatsheet.updated and cheatsheet.tool_version.
// It has no clock and never reformats, so a future date is accepted as authored.
// A cheatsheet that is not a mapping is ignored here (VAL003 owns that on the
// source path). Locations may be nil (direct IR); diagnostics then sit at the
// document start. Any error yields empty settings.
func Freshness(m map[string]any, locations map[string]diagnostic.Span) (FreshnessSettings, []diagnostic.Diagnostic) {
	cs, ok := m["cheatsheet"].(map[string]any)
	if !ok {
		return FreshnessSettings{}, nil
	}
	var settings FreshnessSettings
	var ds []diagnostic.Diagnostic
	add := func(msg, key string) {
		span := diagnostic.Span{Start: diagnostic.Position{Line: 1, Column: 1}, End: diagnostic.Position{Line: 1, Column: 1}}
		for _, k := range []string{key, "cheatsheet", ""} {
			if s, ok := locations[k]; ok {
				span = s
				break
			}
		}
		ds = append(ds, diagnostic.Diagnostic{Code: "VAL004", Message: msg, Span: span})
	}
	if v, ok := cs["updated"]; ok {
		s, isString := v.(string)
		if isString && validUpdated(s) {
			settings.Updated = s
		} else if !isString {
			add("cheatsheet.updated must be a quoted YAML string, e.g. \"2026-10-14\"", "cheatsheet.updated")
		} else {
			add("cheatsheet.updated must be a real calendar date written YYYY-MM-DD (years 1970 to 9999)", "cheatsheet.updated")
		}
	}
	if v, ok := cs["tool_version"]; ok {
		s, isString := v.(string)
		if isString && validToolVersion(s) {
			settings.ToolVersion = s
		} else if !isString {
			add("cheatsheet.tool_version must be a quoted YAML string, e.g. \"3.81\": an unquoted 3.81 or 3.10 is a number", "cheatsheet.tool_version")
		} else {
			add("cheatsheet.tool_version must be one line of 1 to 64 characters without surrounding whitespace, control or invisible format characters", "cheatsheet.tool_version")
		}
	}
	if len(ds) > 0 {
		return FreshnessSettings{}, ds
	}
	return settings, nil
}

func validUpdated(s string) bool {
	if len(s) != len("2006-01-02") {
		return false
	}
	for i := 0; i < len(s); i++ {
		if i == 4 || i == 7 {
			if s[i] != '-' {
				return false
			}
		} else if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	t, err := time.Parse("2006-01-02", s)
	return err == nil && t.Year() >= minFreshnessYear
}

// blankFiller reports runes that look blank but are letters or symbols; a version made only of them
// would show an empty fact.
func blankFiller(r rune) bool {
	switch r {
	case 0x115F, 0x1160, 0x3164, 0xFFA0, 0x2800:
		return true
	}
	return false
}

func validToolVersion(s string) bool {
	if !utf8.ValidString(s) || strings.TrimSpace(s) != s {
		return false
	}
	if n := utf8.RuneCountInString(s); n < 1 || n > maxToolVersionPoints {
		return false
	}
	visible := false
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return false
		}
		if !blankFiller(r) {
			visible = true
		}
	}
	return visible
}
