// SPDX-License-Identifier: EUPL-1.2
package navi

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// varCard builds a sheet with one .command card whose parameters are written as
// YAML flow mappings, one per line, and renders it for the given shell.
func varCard(t *testing.T, shell, params, command string) (Project, []Issue) {
	t.Helper()
	return Render([]Sheet{sheet(t, varSource(shell, params, command))}, Options{Shell: shell})
}

func varSource(shell, params, command string) string {
	src := "---\ntitle: T\nlicense: x\ncheatsheet:\n  cards:\n    c:\n      params:\n"
	for _, line := range strings.Split(params, "\n") {
		src += "        " + line + "\n"
	}
	src += "---\n## S\n\n### C {.command #c}\n\n```" + shell + "\n" + command + "\n```\n"
	return src
}

// variableLinesOf returns the exported `$` lines of the demo sheet.
func variableLinesOf(t *testing.T, p Project) []string {
	t.Helper()
	var lines []string
	for _, l := range strings.Split(string(p.Files["demo.cheat"]), "\n") {
		if strings.HasPrefix(l, "$") {
			lines = append(lines, l)
		}
	}
	return lines
}

func mustRender(t *testing.T, shell, params, command string) Project {
	t.Helper()
	p, errs := varCard(t, shell, params, command)
	if len(errs) > 0 || len(p.Warnings) > 0 {
		t.Fatalf("errors=%v warnings=%v", errs, p.Warnings)
	}
	return p
}

func TestSourceExportedVerbatim(t *testing.T) {
	for _, source := range []string{
		`git branch --format='%(refname:short)'`,
		`ls | awk '{print $1}' # not a comment to navi`,
		`echo "a:b" && printf '%s\n' ü`,
		`cat "$HOME/x" || true`,
	} {
		p := mustRender(t, "bash", "a: {type: string, source: \""+strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(source)+"\"}", "run <a>")
		if got := variableLinesOf(t, p); len(got) != 1 || got[0] != "$ a: "+source {
			t.Errorf("source %q exported as %q", source, got)
		}
	}
}

func TestSourceWithOptionsFixedOrder(t *testing.T) {
	// Authored in reverse order; the export order is fixed.
	p := mustRender(t, "bash", `a: {type: string, source: ls -l, options: {multi: true, header_lines: 1, delimiter: "\\s\\s+", column: 2}}`, "run <a>")
	want := `$ a: ls -l --- --column 2 --delimiter '\s\s+' --header-lines 1 --multi`
	if got := variableLinesOf(t, p); len(got) != 1 || got[0] != want {
		t.Fatalf("got %q want %q", got, want)
	}
	p = mustRender(t, "bash", `a: {type: string, choices: [x, y], options: {prevent_extra: true}}`, "run <a>")
	if got := variableLinesOf(t, p); len(got) != 1 || !strings.HasSuffix(got[0], " --- --prevent-extra") {
		t.Fatalf("got %q", got)
	}
	// Written zero and false: header_lines 0 is kept, false booleans are omitted.
	p = mustRender(t, "bash", `a: {type: string, source: ls, options: {header_lines: 0, multi: false}}`, "run <a>")
	if got := variableLinesOf(t, p); len(got) != 1 || got[0] != "$ a: ls --- --header-lines 0" {
		t.Fatalf("got %q", got)
	}
}

func TestVariableLinesFollowFirstUse(t *testing.T) {
	// Declared alphabetically a, b, c but used c, a, b.
	params := "a: {type: string, source: echo a}\nb: {type: string, source: echo b}\nc: {type: string, choices: [x]}"
	p := mustRender(t, "bash", params, "run <c> <a> <b>")
	got := variableLinesOf(t, p)
	if len(got) != 3 || !strings.HasPrefix(got[0], "$ c:") || !strings.HasPrefix(got[1], "$ a:") || !strings.HasPrefix(got[2], "$ b:") {
		t.Fatalf("got %q", got)
	}
}

