// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/validate/freshcases"
)

func freshnessSheet(t *testing.T, id string, cs map[string]any) Sheet {
	t.Helper()
	s := sheet(t, id, "## S\n### C\nText.\n")
	s.Document.Metadata["cheatsheet"] = cs
	return s
}

func TestFreshnessDirectIRValidation(t *testing.T) {
	for _, field := range []struct {
		key   string
		cases []freshcases.Case
	}{{"updated", freshcases.Dates}, {"tool_version", freshcases.Versions}} {
		for _, c := range field.cases {
			t.Run(field.key+"/"+c.Name, func(t *testing.T) {
				p, ds := Render([]Sheet{freshnessSheet(t, "a", map[string]any{field.key: c.Value})}, Options{})
				if c.Valid {
					if len(ds) != 0 || len(p.Files) == 0 {
						t.Fatalf("want accepted, got %+v", ds)
					}
					return
				}
				if len(ds) != 1 || ds[0].Code != "VAL004" || ds[0].Source != "a.md" || len(p.Files) != 0 {
					t.Fatalf("want one VAL004 and an empty project, got %+v with %d files", ds, len(p.Files))
				}
				if ds[0].Span.Start.Line != 1 || ds[0].Span.Start.Column != 1 {
					t.Fatalf("direct IR has no source positions, want document start: %+v", ds[0])
				}
			})
		}
	}
}

func TestFreshnessDirectIRIgnoresNonMappingCheatsheet(t *testing.T) {
	s := sheet(t, "a", "## S\n### C\nText.\n")
	s.Document.Metadata["cheatsheet"] = "2026-10-14"
	if p, ds := Render([]Sheet{s}, Options{}); len(ds) != 0 || len(p.Files) == 0 {
		t.Fatalf("a non-mapping cheatsheet is ignored as for cheatsheet.tool: %+v", ds)
	}
}

// freshLibrary builds three tagged sheets sharing tags so ordering and related
// scoring are exercised; the dates and versions are parameters.
func freshLibrary(t *testing.T, dates [3]string, versions [3]string, order [3]int) []Sheet {
	t.Helper()
	ids := [3]string{"alpha", "beta", "gamma"}
	var out []Sheet
	for _, i := range order {
		cs := map[string]any{}
		if dates[i] != "" {
			cs["updated"] = dates[i]
		}
		if versions[i] != "" {
			cs["tool_version"] = versions[i]
		}
		s := freshnessSheet(t, ids[i], cs)
		s.Document.Metadata["title"] = []string{"Alpha", "Beta", "Gamma"}[i]
		s.Document.Metadata["category"] = "tools"
		s.Document.Metadata["tags"] = []any{"cli", "shared", ids[i]}
		out = append(out, s)
	}
	return out
}

func dataMap(t *testing.T, p Project, name string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(p.Files[name], &m); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return m
}

func dataList(t *testing.T, p Project, name string) []any {
	t.Helper()
	var l []any
	if err := json.Unmarshal(p.Files[name], &l); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return l
}

