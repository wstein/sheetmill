// SPDX-License-Identifier: EUPL-1.2
// Package hugo renders semantic IR into a self-contained, inert Hugo project.
package hugo

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"net/url"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	brand "github.com/wstein/sheetmill/assets"
	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/model"
	"github.com/wstein/sheetmill/internal/validate"
	goldhtml "github.com/yuin/goldmark/renderer/html"
	"go.yaml.in/yaml/v3"
)

const Version = "0.167.0"

//go:embed theme
var theme embed.FS

// Embedded sources are variables so tests can prove read failures are fatal.
var themeFiles, brandFiles fs.FS = theme, brand.Files

type Sheet struct {
	ID, SourceName string
	// Path is the slash-separated source path relative to the selection root,
	// such as "format/gfm-plus.md". Empty means SourceName. Relative links resolve
	// against it; it never contains host path components.
	Path     string
	Document *model.Document
}
type Options struct {
	Title string
	// BasePath is the URL path the site is served under, such as "/sheetmill/".
	// Empty means the origin root.
	BasePath string
	// EditBaseURL is an explicit GitHub /owner/repository/edit/ref/directory/ base.
	// It refers to authored sheets, never to upstream source provenance.
	EditBaseURL string
	// Description introduces the library on the home page.
	Description string
	// SiteURL is the canonical https://host/<base path>/ of the published site.
	// It enables canonical links, og:url/og:image and a sitemap.
	SiteURL string
	// SourceBaseURL is an explicit GitHub /owner/repository/blob/ref/directory/ base
	// for the selection root. Relative links that name no selected sheet resolve to
	// the repository file instead of being reported as unresolved.
	SourceBaseURL string
	// Theme selects the page layout: "" or "editorial" (default), "classic-compact" or "reference-cards".
	Theme string
	// Palette selects a color palette independently of the layout; empty uses graphite-coral.
	Palette string
	// HeaderLinks adds up to four validated links to the top bar and the home page
	// introduction, such as a sibling site. See ValidateHeaderLinks.
	HeaderLinks []HeaderLink
}

// HeaderLink is an optional navigation link shown in the top bar and on the home page.
type HeaderLink struct{ Label, URL string }

// Themes lists the accepted Options.Theme values.
var Themes = []string{"editorial", "classic-compact", "reference-cards"}

// basePath accepts "/" or "/seg/.../" with unreserved URL characters only, so the
// value is safe in HTML attributes, CSS, JSON and JavaScript without escaping.
var basePath = regexp.MustCompile(`^/(?:[A-Za-z0-9_~-][A-Za-z0-9._~-]*/)*$`)

// Issue associates a renderer diagnostic with the explicitly selected basename.
// The CLI resolves this identity to the selected path; host paths never enter output.
type Issue struct {
	Source string
	diagnostic.Diagnostic
}

func issues(source string, ds []diagnostic.Diagnostic) []Issue {
	out := make([]Issue, 0, len(ds))
	for _, d := range ds {
		out = append(out, Issue{source, d})
	}
	return out
}

type Project struct {
	Files    map[string][]byte
	Warnings []Issue
}
type entry struct {
	Title    string   `json:"title"`
	Sheet    string   `json:"sheet"`
	Section  string   `json:"section"`
	Text     string   `json:"text"`
	URL      string   `json:"url"`
	Tags     []string `json:"tags,omitempty"`
	Category string   `json:"category,omitempty"`
	License  string   `json:"license,omitempty"`
}

// sheetMetadata reads the displayed fields; validation already enforced their types.
func sheetMetadata(m map[string]any) (description, category, license string, tags []string) {
	if source, ok := m["source"].(map[string]any); ok {
		license, _ = source["license"].(string)
	}
	description, _ = m["description"].(string)
	category, _ = m["category"].(string)
	if list, ok := m["tags"].([]any); ok {
		for _, tag := range list {
			if tag, ok := tag.(string); ok {
				tags = append(tags, tag)
			}
		}
	}
	return description, category, license, tags
}

// maxShelfTags bounds the tag pills on one home page card.
const maxShelfTags = 3

// shelfTags returns up to maxShelfTags distinct tags in source order.
func shelfTags(tags []string) []string {
	var out []string
	for _, tag := range tags {
		if len(out) == maxShelfTags {
			break
		}
		if !slices.Contains(out, tag) {
			out = append(out, tag)
		}
	}
	return out
}

// documentedTool reads the validated cheatsheet.tool field.
func documentedTool(m map[string]any) string {
	if cs, ok := m["cheatsheet"].(map[string]any); ok {
		tool, _ := cs["tool"].(string)
		return tool
	}
	return ""
}

