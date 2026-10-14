// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func luminance(hex string) float64 {
	var c [3]float64
	for i := range c {
		v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		x := float64(v) / 255
		if x <= 0.04045 {
			c[i] = x / 12.92
		} else {
			c[i] = math.Pow((x+0.055)/1.055, 2.4)
		}
	}
	return c[0]*0.2126 + c[1]*0.7152 + c[2]*0.0722
}

func contrast(a, b string) float64 {
	hi, lo := luminance(a), luminance(b)
	if lo > hi {
		hi, lo = lo, hi
	}
	return (hi + 0.05) / (lo + 0.05)
}

// The description rows and badges reuse palette tokens; every palette and mode
// must keep those pairs readable (WCAG AA, 4.5:1).
func TestReferenceCardsPaletteContrast(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{Theme: "reference-cards"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var site string
	for name, data := range p.Files {
		if strings.HasPrefix(name, "static/assets/site.") && strings.HasSuffix(name, ".css") {
			site = string(data)
		}
	}
	block := regexp.MustCompile(`(?m)^:root\[data-palette="([a-z-]+)"\](\[data-theme="dark"\])?\{([^}]*)\}`)
	token := regexp.MustCompile(`--([a-z]+):(#[0-9A-Fa-f]{6})`)
	found := 0
	for _, m := range block.FindAllStringSubmatch(site, -1) {
		values := map[string]string{}
		for _, kv := range token.FindAllStringSubmatch(m[3], -1) {
			values[kv[1]] = kv[2]
		}
		for _, pair := range [][2]string{{"text", "wash"}, {"link", "wash"}, {"bg", "link"}, {"text", "surface"}, {"link", "surface"}, {"muted", "surface"}, {"accent", "surface"}} {
			if c := contrast(values[pair[0]], values[pair[1]]); c < 4.5 {
				t.Errorf("%s %s: %s on %s is %.2f:1", m[1], m[2], pair[0], pair[1], c)
			}
		}
		found++
	}
	if found != 12 {
		t.Fatalf("expected 6 palettes in 2 modes, parsed %d", found)
	}
}

// The stylesheet's contract: prompt is generated content, spans match the grid.
func TestReferenceCardsStylesheetContract(t *testing.T) {
	p, _ := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{Theme: "reference-cards"})
	var css string
	for name, data := range p.Files {
		if strings.HasPrefix(name, "static/assets/cards.") {
			css = string(data)
		}
	}
	for _, want := range []string{
		`::before{content:"$ "`, "user-select:none", ".card>table{display:block",
		"@container(min-width:620px)", "@container(min-width:700px)", "@container(min-width:940px)", "@container(min-width:1060px)",
		".card>p:has(+pre):not(.card-kind){background:var(--wash)}",
		".card>:last-child:not(h3){border-radius:0 0 9px 9px}",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("cards.css lacks %q", want)
		}
	}
	// Print: no tinted rows, a three-column grid and spans that fit it.
	print := css[strings.Index(css, "@media print{"):]
	for _, want := range []string{
		".card>p:has(+pre):not(.card-kind){background:none}",
		"grid-template-columns:repeat(3,1fr)",
		".card.col-span-2{grid-column:span 2}", ".card.col-span-3{grid-column:span 3}",
		".browse-panel{display:none}", "box-shadow:none",
	} {
		if !strings.Contains(print, want) {
			t.Errorf("print rules lack %q", want)
		}
	}
	if strings.Contains(css, "overflow:hidden") || strings.Contains(css, "position:absolute") {
		t.Error("badge must stay in flow and tiles must not clip content")
	}
}
