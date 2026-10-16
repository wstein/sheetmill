// SPDX-License-Identifier: EUPL-1.2

package pagesworkflow

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
	"mvdan.cc/sh/v3/syntax"
)

type workflow struct {
	Name        string         `yaml:"name"`
	On          map[string]any `yaml:"on"`
	Permissions map[string]string
	Concurrency map[string]any
	Jobs        map[string]job
}

type job struct {
	Name           string
	RunsOn         string `yaml:"runs-on"`
	TimeoutMinutes int    `yaml:"timeout-minutes"`
	Needs          string
	If             string `yaml:"if"`
	Permissions    map[string]string
	Environment    map[string]string
	Steps          []step
}

type step struct {
	ID               string
	Name             string
	Uses             string
	Shell            string
	Run              string
	With             map[string]any
	If               any    `yaml:"if"`
	ContinueOnError  any    `yaml:"continue-on-error"`
	WorkingDirectory string `yaml:"working-directory"`
	Env              map[string]any
	TimeoutMinutes   any `yaml:"timeout-minutes"`
}

func loadWorkflow(t *testing.T) (workflow, string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "pages.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var w workflow
	if err := yaml.Unmarshal(raw, &w); err != nil {
		t.Fatal(err)
	}
	return w, string(raw)
}

func buildSteps(t *testing.T) []step {
	t.Helper()
	w, _ := loadWorkflow(t)
	j, ok := w.Jobs["build"]
	if !ok {
		t.Fatal("no build job")
	}
	return j.Steps
}

// words renders a shell word without quotes; parameter expansions stay as $NAME.
func words(w *syntax.Word) string {
	var b strings.Builder
	var walk func(parts []syntax.WordPart)
	walk = func(parts []syntax.WordPart) {
		for _, p := range parts {
			switch p := p.(type) {
			case *syntax.Lit:
				b.WriteString(p.Value)
			case *syntax.SglQuoted:
				b.WriteString(p.Value)
			case *syntax.DblQuoted:
				walk(p.Parts)
			case *syntax.ParamExp:
				b.WriteString("$" + p.Param.Value)
			default:
				var sb strings.Builder
				_ = syntax.NewPrinter().Print(&sb, p)
				b.WriteString(sb.String())
			}
		}
	}
	walk(w.Parts)
	return b.String()
}

func parseRun(t *testing.T, run string) *syntax.File {
	t.Helper()
	f, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(run), "run")
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// calls returns every simple command of the script, in source order, as unquoted words.
// Function bodies are listed separately by funcs.
func calls(f *syntax.File) [][]string {
	var out [][]string
	syntax.Walk(f, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.CallExpr:
			if len(n.Args) > 0 {
				var args []string
				for _, a := range n.Args {
					args = append(args, words(a))
				}
				out = append(out, args)
			}
		}
		return true
	})
	return out
}

func funcBody(t *testing.T, f *syntax.File, name string) string {
	t.Helper()
	for _, st := range f.Stmts {
		if fd, ok := st.Cmd.(*syntax.FuncDecl); ok && fd.Name.Value == name {
			var sb strings.Builder
			_ = syntax.NewPrinter().Print(&sb, fd.Body)
			return sb.String()
		}
	}
	t.Fatalf("no function %s", name)
	return ""
}

type export struct {
	out   string
	flags map[string][]string
	args  []string
}

func parseExport(args []string) export {
	e := export{flags: map[string][]string{}}
	rest := args[5:] // go run ./cmd/smill build hugo
	for i := 0; i < len(rest); i++ {
		if strings.HasPrefix(rest[i], "--") {
			e.flags[rest[i]] = append(e.flags[rest[i]], rest[i+1])
			i++
			continue
		}
		e.args = append(e.args, rest[i])
	}
	e.out = strings.Join(e.flags["--out"], "")
	return e
}

func isExport(c []string) bool {
	return len(c) > 5 && strings.Join(c[:4], " ") == "go run ./cmd/smill build" && c[4] == "hugo"
}