type page struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	TagLinks    []link   `json:"tag_links,omitempty"`
	License     string   `json:"license,omitempty"`
	// CardCount, Tool and ShelfTags feed the home page shelf cards. ShelfTags is the
	// first maxShelfTags distinct tags in source order; Tags keeps the full list.
	CardCount int      `json:"card_count,omitempty"`
	Tool      string   `json:"tool,omitempty"`
	ShelfTags []string `json:"shelf_tags,omitempty"`
	Body      string   `json:"body"`
	EditURL   string   `json:"edit_url,omitempty"`
	Related   []link   `json:"related,omitempty"`
	nav       validate.NavigationSettings
}

type link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// related ranks other pages by shared tags (2 points each) and a shared
// category (1 point) and keeps the best three; ties keep route order.
func related(pages []page, self int) []link {
	type candidate struct{ index, score int }
	var candidates []candidate
	for i, other := range pages {
		if i == self {
			continue
		}
		score := 0
		for _, tag := range other.Tags {
			if slices.Contains(pages[self].Tags, tag) {
				score += 2
			}
		}
		if other.Category != "" && other.Category == pages[self].Category {
			score++
		}
		if score > 0 {
			candidates = append(candidates, candidate{i, score})
		}
	}
	sort.SliceStable(candidates, func(a, b int) bool { return candidates[a].score > candidates[b].score })
	var out []link
	for _, c := range candidates[:min(3, len(candidates))] {
		out = append(out, link{pages[c.index].Title, pages[c.index].URL})
	}
	return out
}

// shelf groups library pages by category for the home page.
type shelf struct {
	Category string `json:"category"`
	Sheets   []page `json:"sheets"`
}

// shelves groups pages by their effective group (navigation.group, else
// category) and sorts named groups by name with the ungrouped shelf last.
// Members of a group order by navigation.order, then title and route; pages keep
// their deterministic route order elsewhere. Bodies are omitted from the index.
func shelves(pages []page) []shelf {
	groups := map[string][]page{}
	for _, p := range pages {
		p.Body, p.Related = "", nil
		name := p.Category
		if p.nav.Group != "" {
			name = p.nav.Group
		}
		groups[name] = append(groups[name], p)
	}
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] != "" && (names[j] == "" || names[i] < names[j]) })
	out := make([]shelf, 0, len(names))
	for _, name := range names {
		sort.SliceStable(groups[name], func(i, j int) bool { return placedBefore(groups[name][i], groups[name][j]) })
		out = append(out, shelf{name, groups[name]})
	}
	return out
}

// placedBefore orders explicitly ordered pages first (numerically), then by
// Unicode-lowercased title, original title bytes and finally the unique route.
func placedBefore(a, b page) bool {
	if (a.nav.Order != nil) != (b.nav.Order != nil) {
		return a.nav.Order != nil
	}
	if a.nav.Order != nil && *a.nav.Order != *b.nav.Order {
		return *a.nav.Order < *b.nav.Order
	}
	if x, y := strings.ToLower(a.Title), strings.ToLower(b.Title); x != y {
		return x < y
	}
	if a.Title != b.Title {
		return a.Title < b.Title
	}
	return a.URL < b.URL
}

