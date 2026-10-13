// SPDX-License-Identifier: EUPL-1.2
package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/frontmatter"
	"github.com/wstein/sheetmill/internal/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const MaxSourceBytes = 8 << 20

var explicitID = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

func Parse(src []byte) (*model.Document, []diagnostic.Diagnostic) {
	source := diagnostic.NewSource(src)
	if len(src) > MaxSourceBytes {
		return nil, []diagnostic.Diagnostic{source.Error("SRC002", "source exceeds 8 MiB", 0, 0)}
	}
	fm, ds := frontmatter.Parse(src)
	if len(ds) > 0 {
		return nil, ds
	}
	d := &model.Document{Schema: 1, Metadata: fm.Metadata, Source: string(fm.Body), Preamble: []model.Node{}, Sections: []model.Section{}}
	root := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(fm.Body))
	cv := converter{fm.Body, fm.BodyOffset, source}
	used := map[string]bool{}
	id := func(title, explicit string, span diagnostic.Span) string {
		if explicit != "" {
			if used[explicit] {
				ds = append(ds, diagnostic.Diagnostic{Code: "ID001", Message: "duplicate content ID: " + explicit, Span: span})
			}
			used[explicit] = true
			return explicit
		}
		base := slug(title)
		candidate := base
		for n := 2; used[candidate]; n++ {
			candidate = fmt.Sprintf("%s-%d", base, n)
		}
		used[candidate] = true
		return candidate
	}
	var section *model.Section
	var card *model.Card
	closeCard := func(end diagnostic.Position) {
		if card != nil {
			card.Span.End = end
			card = nil
		}
	}
	closeSection := func(end diagnostic.Position) {
		closeCard(end)
		if section != nil {
			section.Span.End = end
			section = nil
		}
	}
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		node := cv.node(n)
		if h, ok := n.(*ast.Heading); ok && (h.Level == 2 || h.Level == 3) {
			title, attrs, errors := heading(h, fm.Body, cv, node.Span)
			ds = append(ds, errors...)
			if title == "" && len(errors) == 0 {
				ds = append(ds, diagnostic.Diagnostic{Code: "MD002", Message: "section and card titles must be nonempty", Span: node.Span})
			}
			contentID := id(title, attrs.ID, node.Span)
			if h.Level == 2 {
				closeSection(node.Span.Start)
				if attrs.Layout.Columns != 0 || attrs.Layout.Rows != 0 || len(attrs.Classes) > 0 || len(attrs.Compatibility) > 0 {
					ds = append(ds, diagnostic.Diagnostic{Code: "ATTR001", Message: "section headings support only an explicit #id", Span: node.Span})
				}
				d.Sections = append(d.Sections, model.Section{ID: contentID, Title: title, Span: node.Span, Heading: node, Content: []model.Node{}, Cards: []model.Card{}})
				section = &d.Sections[len(d.Sections)-1]
			} else {
				closeCard(node.Span.Start)
				if section == nil {
					ds = append(ds, diagnostic.Diagnostic{Code: "MD001", Message: "a ### card requires a preceding ## section", Span: node.Span})
					continue
				}
				section.Cards = append(section.Cards, model.Card{ID: contentID, IDExplicit: attrs.ID != "", Title: title, Span: node.Span, Heading: node, Layout: attrs.Layout, Classes: attrs.Classes, Compatibility: attrs.Compatibility, Content: []model.Node{}})
				card = &section.Cards[len(section.Cards)-1]
			}
			continue
		}
		if card != nil {
			card.Content = append(card.Content, node)
		} else if section != nil {
			section.Content = append(section.Content, node)
		} else {
			d.Preamble = append(d.Preamble, node)
		}
	}
	closeSection(source.Position(len(src)))
	if len(ds) > 0 {
		return nil, ds
	}
	return d, nil
}

func slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			dash = false
		} else {
			dash = true
		}
	}
	if b.Len() == 0 {
		return "content"
	}
	return b.String()
}

type attributes struct {
	ID                     string
	Layout                 model.Layout
	Classes, Compatibility []string
}

