// SPDX-License-Identifier: EUPL-1.2
package navi

import (
	"reflect"
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
