// SPDX-License-Identifier: EUPL-1.2
package navi

import (
	"reflect"
	"strings"
	"testing"
)

// FuzzVariableLine drives arbitrary source, delimiter and choice text through the
// whole renderer. Text the shared rules reject is not exported; text navi's line
// grammar cannot carry warns and skips the card; everything else must come back
// from the grammar model as exactly one variable with the same command and options.
func FuzzVariableLine(f *testing.F) {
	for _, s := range []struct {
		shell uint8
		text  string
	}{{0, "git remote"}, {1, `echo "a`}, {2, "it's"}, {3, "a---b"}, {0, "<x>"}, {1, `a\`}, {2, `x\'y`}, {3, `"\`}, {0, "süd -n %s $HOME"}} {
		f.Add(s.shell, s.text, s.text, s.text)
	}
	f.Fuzz(func(t *testing.T, shell uint8, source, delimiter, choice string) {
		profile := ProfileNames()[int(shell)%len(ProfileNames())]
		for _, byChoice := range []bool{false, true} {
			s := sheet(t, varSource(profile, "a: {type: string, source: x}", "run <a>"))
			a := s.Document.Metadata["cheatsheet"].(map[string]any)["cards"].(map[string]any)["c"].(map[string]any)["params"].(map[string]any)["a"].(map[string]any)
			if byChoice {
				delete(a, "source")
				a["choices"] = []any{choice}
			} else {
				a["source"] = source
				a["options"] = map[string]any{"delimiter": delimiter, "column": int64(2)}
			}
			p, errs := Render([]Sheet{s}, Options{Shell: profile})
			if len(errs) > 0 {
				if errs[0].Code != "CMD001" {
					t.Fatalf("unexpected error %v", errs)
				}
				continue
			}
			text, exported := p.Files["demo.cheat"]
			if !exported {
				if len(p.Warnings) == 0 || (p.Warnings[0].Code != "NAV108" && p.Warnings[0].Code != "NAV109") {
					t.Fatalf("skipped without a variable warning: %v", p.Warnings)
				}
				continue
			}
			var lines []string
			for _, l := range strings.Split(string(text), "\n") {
				if strings.HasPrefix(l, "$") {
					lines = append(lines, l)
				}
			}
			v, ok := parseVariableLine(strings.Join(lines, "\n"))
			if len(lines) != 1 || !ok || v.name != "a" {
				t.Fatalf("model recovered %+v (ok=%v) from %q", v, ok, lines)
			}
			checkEntries(t, string(text))
			if byChoice {
				if got := unquoteChoices(t, profile, v.command); !reflect.DeepEqual(got, []string{choice}) || v.words != nil {
					t.Fatalf("%s: choice %q recovered as %q (%v)", profile, choice, got, v.words)
				}
			} else if v.command != source || !reflect.DeepEqual(v.words, []string{"--column", "2", "--delimiter", delimiter}) || v.column != 2 {
				t.Fatalf("%s: source %q delimiter %q recovered as %+v", profile, source, delimiter, v)
			}
		}
	})
}