func TestPagesWorkflowStructure(t *testing.T) {
	w, raw := loadWorkflow(t)

	// Unchanged: triggers, permissions, concurrency, deploy gating, action pins.
	wantOn := map[string]any{"push": map[string]any{"branches": []any{"main"}}, "workflow_dispatch": nil}
	if !reflect.DeepEqual(w.On, wantOn) {
		t.Errorf("triggers must stay exactly push to main and workflow_dispatch, got %v", w.On)
	}
	if checkout := w.Jobs["build"].Steps[0]; checkout.With["persist-credentials"] != false {
		t.Errorf("checkout must keep persist-credentials: false, got %v", checkout.With)
	}
	if !reflect.DeepEqual(w.Permissions, map[string]string{"contents": "read"}) {
		t.Errorf("top-level permissions changed: %v", w.Permissions)
	}
	if w.Concurrency["group"] != "pages" || w.Concurrency["cancel-in-progress"] != false {
		t.Errorf("concurrency changed: %v", w.Concurrency)
	}
	if len(w.Jobs) != 2 {
		t.Fatalf("want build and deploy jobs, got %d", len(w.Jobs))
	}
	deploy := w.Jobs["deploy"]
	if deploy.Needs != "build" || deploy.If != "github.ref == 'refs/heads/main'" || deploy.Environment["name"] != "github-pages" ||
		!reflect.DeepEqual(deploy.Permissions, map[string]string{"pages": "write", "id-token": "write"}) {
		t.Errorf("deploy gating changed: %+v", deploy)
	}
	wantUses := []string{
		"actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09",
		"actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16",
		"actions/upload-pages-artifact@fc324d3547104276b827a68afc52ff2a11cc49c9",
	}
	var gotUses []string
	for _, s := range w.Jobs["build"].Steps {
		if s.Uses != "" {
			gotUses = append(gotUses, s.Uses)
		}
	}
	if !reflect.DeepEqual(gotUses, wantUses) {
		t.Errorf("action pins changed: %v", gotUses)
	}
	if deploy.Steps[0].Uses != "actions/deploy-pages@368f82528645a54fb793d4d04e342629a3f51346" {
		t.Errorf("deploy action changed: %s", deploy.Steps[0].Uses)
	}

	// The header comment names the three published demo paths.
	var header []string
	for _, line := range strings.Split(raw, "\n") {
		if !strings.HasPrefix(line, "#") {
			break
		}
		header = append(header, line)
	}
	comment := strings.Join(header, "\n")
	for _, path := range []string{"/sheetmill/demo/", "/sheetmill/demo-compact/", "/sheetmill/demo-cards/"} {
		if !strings.Contains(comment, path) {
			t.Errorf("header comment does not name %s:\n%s", path, comment)
		}
	}
}

func TestPagesWorkflowBuildCommands(t *testing.T) {
	steps := buildSteps(t)
	var build *step
	for i := range steps {
		if strings.HasPrefix(steps[i].Name, "Export and build") {
			build = &steps[i]
		}
	}
	if build == nil {
		t.Fatal("no Export and build step")
	}
	if build.Shell != "bash" || !strings.Contains(build.Run, "set -euo pipefail") {
		t.Error("the build step must stay strict bash")
	}
	if strings.Contains(build.Run, "verify-pages") {
		t.Error("verification must be its own step, not the tail of the build step")
	}
	script := parseRun(t, build.Run)

	// All Hugo invocations share one helper with the fixed clock and flags.
	helper := funcBody(t, script, "hugo")
	for _, want := range []string{`--clock 2000-01-01T00:00:00Z`, `--noBuildLock`, `--quiet`, `HUGO_CACHEDIR=`, `sheetmill-hugo/hugo`} {
		if !strings.Contains(helper, want) {
			t.Errorf("hugo helper lacks %s: %s", want, helper)
		}
	}

	// Ordered command list: exports, then the manual build, then guarded mounts.
	var order []string
	exports := map[string]export{}
	for _, c := range calls(script) {
		switch {
		case isExport(c):
			e := parseExport(c)
			exports[e.out] = e
			order = append(order, "export "+e.out)
		case c[0] == "hugo":
			order = append(order, strings.Join(c, " "))
		case c[0] == "test" && len(c) == 3 && c[1] == "!" || c[0] == "test" && len(c) == 4 && c[1] == "!":
			order = append(order, strings.Join(c, " "))
		case c[0] == "cp":
			order = append(order, strings.Join(c, " "))
		}
	}
	tmp := "$RUNNER_TEMP/"
	index := func(item string) int {
		t.Helper()
		for i, o := range order {
			if o == item {
				return i
			}
		}
		t.Errorf("command %q missing from %q", item, order)
		return -1
	}
	prev := -1
	for _, item := range []string{
		"export " + tmp + "manual",
		"hugo " + tmp + "manual " + tmp + "public",
		"test ! -e " + tmp + "public/demo",
		"hugo " + tmp + "demo " + tmp + "public/demo",
		"test ! -e " + tmp + "public/demo-compact",
		"hugo " + tmp + "demo-compact " + tmp + "public/demo-compact",
		"test ! -e " + tmp + "public/demo-cards",
		"hugo " + tmp + "demo-cards " + tmp + "public/demo-cards",
		"cp llms.txt " + tmp + "public/llms.txt",
		"cp docs/agent-guide.md " + tmp + "public/agent-guide.md",
	} {
		i := index(item)
		if i <= prev {
			t.Errorf("%q is out of order (position %d after %d) in %q", item, i, prev, order)
		}
		prev = i
	}
	for _, name := range []string{"demo", "demo-compact", "demo-cards"} {
		if i, j := index("export "+tmp+name), index("hugo "+tmp+name+" "+tmp+"public/"+name); i >= j {
			t.Errorf("%s must be exported before it is built", name)
		}
	}
	if len(exports) != 4 {
		t.Fatalf("want the manual plus three demo exports, got %d", len(exports))
	}

	type want struct {
		theme, base, title string
		links              []string
	}
	demos := map[string]want{
		"demo": {"", "/sheetmill/demo/", "SheetMill Demo",
			[]string{"Compact layout=/sheetmill/demo-compact/", "Cards layout=/sheetmill/demo-cards/"}},
		"demo-compact": {"classic-compact", "/sheetmill/demo-compact/", "SheetMill Demo (compact layout)",
			[]string{"Editorial layout=/sheetmill/demo/", "Cards layout=/sheetmill/demo-cards/"}},
		"demo-cards": {"reference-cards", "/sheetmill/demo-cards/", "SheetMill Demo (cards layout)",
			[]string{"Editorial layout=/sheetmill/demo/", "Compact layout=/sheetmill/demo-compact/"}},
	}
	varying := map[string]bool{"--theme": true, "--base-path": true, "--site-url": true, "--out": true, "--title": true, "--header-link": true}
	var shared map[string][]string
	for name, w := range demos {
		e := exports[tmp+name]
		if !reflect.DeepEqual(e.args, []string{"examples/sheets"}) {
			t.Errorf("%s: source must be examples/sheets only, got %v", name, e.args)
		}
		if got := e.flags["--theme"]; w.theme == "" && got != nil || w.theme != "" && !reflect.DeepEqual(got, []string{w.theme}) {
			t.Errorf("%s: theme flags %v, want %q (the default build passes none)", name, got, w.theme)
		}
		if got := e.flags["--base-path"]; !reflect.DeepEqual(got, []string{w.base}) {
			t.Errorf("%s: base path %v", name, got)
		}
		if got := e.flags["--site-url"]; !reflect.DeepEqual(got, []string{"https://wstein.github.io" + w.base}) {
			t.Errorf("%s: site URL %v must equal the base path on the Pages host", name, got)
		}
		if got := e.flags["--title"]; !reflect.DeepEqual(got, []string{w.title}) {
			t.Errorf("%s: title %v", name, got)
		}
		links := e.flags["--header-link"]
		if !reflect.DeepEqual(links, w.links) {
			t.Errorf("%s: header links %v, want %v", name, links, w.links)
		}
		seen := map[string]bool{}
		for _, l := range links {
			label, url, _ := strings.Cut(l, "=")
			if seen[label] || !strings.HasPrefix(url, "/sheetmill/") || url == w.base {
				t.Errorf("%s: header link %q repeats a label, lacks /sheetmill/ or names its own layout", name, l)
			}
			seen[label] = true
		}
		own := map[string][]string{}
		for flag, v := range e.flags {
			if !varying[flag] {
				own[flag] = v
			}
		}
		if own["--description"] == nil || own["--edit-base-url"] == nil {
			t.Errorf("%s: description and edit base URL must be kept", name)
		}
		if shared == nil {
			shared = own
		} else if !reflect.DeepEqual(shared, own) {
			t.Errorf("%s differs from the other demos in more than the allowed flags: %v vs %v", name, own, shared)
		}
	}
	if len(shared) == 0 {
		t.Error("no shared flags found")
	}

	manual := exports[tmp+"manual"]
	if got := manual.flags["--header-link"]; !reflect.DeepEqual(got, []string{"Example library=/sheetmill/demo/"}) {
		t.Errorf("the manual must keep its Example library link: %v", got)
	}
	if got := manual.flags["--base-path"]; !reflect.DeepEqual(got, []string{"/sheetmill/"}) {
		t.Errorf("manual base path: %v", got)
	}
}

