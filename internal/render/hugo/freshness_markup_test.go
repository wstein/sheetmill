// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/markdown"
)

const (
	bothFacts = `<span>Version 2.56.0</span> <span aria-hidden="true">·</span> <span>Updated <time datetime="2026-10-14">2026-10-14</time></span>`
	onlyVer   = `<span>Version 2.56.0</span>`
	onlyDate  = `<span>Updated <time datetime="2026-10-14">2026-10-14</time></span>`
)

func themeFile(t *testing.T, name string) string {
	t.Helper()
	p, ds := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{Theme: "reference-cards"})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	b, ok := p.Files[name]
	if !ok {
		t.Fatalf("no %s in the project", name)
	}
	return string(b)
}

// The always-run layer: exact template snippets, independent of an installed Hugo.
func TestFreshnessMarkupInTemplates(t *testing.T) {
	partial := themeFile(t, "layouts/partials/freshness.html")
	for _, want := range []string{
		`with .tool_version }}<span>Version {{ . }}</span>{{ end }}`,
		`{{ if and .tool_version .updated }} <span aria-hidden="true">·</span> {{ end }}`,
		`with .updated }}<span>Updated <time datetime="{{ . }}">{{ . }}</time></span>{{ end -}}`,
	} {
		if !strings.Contains(partial, want) {
			t.Errorf("freshness partial lacks %s", want)
		}
	}
	home := themeFile(t, "layouts/home.html")
	for _, want := range []string{
		`{{ with .license }}<span class="index-license">{{ . }}</span>{{ end }}{{ if or .tool_version .updated }}<span class="index-fresh">{{ partial "freshness.html" . }}</span>{{ end }}</li>`,
		`{{ with .tool }}<p class="library-tool">Tool: {{ . }}</p>{{ end }}{{ if or .tool_version .updated }}<p class="library-fresh">{{ partial "freshness.html" . }}</p>{{ end }}{{ with .description }}`,
	} {
		if !strings.Contains(home, want) {
			t.Errorf("home layout lacks %s", want)
		}
	}
	// Row order: link, description, count, license, then the freshness span last.
	row := home[strings.Index(home, `<li><a href="{{ .url }}">`):]
	row = row[:strings.Index(row, "</li>")]
	last := -1
	for _, marker := range []string{`<a href=`, `index-desc`, `index-count`, `index-license`, `index-fresh`} {
		i := strings.Index(row, marker)
		if i <= last {
			t.Errorf("index row order broken at %s", marker)
		}
		last = i
	}
	page := themeFile(t, "layouts/page.html")
	fresh := `{{ if or $sheet.tool_version $sheet.updated }}<p class="sheet-fresh">{{ partial "freshness.html" $sheet }}</p>{{ end }}`
	for _, want := range []string{
		`</div>{{ with $sheet.description }}<p class="sheet-description">{{ . }}</p>{{ end }}` + fresh,
		`{{ with $sheet.description }}<p class="sheet-description">{{ . }}</p>{{ end }}` + fresh + `{{ partial "tag-chips.html" $sheet }}</div>`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page layout lacks %s", want)
		}
	}
	for _, bad := range []string{"<script", "title=", "aria-label", "now", "Now", "dateFormat", ".Format", "time."} {
		if strings.Contains(partial, bad) {
			t.Errorf("freshness partial contains %q", bad)
		}
	}
	for _, name := range []string{"layouts/tag.html", "layouts/tags.html", "layouts/partials/library.html", "layouts/partials/tag-chips.html", "layouts/baseof.html"} {
		if strings.Contains(themeFile(t, name), "fresh") {
			t.Errorf("%s must not show freshness", name)
		}
	}
}

var datetimeAttr = regexp.MustCompile(`datetime="([^"]*)"`)
var isoDate = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func freshnessSheets(t *testing.T) []Sheet {
	t.Helper()
	mk := func(id, title string, cs map[string]any) Sheet {
		s := freshnessSheet(t, id, cs)
		s.Document.Metadata["title"] = title
		s.Document.Metadata["description"] = "About " + title + "."
		s.Document.Metadata["tags"] = []any{"cli"}
		return s
	}
	return []Sheet{
		mk("both", "Both", map[string]any{"tool_version": "2.56.0", "updated": "2026-10-14"}),
		mk("version", "Solo-v", map[string]any{"tool_version": "2.56.0"}),
		mk("date", "Solo-d", map[string]any{"updated": "2026-10-14"}),
		mk("none", "Neither", map[string]any{"tool": "plain"}),
		mk("evil", "Evil", map[string]any{"tool_version": `<script>alert(1)</script> & "q"`, "updated": "2026-10-14"}),
	}
}

