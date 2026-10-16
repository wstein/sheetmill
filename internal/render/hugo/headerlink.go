// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxHeaderLinks      = 4
	maxHeaderLabelRunes = 40
	maxHeaderURLBytes   = 2048
)

// validateHeaderLinks accepts at most four links whose labels are short, trimmed,
// printable text and whose URLs are either root-relative paths or https URLs
// without credentials. Nothing else (javascript:, data:, http:, protocol-relative
// URLs, whitespace, control or non-ASCII characters) is allowed.
func validateHeaderLinks(links []HeaderLink) error {
	if len(links) > maxHeaderLinks {
		return fmt.Errorf("invalid header link: at most %d links are allowed", maxHeaderLinks)
	}
	seen := map[string]bool{}
	for _, l := range links {
		if err := validateHeaderLink(l); err != nil {
			return fmt.Errorf("invalid header link %q: %w", l.Label, err)
		}
		if seen[l.Label] {
			return fmt.Errorf("invalid header link %q: duplicate label", l.Label)
		}
		seen[l.Label] = true
	}
	return nil
}

func validateHeaderLink(l HeaderLink) error {
	if l.Label == "" || strings.TrimSpace(l.Label) != l.Label {
		return fmt.Errorf("the label must be nonempty without surrounding whitespace")
	}
	if utf8.RuneCountInString(l.Label) > maxHeaderLabelRunes {
		return fmt.Errorf("the label is longer than %d characters", maxHeaderLabelRunes)
	}
	for _, r := range l.Label {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == utf8.RuneError || (unicode.IsSpace(r) && r != ' ') || unicode.In(r, unicode.Zl, unicode.Zp, unicode.Zs) && r != ' ' {
			return fmt.Errorf("the label contains a control, invisible or non-ASCII-space character")
		}
	}
	if l.URL == "" || len(l.URL) > maxHeaderURLBytes {
		return fmt.Errorf("the URL must be 1 to %d bytes", maxHeaderURLBytes)
	}
	for _, r := range l.URL {
		if r <= ' ' || r >= 0x7f || strings.ContainsRune(`"'<>\`+"`", r) {
			return fmt.Errorf("the URL contains a space, control, non-ASCII or unsafe character")
		}
	}
	u, err := url.Parse(l.URL)
	if err != nil {
		return fmt.Errorf("the URL does not parse")
	}
	switch {
	case strings.HasPrefix(l.URL, "/"):
		if strings.HasPrefix(l.URL, "//") || u.Scheme != "" || u.Host != "" {
			return fmt.Errorf("a root-relative URL must start with a single slash")
		}
	case u.Scheme == "https" && u.Host != "" && u.User == nil && strings.HasPrefix(l.URL, "https://"):
	default:
		return fmt.Errorf("the URL must be https or root-relative")
	}
	return nil
}

// headerLinkParams is the template data; keys are lowercase for Hugo.
func headerLinkParams(links []HeaderLink) []map[string]string {
	out := make([]map[string]string, 0, len(links))
	for _, l := range links {
		out = append(out, map[string]string{"label": l.Label, "url": l.URL})
	}
	return out
}
