// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type shelfEntry struct {
	Title     string   `json:"title"`
	CardCount int      `json:"card_count"`
	Tool      string   `json:"tool"`
	ShelfTags []string `json:"shelf_tags"`
	Tags      []string `json:"tags"`
}

func libraryEntries(t *testing.T, p Project) map[string]shelfEntry {
	t.Helper()
	var all []shelfEntry
	if err := json.Unmarshal(p.Files["data/library.json"], &all); err != nil {
		t.Fatal(err)
	}
	out := map[string]shelfEntry{}
	for _, e := range all {
		out[e.Title] = e
	}
	return out
}

func TestShelfCardData(t *testing.T) {
	p, ds := Render([]Sheet{
		described(t, "git", "title: Git\ncheatsheet: {tool: git}\ntags: [vcs, cli, history, extra, more]\n", "## A\n### One\nx\n### Two\nx\n## B\n### Three\nx\n"),
		described(t, "intro", "title: Intro\n", "Only a preamble.\n"),
		described(t, "dups", "title: Dups\ntags: [b, a, b, a, c, d]\n", "## A\n### One\nx\n"),
		described(t, "fenced", "title: Fenced\n", "## A\n### Real\n```md\n## Not a section\n### Not a card\n```\n"),
		described(t, "hostile", "title: Hostile\ncheatsheet: {tool: '<script>alert(1)</script>'}\ntags: ['<img src=x onerror=1>', 'a&b']\n", "## A\n### One\nx\n"),
	}, Options{})
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	got := libraryEntries(t, p)
	git := got["Git"]
	if git.CardCount != 3 || git.Tool != "git" {
		t.Errorf("git: %+v", git)
	}
	if want := []string{"vcs", "cli", "history"}; !reflect.DeepEqual(git.ShelfTags, want) {
		t.Errorf("shelf tags keep source order and stop at three: %v", git.ShelfTags)
	}
	if len(git.Tags) != 5 {
		t.Errorf("full tag list must stay intact for tag pages: %v", git.Tags)
	}
	if f := got["Fenced"]; f.CardCount != 1 {
		t.Errorf("headings inside a fenced block are not cards: %+v", f)
	}
	if intro := got["Intro"]; intro.CardCount != 0 || intro.Tool != "" || len(intro.ShelfTags) != 0 {
		t.Errorf("preamble-only sheet: %+v", intro)
	}
	if want := []string{"b", "a", "c"}; !reflect.DeepEqual(got["Dups"].ShelfTags, want) {
		t.Errorf("duplicates must not consume pills: %v", got["Dups"].ShelfTags)
	}
	if h := got["Hostile"]; h.Tool != "<script>alert(1)</script>" || len(h.ShelfTags) != 2 {
		t.Errorf("data keeps text verbatim for the template to escape: %+v", h)
	}
	again, _ := Render([]Sheet{
		described(t, "git", "title: Git\ncheatsheet: {tool: git}\ntags: [vcs, cli, history, extra, more]\n", "## A\n### One\nx\n### Two\nx\n## B\n### Three\nx\n"),
	}, Options{})
	only, _ := Render([]Sheet{
		described(t, "git", "title: Git\ncheatsheet: {tool: git}\ntags: [vcs, cli, history, extra, more]\n", "## A\n### One\nx\n### Two\nx\n## B\n### Three\nx\n"),
	}, Options{})
	if !bytes.Equal(again.Files["data/library.json"], only.Files["data/library.json"]) {
		t.Error("library data must be deterministic")
	}
}

func TestShelfCardMarkupInTemplate(t *testing.T) {
	p, _ := Render([]Sheet{sheet(t, "a", "## S\n### C\nText.\n")}, Options{})
	home := string(p.Files["layouts/home.html"])
	for _, want := range []string{"shelf-number", "card_count", "library-tool", "library-tags", "shelf_tags"} {
		if !strings.Contains(home, want) {
			t.Errorf("home layout lacks %s", want)
		}
	}
	if strings.Contains(home, "<script") || strings.Contains(home, "safeHTML") {
		t.Error("home layout must stay static and escaped")
	}
}

func TestHugoIntegrationShelfCards(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("SKIPPED, NOT PASSED: set SHEETMILL_HUGO to an installed pinned Hugo to render shelf cards")
	}
	sheets := func() []Sheet {
		return []Sheet{
			described(t, "git", "title: Git\ndescription: Track changes.\ncheatsheet: {tool: git}\ntags: [vcs, cli, history, extra]\n", "## A\n### One\nx\n### Two\nx\n## B\n### Three\nx\n"),
			described(t, "solo", "title: Solo\ncategory: files\n", "## A\n### Only\nx\n"),
			described(t, "intro", "title: Intro\n", "Only a preamble.\n"),
			described(t, "evil", "title: Evil\ncheatsheet: {tool: '<script>alert(1)</script>'}\ntags: ['<b>x</b>']\n", "## A\n### One\nx\n"),
		}
	}
	for _, theme := range []string{"editorial", "classic-compact", "reference-cards"} {
		t.Run(theme, func(t *testing.T) {
			p, ds := Render(sheets(), Options{Theme: theme})
			if len(ds) > 0 {
				t.Fatal(ds)
			}
			out := hugoBuild(t, binary, p)
			raw, err := os.ReadFile(filepath.Join(out, "index.html"))
			if err != nil {
				t.Fatal(err)
			}
			home := string(raw)
			if bytes.Contains(raw, []byte("<script>alert(1)")) || bytes.Contains(raw, []byte("<b>x</b>")) {
				t.Fatal("hostile text reached the home page unescaped")
			}
			if theme != "editorial" {
				if !strings.Contains(home, "3 cards") || !strings.Contains(home, "1 card<") {
					t.Errorf("classic list lacks card counts: %s", home)
				}
				return
			}
			start := strings.Index(home, `<h3>Git</h3>`)
			if start < 0 {
				t.Fatal("no Git card")
			}
			card := home[strings.LastIndex(home[:start], `<a class="library-item"`):]
			card = card[:strings.Index(card, "</a>")]
			// The link's accessible name starts with the title: facts follow the heading.
			if h, n := strings.Index(card, "<h3>"), strings.Index(card, `class="shelf-number"`); h < 0 || n < h {
				t.Errorf("shelf-number must follow the title in the DOM: %s", card)
			}
			for _, want := range []string{
				`<p class="shelf-number">3 cards</p>`,
				`<p class="library-tool">Tool: git</p>`,
				`<ul class="library-tags" aria-label="Tags"><li>vcs</li><li>cli</li><li>history</li></ul>`,
				"Track changes.",
			} {
				if !strings.Contains(card, want) {
					t.Errorf("Git card lacks %s in %s", want, card)
				}
			}
			if strings.Contains(card, "extra") {
				t.Error("a fourth tag must not appear on the card")
			}
			for _, want := range []string{`<p class="shelf-number">1 card</p>`, `Tool: &lt;script&gt;alert(1)&lt;/script&gt;`, `<li>&lt;b&gt;x&lt;/b&gt;</li>`} {
				if !strings.Contains(home, want) {
					t.Errorf("home lacks %s", want)
				}
			}
			intro := home[strings.Index(home, `<h3>Intro</h3>`):]
			intro = intro[:strings.Index(intro, "</a>")]
			if strings.Contains(intro, "library-tool") || strings.Contains(intro, "library-tags") || strings.Contains(intro, "0 cards") {
				t.Errorf("empty metadata must render nothing: %s", intro)
			}
		})
	}
}