func TestFreshnessDataCarriesFieldsOnlyWhenSet(t *testing.T) {
	hostile := `<script>alert("x")</script> & 'q'`
	sheets := freshLibrary(t, [3]string{"2026-10-14", "2026-10-15", ""}, [3]string{hostile, "", ""}, [3]int{0, 1, 2})
	p, ds := Render(sheets, Options{})
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	alpha := dataMap(t, p, "data/sheets/alpha.json")
	if alpha["updated"] != "2026-10-14" || alpha["tool_version"] != hostile {
		t.Fatalf("values must be carried verbatim and unescaped: %v", alpha)
	}
	beta := dataMap(t, p, "data/sheets/beta.json")
	if _, ok := beta["tool_version"]; ok || beta["updated"] != "2026-10-15" {
		t.Fatalf("an unset field must be omitted: %v", beta)
	}
	gamma := dataMap(t, p, "data/sheets/gamma.json")
	if _, ok := gamma["updated"]; ok {
		t.Fatalf("a sheet without fields must carry no keys: %v", gamma)
	}
	if _, ok := gamma["tool_version"]; ok {
		t.Fatalf("a sheet without fields must carry no keys: %v", gamma)
	}
	byTitle := func(list []any) map[string]map[string]any {
		out := map[string]map[string]any{}
		for _, e := range list {
			m := e.(map[string]any)
			out[m["title"].(string)] = m
		}
		return out
	}
	lib := byTitle(dataList(t, p, "data/library.json"))
	if lib["Alpha"]["updated"] != "2026-10-14" || lib["Alpha"]["tool_version"] != hostile {
		t.Fatalf("library: %v", lib["Alpha"])
	}
	if _, ok := lib["Gamma"]["updated"]; ok {
		t.Fatalf("library gamma: %v", lib["Gamma"])
	}
	shelves := dataList(t, p, "data/shelves.json")
	found := false
	for _, sh := range shelves {
		for _, m := range sh.(map[string]any)["sheets"].([]any) {
			if e := m.(map[string]any); e["title"] == "Alpha" {
				found = e["updated"] == "2026-10-14" && e["tool_version"] == hostile
			}
		}
	}
	if !found {
		t.Fatalf("shelves must carry the fields: %s", p.Files["data/shelves.json"])
	}
	tags := dataList(t, p, "data/tagindex.json")
	seen := false
	for _, tag := range tags {
		for _, m := range tag.(map[string]any)["sheets"].([]any) {
			if e := m.(map[string]any); e["title"] == "Alpha" {
				seen = e["updated"] == "2026-10-14"
			}
		}
	}
	if !seen {
		t.Fatalf("tag data embeds the page and carries the field: %s", p.Files["data/tagindex.json"])
	}
	for name, data := range p.Files {
		if strings.HasPrefix(name, "data/tags/") && !strings.Contains(string(data), "updated") {
			// Only tags that list Alpha or Beta must carry a date; gamma-only does not.
			if !strings.Contains(name, "gamma") {
				t.Errorf("%s lacks the embedded date", name)
			}
		}
	}
}

// stripFreshness removes the two keys everywhere so the remainder can be compared.
func stripFreshness(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, e := range x {
			if k != "updated" && k != "tool_version" {
				out[k] = stripFreshness(e)
			}
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = stripFreshness(e)
		}
		return out
	}
	return v
}

func TestFreshnessDoesNotAffectOrderingRelatedOrSearch(t *testing.T) {
	plain, ds := Render(freshLibrary(t, [3]string{}, [3]string{}, [3]int{0, 1, 2}), Options{SiteURL: "https://example.com/"})
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	variants := []Project{}
	for _, c := range []struct {
		dates, versions [3]string
		order           [3]int
	}{
		{[3]string{"2026-10-14", "1970-01-01", "9999-12-31"}, [3]string{"3", "2", "1"}, [3]int{0, 1, 2}},
		{[3]string{"9999-12-31", "2026-10-14", "1970-01-01"}, [3]string{"1", "2", "3"}, [3]int{2, 0, 1}},
		{[3]string{"2000-01-01", "2000-01-01", "2000-01-01"}, [3]string{"x", "x", "x"}, [3]int{1, 2, 0}},
	} {
		p, ds := Render(freshLibrary(t, c.dates, c.versions, c.order), Options{SiteURL: "https://example.com/"})
		if len(ds) != 0 {
			t.Fatal(ds)
		}
		variants = append(variants, p)
	}
	for i, p := range variants {
		for _, name := range []string{"static/search.json", "static/sitemap.xml", "content/alpha/index.md", "content/beta/index.md", "content/gamma/index.md"} {
			if !bytes.Equal(p.Files[name], plain.Files[name]) {
				t.Errorf("variant %d: %s depends on freshness or input order", i, name)
			}
		}
		for _, name := range []string{"data/library.json", "data/shelves.json", "data/tagindex.json", "data/sheets/alpha.json", "data/sheets/beta.json", "data/sheets/gamma.json"} {
			var a, b any
			json.Unmarshal(plain.Files[name], &a)
			json.Unmarshal(p.Files[name], &b)
			if !reflect.DeepEqual(stripFreshness(a), stripFreshness(b)) {
				t.Errorf("variant %d: %s differs beyond the freshness keys", i, name)
			}
		}
	}
}

