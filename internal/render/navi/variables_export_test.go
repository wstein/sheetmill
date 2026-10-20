// SPDX-License-Identifier: EUPL-1.2
package navi

import (
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