var route = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Render does not open files, execute tools or evaluate source Markdown.
// No partial project is returned on a fatal diagnostic.
func Render(sheets []Sheet, options Options) (Project, []Issue) {
	p := Project{Files: map[string][]byte{}}
	base := options.BasePath
	if base == "" {
		base = "/"
	}
	if !basePath.MatchString(base) || strings.Contains(base, "/../") || strings.Contains(base, "/./") {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: "invalid base path: " + options.BasePath}})
	}
	if len(sheets) == 0 {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: "select at least one sheet"}})
	}
	if err := validateHeaderLinks(options.HeaderLinks); err != nil {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: err.Error()}})
	}
	editBase, err := githubBase(options.EditBaseURL, "edit")
	if err != nil {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: err.Error()}})
	}
	origin, err := siteOrigin(options.SiteURL, base)
	if err != nil {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: err.Error()}})
	}
	sourceBase, err := githubBase(options.SourceBaseURL, "blob")
	if err != nil {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: err.Error()}})
	}
	sheets = append([]Sheet(nil), sheets...)
	sort.SliceStable(sheets, func(i, j int) bool { return sheets[i].ID < sheets[j].ID })
	routes := map[string]string{}
	paths := map[string]string{}
	names := map[string]bool{}
	ids := map[string]string{}
	for _, s := range sheets {
		if first, duplicate := ids[s.ID]; duplicate {
			// Name every colliding source so neither file is left unexplained.
			d := []diagnostic.Diagnostic{{Code: "REN003", Message: "duplicate sheet route: " + s.ID}}
			return Project{}, append(issues(first, d), issues(s.SourceName, d)...)
		}
		if !route.MatchString(s.ID) || s.ID == "assets" || s.ID == "brand" || s.ID == "tags" {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN003", Message: "invalid or reserved sheet route: " + s.ID}})
		}
		if s.Document == nil || s.Document.Schema != 1 {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN001", Message: "renderer requires schema-1 IR"}})
		}
		if commands := validate.Commands(s.Document, nil); len(commands) > 0 {
			return Project{}, issues(s.SourceName, commands)
		}
		if _, nav := validate.Navigation(s.Document.Metadata, nil); len(nav) > 0 {
			return Project{}, issues(s.SourceName, nav)
		}
		if path.Base(s.SourceName) != s.SourceName || s.SourceName == "" || s.SourceName == "." || s.SourceName == ".." {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN003", Message: "source name must be a basename"}})
		}
		if s.Path == "" {
			s.Path = s.SourceName
		}
		if !localSheetPath(s.Path) || path.Base(s.Path) != s.SourceName {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN003", Message: "sheet path must be a clean relative path ending in its source name: " + s.Path}})
		}
		if _, exists := names[s.SourceName]; exists {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN003", Message: "ambiguous source basename: " + s.SourceName}})
		}
		names[s.SourceName] = true
		routes[s.Path] = s.ID
		paths[s.ID] = s.Path
		ids[s.ID] = s.SourceName
	}
	putJSON := func(name string, v any) error {
		b, e := json.MarshalIndent(v, "", "  ")
		if e == nil {
			p.Files[name] = append(b, '\n')
		}
		return e
	}
	var pages []page
	var sources []string
	var index []entry
	var errors []Issue
	// Anchors of every selected sheet, so cross-sheet fragments can be resolved.
	anchors := map[string]map[string]bool{}
	for _, s := range sheets {
		anchors[s.ID] = map[string]bool{}
		for _, section := range s.Document.Sections {
			anchors[s.ID][section.ID] = true
			for _, card := range section.Cards {
				anchors[s.ID][card.ID] = true
			}
		}
	}
	for _, s := range sheets {
		r := renderer{self: s.ID, path: paths[s.ID], base: base, source: sourceBase, routes: routes, anchors: anchors}
		title, _ := s.Document.Metadata["title"].(string)
		description, category, license, tags := sheetMetadata(s.Document.Metadata)
		navigation, _ := validate.Navigation(s.Document.Metadata, nil)
		var b strings.Builder
		r.nodes(&b, s.Document.Preamble, false)
		// A jump list lets readers reach any section without JavaScript.
		if len(s.Document.Sections) > 1 {
			b.WriteString(`<nav class="sheet-toc" aria-label="On this page"><ul>`)
			for _, section := range s.Document.Sections {
				fmt.Fprintf(&b, `<li><a href="#ref-%s">%s</a></li>`, esc(section.ID), esc(section.Title))
			}
			b.WriteString(`</ul></nav>`)
		}
		cardCount := 0
		for _, section := range s.Document.Sections {
			cardCount += len(section.Cards)
			fmt.Fprintf(&b, `<section class="sheet-section" id="ref-%s"><h2>%s</h2>`, esc(section.ID), r.heading(section.Heading))
			r.nodes(&b, section.Content, false)
			// Cards never span more columns than their section's grid provides.
			columns := section.Columns
			if columns == 0 {
				columns = 3
			}
			fmt.Fprintf(&b, `<div class="cards cols-%d">`, columns)
			for _, card := range section.Cards {
				class := fmt.Sprintf("card col-span-%d row-span-%d", min(max(1, card.Layout.Columns), columns), max(1, card.Layout.Rows))
				for _, c := range card.Classes {
					class += " " + c
				}
				fmt.Fprintf(&b, `<article class="%s" id="ref-%s"><h3>%s <a class="permalink" href="#ref-%s" aria-label="Permalink to %s">#</a></h3>`, esc(class), esc(card.ID), r.heading(card.Heading), esc(card.ID), esc(card.Title))
				r.keys = slices.Contains(card.Classes, "keys")
				// Card kinds carry a visible text label, never colour alone.
				for _, kind := range []string{"warning", "tip", "note"} {
					if slices.Contains(card.Classes, kind) {
						fmt.Fprintf(&b, `<p class="card-kind">%s</p>`, strings.ToUpper(kind[:1])+kind[1:])
						break
					}
				}
				r.nodes(&b, card.Content, false)
				r.keys = false
				b.WriteString(`</article>`)
				index = append(index, entry{Title: card.Title, Sheet: title, Section: section.Title, Text: plain(card.Content), URL: base + s.ID + "/#ref-" + card.ID, Tags: tags, Category: category, License: license})
			}
			b.WriteString(`</div></section>`)
		}
		if source, ok := s.Document.Metadata["source"]; ok {
			disclosure, e := sourceDetails(source)
			if e != nil {
				errors = append(errors, Issue{s.SourceName, diagnostic.Diagnostic{Code: "REN001", Message: e.Error()}})
			} else {
				b.WriteString(`<aside class="provenance" aria-label="Source and license"><details class="source-details"><summary>`)
				if fields, ok := source.(map[string]any); ok {
					for _, key := range []string{"attribution", "license"} {
						if value, ok := fields[key].(string); ok && value != "" {
							fmt.Fprintf(&b, `<span class="source-facts">%s</span><span aria-hidden="true"> · </span>`, esc(value))
						}
					}
				}
				fmt.Fprintf(&b, `Source details</summary>%s<p>Source attribution is preserved. Import does not grant redistribution rights.</p></details></aside>`, disclosure)
			}
		}
		p.Warnings = append(p.Warnings, issues(s.SourceName, r.warnings)...)
		errors = append(errors, issues(s.SourceName, r.errors)...)
		page := page{Title: title, URL: base + s.ID + "/", Description: description, Category: category, Tags: tags, License: license, Body: b.String(), nav: navigation, CardCount: cardCount, Tool: documentedTool(s.Document.Metadata), ShelfTags: shelfTags(tags)}
		if editBase != "" {
			page.EditURL = editBase + escapePath(paths[s.ID])
		}
		pages = append(pages, page)
		sources = append(sources, s.SourceName)
		ir, e := s.Document.JSON()
		if e != nil {
			errors = append(errors, Issue{s.SourceName, diagnostic.Diagnostic{Code: "REN001", Message: e.Error()}})
		} else {
			p.Files["data/ir/"+s.ID+".json"] = ir
		}
		front, _ := json.Marshal(map[string]any{"title": title, "url": "/" + s.ID + "/", "sheetid": s.ID})
		p.Files["content/"+s.ID+"/index.md"] = []byte("---\n" + string(front) + "\n---\n")
		// Preamble-only sheets must also be discoverable by search.
		index = append(index, entry{Title: title, Sheet: title, Text: plain(s.Document.Preamble), URL: page.URL, Tags: tags, Category: category, License: license})
	}
	tagPages, tagErr := tagCatalog(pages, base)
	if tagErr != nil {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: tagErr.Error()}})
	}
	tagURLs := map[string]string{}
	for _, tag := range tagPages {
		tagURLs[tag.Title] = tag.URL
		if e := putJSON("data/tags/"+tag.ID+".json", tag); e != nil {
			return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN001", Message: e.Error()}})
		}
		front, _ := json.Marshal(map[string]any{"title": tag.Title, "layout": "tag", "tagid": tag.ID, "url": "/tags/" + tag.ID + "/"})
		p.Files["content/tags/"+tag.ID+"/index.md"] = []byte("---\n" + string(front) + "\n---\n")
	}
	if len(tagPages) > 0 {
		p.Files["content/tags/_index.md"] = []byte("---\n{\"title\":\"Tags\",\"layout\":\"tags\",\"url\":\"/tags/\",\"tagindex\":true}\n---\n")
	}
	for i := range pages {
		seen := map[string]bool{}
		for _, tag := range pages[i].Tags {
			if !seen[tag] {
				pages[i].TagLinks = append(pages[i].TagLinks, link{tag, tagURLs[tag]})
				seen[tag] = true
			}
		}
	}
	// Related sheets need every page, so per-sheet data is written afterwards.
	for i := range pages {
		pages[i].Related = related(pages, i)
		id := strings.TrimSuffix(strings.TrimPrefix(pages[i].URL, base), "/")
		if e := putJSON("data/sheets/"+id+".json", pages[i]); e != nil {
			errors = append(errors, Issue{sources[i], diagnostic.Diagnostic{Code: "REN001", Message: e.Error()}})
		}
	}
	if len(errors) > 0 {
		return Project{}, errors
	}
	cards := options.Theme == "reference-cards"
	classic := options.Theme == "classic-compact" || cards
	if options.Theme != "" && !classic && options.Theme != "editorial" {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: "unknown theme: " + options.Theme}})
	}
	paletteName := options.Palette
	if paletteName == "" {
		paletteName = "graphite-coral"
	}
	if !slices.Contains(PaletteNames(), paletteName) {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: "unknown palette: " + paletteName}})
	}
	title := options.Title
	if title == "" {
		title = "SheetMill References"
	}
	params := map[string]any{"base": base, "description": options.Description, "palette": paletteName}
	if len(options.HeaderLinks) > 0 {
		params["headerlinks"] = headerLinkParams(options.HeaderLinks)
	}
	if classic {
		params["classic"] = true
	}
	if cards {
		params["cards"] = true
	}
	if origin != "" {
		params["origin"] = origin
		var sitemap strings.Builder
		sitemap.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
		locations := append([]string{base}, urls(pages)...)
		if len(tagPages) > 0 {
			locations = append(locations, base+"tags/")
		}
		for _, tag := range tagPages {
			locations = append(locations, tag.URL)
		}
		for _, loc := range locations {
			fmt.Fprintf(&sitemap, "<url><loc>%s</loc></url>\n", esc(origin+loc))
		}
		sitemap.WriteString("</urlset>\n")
		p.Files["static/sitemap.xml"] = []byte(sitemap.String())
	}
	config := map[string]any{"baseURL": "https://example.invalid/", "title": title, "languageCode": "en", "taxonomies": map[string]string{}, "disableKinds": []string{"taxonomy", "term", "RSS", "sitemap"}, "enableRobotsTXT": false, "disableHugoGeneratorInject": true, "params": params, "outputs": map[string]any{"home": []string{"HTML"}, "page": []string{"HTML"}}, "security": map[string]any{"enableInlineShortcodes": false, "exec": map[string]any{"allow": []string{"none"}, "osEnv": []string{"none"}}, "http": map[string]any{"methods": []string{"none"}, "urls": []string{"none"}}}}
	// Every remaining file is required; never return a partial project.
	copyTree := func(files fs.FS, from, to string) error {
		return fs.WalkDir(files, from, func(name string, d fs.DirEntry, e error) error {
			if e != nil || d.IsDir() {
				return e
			}
			b, e := fs.ReadFile(files, name)
			if from != "." {
				name = strings.TrimPrefix(name, from+"/")
			}
			p.Files[path.Join(to, name)] = b
			return e
		})
	}
	for _, e := range []error{putJSON("data/palettes.json", palettes), putJSON("data/tagindex.json", tagPages), putJSON("data/library.json", pages), putJSON("data/shelves.json", shelves(pages)), putJSON("static/search.json", index), copyTree(themeFiles, "theme", ""), copyTree(brandFiles, ".", "static/brand")} {
		if e != nil {
			return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN001", Message: "project assembly failed: " + e.Error()}})
		}
	}
	if !classic {
		delete(p.Files, "static/assets/classic.css")
	}
	if !cards {
		delete(p.Files, "static/assets/cards.css")
	}
	// Hash the exact exported bytes; keep assets in the same directory so CSS's
	// relative local-font URLs continue to resolve. No unversioned aliases remain.
	for _, asset := range []struct{ key, stem, ext string }{{"stylesheet", "site", "css"}, {"script", "app", "js"}, {"classicstylesheet", "classic", "css"}, {"cardsstylesheet", "cards", "css"}} {
		original := "static/assets/" + asset.stem + "." + asset.ext
		data, ok := p.Files[original]
		if !ok && (asset.key == "classicstylesheet" || asset.key == "cardsstylesheet") {
			continue
		}
		if !ok {
			return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN001", Message: "project assembly failed: missing asset " + original}})
		}
		name := fmt.Sprintf("assets/%s.%x.%s", asset.stem, sha256.Sum256(data), asset.ext)
		p.Files["static/"+name] = data
		delete(p.Files, original)
		params[asset.key] = name
	}
	if err := putJSON("hugo.json", config); err != nil {
		panic(err)
	} // strings and literals only
	return p, nil
}

