// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/markdown"
)

func navSheet(t *testing.T, id, title, category, nav string) Sheet {
	t.Helper()
	front := "title: " + title + "\n"
	if category != "" {
		front += "category: " + category + "\n"
	}
	if nav != "" {
		front += "navigation: " + nav + "\n"
	}
	d, ds := markdown.Parse([]byte("---\n" + front + "---\n## Basics\n### Card {#card}\nText.\n"))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	return Sheet{ID: id, SourceName: id + ".md", Document: d}
}

// shelfLayout renders data/shelves.json (the single source of the sidebar and
// the home page) as group:route,route|group:route.
func shelfLayout(t *testing.T, p Project) string {
	t.Helper()
	var shelves []struct {
		Category string
		Sheets   []struct{ URL string }
	}
	if err := json.Unmarshal(p.Files["data/shelves.json"], &shelves); err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, shelf := range shelves {
		routes := []string{}
		for _, s := range shelf.Sheets {
			routes = append(routes, strings.Trim(s.URL, "/"))
		}
		out = append(out, shelf.Category+":"+strings.Join(routes, ","))
	}
	return strings.Join(out, "|")
}

func navigationFixture(t *testing.T) []Sheet {
	t.Helper()
	var sheets []Sheet
	for _, id := range []string{"cli", "essentials", "zero", "alpha", "beta", "tie-b", "tie-a", "net", "plain"} {
		src, err := os.ReadFile(filepath.Join("../../../testdata/hugo/navigation", id+".md"))
		if err != nil {
			t.Fatal(err)
		}
		doc, ds := markdown.Parse(src)
		if len(ds) > 0 {
			t.Fatal(ds)
		}
		sheets = append(sheets, Sheet{ID: id, SourceName: id + ".md", Document: doc})
	}
	return sheets
}

func TestNavigationPlacement(t *testing.T) {
	p, ds := Render(navigationFixture(t), Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	want := "Networking:net|SheetMill:zero,essentials,cli,alpha,beta|reference:tie-a,tie-b|:plain"
	if got := shelfLayout(t, p); got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}

func TestNavigationInputPermutations(t *testing.T) {
	base, ds := Render(navigationFixture(t), Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	sheets := navigationFixture(t)
	rng := rand.New(rand.NewSource(14))
	for i := 0; i < 50; i++ {
		rng.Shuffle(len(sheets), func(a, b int) { sheets[a], sheets[b] = sheets[b], sheets[a] })
		p, ds := Render(sheets, Options{})
		if len(ds) > 0 || !reflect.DeepEqual(p.Files, base.Files) {
			t.Fatal("output depends on input order", ds)
		}
	}
}

func TestNavigationLeavesOtherOutputAlone(t *testing.T) {
	with, _ := Render([]Sheet{navSheet(t, "a", "A", "kind", "{group: Elsewhere, order: 1}"), navSheet(t, "b", "B", "kind", "")}, Options{})
	without, _ := Render([]Sheet{navSheet(t, "a", "A", "kind", ""), navSheet(t, "b", "B", "kind", "")}, Options{})
	for _, name := range []string{"data/sheets/a.json", "data/sheets/b.json", "static/search.json", "data/tagindex.json", "data/library.json"} {
		if !bytes.Equal(with.Files[name], without.Files[name]) {
			t.Errorf("%s changed because of navigation", name)
		}
	}
	if got := shelfLayout(t, with); got != "Elsewhere:a|kind:b" {
		t.Fatal(got)
	}
}

// With the explicitly authorized Hugo binary, the sidebar and home page of both
// layouts list library members in the same, navigation-driven order.
func TestNavigationHugoIntegration(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("set SHEETMILL_HUGO to explicitly authorize installed tool")
	}
	for _, theme := range []string{"editorial", "classic-compact"} {
		p, ds := Render(navigationFixture(t), Options{Theme: theme})
		if len(ds) > 0 {
			t.Fatal(ds)
		}
		out := hugoBuild(t, binary, p)
		for _, page := range []string{"index.html", "cli/index.html"} {
			html, err := os.ReadFile(filepath.Join(out, page))
			if err != nil {
				t.Fatal(err)
			}
			last := -1
			for _, route := range []string{"/zero/", "/essentials/", "/cli/", "/alpha/", "/beta/"} {
				at := bytes.Index(html, []byte(`<a href="`+route+`"`))
				if at < 0 || at < last {
					t.Fatalf("%s %s: %s out of order or missing", theme, page, route)
				}
				last = at
			}
		}
	}
}

func TestNavigationDirectIRValidation(t *testing.T) {
	for name, nav := range map[string]any{
		"float order": map[string]any{"order": 1.0},
		"unknown":     map[string]any{"title": "x"},
		"scalar":      "x",
		"blank group": map[string]any{"group": " "},
	} {
		s := navSheet(t, "a", "A", "", "")
		s.Document.Metadata["navigation"] = nav
		p, ds := Render([]Sheet{s}, Options{})
		if len(ds) != 1 || ds[0].Code != "VAL003" || len(p.Files) != 0 {
			t.Errorf("%s: %+v", name, ds)
		}
	}
}

func TestNavigationAllFourSheetPermutations(t *testing.T) {
	sheets := navigationFixture(t)[:4]
	baseline, ds := Render(sheets, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	count := 0
	var permute func(int)
	permute = func(i int) {
		if i == len(sheets) {
			count++
			got, ds := Render(sheets, Options{})
			if len(ds) > 0 || !reflect.DeepEqual(got.Files, baseline.Files) {
				t.Fatal("permutation differs", ds)
			}
			return
		}
		for j := i; j < len(sheets); j++ {
			sheets[i], sheets[j] = sheets[j], sheets[i]
			permute(i + 1)
			sheets[i], sheets[j] = sheets[j], sheets[i]
		}
	}
	permute(0)
	if count != 24 {
		t.Fatal(count)
	}
}

func TestNavigationTitleTiesAndNamedFallback(t *testing.T) {
	sheets := []Sheet{
		navSheet(t, "z", "éclair", "", "{group: G}"),
		navSheet(t, "a", "Éclair", "", "{group: G}"),
		navSheet(t, "b", "Éclair", "", "{group: G}"),
		navSheet(t, "named", "Named", "", "{group: Other references}"),
		navSheet(t, "fallback", "Fallback", "", ""),
	}
	p, ds := Render(sheets, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if got := shelfLayout(t, p); got != "G:a,b,z|Other references:named|:fallback" {
		t.Fatal(got)
	}
}

func TestNavigationExactGroupIdentity(t *testing.T) {
	p, ds := Render([]Sheet{
		navSheet(t, "spaced", "Spaced", "", "{group: ' G '}"),
		navSheet(t, "upper", "Upper", "", "{group: G}"),
		navSheet(t, "lower", "Lower", "", "{group: g}"),
	}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	if got := shelfLayout(t, p); got != " G :spaced|G:upper|g:lower" {
		t.Fatal(got)
	}
}
