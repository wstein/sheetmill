// SPDX-License-Identifier: EUPL-1.2
// Package hugo renders semantic IR into a self-contained, inert Hugo project.
package hugo

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"

	brand "github.com/wstein/sheetmill/assets"
	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/model"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

const Version = "0.167.0"

//go:embed theme
var theme embed.FS

// Embedded sources are variables so tests can prove read failures are fatal.
var themeFiles, brandFiles fs.FS = theme, brand.Files

type Sheet struct {
	ID, SourceName string
	Document       *model.Document
}
type Options struct {
	Title string
	// BasePath is the URL path the site is served under, such as "/sheetmill/".
	// Empty means the origin root.
	BasePath string
	// EditBaseURL is an explicit GitHub /owner/repository/edit/ref/directory/ base.
	// It refers to authored sheets, never to upstream source provenance.
	EditBaseURL string
}

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
	Title   string `json:"title"`
	Sheet   string `json:"sheet"`
	Section string `json:"section"`
	Text    string `json:"text"`
	URL     string `json:"url"`
}
type page struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Body    string `json:"body"`
	EditURL string `json:"edit_url,omitempty"`
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
	editBase, err := githubEditBase(options.EditBaseURL)
	if err != nil {
		return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN003", Message: err.Error()}})
	}
	sheets = append([]Sheet(nil), sheets...)
	sort.SliceStable(sheets, func(i, j int) bool { return sheets[i].ID < sheets[j].ID })
	routes := map[string]string{}
	ids := map[string]string{}
	for _, s := range sheets {
		if first, duplicate := ids[s.ID]; duplicate {
			// Name every colliding source so neither file is left unexplained.
			d := []diagnostic.Diagnostic{{Code: "REN003", Message: "duplicate sheet route: " + s.ID}}
			return Project{}, append(issues(first, d), issues(s.SourceName, d)...)
		}
		if !route.MatchString(s.ID) || s.ID == "assets" || s.ID == "brand" {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN003", Message: "invalid or reserved sheet route: " + s.ID}})
		}
		if s.Document == nil || s.Document.Schema != 1 {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN001", Message: "renderer requires schema-1 IR"}})
		}
		if path.Base(s.SourceName) != s.SourceName || s.SourceName == "" || s.SourceName == "." || s.SourceName == ".." {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN003", Message: "source name must be a basename"}})
		}
		if _, exists := routes[s.SourceName]; exists {
			return Project{}, issues(s.SourceName, []diagnostic.Diagnostic{{Code: "REN003", Message: "ambiguous source basename: " + s.SourceName}})
		}
		routes[s.SourceName] = s.ID
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
		r := renderer{self: s.ID, base: base, routes: routes, anchors: anchors}
		title, _ := s.Document.Metadata["title"].(string)
		var b strings.Builder
		r.nodes(&b, s.Document.Preamble, false)
		for _, section := range s.Document.Sections {
			fmt.Fprintf(&b, `<section class="sheet-section" id="ref-%s"><h2>%s</h2>`, esc(section.ID), r.heading(section.Heading))
			r.nodes(&b, section.Content, false)
			b.WriteString(`<div class="cards">`)
			for _, card := range section.Cards {
				class := fmt.Sprintf("card col-span-%d row-span-%d", max(1, card.Layout.Columns), max(1, card.Layout.Rows))
				for _, c := range card.Classes {
					class += " " + c
				}
				fmt.Fprintf(&b, `<article class="%s" id="ref-%s"><h3>%s <a class="permalink" href="#ref-%s" aria-label="Permalink to %s">#</a></h3>`, esc(class), esc(card.ID), r.heading(card.Heading), esc(card.ID), esc(card.Title))
				r.nodes(&b, card.Content, false)
				b.WriteString(`</article>`)
				index = append(index, entry{card.Title, title, section.Title, plain(card.Content), base + s.ID + "/#ref-" + card.ID})
			}
			b.WriteString(`</div></section>`)
		}
		if source, ok := s.Document.Metadata["source"]; ok {
			disclosure, e := json.MarshalIndent(source, "", "  ")
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
				fmt.Fprintf(&b, `Source details</summary><pre>%s</pre><p>Source attribution is preserved. Import does not grant redistribution rights.</p></details></aside>`, esc(string(disclosure)))
			}
		}
		p.Warnings = append(p.Warnings, issues(s.SourceName, r.warnings)...)
		errors = append(errors, issues(s.SourceName, r.errors)...)
		page := page{Title: title, URL: base + s.ID + "/", Body: b.String()}
		if editBase != "" {
			page.EditURL = editBase + url.PathEscape(s.SourceName)
		}
		pages = append(pages, page)
		if e := putJSON("data/sheets/"+s.ID+".json", page); e != nil {
			errors = append(errors, Issue{s.SourceName, diagnostic.Diagnostic{Code: "REN001", Message: e.Error()}})
		}
		ir, e := s.Document.JSON()
		if e != nil {
			errors = append(errors, Issue{s.SourceName, diagnostic.Diagnostic{Code: "REN001", Message: e.Error()}})
		} else {
			p.Files["data/ir/"+s.ID+".json"] = ir
		}
		front, _ := json.Marshal(map[string]any{"title": title, "url": "/" + s.ID + "/", "sheetid": s.ID})
		p.Files["content/"+s.ID+"/index.md"] = []byte("---\n" + string(front) + "\n---\n")
		// Preamble-only sheets must also be discoverable by search.
		index = append(index, entry{title, title, "", plain(s.Document.Preamble), page.URL})
	}
	if len(errors) > 0 {
		return Project{}, errors
	}
	title := options.Title
	if title == "" {
		title = "SheetMill References"
	}
	config := map[string]any{"baseURL": "https://example.invalid/", "title": title, "languageCode": "en", "disableKinds": []string{"taxonomy", "term", "RSS", "sitemap"}, "enableRobotsTXT": false, "disableHugoGeneratorInject": true, "params": map[string]any{"base": base}, "outputs": map[string]any{"home": []string{"HTML"}, "page": []string{"HTML"}}, "security": map[string]any{"enableInlineShortcodes": false, "exec": map[string]any{"allow": []string{"none"}, "osEnv": []string{"none"}}, "http": map[string]any{"methods": []string{"none"}, "urls": []string{"none"}}}}
	if err := putJSON("hugo.json", config); err != nil {
		panic(err)
	} // strings and literals only
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
	for _, e := range []error{putJSON("data/library.json", pages), putJSON("static/search.json", index), copyTree(themeFiles, "theme", ""), copyTree(brandFiles, ".", "static/brand")} {
		if e != nil {
			return Project{}, issues("", []diagnostic.Diagnostic{{Code: "REN001", Message: "project assembly failed: " + e.Error()}})
		}
	}
	return p, nil
}

type renderer struct {
	inLink           bool
	routes           map[string]string
	self, base       string
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
		wrap("tr")
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
		if id, ok := r.routes[path.Clean(u.Path)]; ok && !routed {
			u.Path, target, routed = r.base+id+"/", id, true
		}
		unresolved := func(message string) {
			r.warnings = append(r.warnings, diagnostic.Diagnostic{Code: "REN103", Message: message + destination, Span: n.Span})
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
func githubEditBase(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", fmt.Errorf("edit base must be an HTTPS github.com repository edit URL without credentials, query or fragment")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "edit" {
		return "", fmt.Errorf("edit base must name /owner/repository/edit/ref/ and an optional sheet directory")
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.Contains(part, "\\") {
			return "", fmt.Errorf("invalid edit base path")
		}
	}
	return strings.TrimRight(u.String(), "/") + "/", nil
}
