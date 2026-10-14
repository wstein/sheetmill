// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTagPages(t *testing.T) {
	a, b := sheet(t, "a", "Text."), sheet(t, "b", "Text.")
	a.Document.Metadata["tags"] = []any{"cli", "cli", "C++", "C#", "CLI", "日本語", `<unsafe>/../x`}
	b.Document.Metadata["tags"] = []any{"cli"}
	p, ds := Render([]Sheet{b, a}, Options{BasePath: "/sheetmill/", SiteURL: "https://example.com/sheetmill/"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	var tags []struct {
		ID, Title, URL string
		Sheets         []struct{ URL string }
	}
	if err := json.Unmarshal(p.Files["data/tagindex.json"], &tags); err != nil {
		t.Fatal(err)
	}
	if len(tags) != 6 {
		t.Fatalf("%+v", tags)
	}
	seen := map[string]bool{}
	for _, tag := range tags {
		if seen[tag.URL] || !strings.HasPrefix(tag.URL, "/sheetmill/tags/") || strings.Contains(tag.ID, "/") {
			t.Fatalf("unsafe/colliding route: %+v", tag)
		}
		seen[tag.URL] = true
		if len(p.Files["content/tags/"+tag.ID+"/index.md"]) == 0 || len(p.Files["data/tags/"+tag.ID+".json"]) == 0 {
			t.Fatal("missing tag page", tag)
		}
		count := 1
		if tag.Title == "cli" {
			count = 2
		}
		if len(tag.Sheets) != count {
			t.Fatalf("tag membership duplicated/lost: %+v", tag)
		}
		if !strings.Contains(string(p.Files["static/sitemap.xml"]), "https://example.com"+tag.URL) {
			t.Fatal("tag absent from sitemap", tag.URL)
		}
	}

	if !strings.Contains(string(p.Files["data/sheets/a.json"]), `"tag_links"`) {
		t.Fatal("missing clickable tag data")
	}
	q, ds := Render([]Sheet{a, b}, Options{BasePath: "/sheetmill/", SiteURL: "https://example.com/sheetmill/"})
	if len(ds) > 0 || !reflect.DeepEqual(p.Files, q.Files) {
		t.Fatal("tag pages depend on input order", ds)
	}
}

func TestTagsRouteReserved(t *testing.T) {
	p, ds := Render([]Sheet{sheet(t, "tags", "Text.")}, Options{})
	if len(ds) != 1 || ds[0].Code != "REN003" || len(p.Files) > 0 {
		t.Fatal("tags route must not overwrite tag index", ds)
	}
}
