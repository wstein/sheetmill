// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"regexp"
	"strings"
	"testing"
)

func stylesheet(t *testing.T, p Project, stem string) string {
	t.Helper()
	for name, data := range p.Files {
		if strings.HasPrefix(name, "static/assets/"+stem+".") && strings.HasSuffix(name, ".css") {
			return string(data)
		}
	}
	t.Fatalf("no %s stylesheet", stem)
	return ""
}

// ruleBody returns the declarations of the first rule whose selector list is
// exactly selector, or fails.
func ruleBody(t *testing.T, css, selector string) string {
	t.Helper()
	re := regexp.MustCompile(`(?:^|[}\n])` + regexp.QuoteMeta(selector) + `\{([^}]*)\}`)
	m := re.FindStringSubmatch(css)
	if m == nil {
		t.Fatalf("no rule %q", selector)
	}
	return m[1]
}

// The CSS-level assertions of FR-008; the rendered 320px wrap needs a browser.
func TestFreshnessStylesheetContract(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{Theme: "reference-cards"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	site, classic, cards := stylesheet(t, p, "site"), stylesheet(t, p, "classic"), stylesheet(t, p, "cards")
	rules := []struct{ css, selector string }{
		{site, ".library-item .library-fresh"},
		{site, ".page-heading .sheet-fresh"},
		{classic, ".classic .sheet-fresh"},
		{classic, ".classic .index-fresh"},
	}
	token := regexp.MustCompile(`#[0-9A-Fa-f]{3,8}|rgba?\(|hsla?\(`)
	for _, r := range rules {
		body := ruleBody(t, r.css, r.selector)
		if !strings.Contains(body, "color:var(--muted)") {
			t.Errorf("%s must use the muted palette token: %s", r.selector, body)
		}
		if token.MatchString(body) {
			t.Errorf("%s uses a literal colour instead of palette tokens: %s", r.selector, body)
		}
		for _, bad := range []string{"white-space:nowrap", "display:none", "visibility:hidden", "width:", "content:", "overflow:hidden", "text-overflow"} {
			if strings.Contains(body, bad) {
				t.Errorf("%s must not use %s: %s", r.selector, bad, body)
			}
		}
	}
	if !strings.Contains(classic, ".classic .index-group li{flex-wrap:wrap}") && !strings.Contains(ruleBody(t, classic, ".classic .index-group li"), "flex-wrap:wrap") {
		t.Error("classic index rows must be able to wrap")
	}
	if !strings.Contains(ruleBody(t, classic, ".classic .index-fresh"), "flex-basis:100%") {
		t.Error("classic index-fresh takes its own full-width line")
	}
	if !strings.Contains(ruleBody(t, cards, ".reference-cards .index-fresh"), "flex-basis:auto") {
		t.Error("cards.css must reset the basis: its rows are columns")
	}
	if body := ruleBody(t, cards, ".reference-cards .sheet-fresh"); !strings.Contains(body, "text-align:center") || !strings.Contains(body, "margin-inline:auto") {
		t.Errorf("cards.css must centre the sheet facts like the description: %s", body)
	}
	// Facts print: no print rule may hide or restyle them.
	for name, css := range map[string]string{"site.css": site, "classic.css": classic, "cards.css": cards} {
		for _, block := range printBlocks(css) {
			if strings.Contains(block, "fresh") {
				t.Errorf("%s: a print rule mentions the freshness facts", name)
			}
		}
	}
}

// printBlocks returns the body of every @media print{...} block, matching braces.
func printBlocks(css string) []string {
	var out []string
	const open = "@media print{"
	for from := 0; ; {
		i := strings.Index(css[from:], open)
		if i < 0 {
			return out
		}
		start := from + i + len(open)
		depth, end := 1, start
		for ; end < len(css) && depth > 0; end++ {
			switch css[end] {
			case '{':
				depth++
			case '}':
				depth--
			}
		}
		out = append(out, css[start:end])
		from = end
	}
}

// The facts use --muted on the page and card backgrounds; every palette and mode must
// keep that pair at WCAG AA.
func TestFreshnessMutedContrastInEveryPalette(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	site := stylesheet(t, p, "site")
	block := regexp.MustCompile(`(?m)^:root\[data-palette="([a-z-]+)"\](\[data-theme="dark"\])?\{([^}]*)\}`)
	tokens := regexp.MustCompile(`--([a-z]+):(#[0-9A-Fa-f]{6})`)
	found := 0
	for _, m := range block.FindAllStringSubmatch(site, -1) {
		values := map[string]string{}
		for _, kv := range tokens.FindAllStringSubmatch(m[3], -1) {
			values[kv[1]] = kv[2]
		}
		for _, background := range []string{"bg", "surface"} {
			if c := contrast(values["muted"], values[background]); c < 4.5 {
				t.Errorf("%s %s: muted on %s is %.2f:1", m[1], m[2], background, c)
			}
		}
		found++
	}
	if found != 12 {
		t.Fatalf("expected 6 palettes in 2 modes, parsed %d", found)
	}
}