// sourceDetails presents all source fields without exposing JSON syntax. Unknown
// structured values retain their YAML representation; every key/value is escaped.
func sourceDetails(source any) (string, error) {
	// Keep the same JSON-compatible metadata contract as canonical IR exports.
	if _, err := json.Marshal(source); err != nil {
		return "", err
	}
	fields, ok := source.(map[string]any)
	if !ok {
		fields = map[string]any{"source": source}
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(`<dl class="source-metadata">`)
	for _, key := range keys {
		label := key
		if key == "attribution" {
			label = "Attribution"
		}
		if key == "license" {
			label = "License"
		}
		label, err := sourceText(label, true)
		if err != nil {
			return "", err
		}
		value, err := sourceText(fields[key], false)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, `<dt>%s:</dt><dd>%s</dd>`, esc(label), esc(value))
	}
	b.WriteString(`</dl>`)
	return b.String(), nil
}

// HTML normalizes NUL and CR, and ordinary key layout collapses whitespace.
// Quote controls using YAML escapes so accepted metadata is not silently changed.
func sourceText(value any, key bool) (string, error) {
	if text, ok := value.(string); ok {
		controls := strings.ContainsFunc(text, func(r rune) bool { return unicode.IsControl(r) && (key || (r != '\n' && r != '\t')) })
		if !controls {
			return text, nil
		}
		value = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Style: yaml.DoubleQuotedStyle, Value: text}
	}
	var encoded bytes.Buffer
	encoder := yaml.NewEncoder(&encoded)
	encoder.SetIndent(2)
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return strings.TrimSuffix(encoded.String(), "\n"), nil
}

