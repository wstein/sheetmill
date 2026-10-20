// SPDX-License-Identifier: EUPL-1.2
package navi

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// unquoteChoices is a test-only inverse of quoteChoices for each profile. It
// reads the command back into the literal list; nothing is executed.
func unquoteChoices(t *testing.T, profile, command string) []string {
	t.Helper()
	var rest string
	var quote byte
	switch profile {
	case "bash", "zsh", "fish":
		rest = strings.TrimPrefix(command, `printf '%s\n' `)
		quote = '\''
	case "nu":
		rest = strings.TrimSuffix(strings.TrimPrefix(command, "[ "), " ] | to text")
		quote = '"'
	}
	if rest == command {
		t.Fatalf("%s: unexpected command shape %q", profile, command)
	}
	var out []string
	var cur strings.Builder
	inQuote, started := false, false
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		switch {
		case inQuote && c == quote:
			inQuote = false
		case inQuote && c == '\\' && profile != "bash" && profile != "zsh":
			i++
			cur.WriteByte(rest[i])
		case inQuote:
			cur.WriteByte(c)
		case c == quote:
			inQuote, started = true, true
		case c == '\\' && (profile == "bash" || profile == "zsh"):
			i++
			cur.WriteByte(rest[i])
		case c == ' ':
			out = append(out, cur.String())
			cur.Reset()
			started = false
		default:
			t.Fatalf("%s: unexpected byte %q outside quotes in %q", profile, c, command)
		}
	}
	if inQuote {
		t.Fatalf("%s: unterminated quote in %q", profile, command)
	}
	if started {
		out = append(out, cur.String())
	}
	return out
}

func TestChoicesQuotingPerProfile(t *testing.T) {
	cases := []struct{ name, value, posix, fish, nu string }{
		{"plain", "main", `'main'`, `'main'`, `"main"`},
		{"quote", "it's", `'it'\''s'`, `'it\'s'`, `"it's"`},
		{"double quote", `a"b`, `'a"b'`, `'a"b'`, `"a\"b"`},
		{"backslash", `a\b`, `'a\b'`, `'a\\b'`, `"a\\b"`},
		{"dollar", "$HOME", `'$HOME'`, `'$HOME'`, `"$HOME"`},
		{"percent", "%s", `'%s'`, `'%s'`, `"%s"`},
		{"backtick", "`x`", "'`x`'", "'`x`'", "\"`x`\""},
		{"leading dash", "-n", `'-n'`, `'-n'`, `"-n"`},
		{"non-ASCII", "süd", `'süd'`, `'süd'`, `"süd"`},
	}
	for _, shell := range ProfileNames() {
		for _, c := range cases {
			t.Run(shell+"/"+c.name, func(t *testing.T) {
				got := quoteChoices(profiles[shell], []string{c.value})
				var want string
				switch shell {
				case "bash", "zsh":
					want = `printf '%s\n' ` + c.posix
				case "fish":
					want = `printf '%s\n' ` + c.fish
				case "nu":
					want = `[ ` + c.nu + ` ] | to text`
				}
				if got != want {
					t.Fatalf("got %q want %q", got, want)
				}
			})
		}
	}
}