func readOut(t *testing.T, out, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(out, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// between returns the text from the first start marker to the next end marker.
func between(t *testing.T, s, start, end string) string {
	t.Helper()
	i := strings.Index(s, start)
	if i < 0 {
		t.Fatalf("no %q", start)
	}
	j := strings.Index(s[i:], end)
	if j < 0 {
		t.Fatalf("no %q after %q", end, start)
	}
	return s[i : i+j]
}

func TestHugoIntegrationFreshnessMarkup(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("SKIPPED, NOT PASSED: set SHEETMILL_HUGO to an installed pinned Hugo to render sheet freshness")
	}
	for _, theme := range []string{"editorial", "classic-compact", "reference-cards"} {
		t.Run(theme, func(t *testing.T) {
			p, ds := Render(freshnessSheets(t), Options{Theme: theme})
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			out := hugoBuild(t, binary, p)
			home := readOut(t, out, "index.html")
			// The sidebar and the tag pages never show freshness.
			sidebar := between(t, home, `<nav class="library-nav"`, `</nav>`)
			if strings.Contains(sidebar, "Version") || strings.Contains(sidebar, "Updated") || strings.Contains(sidebar, "<time") {
				t.Errorf("sidebar shows freshness: %s", sidebar)
			}
			tagFiles, _ := filepath.Glob(filepath.Join(out, "tags", "*", "index.html"))
			tagFiles = append(tagFiles, filepath.Join(out, "tags", "index.html"))
			for _, f := range tagFiles {
				text, _ := os.ReadFile(f)
				if strings.Contains(string(text), "-fresh") || strings.Contains(string(text), "<time") || strings.Contains(string(text), "Version 2.56") {
					t.Errorf("%s shows freshness", f)
				}
			}
			// datetime is only ever a validated digits-and-hyphens value, on every page.
			filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") {
					return err
				}
				text, _ := os.ReadFile(path)
				for _, m := range datetimeAttr.FindAllStringSubmatch(string(text), -1) {
					if !isoDate.MatchString(m[1]) {
						t.Errorf("%s: datetime %q is not YYYY-MM-DD", path, m[1])
					}
				}
				if strings.Contains(string(text), "<script>alert(1)") {
					t.Errorf("%s: hostile text unescaped", path)
				}
				return nil
			})
			escaped := `<span>Version &lt;script&gt;alert(1)&lt;/script&gt; &amp; &#34;q&#34;</span>`
			if !strings.Contains(home, escaped) {
				t.Errorf("hostile version must be escaped text: %s", home)
			}
			facts := map[string]string{"both": bothFacts, "version": onlyVer, "date": onlyDate}
			if theme == "editorial" {
				if !strings.Contains(home, `<p class="library-fresh">`+bothFacts+`</p>`) ||
					!strings.Contains(home, `<p class="library-fresh">`+onlyVer+`</p>`) ||
					!strings.Contains(home, `<p class="library-fresh">`+onlyDate+`</p>`) {
					t.Errorf("home cards lack the expected facts: %s", home)
				}
				card := between(t, home, `<h3>Neither</h3>`, `</a>`)
				if strings.Contains(card, "fresh") || strings.Contains(card, "<time") || strings.Contains(card, "Version") {
					t.Errorf("a sheet without fields must render no wrapper: %s", card)
				}
				if tool := strings.Index(card, "library-tool"); tool < 0 {
					t.Errorf("tool line missing: %s", card)
				}
				both := between(t, home, `<h3>Both</h3>`, `</a>`)
				if strings.Index(both, "library-fresh") < 0 {
					t.Errorf("no freshness on the both card")
				}
			} else {
				index := home[strings.Index(home, `<section class="index-group"`):]
				for _, id := range []string{"both", "version", "date"} {
					row := between(t, index, `<li><a href="/`+id+`/">`, `</li>`)
					want := `<span class="index-fresh">` + facts[id] + `</span>`
					if !strings.HasSuffix(row, want) {
						t.Errorf("%s: index-fresh must be the last child: %s", id, row)
					}
					if !strings.Contains(row, "index-license") || strings.Index(row, "index-license") > strings.Index(row, "index-fresh") {
						t.Errorf("%s: license must precede the freshness span: %s", id, row)
					}
				}
				row := between(t, index, `<li><a href="/none/">`, `</li>`)
				if strings.Contains(row, "fresh") || strings.Contains(row, "<time") {
					t.Errorf("none: %s", row)
				}
			}
			for id, want := range facts {
				pg := readOut(t, out, id+"/index.html")
				fresh := `<p class="sheet-fresh">` + want + `</p>`
				if strings.Count(pg, fresh) != 1 {
					t.Errorf("%s sheet header lacks %s", id, fresh)
				}
				desc := strings.Index(pg, `class="sheet-description"`)
				at := strings.Index(pg, fresh)
				if desc < 0 || at < desc {
					t.Errorf("%s: freshness must follow the description", id)
				}
				if chips := strings.Index(pg, `class="sheet-tags"`); theme == "editorial" && chips < at {
					t.Errorf("%s: editorial freshness precedes the tag chips", id)
				}
				if theme != "editorial" {
					if head := strings.Index(pg, `class="sheet-head"`); !(head < desc && desc < at) {
						t.Errorf("%s: classic order is head, description, freshness", id)
					}
				}
			}
			none := readOut(t, out, "none/index.html")
			if strings.Contains(none, "sheet-fresh") || strings.Contains(none, "<time") {
				t.Error("a sheet without fields must render no freshness")
			}
			evil := readOut(t, out, "evil/index.html")
			if !strings.Contains(evil, `<p class="sheet-fresh"><span>Version &lt;script&gt;`) {
				t.Errorf("evil sheet header: %s", evil)
			}
		})
	}
}