func TestRepeatedPlaceholderOneLine(t *testing.T) {
	p := mustRender(t, "bash", "a: {type: string, source: echo a}\nb: {type: string, source: echo b}", "cp <a> <b>\ncp <a> <a>")
	if got := variableLinesOf(t, p); len(got) != 2 || !strings.HasPrefix(got[0], "$ a:") || !strings.HasPrefix(got[1], "$ b:") {
		t.Fatalf("got %q", got)
	}
}

func TestNoVariableLineWithoutSourceOrChoices(t *testing.T) {
	p := mustRender(t, "bash", "a: {type: string, description: free input}\nb: {type: int}", "run <a> <b>")
	if got := variableLinesOf(t, p); len(got) != 0 {
		t.Fatalf("got %q", got)
	}
	if !strings.HasSuffix(string(p.Files["demo.cheat"]), "# C\nrun <a> <b>\n") {
		t.Fatalf("card text changed:\n%s", p.Files["demo.cheat"])
	}
}

// skipped renders one card and asserts it was skipped with one warning of code,
// positioned on the card's code block, naming the parameter.
func skipped(t *testing.T, shell, params, command, code string) {
	t.Helper()
	src := varSource(shell, params, command)
	p, errs := Render([]Sheet{sheet(t, src)}, Options{Shell: shell})
	if len(errs) > 0 {
		t.Fatalf("%s %q: errors %v", shell, params, errs)
	}
	if len(p.Files) != 0 || len(p.Warnings) == 0 || p.Warnings[0].Code != code {
		t.Fatalf("%s %q: files=%d warnings=%v, want a skipped card with %s", shell, params, len(p.Files), p.Warnings, code)
	}
	fence := 1 + strings.Count(src[:strings.Index(src, "```")], "\n")
	if w := p.Warnings[0]; w.Span.Start.Line != fence || !strings.Contains(w.Message, `"a"`) || !strings.Contains(w.Message, `"c"`) {
		t.Fatalf("%s %q: warning %v, want the code block at line %d naming card and parameter", shell, params, w, fence)
	}
}

