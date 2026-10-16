// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHeaderLinksRejectHostileValues(t *testing.T) {
	long := strings.Repeat("a", 41)
	for _, tt := range []struct{ name, label, url string }{
		{"empty label", "", "/demo/"},
		{"blank label", "  ", "/demo/"},
		{"padded label", " Demo", "/demo/"},
		{"long label", long, "/demo/"},
		{"newline in label", "De\nmo", "/demo/"},
		{"nul in label", "De\x00mo", "/demo/"},
		{"bidi override in label", "De‮mo", "/demo/"},
		{"empty url", "Demo", ""},
		{"javascript", "Demo", "javascript:alert(1)"},
		{"mixed case javascript", "Demo", "JaVaScRiPt:alert(1)"},
		{"data", "Demo", "data:text/html,x"},
		{"http", "Demo", "http://example.org/"},
		{"mailto", "Demo", "mailto:a@example.org"},
		{"protocol relative", "Demo", "//example.org/"},
		{"backslash host", "Demo", `/\example.org/`},
		{"relative", "Demo", "demo/"},
		{"leading space", "Demo", " /demo/"},
		{"tab in url", "Demo", "/de\tmo/"},
		{"newline in url", "Demo", "/demo/\n"},
		{"nul in url", "Demo", "/demo/\x00"},
		{"space in url", "Demo", "/de mo/"},
		{"quote in url", "Demo", `/demo/"onmouseover=1`},
		{"angle in url", "Demo", "/demo/<x>"},
		{"credentials", "Demo", "https://user:pw@example.org/"},
		{"no host", "Demo", "https:///demo/"},
		{"https relative", "Demo", "https:demo/"},
		{"url too long", "Demo", "/" + strings.Repeat("a", 2048)},
		{"non ascii", "Demo", "/démo/"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{HeaderLinks: []HeaderLink{{tt.label, tt.url}}})
			if len(ds) == 0 || ds[0].Code != "REN003" || !strings.Contains(ds[0].Message, "header link") {
				t.Fatalf("accepted: %v", ds)
			}
		})
	}
	if _, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{HeaderLinks: make([]HeaderLink, 5)}); len(ds) == 0 {
		t.Error("more than four links accepted")
	}
}

func TestHeaderLinksAcceptSafeValues(t *testing.T) {
	links := []HeaderLink{{"Example library", "/sheetmill/demo/"}, {"Source & more", "https://example.org/a/b?x=1&y=2#top"}, {strings.Repeat("é", 40), "/"}}
	p, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{HeaderLinks: links})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	config := string(p.Files["hugo.json"])
	for _, want := range []string{`"headerlinks"`, `"label": "Example library"`, `"url": "/sheetmill/demo/"`} {
		if !strings.Contains(config, want) {
			t.Errorf("hugo.json lacks %s", want)
		}
	}
	plain, _ := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{})
	if strings.Contains(string(plain.Files["hugo.json"]), "headerlinks") {
		t.Error("no links must add no parameter")
	}
}

func TestHugoIntegrationHeaderLinks(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("SKIPPED, NOT PASSED: set SHEETMILL_HUGO to an installed pinned Hugo to render header links")
	}
	for _, theme := range []string{"editorial", "classic-compact", "reference-cards"} {
		t.Run(theme, func(t *testing.T) {
			p, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{Theme: theme, HeaderLinks: []HeaderLink{{"Example <library>", "/sheetmill/demo/"}}})
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			out := hugoBuild(t, binary, p)
			home, _ := os.ReadFile(filepath.Join(out, "index.html"))
			sheetPage, _ := os.ReadFile(filepath.Join(out, "a", "index.html"))
			want := []byte(`href="/sheetmill/demo/">Example &lt;library&gt;</a>`)
			if n := bytes.Count(home, want); n != 2 {
				t.Errorf("home must show the link in the top bar and the hero, got %d", n)
			}
			if n := bytes.Count(sheetPage, want); n != 1 {
				t.Errorf("sheet pages show the link in the top bar only, got %d", n)
			}
			if bytes.Contains(home, []byte("<library>")) {
				t.Error("label was not escaped")
			}
		})
	}
}