type renderer struct {
	inLink           bool
	keys, keyCell    bool // inside a .keys card; inside its first table column
	routes           map[string]string
	self, path, base string
	source           string
	anchors          map[string]map[string]bool
	errors, warnings []diagnostic.Diagnostic
}

func esc(s string) string { return html.EscapeString(s) }
func normal(s string) string {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	goldhtml.NewWriter().Write(w, []byte(s))
	w.Flush()
	return b.String()
}
func (r *renderer) children(n model.Node, header bool) string {
	var b strings.Builder
	r.nodes(&b, n.Children, header)
	return b.String()
}
func (r *renderer) nodes(b *strings.Builder, nodes []model.Node, header bool) {
	for _, n := range nodes {
		r.node(b, n, header)
	}
}
func (r *renderer) node(b *strings.Builder, n model.Node, header bool) {
	wrap := func(tag string) {
		b.WriteString("<" + tag + ">")
		r.nodes(b, n.Children, header)
		b.WriteString("</" + strings.Fields(tag)[0] + ">")
	}
	prop := func(key string) string { v, _ := n.Properties[key].(string); return v }
	yes := func(key string) bool { v, _ := n.Properties[key].(bool); return v }
	switch n.Kind {
	case "Paragraph":
		wrap("p")
	case "TextBlock":
		r.nodes(b, n.Children, header)
		b.WriteByte('\n')
	case "Text", "String":
		if verbatim(n) {
			b.WriteString(esc(n.Text))
		} else {
			b.WriteString(normal(n.Text))
		}
		if yes("hard_break") {
			b.WriteString("<br>\n")
		} else if yes("soft_break") {
			b.WriteByte('\n')
		}
	case "Heading":
		level, _ := n.Properties["level"].(int)
		wrap(fmt.Sprintf("h%d", max(1, min(6, level))))
	case "Blockquote":
		wrap("blockquote")
	case "List":
		if yes("ordered") {
			start, _ := n.Properties["start"].(int)
			wrap(fmt.Sprintf(`ol start="%d"`, start))
		} else {
			wrap("ul")
		}
	case "ListItem":
		wrap("li")
	case "ThematicBreak":
		b.WriteString("<hr>")
	case "CodeSpan":
		if r.keyCell && !header {
			b.WriteString(keys(n))
			return
		}
		wrap("code")
	case "Emphasis":
		if n.Properties["level"] == 2 {
			wrap("strong")
		} else {
			wrap("em")
		}
	case "Strikethrough":
		wrap("del")
	case "CodeBlock", "FencedCodeBlock":
		fmt.Fprintf(b, `<pre><code class="language-%s">%s</code></pre>`, esc(prop("language")), esc(n.Text))
	case "TaskCheckBox":
		b.WriteString(`<input type="checkbox" disabled`)
		if yes("checked") {
			b.WriteString(` checked`)
		}
		b.WriteString(` aria-label="Task status">`)
	case "Table":
		wrap("table")
	case "TableHeader":
		b.WriteString("<thead><tr>")
		r.nodes(b, n.Children, true)
		b.WriteString("</tr></thead>")
	case "TableRow":
		b.WriteString("<tr>")
		for i, cell := range n.Children {
			r.keyCell = r.keys && i == 0
			r.node(b, cell, header)
		}
		r.keyCell = false
		b.WriteString("</tr>")
	case "TableCell":
		tag := "td"
		if header {
			tag = "th"
		}
		alignment := strings.ToLower(prop("alignment"))
		if alignment == "left" || alignment == "center" || alignment == "right" {
			tag += ` class="align-` + alignment + `"`
		}
		wrap(tag)
	case "Link", "Image", "AutoLink":
		destination := r.url(n, prop("destination"))
		title := normal(prop("title"))
		if n.Kind == "Image" {
			r.warnings = append(r.warnings, diagnostic.Diagnostic{Code: "REN102", Message: "image retained as a link; resource not fetched", Span: n.Span})
		}
		if r.inLink && n.Kind != "Image" {
			// Image alternative text may contain parsed links. Keep its visible label
			// inside the surrounding image/ordinary link without nesting anchors.
			if n.Kind == "AutoLink" {
				b.WriteString(esc(n.Text))
			} else {
				r.nodes(b, n.Children, header)
			}
			return
		}
		// Without label text an anchor is invisible and has no accessible name.
		unlabeled := n.Kind != "AutoLink" && !labeled(n.Children)
		if n.Kind == "Image" && r.inLink {
			b.WriteString(`<span class="image-reference">`)
			if unlabeled {
				fmt.Fprintf(b, "image: %s</span>", esc(destination))
				return
			}
			r.nodes(b, n.Children, header)
			fmt.Fprintf(b, " (image: %s)</span>", esc(destination))
			return
		}
		fmt.Fprintf(b, `<a href="%s" title="%s">`, esc(destination), title)
		previous := r.inLink
		r.inLink = true
		switch {
		case n.Kind == "AutoLink":
			b.WriteString(esc(n.Text))
		case unlabeled && destination != "":
			b.WriteString(esc(destination))
		case unlabeled:
			b.WriteString("link")
		default:
			r.nodes(b, n.Children, header)
		}
		b.WriteString("</a>")
		r.inLink = previous
	case "RawHTML", "HTMLBlock":
		r.warnings = append(r.warnings, diagnostic.Diagnostic{Code: "REN101", Message: "raw HTML displayed inertly", Span: n.Span})
		tag := "code"
		if n.Kind == "HTMLBlock" {
			tag = "pre"
		}
		fmt.Fprintf(b, "<%s>%s</%s>", tag, esc(n.Text), tag)
	default:
		r.errors = append(r.errors, diagnostic.Diagnostic{Code: "REN001", Message: "unsupported IR node: " + n.Kind, Span: n.Span})
	}
}
func (r *renderer) url(n model.Node, destination string) string {
	// Decode CommonMark entity/escape syntax before URL validation. Unescape only
	// this text result; never treat it as HTML markup.
	destination = html.UnescapeString(normal(destination))
	u, e := url.Parse(destination)
	invalid := e != nil || strings.HasPrefix(destination, "//") || strings.Contains(destination, "\\") || strings.IndexFunc(destination, func(c rune) bool { return unicode.IsControl(c) }) >= 0
	if e == nil {
		scheme := strings.ToLower(u.Scheme)
		invalid = invalid || (u.Opaque != "" && strings.IndexFunc(u.Opaque, unicode.IsSpace) >= 0)
		u.RawQuery = strings.ReplaceAll(u.RawQuery, " ", "%20")
		invalid = invalid || (scheme != "" && scheme != "http" && scheme != "https" && scheme != "mailto")
	}
	if invalid {
		r.errors = append(r.errors, diagnostic.Diagnostic{Code: "REN002", Message: "unsafe URL: " + destination, Span: n.Span})
		return ""
	}
	if u.Scheme == "" {
		// Internal links resolve to a selected sheet and a section or card ID.
		// Anything else is kept as written and reported, never silently broken.
		target, routed := r.self, u.Path == ""
		relative := !routed && !strings.HasPrefix(u.Path, "/")
		joined := path.Join(path.Dir(r.path), u.Path)
		if id, ok := r.routes[joined]; ok && relative {
			u.Path, target, routed = r.base+id+"/", id, true
		}
		unresolved := func(message string) {
			r.warnings = append(r.warnings, diagnostic.Diagnostic{Code: "REN103", Message: message + destination, Span: n.Span})
		}
		if !routed && relative && r.source != "" {
			if resolved, ok := r.repositoryURL(joined, u); ok {
				return resolved
			}
		}
		switch {
		case !routed && n.Kind != "Image": // images already report REN102
			unresolved("link target is not a selected sheet: ")
		case !routed || u.Fragment == "":
		case r.anchors[target][u.Fragment]:
			u.Fragment = "ref-" + u.Fragment
		default:
			unresolved("link fragment matches no section or card: ")
		}
	}
	// url.URL cannot represent an empty fragment; keep "#" as the author wrote it.
	if u.Fragment == "" && strings.HasSuffix(destination, "#") {
		return u.String() + "#"
	}
	return u.String()
}