func TestSourceRejectedSkipsCard(t *testing.T) {
	rows := []struct {
		name, source string
		quoteChecked bool // false: not rejected under nu, which has no quote check
	}{
		{"options separator", "echo x --- y", false},
		{"navi token", "echo <other>", false},
		{"trailing backslash", `echo x \`, false},
		{"unterminated double quote", `echo "x`, true},
		{"unterminated single quote", `echo 'x`, true},
	}
	for _, shell := range ProfileNames() {
		for _, r := range rows {
			t.Run(shell+"/"+r.name, func(t *testing.T) {
				params := "a: {type: string, source: '" + strings.ReplaceAll(r.source, "'", "''") + "'}"
				if r.quoteChecked && shell == "nu" {
					mustRender(t, shell, params, "run <a>")
					return
				}
				skipped(t, shell, params, "run <a>", "NAV108")
			})
		}
	}
}

func TestChoicesRejectedSkipsCard(t *testing.T) {
	for _, choices := range []string{`["a---b"]`, `[ok, "x---"]`, `["<x>"]`, `["pre<name-1>post"]`} {
		for _, shell := range ProfileNames() {
			skipped(t, shell, "a: {type: string, choices: "+choices+"}", "run <a>", "NAV109")
		}
	}
	// Text that merely resembles a token or separator is carried.
	mustRender(t, "bash", `a: {type: string, choices: ["a--b", "<", "<>", "< x>", "a>b"]}`, "run <a>")
}

func TestOptionStringRejected(t *testing.T) {
	for _, delimiter := range []string{`"it's"`, `"a---b"`, `"---"`} {
		skipped(t, "bash", "a: {type: string, source: ls, options: {delimiter: "+delimiter+"}}", "run <a>", "NAV108")
	}
	mustRender(t, "bash", `a: {type: string, source: ls, options: {delimiter: "a--b\\s"}}`, "run <a>")
}

func TestUnusedParamWithSourceStaysNAV104(t *testing.T) {
	p, errs := varCard(t, "bash", "a: {type: string, source: ls}\nb: {type: string, choices: [x]}", "run <a>")
	if len(errs) > 0 || len(p.Files) != 0 || len(p.Warnings) == 0 || p.Warnings[0].Code != "NAV104" || !strings.Contains(p.Warnings[0].Message, `"b"`) {
		t.Fatalf("errors=%v files=%d warnings=%v", errs, len(p.Files), p.Warnings)
	}
}

// checkEntries asserts the line invariants of the data model on .cheat text:
// no blank line inside an entry, every `$` name appears once per card and is
// referenced in the card's command, and every referenced name with a `$` line
// is declared exactly once.
func checkEntries(t *testing.T, text string) {
	t.Helper()
	var entries [][]string
	for _, block := range strings.Split(text, "\n\n") {
		var lines []string
		for _, l := range strings.Split(strings.TrimSuffix(block, "\n"), "\n") {
			if l == "" {
				t.Errorf("blank line inside an entry:\n%s", block)
			}
			lines = append(lines, l)
		}
		if strings.HasPrefix(lines[0], "; card:") {
			entries = append(entries, lines)
		}
	}
	for _, lines := range entries {
		var command []string
		seen := map[string]bool{}
		sawVariable := false
		for _, l := range lines {
			switch {
			case strings.HasPrefix(l, "$"):
				sawVariable = true
				v, ok := parseVariableLine(l)
				if !ok || seen[v.name] {
					t.Errorf("unparsable or repeated variable line %q in %v", l, lines)
				}
				seen[v.name] = true
			case strings.HasPrefix(l, "%"), strings.HasPrefix(l, "#"), strings.HasPrefix(l, ";"):
			case sawVariable:
				t.Errorf("command text after a variable line in %v", lines)
			default:
				command = append(command, l)
			}
		}
		joined := strings.Join(command, "\n")
		for name := range seen {
			if !strings.Contains(joined, "<"+name+">") {
				t.Errorf("variable %q is not referenced in %q", name, joined)
			}
		}
	}
}

func TestEveryDollarLineIsReferencedAndEntryHasNoBlankLine(t *testing.T) {
	for _, shell := range ProfileNames() {
		checkEntries(t, readGolden(t, shell))
	}
	// A deterministic corpus: value shapes x parameter counts x profiles.
	values := []string{"main", "it's", `a\b`, `a"b`, "$HOME", "%s", "`x`", "-n", "süd", "a--b", "<", "x y", "[1]", "~", "#c", ";", "@t"}
	for _, shell := range ProfileNames() {
		for n := 1; n <= 3; n++ {
			var params, command []string
			for i := 0; i < n; i++ {
				name := string(rune('a' + i))
				var items []string
				for _, v := range values[i*3 : i*3+6] {
					items = append(items, `"`+strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v)+`"`)
				}
				params = append(params, name+": {type: string, choices: ["+strings.Join(items, ", ")+"]}")
				command = append(command, "<"+name+">")
			}
			// Use the parameters in reverse order, repeated, across two lines.
			for i, j := 0, len(command)-1; i < j; i, j = i+1, j-1 {
				command[i], command[j] = command[j], command[i]
			}
			p := mustRender(t, shell, strings.Join(params, "\n"), "run "+strings.Join(command, " ")+"\nrun "+command[0])
			checkEntries(t, string(p.Files["demo.cheat"]))
			if got := variableLinesOf(t, p); len(got) != n {
				t.Fatalf("%s n=%d: %d variable lines", shell, n, len(got))
			}
		}
	}
}

func TestSourceAndChoicesBothSetErrorsAtomically(t *testing.T) {
	good := sheet(t, card("bash", "echo <sheet>"))
	good.ID, good.SourceName, good.Path = "good", "good.md", "good.md"
	// The source path rejects the pair earlier; direct IR reaches the renderer.
	bad := sheet(t, varSource("bash", "a: {type: string, source: ls}", "run <a>"))
	a := bad.Document.Metadata["cheatsheet"].(map[string]any)["cards"].(map[string]any)["c"].(map[string]any)["params"].(map[string]any)["a"].(map[string]any)
	a["choices"] = []any{"x"}
	bad.ID, bad.SourceName, bad.Path = "bad", "bad.md", "bad.md"
	p, errs := Render([]Sheet{good, bad}, Options{})
	if len(errs) == 0 || errs[0].Code != "CMD001" || len(p.Files) != 0 || p.SheetCount != 0 || len(p.Warnings) != 0 {
		t.Fatalf("errors=%v files=%d", errs, len(p.Files))
	}
}