// demoVersions are the documented versions recorded in specs/017-freshness/research.md;
// tar is deliberately absent (a bare bsdtar number would read as GNU tar's).
var demoVersions = map[string]string{"curl": "8.7.1", "git": "2.56.0", "jq": "1.8.2", "make": "3.81", "grep": "2.6.0-FreeBSD", "ssh": "10.3p1"}

func TestHugoIntegrationDemoSheetsShowTheirFreshness(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("SKIPPED, NOT PASSED: set SHEETMILL_HUGO to an installed pinned Hugo to render the demo sheets")
	}
	files, err := filepath.Glob(filepath.Join("..", "..", "..", "examples", "sheets", "*.md"))
	if err != nil || len(files) != 15 {
		t.Fatalf("want the 15 demo sheets, got %d %v", len(files), err)
	}
	var sheets []Sheet
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		d, ds := markdown.Parse(b)
		if len(ds) > 0 {
			t.Fatal(f, ds)
		}
		name := filepath.Base(f)
		sheets = append(sheets, Sheet{ID: strings.TrimSuffix(name, ".md"), SourceName: name, Document: d})
	}
	date := regexp.MustCompile(`^20[0-9]{2}-[0-9]{2}-[0-9]{2}$`)
	for _, theme := range []string{"editorial", "classic-compact", "reference-cards"} {
		t.Run(theme, func(t *testing.T) {
			p, ds := Render(sheets, Options{Theme: theme})
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			out := hugoBuild(t, binary, p)
			home := readOut(t, out, "index.html")
			for _, s := range sheets {
				cs, _ := s.Document.Metadata["cheatsheet"].(map[string]any)
				updated, _ := cs["updated"].(string)
				version, _ := cs["tool_version"].(string)
				if !date.MatchString(updated) {
					t.Errorf("%s: every demo sheet needs a real updated date, got %q", s.ID, updated)
				}
				if version != demoVersions[s.ID] {
					t.Errorf("%s: tool_version %q, research.md records %q", s.ID, version, demoVersions[s.ID])
				}
				if _, ok := cs["tool"]; ok {
					t.Errorf("%s: adding cheatsheet.tool would add a navi tag", s.ID)
				}
				page := readOut(t, out, s.ID+"/index.html")
				want := `Updated <time datetime="` + updated + `">` + updated + `</time>`
				if !strings.Contains(page, want) || !strings.Contains(home, want) {
					t.Errorf("%s: missing %s", s.ID, want)
				}
				if version != "" && (!strings.Contains(page, "<span>Version "+version+"</span>") || !strings.Contains(home, "<span>Version "+version+"</span>")) {
					t.Errorf("%s: missing version %s", s.ID, version)
				}
				if version == "" && strings.Contains(page, "<span>Version ") {
					t.Errorf("%s: no version expected", s.ID)
				}
			}
		})
	}
}