func TestFreshnessStaysOutOfContentFrontMatterAndSitemap(t *testing.T) {
	p, ds := Render(freshLibrary(t, [3]string{"2999-01-01", "2026-10-14", ""}, [3]string{"1.0", "", ""}, [3]int{0, 1, 2}), Options{SiteURL: "https://example.com/"})
	if len(ds) != 0 {
		t.Fatal(ds)
	}
	for _, id := range []string{"alpha", "beta", "gamma"} {
		raw := strings.TrimSuffix(strings.TrimPrefix(string(p.Files["content/"+id+"/index.md"]), "---\n"), "\n---\n")
		var front map[string]any
		if err := json.Unmarshal([]byte(raw), &front); err != nil {
			t.Fatal(err)
		}
		keys := []string{}
		for k := range front {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		if !reflect.DeepEqual(keys, []string{"sheetid", "title", "url"}) {
			t.Errorf("%s front matter keys must stay title, url, sheetid (Hugo would read updated as the page date): %v", id, keys)
		}
	}
	sitemap := string(p.Files["static/sitemap.xml"])
	if !strings.Contains(sitemap, "<loc>") || strings.Contains(sitemap, "lastmod") || strings.Contains(sitemap, "2999") || strings.Contains(sitemap, "2026-10-14") {
		t.Errorf("sitemap must stay <loc>-only: %s", sitemap)
	}
}

// hugoBuildClock is hugoBuild with an explicit Hugo --clock value.
func hugoBuildClock(t *testing.T, binary string, p Project, clock string) string {
	t.Helper()
	root := t.TempDir()
	for name, data := range p.Files {
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(root, "public")
	cmd := exec.Command(binary, "--source", root, "--destination", out, "--clock", clock, "--noBuildLock", "--quiet")
	cmd.Env = append(os.Environ(), "HUGO_CACHEDIR="+filepath.Join(root, "cache"))
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("hugo: %v %s", err, b)
	}
	return out
}

func readTree(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		rel, _ := filepath.Rel(root, path)
		files[filepath.ToSlash(rel)] = b
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestHugoIntegrationFreshnessIgnoresTheClock(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("SKIPPED, NOT PASSED: set SHEETMILL_HUGO to an installed pinned Hugo to compare clock-independent builds")
	}
	dates := [3]string{"2999-01-01", "1970-01-01", "2026-10-14"}
	versions := [3]string{"2.56.0", "", "10.3p1"}
	for _, theme := range []string{"editorial", "classic-compact", "reference-cards"} {
		t.Run(theme, func(t *testing.T) {
			a, ds := Render(freshLibrary(t, dates, versions, [3]int{0, 1, 2}), Options{Theme: theme, SiteURL: "https://example.com/"})
			if len(ds) != 0 {
				t.Fatal(ds)
			}
			b, ds := Render(freshLibrary(t, dates, versions, [3]int{2, 0, 1}), Options{Theme: theme, SiteURL: "https://example.com/"})
			if len(ds) != 0 {
				t.Fatal(ds)
			}
			first := readTree(t, hugoBuildClock(t, binary, a, "2000-01-01T00:00:00Z"))
			second := readTree(t, hugoBuildClock(t, binary, b, "2031-06-01T00:00:00Z"))
			if len(first) == 0 || len(first) != len(second) {
				t.Fatalf("file sets differ: %d vs %d", len(first), len(second))
			}
			for name, data := range first {
				if !bytes.Equal(data, second[name]) {
					t.Errorf("%s differs between clocks and input orders", name)
				}
			}
			if !strings.Contains(string(first["alpha/index.html"]), `<time datetime="2999-01-01">2999-01-01</time>`) {
				t.Error("a date after the build clock must still be shown as authored")
			}
		})
	}
}
