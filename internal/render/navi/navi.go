// SPDX-License-Identifier: EUPL-1.2

// Package navi renders explicitly marked cards as navi .cheat files. It generates
// text only: neither navi, a shell nor any exported command is ever executed.
package navi

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/model"
)

// Version is the export format version recorded in every file header.
const Version = 1

// Profile names the one shell an export targets. navi runs every snippet through
// its single configured shell, so SheetMill validates against one profile per build.
type Profile struct {
	Name string
	// Languages are the fence language tags accepted for the profile.
	Languages []string
	// quotes enables the quote-balance check; nushell has more quoting forms than
	// are checked here, so its commands are not quote-checked.
	quotes bool
	// fishQuotes lets a backslash escape inside single quotes.
	fishQuotes bool
	// ansiC enables $'...' strings whose backslashes escape.
	ansiC bool
}

var profiles = map[string]Profile{
	"bash": {Name: "bash", Languages: []string{"bash", "sh"}, quotes: true, ansiC: true},
	"zsh":  {Name: "zsh", Languages: []string{"zsh"}, quotes: true, ansiC: true},
	"fish": {Name: "fish", Languages: []string{"fish"}, quotes: true, fishQuotes: true},
	"nu":   {Name: "nu", Languages: []string{"nu", "nushell"}},
}

// ProfileNames lists the supported shell profiles in a stable order.
func ProfileNames() []string { return []string{"bash", "zsh", "fish", "nu"} }

type Sheet struct {
	ID, SourceName string
	Path           string
	Document       *model.Document
}
type Options struct{ Shell string }

// Issue associates a diagnostic with the explicitly selected basename, like the
// other renderers; the CLI resolves it to the selected path.
type Issue struct {
	Source string
	diagnostic.Diagnostic
}
type Project struct {
	Files    map[string][]byte
	Warnings []Issue
}

type param struct{ Type, Description string }

var (
	paramName   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
	placeholder = regexp.MustCompile(`<([A-Za-z_][A-Za-z0-9_-]*)>`)
	cardID      = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)
)

