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

// braille is U+2800, which navi's finder serialization uses as a field separator.
const braille = '⠀'

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
	paramName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
	// placeholder follows navi v2.24.0's VAR_REGEX, \\?<(\w[\w\d\-_]*)>, whose \w is
	// Unicode in Rust's regex crate. \p{N} is a deliberately wider superset of \d.
	placeholder = regexp.MustCompile(`\\?<([\p{L}\p{M}\p{N}\p{Pc}][\p{L}\p{M}\p{N}\p{Pc}-]*)>`)
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
	hashes := map[uint64]string{}
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
		body, entries, warnings, es := renderSheet(s, profile)
		errs = append(errs, es...)
		p.Warnings = append(p.Warnings, warnings...)
		for _, e := range entries {
			// navi keeps only the first entry per identity hash and says nothing about
			// the rest, so a batch with equal hashes would lose cards silently.
			h := itemHash(e.prehash)
			if prior, seen := hashes[h]; seen {
				errs = append(errs, issue(s, "NAV008", fmt.Sprintf("card %q has the same navi identity hash as %s; navi would silently drop one of them", e.card, prior), diagnostic.Span{}))
			} else {
				hashes[h] = s.SourceName + "#" + e.card
			}
		}
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

// entry identifies one exported card for the batch-wide identity check.
type entry struct{ card, prehash string }

// itemHash models navi v2.24.0 common/hash.rs: FNV-1a (64 bit) over the Rust
// string hash input, which is the UTF-8 bytes followed by 0xff.
func itemHash(prehash string) uint64 {
	h := uint64(0x811C9DC5)
	for _, b := range append([]byte(prehash), 0xff) {
		h ^= uint64(b)
		h *= 0x100000001b3
	}
	return h
}

// context is the unique first tag of one card: sheetmill:<path>#<id>!. The final
// ! terminates it so tags and title cannot concatenate into another card's text.
func context(s Sheet, card model.Card) string {
	return "sheetmill:" + encode(firstNonEmpty(s.Path, s.SourceName)) + "#" + encode(card.ID) + "!"
}