// legacyVariableSheet states the same parameters as the contract sheet in the
// legacy cheatsheet.navi.cards namespace with the .navi marker.
func legacyVariableSheet(t *testing.T, source string) Sheet {
	t.Helper()
	s := sheet(t, source)
	cs := s.Document.Metadata["cheatsheet"].(map[string]any)
	cs["navi"] = map[string]any{"cards": cs["cards"]}
	delete(cs, "cards")
	for i := range s.Document.Sections {
		for j := range s.Document.Sections[i].Cards {
			s.Document.Sections[i].Cards[j].Classes = []string{"navi"}
		}
	}
	s.ID, s.SourceName, s.Path = "git", "git.md", "sheets/git.md"
	return s
}

func TestLegacyNamespaceAcceptsVariableSources(t *testing.T) {
	for _, shell := range ProfileNames() {
		t.Run(shell, func(t *testing.T) {
			p, errs := Render([]Sheet{legacyVariableSheet(t, strings.ReplaceAll(variableSheetSource, "LANG", shell))}, Options{Shell: shell})
			if len(errs) > 0 || len(p.Warnings) > 0 {
				t.Fatalf("errors=%v warnings=%v", errs, p.Warnings)
			}
			if got, want := string(p.Files["git.cheat"]), readGolden(t, shell); got != want {
				t.Fatalf("legacy namespace differs from the shared one:\n%s", got)
			}
		})
	}
	for name, fields := range map[string]map[string]any{
		"both":          {"type": "string", "source": "ls", "choices": []any{"x"}},
		"empty choices": {"type": "string", "choices": []any{}},
		"lone options":  {"type": "string", "options": map[string]any{"multi": true}},
		"unknown":       {"type": "string", "source": "ls", "options": map[string]any{"map": "x"}},
	} {
		t.Run(name, func(t *testing.T) {
			s := legacyVariableSheet(t, strings.ReplaceAll(variableSheetSource, "LANG", "bash"))
			cs := s.Document.Metadata["cheatsheet"].(map[string]any)
			cs["navi"].(map[string]any)["cards"].(map[string]any)["fetch-branch"] = map[string]any{"params": map[string]any{"remote": fields, "branch": map[string]any{"type": "string"}}}
			p, errs := Render([]Sheet{s}, Options{})
			if len(errs) == 0 || errs[0].Code != "NAV006" || len(p.Files) != 0 {
				t.Fatalf("errors=%v files=%d", errs, len(p.Files))
			}
		})
	}
}

func TestConflictingVariableDeclarationsAcrossNamespaces(t *testing.T) {
	build := func(shared, legacy string) ([]Issue, Project) {
		src := "---\ntitle: T\nlicense: x\ncheatsheet:\n  cards:\n    c: {params: {a: " + shared + "}}\n  navi:\n    cards:\n      c: {params: {a: " + legacy + "}}\n---\n## S\n\n### C {.command .navi #c}\n\n```bash\nrun <a>\n```\n"
		p, errs := Render([]Sheet{sheet(t, src)}, Options{})
		return errs, p
	}
	same := "{type: string, source: ls, options: {column: 2}}"
	if errs, p := build(same, same); len(errs) > 0 || len(variableLinesOf(t, p)) != 1 {
		t.Fatalf("equal declarations: %v", errs)
	}
	// Absent and empty options are the same declaration.
	if errs, _ := build("{type: string, source: ls}", "{type: string, source: ls, options: {}}"); len(errs) > 0 {
		t.Fatalf("empty options: %v", errs)
	}
	for name, legacy := range map[string]string{
		"source":  "{type: string, source: ls -a, options: {column: 2}}",
		"option":  "{type: string, source: ls, options: {column: 3}}",
		"choices": "{type: string, choices: [x]}",
		"none":    "{type: string}",
	} {
		t.Run(name, func(t *testing.T) {
			errs, p := build(same, legacy)
			if len(errs) == 0 || errs[0].Code != "NAV006" || !strings.Contains(errs[0].Message, "conflicts") || len(p.Files) != 0 {
				t.Fatalf("errors=%v files=%d", errs, len(p.Files))
			}
		})
	}
}