func heading(h *ast.Heading, body []byte, cv converter, span diagnostic.Span) (string, attributes, []diagnostic.Diagnostic) {
	a := attributes{Classes: []string{}, Compatibility: []string{}}
	title := plain(h, body)
	last, ok := h.LastChild().(*ast.Text)
	if !ok {
		return strings.TrimSpace(title), a, nil
	}
	tail := string(last.Value(body))
	tailStart := last.Segment.Start
	for prev := last.PreviousSibling(); prev != nil; prev = prev.PreviousSibling() {
		t, ok := prev.(*ast.Text)
		if !ok || t.Segment.Stop != tailStart {
			break
		}
		tail = string(t.Value(body)) + tail
		tailStart = t.Segment.Start
	}
	open := strings.LastIndex(tail, "{")
	if open < 0 || escapedPunctuation(tail, open) {
		return strings.TrimSpace(title), a, nil
	}
	suffix := strings.TrimSpace(tail[open:])
	if len(suffix) < 2 || (suffix[1] != '.' && suffix[1] != '#') {
		return strings.TrimSpace(title), a, nil
	}
	off := cv.offset + tailStart + open
	attrSpan := cv.source.Span(off, off+len(suffix))
	bad := func(code, msg string) (string, attributes, []diagnostic.Diagnostic) {
		return "", a, []diagnostic.Diagnostic{{Code: code, Message: msg, Span: attrSpan}}
	}
	for before := 0; before < open; {
		relative := strings.IndexByte(tail[before:open], '{')
		if relative < 0 {
			break
		}
		at := before + relative
		if at+1 < open && !escapedPunctuation(tail, at) && (tail[at+1] == '.' || tail[at+1] == '#') {
			return bad("ATTR001", "use one trailing heading attribute group")
		}
		before = at + 1
	}
	if !strings.HasSuffix(suffix, "}") || strings.Count(suffix, "{") != 1 || strings.Count(suffix, "}") != 1 {
		return bad("ATTR001", "malformed heading attributes; expected {.class #id}")
	}
	seen := map[string]bool{}
	tokens := strings.Fields(suffix[1 : len(suffix)-1])
	for _, token := range tokens {
		key := token
		switch {
		case strings.HasPrefix(token, "#"):
			key = "id"
			if !explicitID.MatchString(token[1:]) {
				return bad("ATTR001", "ID must start with an ASCII letter and contain letters, digits, _ or -")
			}
			a.ID = token[1:]
		case strings.HasPrefix(token, ".col-span-") || strings.HasPrefix(token, ".row-span-"):
			prefix := ".col-span-"
			key = "columns"
			if strings.HasPrefix(token, ".row-span-") {
				prefix = ".row-span-"
				key = "rows"
			}
			v, err := strconv.Atoi(strings.TrimPrefix(token, prefix))
			if err != nil || v < 1 || v > 12 || strconv.Itoa(v) != strings.TrimPrefix(token, prefix) {
				return bad("ATTR001", "span must be a decimal integer from 1 to 12")
			}
			if key == "columns" {
				a.Layout.Columns = v
			} else {
				a.Layout.Rows = v
			}
		case token == ".navi" || token == ".note" || token == ".tip" || token == ".warning":
			a.Classes = append(a.Classes, token[1:])
		case token == ".wrap":
			a.Compatibility = append(a.Compatibility, token[1:])
		default:
			return bad("ATTR001", "unsupported heading attribute: "+token)
		}
		if seen[key] {
			return bad("ATTR002", "duplicate or conflicting heading attribute: "+key)
		}
		seen[key] = true
	}
	title = strings.TrimSpace(strings.TrimSuffix(title, tail[open:]))
	return title, a, nil
}
func escapedPunctuation(s string, index int) bool {
	slashes := 0
	for i := index - 1; i >= 0 && s[i] == '\\'; i-- {
		slashes++
	}
	return slashes%2 == 1
}
func plain(n ast.Node, b []byte) string {
	var out strings.Builder
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		switch x := n.(type) {
		case *ast.Text:
			if x.IsRaw() {
				out.Write(x.Value(b))
			} else {
				out.Write(util.UnescapePunctuations(x.Value(b)))
			}
		case *ast.String:
			out.Write(x.Value)
		case *ast.AutoLink:
			out.Write(x.Label(b))
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			walk(c)
		}
	}
	walk(n)
	return out.String()
}

type converter struct {
	body   []byte
	offset int
	source *diagnostic.Source
}