// repositoryURL maps a selection-relative path to the source repository, refusing
// paths that leave the repository (above /owner/repository/blob/ref/).
func (r *renderer) repositoryURL(joined string, link *url.URL) (string, bool) {
	base, _ := url.Parse(r.source)
	root := strings.Join(strings.SplitN(strings.Trim(base.Path, "/"), "/", 5)[:4], "/")
	target := path.Join(base.Path, joined)
	if !strings.HasPrefix(target, "/"+root+"/") {
		return "", false
	}
	resolved := url.URL{Scheme: base.Scheme, Host: base.Host, Path: target, RawQuery: link.RawQuery, Fragment: link.Fragment}
	return resolved.String(), true
}

// keys renders a key-table code span as keyboard input; "Ctrl+C" becomes a
// combination of two keys, while "+" alone stays one key.
func keys(n model.Node) string {
	var text strings.Builder
	for _, c := range n.Children {
		text.WriteString(c.Text)
	}
	parts := strings.Split(text.String(), "+")
	if len(parts) == 1 || slices.Contains(parts, "") {
		return "<kbd>" + esc(text.String()) + "</kbd>"
	}
	for i, part := range parts {
		parts[i] = "<kbd>" + esc(part) + "</kbd>"
	}
	return "<kbd>" + strings.Join(parts, "+") + "</kbd>"
}

