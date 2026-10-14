// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"reflect"
	"strings"
	"testing"
)

func TestClassicCompactTheme(t *testing.T) {
	a := sheet(t, "a", "## S\n### C\nText.\n")
	for _, theme := range []string{"", "editorial"} {
		p, ds := Render([]Sheet{a}, Options{Theme: theme})
		if len(ds) > 0 {
			t.Fatal(ds)
		}
		for name := range p.Files {
			if strings.Contains(name, "classic") {
				t.Errorf("theme %q ships %s", theme, name)
			}
		}
		if strings.Contains(string(p.Files["hugo.json"]), "classic") {
			t.Errorf("theme %q enables the classic layout", theme)
		}
	}
	p, ds := Render([]Sheet{a}, Options{Theme: "classic-compact"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var css string
	for name := range p.Files {
		if strings.HasPrefix(name, "static/assets/classic.") && strings.HasSuffix(name, ".css") {
			css = name
		}
	}
	if css == "" || !strings.Contains(string(p.Files["hugo.json"]), `"classicstylesheet"`) {
		t.Fatal("classic stylesheet is missing")
	}
	if strings.Contains(string(p.Files[css]), "@import") || strings.Contains(string(p.Files[css]), "url(") {
		t.Error("classic stylesheet must stay local and self-contained")
	}
	q, _ := Render([]Sheet{a}, Options{Theme: "classic-compact"})
	if !reflect.DeepEqual(p.Files, q.Files) {
		t.Error("classic output is not deterministic")
	}
	if _, ds := Render([]Sheet{a}, Options{Theme: "fancy"}); len(ds) == 0 {
		t.Error("unknown theme was accepted")
	}
}