func (cv converter) node(n ast.Node) model.Node {
	start := max(0, n.Pos())
	end := start
	out := model.Node{Kind: n.Kind().String()}
	props := map[string]any{}
	if n.Type() == ast.TypeBlock {
		lines := n.Lines()
		if lines.Len() > 0 {
			end = lines.At(lines.Len() - 1).Stop
		}
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		child := cv.node(c)
		out.Children = append(out.Children, child)
		end = max(end, child.Span.End.Offset-cv.offset)
	}
	switch x := n.(type) {
	case *ast.Text:
		out.Text = string(x.Value(cv.body))
		start = x.Segment.Start
		end = x.Segment.Stop
		if x.SoftLineBreak() {
			props["soft_break"] = true
		}
		if x.HardLineBreak() {
			props["hard_break"] = true
		}
		if x.IsRaw() {
			props["raw"] = true
		}
	case *ast.String:
		out.Text = string(x.Value)
		props["code"] = x.IsCode()
		props["raw"] = x.IsRaw()
	case *ast.Heading:
		props["level"] = x.Level
	case *ast.FencedCodeBlock:
		out.Text = string(x.Lines().Value(cv.body))
		props["language"] = string(x.Language(cv.body))
		if x.Info != nil {
			props["info"] = string(x.Info.Value(cv.body))
		}
		end = cv.fenceEnd(n, start, end)
	case *ast.CodeBlock:
		out.Text = string(x.Lines().Value(cv.body))
	case *ast.List:
		props["ordered"] = x.IsOrdered()
		props["start"] = x.Start
		props["tight"] = x.IsTight
		props["marker"] = string(x.Marker)
	case *ast.ListItem:
		props["offset"] = x.Offset
	case *ast.Link:
		props["destination"] = string(x.Destination)
		props["title"] = string(x.Title)
	case *ast.Image:
		props["destination"] = string(x.Destination)
		props["title"] = string(x.Title)
	case *ast.AutoLink:
		out.Text = string(x.Label(cv.body))
		destination := string(x.URL(cv.body))
		if x.AutoLinkType == ast.AutoLinkEmail && !strings.HasPrefix(strings.ToLower(destination), "mailto:") {
			destination = "mailto:" + destination
		}
		props["destination"] = destination
		end = start + len(out.Text)
	case *ast.Emphasis:
		props["level"] = x.Level
	case *ast.RawHTML:
		out.Text = string(x.Segments.Value(cv.body))
		if x.Segments.Len() > 0 {
			start = x.Segments.At(0).Start
			end = x.Segments.At(x.Segments.Len() - 1).Stop
		}
	case *ast.HTMLBlock:
		out.Text = string(x.Lines().Value(cv.body))
		if x.HasClosure() {
			out.Text += string(x.ClosureLine.Value(cv.body))
			end = max(end, x.ClosureLine.Stop)
		}
	case *east.Table:
		align := make([]string, len(x.Alignments))
		for i, a := range x.Alignments {
			align[i] = a.String()
		}
		props["alignments"] = align
	case *east.TableCell:
		props["alignment"] = x.Alignment.String()
	case *east.TaskCheckBox:
		props["checked"] = x.IsChecked
		end = start + 3
	}
	if len(props) > 0 {
		out.Properties = props
	}
	out.Span = cv.source.Span(cv.offset+start, cv.offset+end)
	return out
}

// Fence spans include explicit delimiters. Parsed content determines where an
// implicit container closure ends; unrelated later source cannot extend the span.
func (cv converter) fenceEnd(n ast.Node, start, contentEnd int) int {
	if start >= len(cv.body) {
		return contentEnd
	}
	lineEnd := bytes.IndexByte(cv.body[start:], '\n')
	if lineEnd < 0 {
		return len(cv.body)
	}
	opening := strings.TrimSpace(string(cv.body[start : start+lineEnd]))
	if len(opening) == 0 {
		return contentEnd
	}
	marker := opening[0]
	width := 0
	for width < len(opening) && opening[width] == marker {
		width++
	}
	if width < 3 {
		return contentEnd
	}
	end := max(contentEnd, start+lineEnd+1)
	limit := len(cv.body)
	for block := n; block != nil; block = block.Parent() {
		if next := block.NextSibling(); next != nil && next.Pos() >= 0 {
			pos := min(next.Pos(), len(cv.body))
			limit = bytes.LastIndexByte(cv.body[:pos], '\n') + 1
			break
		}
	}
	if end >= limit {
		return end
	}
	// Goldmark's content segments stop immediately before an explicit closing
	// fence. Only that next line can be the delimiter; do not scan beyond it.
	next := len(cv.body)
	if j := bytes.IndexByte(cv.body[end:], '\n'); j >= 0 {
		next = end + j + 1
	}
	next = min(next, limit)
	line := strings.TrimSpace(string(cv.body[end:next]))
	for strings.HasPrefix(line, ">") {
		line = strings.TrimSpace(strings.TrimPrefix(line, ">"))
	}
	count := 0
	for count < len(line) && line[count] == marker {
		count++
	}
	if count >= width && strings.TrimSpace(line[count:]) == "" {
		return next
	}
	return end
}