func TestQuoteChoicesRoundTrips(t *testing.T) {
	lists := [][]string{
		{"main", "it's", "süd"},
		{`a\`, `\'`, `'\`, `"`, `\"`, `''`, "x y", "$(id)", "%s", "-n", "-"},
		{"z", "a", "m"}, // authored order, never sorted
	}
	for _, shell := range ProfileNames() {
		for _, list := range lists {
			got := unquoteChoices(t, shell, quoteChoices(profiles[shell], list))
			if !reflect.DeepEqual(got, list) {
				t.Errorf("%s: recovered %q from %q", shell, got, list)
			}
		}
	}
}

// modelVariable is what navi's parser.rs recovers from one variable line.
type modelVariable struct {
	name, command  string
	words          []string // option words after shellwords splitting
	column         int      // -1 when absent
	headerLines    int      // -1 when absent
	multi, prevent bool
}

var (
	modelLine  = regexp.MustCompile(`^\$\s*([^:]+):(.*)`)
	modelWords = regexp.MustCompile(`'([^']*)'|\\.?|[^\s'\\]+`)
)

// parseVariableLine is a pure-Go model of navi's variable-line grammar: the name
// is everything up to the first colon, the remainder is split on every `---` and
// only the first two parts count, options are shell words. A trailing backslash
// continues the line in navi, so the model reports it as not a single line.
func parseVariableLine(line string) (modelVariable, bool) {
	if strings.HasSuffix(line, `\`) {
		return modelVariable{}, false
	}
	m := modelLine.FindStringSubmatch(line)
	if m == nil {
		return modelVariable{}, false
	}
	parts := strings.Split(m[2], "---")
	v := modelVariable{name: strings.TrimSpace(m[1]), command: strings.TrimSpace(parts[0]), column: -1, headerLines: -1}
	if len(parts) < 2 {
		return v, true
	}
	for _, w := range modelWords.FindAllStringSubmatch(parts[1], -1) {
		word := w[0]
		if strings.HasPrefix(word, "'") {
			word = w[1]
		}
		v.words = append(v.words, word)
	}
	for i := 0; i < len(v.words); i++ {
		flag := v.words[i]
		switch flag {
		case "--multi":
			v.multi = true
			continue
		case "--prevent-extra":
			v.prevent = true
			continue
		}
		if i+1 >= len(v.words) {
			break
		}
		i++
		switch flag {
		case "--column", "--header-lines":
			n, err := strconv.ParseUint(v.words[i], 10, 8)
			if err != nil {
				return modelVariable{}, false
			}
			if flag == "--column" {
				v.column = int(n)
			} else {
				v.headerLines = int(n)
			}
		}
	}
	return v, true
}

func TestVariableLineGrammarModel(t *testing.T) {
	cases := []struct {
		name string
		line string
		want modelVariable
	}{
		{"command only", variableLine("branch", "git branch", options{}), modelVariable{name: "branch", command: "git branch", column: -1, headerLines: -1}},
		{"all options", variableLine("f", "ls -l", options{Column: 2, HeaderLines: 1, Delimiter: `\s\s+`, Multi: true, hasColumn: true, hasHeaderLines: true}),
			modelVariable{name: "f", command: "ls -l", words: []string{"--column", "2", "--delimiter", `\s\s+`, "--header-lines", "1", "--multi"}, column: 2, headerLines: 1, multi: true}},
		{"prevent extra", variableLine("f", "ls", options{PreventExtra: true}), modelVariable{name: "f", command: "ls", words: []string{"--prevent-extra"}, column: -1, headerLines: -1, prevent: true}},
		{"header lines zero", variableLine("f", "ls", options{hasHeaderLines: true}), modelVariable{name: "f", command: "ls", words: []string{"--header-lines", "0"}, column: -1, headerLines: 0}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parseVariableLine(c.line)
			if !ok || !reflect.DeepEqual(got, c.want) {
				t.Fatalf("%q parsed as %+v (ok=%v), want %+v", c.line, got, ok, c.want)
			}
		})
	}
}

func TestVariableLineModelRejectsTrailingBackslash(t *testing.T) {
	if _, ok := parseVariableLine(variableLine("a", `echo x \`, options{})); ok {
		t.Fatal("a trailing backslash continues the line in navi; the model must not accept it")
	}
}

func TestParseCardReadsVariableFields(t *testing.T) {
	var problems []string
	bad := func(m string) { problems = append(problems, m) }
	meta := parseCard("c", map[string]any{"params": map[string]any{
		"a": map[string]any{"type": "string", "choices": []any{"x", "y"}},
		"b": map[string]any{"type": "string", "source": "ls", "options": map[string]any{"column": int64(2), "multi": true}},
	}}, bad)
	want := map[string]param{
		"a": {Type: "string", Choices: []string{"x", "y"}},
		"b": {Type: "string", Source: "ls", Options: options{Column: 2, hasColumn: true, Multi: true}},
	}
	if len(problems) > 0 || !reflect.DeepEqual(meta.params, want) {
		t.Fatalf("problems=%v params=%+v", problems, meta.params)
	}
	// The legacy namespace applies the shared rules and reports them as NAV006 text.
	problems = nil
	parseCard("c", map[string]any{"params": map[string]any{"a": map[string]any{"type": "string", "choices": []any{}, "source": "ls"}}}, bad)
	if len(problems) == 0 {
		t.Fatal("invalid variable fields must be reported")
	}
}

// variableSheetSource is the contract's example sheet; only the fence language
// (LANG) differs between profiles. The CLI test reads the same file.
var variableSheetSource = func() string {
	b, err := os.ReadFile(filepath.Join("testdata", "variable-sources", "git.md.tmpl"))
	if err != nil {
		panic(err)
	}
	return string(b)
}()

func variableSheet(t *testing.T, shell string) Sheet {
	t.Helper()
	s := sheet(t, strings.ReplaceAll(variableSheetSource, "LANG", shell))
	s.ID, s.SourceName, s.Path = "git", "git.md", "sheets/git.md"
	return s
}

func readGolden(t *testing.T, shell string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "variable-sources", shell+".cheat"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestVariableSourceGolden(t *testing.T) {
	for _, shell := range ProfileNames() {
		t.Run(shell, func(t *testing.T) {
			p, errs := Render([]Sheet{variableSheet(t, shell)}, Options{Shell: shell})
			if len(errs) > 0 || len(p.Warnings) > 0 {
				t.Fatalf("errors=%v warnings=%v", errs, p.Warnings)
			}
			if got, want := string(p.Files["git.cheat"]), readGolden(t, shell); got != want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestChoicesLineInGolden(t *testing.T) {
	for _, shell := range ProfileNames() {
		t.Run(shell, func(t *testing.T) {
			vars := map[string]modelVariable{}
			for _, line := range strings.Split(readGolden(t, shell), "\n") {
				if strings.HasPrefix(line, "$") {
					v, ok := parseVariableLine(line)
					if !ok {
						t.Fatalf("model rejects %q", line)
					}
					vars[v.name] = v
				}
			}
			if len(vars) != 3 {
				t.Fatalf("model recovered %d variables, want 3", len(vars))
			}
			if got, want := unquoteChoices(t, shell, vars["branch"].command), []string{"main", "it's", "süd"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("choices recovered as %q", got)
			}
			if vars["remote"].column != 1 || vars["remote"].command != "git remote" || !vars["branches"].multi {
				t.Fatalf("options not recovered: %+v %+v", vars["remote"], vars["branches"])
			}
		})
	}
}