// Render validates every sheet before producing any file. Errors (metadata,
// profile, identity) fail the whole export; card problems skip only that card
// with a warning.
func Render(sheets []Sheet, options Options) (Project, []Issue) {
	shell := options.Shell
	if shell == "" {
		shell = "bash"
	}
	profile, ok := profiles[shell]
	if !ok {
		return Project{}, []Issue{{Diagnostic: diagnostic.Diagnostic{Code: "NAV001", Message: fmt.Sprintf("unsupported shell profile %q; use one of %s", shell, strings.Join(ProfileNames(), ", "))}}}
	}
	p := Project{Files: map[string][]byte{}}
	var errs []Issue
	ordered := append([]Sheet(nil), sheets...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	for _, s := range ordered {
		if s.Document == nil || s.Document.Schema != 1 {
			errs = append(errs, issue(s, "REN001", "renderer requires schema-1 IR", diagnostic.Span{}))
			continue
		}
		name := s.ID + ".cheat"
		if _, dup := p.Files[name]; dup {
			errs = append(errs, issue(s, "REN001", "duplicate output file "+name+"; sheet IDs must be unique", diagnostic.Span{}))
			continue
		}
		body, warnings, es := renderSheet(s, profile)
		errs = append(errs, es...)
		p.Warnings = append(p.Warnings, warnings...)
		if len(es) == 0 && body != nil {
			p.Files[name] = body
		}
	}
	if len(errs) > 0 {
		return Project{}, errs
	}
	return p, nil
}

func issue(s Sheet, code, message string, span diagnostic.Span) Issue {
	return Issue{s.SourceName, diagnostic.Diagnostic{Code: code, Message: message, Span: span}}
}

type cardMeta struct{ params map[string]param }

func renderSheet(s Sheet, profile Profile) ([]byte, []Issue, []Issue) {
	d := s.Document
	var errs, warns []Issue
	bad := func(msg string) { errs = append(errs, issue(s, "NAV006", msg, diagnostic.Span{})) }
	cheat, _ := d.Metadata["cheatsheet"].(map[string]any)
	tool, _ := cheat["tool"].(string)
	var cards map[string]*cardMeta
	if raw, present := cheat["navi"]; present {
		cards = parseMetadata(raw, profile, bad)
	}
	tags := tagList(d.Metadata, tool, s.ID, bad)
	// Declarations of one name must agree across a sheet: all its cards share one
	// navi tag context, where a variable name has one definition.
	declared := map[string]param{}
	ids := make([]string, 0, len(cards))
	for id := range cards {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		for _, name := range sortedParams(cards[id].params) {
			if prior, seen := declared[name]; seen && prior != cards[id].params[name] {
				errs = append(errs, issue(s, "NAV007", fmt.Sprintf("parameter %q is declared differently in several cards; one tag context has one definition per name", name), diagnostic.Span{}))
			}
			declared[name] = cards[id].params[name]
		}
	}
	existing := map[string]bool{}
	for _, sec := range d.Sections {
		for _, c := range sec.Cards {
			existing[c.ID] = true
		}
	}
	for _, id := range ids {
		if !existing[id] {
			bad(fmt.Sprintf("cheatsheet.navi.cards.%s refers to no card", id))
		}
	}
	title, _ := d.Metadata["title"].(string)
	var out strings.Builder
	line := func(prefix, value string) { out.WriteString(prefix + oneLine(value) + "\n") }
	line("; Generated by SheetMill smill build navi; format ", fmt.Sprint(Version))
	line("; Source: ", firstNonEmpty(s.Path, s.SourceName))
	line("; Title: ", title)
	line("; Shell: ", profile.Name)
	if v, ok := d.Metadata["attribution"].(string); ok && strings.TrimSpace(v) != "" {
		line("; Attribution: ", v)
	}
	if v, ok := d.Metadata["license"].(string); ok && strings.TrimSpace(v) != "" {
		line("; License: ", v)
	} else {
		line("; License: ", "not stated")
		warns = append(warns, issue(s, "NAV105", "license is not stated; the export header says so", diagnostic.Span{}))
	}
	out.WriteString("% " + strings.Join(tags, ", ") + "\n")
	exported := 0
	for _, section := range d.Sections {
		for _, card := range section.Cards {
			if !contains(card.Classes, "navi") {
				continue
			}
			text, ws := renderCard(s, card, profile, cards)
			warns = append(warns, ws...)
			if text != "" {
				out.WriteString("\n" + text)
				exported++
			}
		}
	}
	if len(errs) > 0 {
		return nil, warns, errs
	}
	if exported == 0 {
		warns = append(warns, issue(s, "NAV106", "no card was exported, so no .cheat file is written for this sheet", diagnostic.Span{}))
		return nil, warns, nil
	}
	return []byte(out.String()), warns, nil
}

func parseMetadata(raw any, profile Profile, bad func(string)) map[string]*cardMeta {
	root, ok := raw.(map[string]any)
	if !ok {
		bad("cheatsheet.navi must be a mapping")
		return nil
	}
	cards := map[string]*cardMeta{}
	for _, key := range sortedKeys(root) {
		switch key {
		case "shell":
			if v, _ := root[key].(string); v != profile.Name {
				bad(fmt.Sprintf("cheatsheet.navi.shell is %v but this build targets %s; navi uses one configured shell, so build each shell separately", root[key], profile.Name))
			}
		case "cards":
			table, ok := root[key].(map[string]any)
			if !ok {
				bad("cheatsheet.navi.cards must be a mapping from card ID to its settings")
				continue
			}
			for _, id := range sortedKeys(table) {
				cards[id] = parseCard(id, table[id], bad)
			}
		default:
			bad("unknown cheatsheet.navi key: " + key)
		}
	}
	return cards
}

func parseCard(id string, raw any, bad func(string)) *cardMeta {
	meta := &cardMeta{params: map[string]param{}}
	if !cardID.MatchString(id) {
		bad(fmt.Sprintf("cheatsheet.navi.cards key %q is not a card ID", id))
	}
	card, ok := raw.(map[string]any)
	if !ok {
		bad(fmt.Sprintf("cheatsheet.navi.cards.%s must be a mapping", id))
		return meta
	}
	for _, key := range sortedKeys(card) {
		if key != "params" {
			bad(fmt.Sprintf("unknown cheatsheet.navi.cards.%s key: %s", id, key))
			continue
		}
		params, ok := card[key].(map[string]any)
		if !ok {
			bad(fmt.Sprintf("cheatsheet.navi.cards.%s.params must be a mapping", id))
			continue
		}
		for _, name := range sortedKeys(params) {
			where := fmt.Sprintf("cheatsheet.navi.cards.%s.params.%s", id, name)
			if !paramName.MatchString(name) {
				bad(where + ": parameter names match [a-z][a-z0-9_]{0,31}")
			}
			fields, ok := params[name].(map[string]any)
			if !ok {
				bad(where + " must be a mapping")
				continue
			}
			var pr param
			for _, f := range sortedKeys(fields) {
				v, isString := fields[f].(string)
				switch {
				case f == "type" && isString && (v == "string" || v == "int" || v == "path"):
					pr.Type = v
				case f == "type":
					bad(where + ".type must be string, int or path (choice lists and dynamic suggestions are not supported yet)")
				case f == "description" && isString && oneLine(v) == v && v != "":
					pr.Description = v
				case f == "description":
					bad(where + ".description must be a nonempty single-line string")
				default:
					bad(where + ": unknown field " + f)
				}
			}
			if pr.Type == "" {
				bad(where + ".type is required")
			}
			meta.params[name] = pr
		}
	}
	return meta
}

// renderCard returns the card's .cheat text, or "" with a warning when the card
// cannot be exported. Nothing is rewritten to make a card pass.
func renderCard(s Sheet, card model.Card, profile Profile, metas map[string]*cardMeta) (string, []Issue) {
	skip := func(code, msg string, span diagnostic.Span) (string, []Issue) {
		return "", []Issue{issue(s, code, fmt.Sprintf("card %q not exported: %s", card.ID, msg), span)}
	}
	if !hasExplicitID(card) {
		return skip("NAV101", "a .navi card needs an explicit {#id}; generated IDs change when titles change", card.Span)
	}
	var blocks []model.Node
	for _, n := range card.Content {
		if n.Kind == "FencedCodeBlock" {
			blocks = append(blocks, n)
		}
	}
	if len(blocks) != 1 {
		return skip("NAV101", fmt.Sprintf("it needs exactly one fenced code block, found %d", len(blocks)), card.Span)
	}
	block := blocks[0]
	language, _ := block.Properties["language"].(string)
	if !contains(profile.Languages, language) {
		return skip("NAV101", fmt.Sprintf("fence language %q is not in the %s profile (%s)", language, profile.Name, strings.Join(profile.Languages, ", ")), block.Span)
	}
	command := strings.TrimRight(block.Text, "\n")
	if code, msg := checkCommand(command, profile); msg != "" {
		return skip(code, msg, block.Span)
	}
	var declared map[string]param
	if meta := metas[card.ID]; meta != nil {
		declared = meta.params
	}
	used := map[string]bool{}
	for _, m := range placeholder.FindAllStringSubmatch(command, -1) {
		name := m[1]
		if !paramName.MatchString(name) {
			return skip("NAV103", fmt.Sprintf("<%s> looks like a navi placeholder but names match [a-z][a-z0-9_]{0,31}", name), block.Span)
		}
		if _, ok := declared[name]; !ok {
			return skip("NAV104", fmt.Sprintf("placeholder <%s> has no declaration under cheatsheet.navi.cards.%s.params", name, card.ID), block.Span)
		}
		used[name] = true
	}
	for _, name := range sortedParams(declared) {
		if !used[name] {
			return skip("NAV104", fmt.Sprintf("parameter %q is declared but never used", name), block.Span)
		}
	}
	return "; card: " + card.ID + "\n# " + oneLine(card.Title) + "\n" + command + "\n", nil
}

// checkCommand applies the verified navi marker rules (parser.rs read_lines reads
// only column-0 markers, even inside fences), byte hygiene and a quote check.
func checkCommand(command string, profile Profile) (string, string) {
	if strings.TrimSpace(command) == "" {
		return "NAV102", "the command is empty"
	}
	for _, r := range command {
		switch {
		case r == '\r':
			return "NAV102", "the command contains a carriage return"
		case r == '\n' || r == '\t':
		case unicode.IsControl(r):
			return "NAV102", fmt.Sprintf("the command contains control character U+%04X", r)
		case unicode.Is(unicode.Cf, r) || r == ' ' || r == ' ':
			return "NAV102", fmt.Sprintf("the command contains invisible or bidirectional control U+%04X", r)
		}
	}
	lines := strings.Split(command, "\n")
	for i, l := range lines {
		switch {
		case l == "":
			return "NAV102", fmt.Sprintf("command line %d is empty; navi turns blank lines into separators", i+1)
		case strings.ContainsRune("%@;#$", rune(l[0])):
			return "NAV102", fmt.Sprintf("command line %d starts with %q, which navi reads as syntax; indent the line or restructure the command", i+1, l[0])
		case strings.HasPrefix(l, "```"):
			return "NAV102", fmt.Sprintf("command line %d starts with a code fence", i+1)
		}
	}
	if strings.HasSuffix(lines[len(lines)-1], `\`) {
		return "NAV102", "the command ends with a line continuation backslash"
	}
	if profile.quotes && !strings.Contains(command, "<<") {
		if open := unterminatedQuote(command, profile); open != 0 {
			return "NAV102", fmt.Sprintf("the command has an unterminated %c quote", open)
		}
	}
	return "", ""
}

// unterminatedQuote scans POSIX-like quoting; it returns the open quote or 0.
// Commands with a heredoc are not scanned because the payload is free text.
func unterminatedQuote(s string, p Profile) rune {
	var quote rune
	ansi := false
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		switch quote {
		case '\'':
			if c == '\\' && (p.fishQuotes || ansi) {
				i++
			} else if c == '\'' {
				quote, ansi = 0, false
			}
		case '"':
			if c == '\\' {
				i++
			} else if c == '"' {
				quote = 0
			}
		default:
			switch c {
			case '\\':
				i++
			case '\'':
				quote = '\''
				ansi = p.ansiC && i > 0 && rs[i-1] == '$'
			case '"':
				quote = '"'
			case '#':
				if i == 0 || unicode.IsSpace(rs[i-1]) {
					for i < len(rs) && rs[i] != '\n' {
						i++
					}
				}
			}
		}
	}
	return quote
}

// hasExplicitID reports whether the heading source carried {#id}; card IDs are
// otherwise generated from the title.
func hasExplicitID(card model.Card) bool {
	var text strings.Builder
	var walk func(model.Node)
	walk = func(n model.Node) {
		text.WriteString(n.Text)
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(card.Heading)
	spaced := strings.Map(func(r rune) rune {
		if r == '{' || r == '}' {
			return ' '
		}
		return r
	}, text.String())
	return contains(strings.Fields(spaced), "#"+card.ID)
}

func tagList(metadata map[string]any, tool, fallback string, bad func(string)) []string {
	var tags []string
	seen := map[string]bool{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || oneLine(v) != v || strings.ContainsRune(v, ',') {
			bad(fmt.Sprintf("tag %q must be nonempty, single-line and free of commas for navi", v))
			return
		}
		if !seen[v] {
			seen[v] = true
			tags = append(tags, v)
		}
	}
	if tool != "" {
		add(tool)
	}
	if list, ok := metadata["tags"].([]any); ok {
		for _, t := range list {
			if v, ok := t.(string); ok {
				add(v)
			}
		}
	}
	if len(tags) == 0 {
		tags = []string{fallback}
	}
	return tags
}

// oneLine replaces every control or line-separator rune with a space so a value
// can never start a new .cheat line.
func oneLine(s string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == ' ' || r == ' ' || unicode.Is(unicode.Cf, r) {
			return ' '
		}
		return r
	}, s))
}
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func sortedParams(m map[string]param) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
