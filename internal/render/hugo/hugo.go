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

type Sheet struct {
	ID, SourceName string
	Document       *model.Document
}
type Options struct{ Title string }
type Project struct {
	Files    map[string][]byte
	Warnings []diagnostic.Diagnostic
}
type entry struct {
	Title   string `json:"title"`
	Sheet   string `json:"sheet"`
	Section string `json:"section"`
	Text    string `json:"text"`
	URL     string `json:"url"`
}
type page struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Body  string `json:"body"`
}

var route = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Render does not open files, execute tools or evaluate source Markdown.
// No partial project is returned on a fatal diagnostic.
func Render(sheets []Sheet, options Options) (Project, []diagnostic.Diagnostic) {
	p := Project{Files: map[string][]byte{}}
	if len(sheets) == 0 {
		return Project{}, []diagnostic.Diagnostic{{Code: "REN003", Message: "select at least one sheet"}}
	}
	sheets = append([]Sheet(nil), sheets...)
	sort.Slice(sheets, func(i, j int) bool { return sheets[i].ID < sheets[j].ID })
	routes := map[string]string{}
	ids := map[string]bool{}
	for _, s := range sheets {
		if !route.MatchString(s.ID) || ids[s.ID] || s.ID == "assets" || s.ID == "brand" {
			return Project{}, []diagnostic.Diagnostic{{Code: "REN003", Message: "invalid, reserved or duplicate sheet route: " + s.ID}}
		}
		if s.Document == nil || s.Document.Schema != 1 {
			return Project{}, []diagnostic.Diagnostic{{Code: "REN001", Message: "renderer requires schema-1 IR"}}
		}
		if path.Base(s.SourceName) != s.SourceName || s.SourceName == "" {
			return Project{}, []diagnostic.Diagnostic{{Code: "REN003", Message: "source name must be a basename"}}
		}
		if _, exists := routes[s.SourceName]; exists {
			return Project{}, []diagnostic.Diagnostic{{Code: "REN003", Message: "ambiguous source basename: " + s.SourceName}}
		}
		routes[s.SourceName] = s.ID
		ids[s.ID] = true
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
	var errors []diagnostic.Diagnostic
	for _, s := range sheets {
		r := renderer{routes: routes, anchors: map[string]bool{}}
		for _, section := range s.Document.Sections {
			r.anchors[section.ID] = true
			for _, card := range section.Cards {
				r.anchors[card.ID] = true
			}
		}
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
				fmt.Fprintf(&b, `<article class="%s" id="ref-%s"><h3><a href="#ref-%s">%s</a></h3>`, esc(class), esc(card.ID), esc(card.ID), r.heading(card.Heading))
				r.nodes(&b, card.Content, false)
				b.WriteString(`</article>`)
				index = append(index, entry{card.Title, title, section.Title, plain(card.Content), "/" + s.ID + "/#ref-" + card.ID})
			}
			b.WriteString(`</div></section>`)
		}
		if source, ok := s.Document.Metadata["source"]; ok {
			disclosure, e := json.MarshalIndent(source, "", "  ")
			if e != nil {
				errors = append(errors, diagnostic.Diagnostic{Code: "REN001", Message: e.Error()})
			} else {
				fmt.Fprintf(&b, `<aside class="provenance"><h2>Source &amp; license</h2><pre>%s</pre><p>Source attribution is preserved. Import does not grant redistribution rights.</p></aside>`, esc(string(disclosure)))
			}
		}
		p.Warnings = append(p.Warnings, r.warnings...)
		errors = append(errors, r.errors...)
		page := page{title, "/" + s.ID + "/", b.String()}
		pages = append(pages, page)
		if e := putJSON("data/sheets/"+s.ID+".json", page); e != nil {
			errors = append(errors, diagnostic.Diagnostic{Code: "REN001", Message: e.Error()})
		}
		ir, e := s.Document.JSON()
		if e != nil {
			errors = append(errors, diagnostic.Diagnostic{Code: "REN001", Message: e.Error()})
		} else {
			p.Files["data/ir/"+s.ID+".json"] = ir
		}
		front, _ := json.Marshal(map[string]any{"title": title, "url": page.URL, "sheetid": s.ID})
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
	config := map[string]any{"baseURL": "https://example.invalid/", "title": title, "languageCode": "en", "disableKinds": []string{"taxonomy", "term", "RSS", "sitemap"}, "enableRobotsTXT": false, "disableHugoGeneratorInject": true, "outputs": map[string]any{"home": []string{"HTML"}, "page": []string{"HTML"}}, "security": map[string]any{"enableInlineShortcodes": false, "exec": map[string]any{"allow": []string{"none"}, "osEnv": []string{"none"}}, "http": map[string]any{"methods": []string{"none"}, "urls": []string{"none"}}}}
	if err := putJSON("hugo.json", config); err != nil {
		panic(err)
	} // strings and literals only
	putJSON("data/library.json", pages)
	putJSON("static/search.json", index)
	fs.WalkDir(theme, "theme", func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			p.Files[strings.TrimPrefix(name, "theme/")], e = theme.ReadFile(name)
		}
		return e
	})
	fs.WalkDir(brand.Files, ".", func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			p.Files["static/brand/"+name], e = brand.Files.ReadFile(name)
		}
		return e
	})
	return p, nil
}

type renderer struct {
	routes           map[string]string
	anchors          map[string]bool
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
		if yes("raw") || yes("code") {
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
		fmt.Fprintf(b, `<a href="%s" title="%s">`, esc(destination), title)
		if n.Kind == "AutoLink" {
			b.WriteString(esc(n.Text))
		} else {
			r.nodes(b, n.Children, header)
		}
		b.WriteString("</a>")
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
	invalid := e != nil || strings.HasPrefix(destination, "//") || strings.Contains(destination, "\\") || strings.IndexFunc(destination, func(c rune) bool { return unicode.IsSpace(c) || unicode.IsControl(c) }) >= 0
	if e == nil {
		scheme := strings.ToLower(u.Scheme)
		invalid = invalid || (scheme != "" && scheme != "http" && scheme != "https" && scheme != "mailto")
	}
	if invalid {
		r.errors = append(r.errors, diagnostic.Diagnostic{Code: "REN002", Message: "unsafe URL: " + destination, Span: n.Span})
		return ""
	}
	if u.Scheme == "" {
		if u.Path == "" && r.anchors[u.Fragment] {
			u.Fragment = "ref-" + u.Fragment
		}
		if id, ok := r.routes[u.Path]; ok {
			u.Path = "/" + id + "/"
			if u.Fragment != "" {
				u.Fragment = "ref-" + u.Fragment
			}
		}
	}
	return u.String()
}
func plain(nodes []model.Node) string {
	var b strings.Builder
	var walk func([]model.Node)
	walk = func(ns []model.Node) {
		for _, n := range ns {
			if n.Kind != "RawHTML" && n.Kind != "HTMLBlock" {
				b.WriteString(html.UnescapeString(normal(n.Text)))
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