// labeled reports visible label content: text, or an image rendered as a reference.
func labeled(nodes []model.Node) bool {
	for _, n := range nodes {
		if n.Kind == "Image" || strings.TrimSpace(n.Text) != "" || labeled(n.Children) {
			return true
		}
	}
	return false
}
func verbatim(n model.Node) bool {
	raw, _ := n.Properties["raw"].(bool)
	code, _ := n.Properties["code"].(bool)
	return raw || code || n.Kind == "CodeBlock" || n.Kind == "FencedCodeBlock"
}
func plain(nodes []model.Node) string {
	var b strings.Builder
	var walk func([]model.Node)
	walk = func(ns []model.Node) {
		for _, n := range ns {
			if n.Kind != "RawHTML" && n.Kind != "HTMLBlock" {
				// Index code exactly as node() displays it; only prose has escapes.
				if verbatim(n) {
					b.WriteString(n.Text)
				} else {
					b.WriteString(html.UnescapeString(normal(n.Text)))
				}
				b.WriteByte(' ')
				walk(n.Children)
			}
		}
	}
	walk(nodes)
	return strings.Join(strings.Fields(b.String()), " ")
}

// Structural titles already have parsed attributes in the IR. Remove their
// literal trailing syntax from a copy of the heading's text nodes, preserving
// inline emphasis/code/link semantics and escaped literal braces.
func (r *renderer) heading(n model.Node) string {
	n.Children = append([]model.Node(nil), n.Children...)
	first := len(n.Children)
	tail := ""
	for first > 0 && n.Children[first-1].Kind == "Text" {
		first--
		tail = n.Children[first].Text + tail
	}
	open := strings.LastIndex(tail, "{")
	if open >= 0 && open+1 < len(tail) && (tail[open+1] == '.' || tail[open+1] == '#') && strings.HasSuffix(strings.TrimSpace(tail), "}") {
		slashes := 0
		for i := open - 1; i >= 0 && tail[i] == '\\'; i-- {
			slashes++
		}
		if slashes%2 == 0 {
			keep := len(strings.TrimRight(tail[:open], " \t\n"))
			for i := first; i < len(n.Children); i++ {
				size := min(keep, len(n.Children[i].Text))
				n.Children[i].Text = n.Children[i].Text[:size]
				keep -= size
			}
		}
	}
	return r.children(n, false)
}