func TestPagesWorkflowVerifyStep(t *testing.T) {
	steps := buildSteps(t)
	verify, build, upload := -1, -1, -1
	for i, s := range steps {
		switch {
		case s.Name == "Verify the Pages tree":
			verify = i
		case strings.HasPrefix(s.Name, "Export and build"):
			build = i
		case strings.HasPrefix(s.Uses, "actions/upload-pages-artifact@"):
			upload = i
		}
	}
	if verify < 0 || build < 0 || upload < 0 {
		t.Fatalf("missing steps: verify %d build %d upload %d", verify, build, upload)
	}
	if !(build < verify && verify+1 == upload) {
		t.Errorf("order must be build (%d), Verify the Pages tree (%d), upload (%d) with nothing between verify and upload", build, verify, upload)
	}
	// The gate must be unconditional and must not be able to fail open.
	for _, i := range []int{build, verify, upload} {
		s := steps[i]
		if s.If != nil || s.ContinueOnError != nil || s.WorkingDirectory != "" || len(s.Env) != 0 || s.TimeoutMinutes != nil {
			t.Errorf("step %q must have no if, continue-on-error, working-directory, env or timeout: %+v", s.Name, s)
		}
	}
	if steps[verify].Shell != "" || steps[verify].Uses != "" {
		t.Errorf("the verify step is a plain run step: %+v", steps[verify])
	}
	got := calls(parseRun(t, steps[verify].Run))
	want := [][]string{{"sh", ".github/scripts/verify-pages.sh", "$RUNNER_TEMP/public"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("verification command %v, want %v", got, want)
	}
	if steps[upload].With["path"] != any("${{ runner.temp }}/public") {
		t.Errorf("upload path changed: %v", steps[upload].With)
	}
}
