// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReferenceCardsTheme(t *testing.T) {
	a := sheet(t, "a", "## S\n### C\nText.\n")
	for _, theme := range []string{"", "editorial", "classic-compact"} {
		p, ds := Render([]Sheet{a}, Options{Theme: theme})
		if len(ds) > 0 {
			t.Fatal(ds)
		}
		for name := range p.Files {
			if strings.Contains(name, "cards.") && strings.HasSuffix(name, ".css") {
				t.Errorf("theme %q ships %s", theme, name)
			}
		}
		if strings.Contains(string(p.Files["hugo.json"]), `"cards"`) {
			t.Errorf("theme %q enables reference-cards", theme)
		}
	}
	p, ds := Render([]Sheet{a}, Options{Theme: "reference-cards"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var css string
	for name := range p.Files {
		if strings.HasPrefix(name, "static/assets/cards.") && strings.HasSuffix(name, ".css") {
			css = name
		}
	}
	config := string(p.Files["hugo.json"])
	if css == "" || !strings.Contains(config, `"cardsstylesheet"`) || !strings.Contains(config, `"cards": true`) || !strings.Contains(config, `"classic": true`) {
		t.Fatalf("reference-cards assets/params missing: %q\n%s", css, config)
	}
	for _, forbidden := range []string{"@import", "url(", "http://", "https://"} {
		if strings.Contains(string(p.Files[css]), forbidden) {
			t.Errorf("cards stylesheet must stay local and self-contained: %s", forbidden)
		}
	}
	q, _ := Render([]Sheet{a}, Options{Theme: "reference-cards"})
	if !reflect.DeepEqual(p.Files, q.Files) {
		t.Error("reference-cards output is not deterministic")
	}
	if _, ds := Render([]Sheet{a}, Options{Theme: "reference-card"}); len(ds) != 1 || ds[0].Code != "REN003" {
		t.Error("near-miss theme name was accepted")
	}
	if got := strings.Join(Themes, ","); got != "editorial,classic-compact,reference-cards" {
		t.Errorf("themes: %s", got)
	}
}

func TestReferenceCardsHugoIntegration(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("set SHEETMILL_HUGO to explicitly authorize installed tool")
	}
	p, ds := Render(navigationFixture(t), Options{Theme: "reference-cards"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	out := hugoBuild(t, binary, p)
	for _, page := range []string{"index.html", "cli/index.html"} {
		html, err := os.ReadFile(filepath.Join(out, page))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{`class="classic reference-cards"`, `<details class="browse-panel"`, `<summary>Browse library</summary>`, `href="/assets/cards.`} {
			if !bytes.Contains(html, []byte(want)) {
				t.Errorf("%s lacks %s", page, want)
			}
		}
		if bytes.Contains(html, []byte(`class="library-sidebar"`)) {
			t.Errorf("%s keeps the permanent sidebar", page)
		}
		last := -1
		for _, route := range []string{"/zero/", "/essentials/", "/cli/", "/alpha/", "/beta/"} {
			at := bytes.Index(html, []byte(`<a href="`+route+`"`))
			if at < 0 || at < last {
				t.Errorf("%s: %s missing or out of the 014 order", page, route)
			}
			last = at
		}
	}
	page, _ := os.ReadFile(filepath.Join(out, "cli", "index.html"))
	if !bytes.Contains(page, []byte(`aria-current="page"`)) {
		t.Error("current sheet is not marked in the Browse panel")
	}
}