// Operator-controlled edit links are intentionally separate from metadata.source.
// siteOrigin validates an explicit https site URL whose path is the base path and
// returns its scheme and host.
func siteOrigin(raw, base string) (string, error) {
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err == nil && u.Path == "" {
		u.Path = "/"
	}
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Path != base {
		return "", fmt.Errorf("site URL must be https://host%s without credentials, query or fragment", base)
	}
	return u.Scheme + "://" + u.Host, nil
}

func urls(pages []page) []string {
	out := make([]string, 0, len(pages))
	for _, p := range pages {
		out = append(out, p.URL)
	}
	return out
}

// githubBase validates an explicit https://github.com/owner/repository/<kind>/ref/
// URL with an optional directory, and returns it with one trailing slash.
func githubBase(raw, kind string) (string, error) {
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", fmt.Errorf("%s base must be an HTTPS github.com repository URL without credentials, query or fragment", kind)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != kind {
		return "", fmt.Errorf("%s base must name /owner/repository/%s/ref/ and an optional sheet directory", kind, kind)
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.Contains(part, "\\") {
			return "", fmt.Errorf("invalid %s base path", kind)
		}
	}
	return strings.TrimRight(u.String(), "/") + "/", nil
}

func localSheetPath(p string) bool {
	return p != "" && path.Clean(p) == p && !path.IsAbs(p) && p != ".." && !strings.HasPrefix(p, "../") && !strings.Contains(p, "\\")
}

func escapePath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