// withoutSources removes the variable fields of the contract sheet's parameters.
func withoutSources(t *testing.T, shell string) Sheet {
	t.Helper()
	s := variableSheet(t, shell)
	cards := s.Document.Metadata["cheatsheet"].(map[string]any)["cards"].(map[string]any)
	for _, card := range cards {
		for _, fields := range card.(map[string]any)["params"].(map[string]any) {
			for _, key := range []string{"choices", "source", "options"} {
				delete(fields.(map[string]any), key)
			}
		}
	}
	return s
}

func TestEntryHashesUnchangedBySources(t *testing.T) {
	for _, shell := range ProfileNames() {
		t.Run(shell, func(t *testing.T) {
			with, errs := Render([]Sheet{variableSheet(t, shell)}, Options{Shell: shell})
			plain, perrs := Render([]Sheet{withoutSources(t, shell)}, Options{Shell: shell})
			if len(errs) > 0 || len(perrs) > 0 {
				t.Fatal(errs, perrs)
			}
			a, b := parseCheat(with.Files["git.cheat"]), parseCheat(plain.Files["git.cheat"])
			if len(a.cards) != 2 || len(b.cards) != 2 || a.dropped != 0 {
				t.Fatalf("entries: %+v %+v", a, b)
			}
			for i := range a.cards {
				if a.cards[i].tags != b.cards[i].tags || a.cards[i].desc != b.cards[i].desc || itemHash(a.cards[i].tags+a.cards[i].desc) != itemHash(b.cards[i].tags+b.cards[i].desc) {
					t.Errorf("identity of card %d changed: %+v vs %+v", i, a.cards[i], b.cards[i])
				}
				if a.cards[i].command != b.cards[i].command {
					t.Errorf("command of card %d changed: %q vs %q", i, a.cards[i].command, b.cards[i].command)
				}
			}
			var wa, wb sidecar
			if json.Unmarshal(with.Files["git.navi.json"], &wa) != nil || json.Unmarshal(plain.Files["git.navi.json"], &wb) != nil || !reflect.DeepEqual(wa.Cards, wb.Cards) {
				t.Fatalf("card identities in the companion changed: %+v vs %+v", wa.Cards, wb.Cards)
			}
		})
	}
	// NAV008 still rejects equal identities when only one of the cards carries
	// sources: the lines are not part of the identity.
	a, b := variableSheet(t, "bash"), withoutSources(t, "bash")
	a.ID, b.ID = "a", "b"
	if p, errs := Render([]Sheet{a, b}, Options{}); len(errs) == 0 || errs[0].Code != "NAV008" || len(p.Files) != 0 {
		t.Fatalf("errors=%v files=%d", errs, len(p.Files))
	}
}

func TestSidecarPreservesVariableSources(t *testing.T) {
	for _, shell := range ProfileNames() {
		s := variableSheet(t, shell)
		p, errs := Render([]Sheet{s}, Options{Shell: shell})
		if len(errs) > 0 {
			t.Fatal(errs)
		}
		var got struct{ Metadata json.RawMessage }
		if err := json.Unmarshal(p.Files["git.navi.json"], &got); err != nil {
			t.Fatal(err)
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, got.Metadata); err != nil {
			t.Fatal(err)
		}
		want, _ := json.Marshal(s.Document.Metadata)
		if !bytes.Equal(compact.Bytes(), want) {
			t.Fatalf("%s: companion metadata differs from the authored metadata", shell)
		}
		for _, text := range []string{`"source":"git remote"`, `"choices":["main","it's","süd"]`, `"options":{"column":1}`, `"options":{"multi":true}`} {
			if !strings.Contains(compact.String(), text) {
				t.Errorf("%s: companion lacks %s", shell, text)
			}
		}
	}
}