// encode percent-escapes every byte outside [A-Za-z0-9_./-].
func encode(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.' || c == '/' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func renderSheet(s Sheet, profile Profile) ([]byte, []entry, []Issue, []Issue) {
	d := s.Document
	var errs, warns []Issue
	bad := func(msg string) { errs = append(errs, issue(s, "NAV006", msg, diagnostic.Span{})) }
	cheat, _ := d.Metadata["cheatsheet"].(map[string]any)
	tool, _ := cheat["tool"].(string)
	var cards map[string]*cardMeta
	if raw, present := cheat["navi"]; present {
		cards = parseMetadata(raw, profile, bad)
	}
	tags := tagList(d.Metadata, tool, bad)
	// Parameters are card-local: every card has its own navi context, so one name
	// may be declared differently in different cards.
	ids := make([]string, 0, len(cards))
	for id := range cards {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if dropped := unexportedMetadata(d.Metadata); len(dropped) > 0 {
		warns = append(warns, issue(s, "NAV107", "metadata not written to the .cheat file (a sidecar is not implemented yet): "+strings.Join(dropped, ", "), diagnostic.Span{}))
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
	var entries []entry
	for _, section := range d.Sections {
		for _, card := range section.Cards {
			if !contains(card.Classes, "navi") {
				continue
			}
			// The context comes first and is unique per card; the sheet's tags follow
			// so navi users can still filter by tool or tag.
			tagLine := strings.Join(append([]string{context(s, card)}, tags...), ", ")
			text, ws := renderCard(s, card, tagLine, profile, cards)
			warns = append(warns, ws...)
			if text != "" {
				out.WriteString("\n" + text)
				entries = append(entries, entry{card.ID, tagLine + oneLine(card.Title)})
			}
		}
	}
	if len(errs) > 0 {
		return nil, nil, warns, errs
	}
	if len(entries) == 0 {
		warns = append(warns, issue(s, "NAV106", "no card was exported, so no .cheat file is written for this sheet", diagnostic.Span{}))
		return nil, nil, warns, nil
	}
	return []byte(out.String()), entries, warns, nil
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
func renderCard(s Sheet, card model.Card, tagLine string, profile Profile, metas map[string]*cardMeta) (string, []Issue) {
	skip := func(code, msg string, span diagnostic.Span) (string, []Issue) {
		return "", []Issue{issue(s, code, fmt.Sprintf("card %q not exported: %s", card.ID, msg), span)}
	}
	if !card.IDExplicit {
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
		if strings.HasPrefix(m[0], `\`) {
			return skip("NAV103", fmt.Sprintf("navi reads \\<%s> as an escaped placeholder; there is no escape for literal placeholders in this version", name), block.Span)
		}
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
	title := oneLine(card.Title)
	if title == "" {
		return skip("NAV102", "the card title is empty once invisible characters are removed; navi skips entries without a description", card.Span)
	}
	if strings.ContainsRune(title, braille) {
		return skip("NAV102", "the card title contains U+2800, navi's finder field separator", card.Span)
	}
	return "; card: " + card.ID + "\n% " + tagLine + "\n# " + title + "\n" + command + "\n", nil
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
		case r == braille:
			return "NAV102", "the command contains U+2800, navi's finder field separator"
		case unicode.Is(unicode.Cf, r) || r == ' ' || r == ' ':
			return "NAV102", fmt.Sprintf("the command contains invisible or bidirectional control U+%04X", r)
		}
	}
	if strings.HasPrefix(command, "navi ") || strings.HasPrefix(command, "navi\t") || command == "navi" {
		return "NAV102", "navi rewrites a leading `navi` word to its own executable path; start the command differently"
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

// unterminatedQuote scans POSIX-like quoting and returns the quote left open, or 0.
// Command substitutions ($( ) and backticks) start a fresh code context, also
// inside double quotes. Only quotes are ever reported: unbalanced parentheses or
// other shell errors are not diagnosed here. Commands with a heredoc are not
// scanned because the payload is free text.
func unterminatedQuote(s string, p Profile) rune {
	rs := []rune(s)
	// scan reads code from i until end of input or the closer of a substitution
	// (')' or '`'); it returns the next index and any quote left open.
	var scan func(i int, closer rune) (int, rune)
	var double func(i int) (int, rune)
	double = func(i int) (int, rune) {
		for i < len(rs) {
			switch {
			case rs[i] == '\\':
				i += 2
			case rs[i] == '"':
				return i + 1, 0
			case rs[i] == '$' && i+1 < len(rs) && rs[i+1] == '(':
				next, open := scan(i+2, ')')
				if open != 0 {
					return next, open
				}
				i = next
			case rs[i] == '`':
				next, open := scan(i+1, '`')
				if open != 0 {
					return next, open
				}
				i = next
			default:
				i++
			}
		}
		return i, '"'
	}
	scan = func(i int, closer rune) (int, rune) {
		depth := 0
		for i < len(rs) {
			c := rs[i]
			switch {
			case c == '\\':
				i += 2
			case c == '\'':
				ansi := p.ansiC && i > 0 && rs[i-1] == '$'
				i++
				for i < len(rs) && rs[i] != '\'' {
					if rs[i] == '\\' && (p.fishQuotes || ansi) {
						i++
					}
					i++
				}
				if i >= len(rs) {
					return i, '\''
				}
				i++
			case c == '"':
				next, open := double(i + 1)
				if open != 0 {
					return next, open
				}
				i = next
			case c == '$' && i+1 < len(rs) && rs[i+1] == '(':
				next, open := scan(i+2, ')')
				if open != 0 {
					return next, open
				}
				i = next
			case c == '`':
				if closer == '`' {
					return i + 1, 0
				}
				next, open := scan(i+1, '`')
				if open != 0 {
					return next, open
				}
				i = next
			case c == '(' && closer == ')':
				depth++
				i++
			case c == ')' && closer == ')':
				if depth == 0 {
					return i + 1, 0
				}
				depth--
				i++
			case c == '#' && (i == 0 || unicode.IsSpace(rs[i-1])):
				for i < len(rs) && rs[i] != '\n' {
					i++
				}
			default:
				i++
			}
		}
		return i, 0
	}
	_, open := scan(0, 0)
	return open
}

// unexportedMetadata lists the frontmatter paths the .cheat file does not carry.
// The export uses title, license, attribution, tags, cheatsheet.tool and the
// cheatsheet.navi namespace; cheatsheet.schema is a format marker.
func unexportedMetadata(metadata map[string]any) []string {
	var dropped []string
	for _, key := range sortedKeys(metadata) {
		switch key {
		case "title", "license", "attribution", "tags":
		case "cheatsheet":
			if cheat, ok := metadata[key].(map[string]any); ok {
				for _, k := range sortedKeys(cheat) {
					if k != "tool" && k != "navi" && k != "schema" {
						dropped = append(dropped, "cheatsheet."+k)
					}
				}
			}
		default:
			dropped = append(dropped, key)
		}
	}
	return dropped
}

func tagList(metadata map[string]any, tool string, bad func(string)) []string {
	var tags []string
	seen := map[string]bool{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || oneLine(v) != v || strings.ContainsRune(v, ',') || strings.ContainsRune(v, braille) {
			bad(fmt.Sprintf("tag %q must be nonempty, single-line and free of commas and U+2800 for navi", v))
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
