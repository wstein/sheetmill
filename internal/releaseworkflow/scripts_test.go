// SPDX-License-Identifier: EUPL-1.2

// Package releaseworkflow tests the release pipeline of specs/019-release-workflow:
// the workflow structure (workflow_test.go) and the scripts it calls (this file).
// It has no non-test code.
package releaseworkflow

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func script(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join(repoRoot(t), ".github", "scripts", name)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("missing script: %v", err)
	}
	return p
}

// need skips (never passes) a test whose tools are missing.
func need(t *testing.T, tools ...string) {
	t.Helper()
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("SKIPPED, NOT PASSED: %s is not available", tool)
		}
	}
}

// gnuTar skips unless tar is GNU tar, which the packaging script requires.
func gnuTar(t *testing.T) {
	t.Helper()
	need(t, "tar", "gzip")
	out, err := exec.Command("tar", "--version").Output()
	if err != nil || !strings.Contains(string(out), "GNU tar") {
		t.Skip("SKIPPED, NOT PASSED: GNU tar is not available")
	}
}

type result struct {
	code   int
	stdout string
	stderr string
}

func run(t *testing.T, dir string, env []string, name string, args ...string) result {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=T", "GIT_AUTHOR_EMAIL=t@example.test",
		"GIT_COMMITTER_NAME=T", "GIT_COMMITTER_EMAIL=t@example.test")
	cmd.Env = append(cmd.Env, env...)
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("%s %v: %v", name, args, err)
		}
		code = ee.ExitCode()
	}
	return result{code, so.String(), se.String()}
}

func sh(t *testing.T, dir string, env []string, name string, args ...string) result {
	t.Helper()
	need(t, "sh")
	return run(t, dir, env, "sh", append([]string{script(t, name)}, args...)...)
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	need(t, "git")
	r := run(t, dir, nil, "git", args...)
	if r.code != 0 {
		t.Fatalf("git %v: %s%s", args, r.stdout, r.stderr)
	}
	return strings.TrimSpace(r.stdout)
}

// newRepo returns a repository with commit subjects on main, oldest first.
func newRepo(t *testing.T, subjects ...string) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	for i, s := range subjects {
		if err := os.WriteFile(filepath.Join(dir, "f"), []byte{byte('a' + i)}, 0o644); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "add", "f")
		git(t, dir, "commit", "-q", "-m", s)
	}
	return dir
}

func TestVerifyVersion(t *testing.T) {
	good := []string{"v0.1.0", "v1.20.300", "v10.0.0"}
	bad := []string{"", "0.1.0", "v1", "v1.2", "v01.2.3", "v1.02.3", "v1.2.03", "v1.2.3-rc1", "v1.2.3+b",
		"v1.2.3\nx", " v1.2.3", "V1.2.3", "v1.2.3.4", "v-1.2.3"}
	for _, v := range good {
		if r := sh(t, repoRoot(t), nil, "release-verify.sh", "version", v); r.code != 0 {
			t.Errorf("version %q rejected: %s", v, r.stderr)
		}
	}
	for _, v := range bad {
		if r := sh(t, repoRoot(t), nil, "release-verify.sh", "version", v); r.code == 0 {
			t.Errorf("version %q accepted", v)
		}
	}
}

func TestVerifyDryRunVersion(t *testing.T) {
	for _, v := range []string{"v0.0.0-dryrun", "v1.2.3-rc.1"} {
		if r := sh(t, repoRoot(t), nil, "release-verify.sh", "dryversion", v); r.code != 0 {
			t.Errorf("dry version %q rejected: %s", v, r.stderr)
		}
	}
	for _, v := range []string{"", "v1.2.3", "1.2.3", "v1.2", "v1.2.3-", "v1.2.3-$(id)", "v1.2.3;x", "v1.2.3 x", "v1.2.3\nv1.2.4", "v1.2.3-a/b"} {
		if r := sh(t, repoRoot(t), nil, "release-verify.sh", "dryversion", v); r.code == 0 {
			t.Errorf("dry version %q accepted", v)
		}
	}
}

func TestVerifyTag(t *testing.T) {
	dir := newRepo(t, "one", "two")
	head := git(t, dir, "rev-parse", "HEAD")
	git(t, dir, "tag", "-a", "-m", "release", "v0.1.0")
	git(t, dir, "tag", "v0.2.0") // lightweight
	git(t, dir, "tag", "-a", "-m", "x", "v0.3.0-rc1")

	if r := sh(t, dir, nil, "release-verify.sh", "tag", "v0.1.0", "main"); r.code != 0 || strings.TrimSpace(r.stdout) != head {
		t.Errorf("annotated tag on main must pass and print the commit: %+v", r)
	}
	if r := sh(t, dir, nil, "release-verify.sh", "tag", "v0.2.0", "main"); r.code == 0 || !strings.Contains(r.stderr, "annotated") {
		t.Errorf("lightweight tag must fail naming 'annotated': %+v", r)
	}
	if r := sh(t, dir, nil, "release-verify.sh", "tag", "v0.3.0-rc1", "main"); r.code == 0 {
		t.Errorf("pre-release tag must fail: %+v", r)
	}
	if r := sh(t, dir, nil, "release-verify.sh", "tag", "v9.9.9", "main"); r.code == 0 {
		t.Errorf("missing tag must fail: %+v", r)
	}

	// A tag on a commit that is not reachable from main.
	git(t, dir, "checkout", "-q", "-b", "side")
	if err := os.WriteFile(filepath.Join(dir, "g"), []byte("g"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "g")
	git(t, dir, "commit", "-q", "-m", "side")
	git(t, dir, "tag", "-a", "-m", "side", "v0.4.0")
	if r := sh(t, dir, nil, "release-verify.sh", "tag", "v0.4.0", "main"); r.code == 0 || !strings.Contains(r.stderr, "main") {
		t.Errorf("tag off main must fail naming main: %+v", r)
	}
	// An annotated tag of a tag object is not a release tag either way; usage errors exit 2.
	if r := sh(t, dir, nil, "release-verify.sh", "tag", "v0.1.0"); r.code != 2 {
		t.Errorf("missing argument must exit 2: %+v", r)
	}
	if r := sh(t, dir, nil, "release-verify.sh", "bogus"); r.code != 2 {
		t.Errorf("unknown subcommand must exit 2: %+v", r)
	}
}

func TestVerifyGoVersionConsistency(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".go-version", "1.27.2\n")
	write("go.mod", "module x\n\ngo 1.27.2\n\nrequire y v1.0.0\n")
	if r := sh(t, dir, nil, "release-verify.sh", "go"); r.code != 0 {
		t.Errorf("matching versions rejected: %+v", r)
	}
	write("go.mod", "module x\n\ngo 1.27.3\n")
	if r := sh(t, dir, nil, "release-verify.sh", "go"); r.code == 0 {
		t.Errorf("mismatch accepted")
	}
	write("go.mod", "module x\n")
	if r := sh(t, dir, nil, "release-verify.sh", "go"); r.code == 0 {
		t.Errorf("missing go line accepted")
	}
	// The real repository must be consistent today.
	if r := sh(t, repoRoot(t), nil, "release-verify.sh", "go"); r.code != 0 {
		t.Errorf("repository .go-version and go.mod disagree: %+v", r)
	}
}

// fakeBinary is a shell stand-in for smill. mode "good" behaves; others misbehave.
func fakeBinary(t *testing.T, dir, version, mode string) string {
	t.Helper()
	body := `#!/bin/sh
case "$1" in
version) echo "smill ` + version + `" ;;
--help) echo "smill help" ;;
check) test -d "$2" ;;
build)
  out=""
  while [ $# -gt 0 ]; do if [ "$1" = --out ]; then out=$2; fi; shift; done
  mkdir "$out"
  echo content > "$out/index.txt"
`
	switch mode {
	case "helpfail":
		body = strings.Replace(body, `--help) echo "smill help" ;;`, `--help) exit 1 ;;`, 1)
	case "checkfail":
		body = strings.Replace(body, `check) test -d "$2" ;;`, `check) exit 1 ;;`, 1)
	case "physical":
		// Like smill's checkedPath, refuse the input and --out paths (made absolute against
		// the logical $PWD) when any component is a symlink (IO003).
		pre := "chk() {\n  case \"$1\" in /*) p=$1 ;; *) p=$PWD/$1 ;; esac\n  while [ \"$p\" != / ] && [ \"$p\" != . ]; do\n    if [ -L \"$p\" ]; then echo \"IO003: symlink paths are unsupported: $p\" >&2; exit 1; fi\n    p=$(dirname \"$p\")\n  done\n}\ncase \"$1\" in check) chk \"$2\" ;; build) chk \"$3\" ;; esac\n"
		body = strings.Replace(body, "case \"$1\" in\nversion)", pre+"case \"$1\" in\nversion)", 1)
		body = strings.Replace(body, "  mkdir \"$out\"\n", "  p=$out\n  while [ \"$p\" != / ] && [ \"$p\" != . ]; do\n    if [ -L \"$p\" ]; then echo \"IO003: symlink paths are unsupported: $p\" >&2; exit 1; fi\n    p=$(dirname \"$p\")\n  done\n  mkdir \"$out\"\n", 1)
	case "random":
		body += "  echo $$ > \"$out/pid.txt\"\n"
	case "empty":
		body += "  rm \"$out/index.txt\"\n"
	}
	body += `  ;;
*) exit 2 ;;
esac
`
	p := filepath.Join(dir, "smill-"+mode)
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// workTree holds the files the packaging script copies and the sheets the smoke test checks.
func workTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range map[string]string{"LICENSE": "license\n", "README.md": "readme\n", "examples/sheets/a.md": "x\n"} {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func pack(t *testing.T, tree, version, binary, out string) result {
	t.Helper()
	gnuTar(t)
	return sh(t, tree, []string{"SOURCE_DATE_EPOCH=1700000000"}, "release-package.sh", version, "linux", "amd64", binary, out)
}

func listing(t *testing.T, archive string) []string {
	t.Helper()
	r := run(t, "", nil, "tar", "-tvzf", archive)
	if r.code != 0 {
		t.Fatal(r.stderr)
	}
	lines := strings.Split(strings.TrimSpace(r.stdout), "\n")
	sort.Slice(lines, func(i, j int) bool {
		a, b := strings.Fields(lines[i]), strings.Fields(lines[j])
		return a[len(a)-1] < b[len(b)-1]
	})
	return lines
}

func TestPackageLayoutAndDeterminism(t *testing.T) {
	tree := workTree(t)
	bin := fakeBinary(t, t.TempDir(), "v0.1.0", "good")
	out1, out2 := filepath.Join(t.TempDir(), "o"), filepath.Join(t.TempDir(), "o")
	if r := pack(t, tree, "v0.1.0", bin, out1); r.code != 0 {
		t.Fatalf("package failed: %+v", r)
	}
	archive := filepath.Join(out1, "smill_0.1.0_linux_amd64.tar.gz")
	lines := listing(t, archive)
	want := []struct{ mode, name string }{
		{"drwxr-xr-x", "smill_0.1.0_linux_amd64/"},
		{"-rw-r--r--", "smill_0.1.0_linux_amd64/LICENSE"},
		{"-rw-r--r--", "smill_0.1.0_linux_amd64/README.md"},
		{"-rwxr-xr-x", "smill_0.1.0_linux_amd64/smill"},
	}
	if len(lines) != len(want) {
		t.Fatalf("unexpected members: %v", lines)
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i], w.mode) || !strings.HasSuffix(lines[i], " "+w.name) || !strings.Contains(lines[i], "0/0") {
			t.Errorf("member %d = %q, want mode %s name %s owner 0/0", i, lines[i], w.mode, w.name)
		}
	}
	sidecar, err := os.ReadFile(archive + ".sha256")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(sidecar), "  smill_0.1.0_linux_amd64.tar.gz\n") || len(strings.Fields(string(sidecar))[0]) != 64 {
		t.Errorf("sidecar = %q", sidecar)
	}

	// Every member carries SOURCE_DATE_EPOCH (1700000000 = 2023-11-14 22:13 UTC), not the file times.
	dates := run(t, "", []string{"TZ=UTC"}, "tar", "-tvzf", archive)
	if n := strings.Count(dates.stdout, " 2023-11-14 22:13 "); n != 4 {
		t.Errorf("want all 4 members dated from SOURCE_DATE_EPOCH, got %d:\n%s", n, dates.stdout)
	}
	// Same inputs with other file times and another umask give the same bytes.
	for _, f := range []string{"LICENSE", "README.md"} {
		if err := os.Chtimes(filepath.Join(tree, f), time.Unix(1600000000, 0), time.Unix(1600000000, 0)); err != nil {
			t.Fatal(err)
		}
	}
	// Members are stored in sorted name order, whatever order the files were created in.
	raw := run(t, "", nil, "tar", "-tzf", archive)
	wantOrder := "smill_0.1.0_linux_amd64/\nsmill_0.1.0_linux_amd64/LICENSE\nsmill_0.1.0_linux_amd64/README.md\nsmill_0.1.0_linux_amd64/smill\n"
	if raw.stdout != wantOrder {
		t.Errorf("members are not stored in sorted order:\n%s", raw.stdout)
	}
	if err := os.Chtimes(bin, time.Unix(1500000000, 0), time.Unix(1500000000, 0)); err != nil {
		t.Fatal(err)
	}
	gnuTar(t)
	if r := run(t, tree, []string{"SOURCE_DATE_EPOCH=1700000000"}, "sh", "-c", "umask 077; exec sh \"$0\" v0.1.0 linux amd64 \"$1\" \"$2\"", script(t, "release-package.sh"), bin, out2); r.code != 0 {
		t.Fatalf("second package failed: %+v", r)
	}
	a, _ := os.ReadFile(archive)
	b, _ := os.ReadFile(filepath.Join(out2, "smill_0.1.0_linux_amd64.tar.gz"))
	if !bytes.Equal(a, b) || len(a) == 0 {
		t.Error("archives differ between runs")
	}
}

func TestPackageRefusals(t *testing.T) {
	tree := workTree(t)
	bin := fakeBinary(t, t.TempDir(), "v0.1.0", "good")
	out := filepath.Join(t.TempDir(), "o")
	if r := pack(t, tree, "0.1.0", bin, out); r.code == 0 {
		t.Error("version without v accepted")
	}
	if r := pack(t, tree, "v0.1.0\nzz", bin, out); r.code == 0 {
		t.Error("version with a newline accepted")
	}
	if r := pack(t, tree, "v0.1.0", filepath.Join(tree, "missing"), out); r.code == 0 {
		t.Error("missing binary accepted")
	}
	if r := sh(t, tree, nil, "release-package.sh", "v0.1.0", "linux", "amd64", bin, out); r.code == 0 {
		t.Error("missing SOURCE_DATE_EPOCH accepted")
	}
	if err := os.Remove(filepath.Join(tree, "LICENSE")); err != nil {
		t.Fatal(err)
	}
	if r := pack(t, tree, "v0.1.0", bin, out); r.code == 0 {
		t.Error("missing LICENSE accepted")
	}
}

func smoke(t *testing.T, version, archive, root string) result {
	t.Helper()
	return sh(t, t.TempDir(), nil, "release-smoke.sh", version, archive, root)
}

// The default dispatch label must survive packaging and the native smoke test, or the
// dry run could never work.
func TestDryRunLabelPackagesAndSmokes(t *testing.T) {
	tree := workTree(t)
	out := filepath.Join(t.TempDir(), "o")
	bin := fakeBinary(t, t.TempDir(), "v0.0.0-dryrun", "good")
	if r := pack(t, tree, "v0.0.0-dryrun", bin, out); r.code != 0 {
		t.Fatalf("dry-run label not packaged: %+v", r)
	}
	if r := smoke(t, "v0.0.0-dryrun", filepath.Join(out, "smill_0.0.0-dryrun_linux_amd64.tar.gz"), tree); r.code != 0 {
		t.Fatalf("dry-run archive not smoke-tested: %+v", r)
	}
	raw := readFile(t, ".github/workflows/release.yml")
	w := parse(t, raw)
	in := w.On["workflow_dispatch"].(map[string]any)["inputs"].(map[string]any)["version"].(map[string]any)
	if r := sh(t, repoRoot(t), nil, "release-verify.sh", "dryversion", in["default"].(string)); r.code != 0 {
		t.Errorf("the dispatch default is not a valid dry-run label: %+v", r)
	}
	if r := sh(t, repoRoot(t), nil, "release-verify.sh", "version", in["default"].(string)); r.code == 0 {
		t.Error("the dispatch default must not be a release version")
	}
}

// macOS runners put TMPDIR under /var, a symlink to /private/var, and smill refuses
// output paths through a symlink. The smoke test must use a physical temp directory.
func TestSmokeUsesAPhysicalTempDir(t *testing.T) {
	tree := workTree(t)
	out := filepath.Join(t.TempDir(), "o")
	if r := pack(t, tree, "v1.2.3", fakeBinary(t, t.TempDir(), "v1.2.3", "physical"), out); r.code != 0 {
		t.Fatalf("package: %+v", r)
	}
	base := t.TempDir()
	real := filepath.Join(base, "real")
	link := filepath.Join(base, "link")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("SKIPPED, NOT PASSED: symlinks unavailable: %v", err)
	}
	archive := filepath.Join(out, "smill_1.2.3_linux_amd64.tar.gz")
	if r := sh(t, t.TempDir(), []string{"TMPDIR=" + link}, "release-smoke.sh", "v1.2.3", archive, tree); r.code != 0 {
		t.Fatalf("smoke failed with TMPDIR through a symlink: %+v", r)
	}
	// The same through a symlinked repository root.
	rootLink := filepath.Join(base, "rootlink")
	if err := os.Symlink(tree, rootLink); err != nil {
		t.Fatal(err)
	}
	if r := sh(t, t.TempDir(), nil, "release-smoke.sh", "v1.2.3", archive, rootLink); r.code != 0 {
		t.Fatalf("smoke failed with a symlinked repository root: %+v", r)
	}
}

func TestSmokeAcceptsAGoodArchiveAndRejectsBadOnes(t *testing.T) {
	tree := workTree(t)
	bins := t.TempDir()
	build := func(version, mode string) string {
		out := filepath.Join(t.TempDir(), "o")
		if r := pack(t, tree, version, fakeBinary(t, bins, version, mode), out); r.code != 0 {
			t.Fatalf("package: %+v", r)
		}
		return filepath.Join(out, "smill_"+strings.TrimPrefix(version, "v")+"_linux_amd64.tar.gz")
	}

	good := build("v1.2.3", "good")
	if r := smoke(t, "v1.2.3", good, tree); r.code != 0 {
		t.Fatalf("good archive rejected: %+v", r)
	}
	if r := smoke(t, "v1.2.4", good, tree); r.code == 0 {
		t.Error("archive with another version accepted")
	}
	if r := smoke(t, "v1.2.3", build("v1.2.3", "random"), tree); r.code == 0 || !strings.Contains(r.stderr, "deterministic") {
		t.Errorf("nondeterministic export accepted or unexplained: %+v", r)
	}
	if r := smoke(t, "v1.2.3", build("v1.2.3", "empty"), tree); r.code == 0 {
		t.Error("empty export accepted")
	}

	// Tampered archive: sidecar no longer matches.
	tampered := build("v1.2.3", "good")
	f, err := os.OpenFile(tampered, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("x")
	_ = f.Close()
	if r := smoke(t, "v1.2.3", tampered, tree); r.code == 0 || !strings.Contains(r.stderr, "checksum") {
		t.Errorf("tampered archive accepted or unexplained: %+v", r)
	}

	// A check failure of the shipped binary.
	if r := smoke(t, "v1.2.3", build("v1.2.3", "checkfail"), tree); r.code == 0 || !strings.Contains(r.stderr, "check") {
		t.Errorf("failing smill check accepted or unexplained: %+v", r)
	}

	if r := smoke(t, "v1.2.3", build("v1.2.3", "helpfail"), tree); r.code == 0 || !strings.Contains(r.stderr, "--help") {
		t.Errorf("failing --help accepted or unexplained: %+v", r)
	}

	// A binary without the executable bit, with a matching checksum.
	noexec := build("v1.2.3", "good")
	ndir := filepath.Dir(noexec)
	if r := run(t, ndir, nil, "tar", "-xzf", noexec); r.code != 0 {
		t.Fatal(r.stderr)
	}
	if err := os.Chmod(filepath.Join(ndir, "smill_1.2.3_linux_amd64", "smill"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := run(t, ndir, nil, "tar", "-czf", noexec, "smill_1.2.3_linux_amd64"); r.code != 0 {
		t.Fatal(r.stderr)
	}
	nsum := run(t, ndir, nil, "sha256sum", filepath.Base(noexec))
	if err := os.WriteFile(noexec+".sha256", []byte(nsum.stdout), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := smoke(t, "v1.2.3", noexec, tree); r.code == 0 || !strings.Contains(r.stderr, "executable") {
		t.Errorf("non-executable smill accepted or unexplained: %+v", r)
	}

	// An archive with an extra member, even with a matching checksum.
	extra := build("v1.2.3", "good")
	dir := filepath.Dir(extra)
	if r := run(t, dir, nil, "tar", "-xzf", extra); r.code != 0 {
		t.Fatal(r.stderr)
	}
	if err := os.WriteFile(filepath.Join(dir, "smill_1.2.3_linux_amd64", "extra"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := run(t, dir, nil, "tar", "-czf", extra, "smill_1.2.3_linux_amd64"); r.code != 0 {
		t.Fatal(r.stderr)
	}
	sum := run(t, dir, nil, "sha256sum", filepath.Base(extra))
	if err := os.WriteFile(extra+".sha256", []byte(sum.stdout), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := smoke(t, "v1.2.3", extra, tree); r.code == 0 || !strings.Contains(r.stderr, "members") {
		t.Errorf("archive with an extra member accepted or unexplained: %+v", r)
	}

	// Missing sidecar.
	nosum := build("v1.2.3", "good")
	if err := os.Remove(nosum + ".sha256"); err != nil {
		t.Fatal(err)
	}
	if r := smoke(t, "v1.2.3", nosum, tree); r.code == 0 {
		t.Error("archive without checksum sidecar accepted")
	}

	// A check failure of the sheets is a failure.
	if err := os.RemoveAll(filepath.Join(tree, "examples")); err != nil {
		t.Fatal(err)
	}
	if r := smoke(t, "v1.2.3", good, tree); r.code == 0 {
		t.Error("missing example sheets accepted")
	}
}

// The smoke script runs the real smill, built here with the release flags, against the
// real example sheets: the same path the native runners take.
func TestSmokeWithTheRealBinary(t *testing.T) {
	gnuTar(t)
	need(t, "go")
	if runtime.GOOS == "windows" {
		t.Skip("SKIPPED, NOT PASSED: shell scripts need a POSIX host")
	}
	root := repoRoot(t)
	bin := filepath.Join(t.TempDir(), "smill")
	r := run(t, root, []string{"CGO_ENABLED=0"}, "go", "build", "-trimpath",
		"-ldflags", "-X github.com/wstein/sheetmill/internal/cli.Version=v9.8.7", "-o", bin, "./cmd/smill")
	if r.code != 0 {
		t.Fatalf("go build: %s", r.stderr)
	}
	out := filepath.Join(t.TempDir(), "o")
	if r := sh(t, root, []string{"SOURCE_DATE_EPOCH=1700000000"}, "release-package.sh", "v9.8.7", "linux", "amd64", bin, out); r.code != 0 {
		t.Fatalf("package: %+v", r)
	}
	archive := filepath.Join(out, "smill_9.8.7_linux_amd64.tar.gz")
	if r := smoke(t, "v9.8.7", archive, root); r.code != 0 {
		t.Fatalf("smoke failed: %+v", r)
	}
	if r := smoke(t, "v9.8.8", archive, root); r.code == 0 {
		t.Error("smoke accepted a wrong version for the real binary")
	}
}

func TestNotes(t *testing.T) {
	dir := newRepo(t, "feat: first", "fix: second")
	git(t, dir, "tag", "-a", "-m", "r", "v0.1.0")
	git(t, dir, "tag", "-a", "-m", "r", "v0.0.9", "HEAD~1") // an older stable tag: not the baseline
	if err := os.WriteFile(filepath.Join(dir, "n"), []byte("n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "n")
	git(t, dir, "commit", "-q", "-m", "feat: third")
	git(t, dir, "tag", "-a", "-m", "r", "v0.1.5-rc1") // a newer pre-release: never the baseline
	if err := os.WriteFile(filepath.Join(dir, "m"), []byte("m"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "m")
	git(t, dir, "commit", "-q", "-m", "fix: fourth")
	git(t, dir, "tag", "-a", "-m", "r", "v0.2.0")
	head := git(t, dir, "rev-parse", "HEAD")

	r := sh(t, dir, nil, "release-notes.sh", "v0.2.0", head, "1.27.2")
	if r.code != 0 {
		t.Fatalf("notes failed: %+v", r)
	}
	for _, want := range []string{"SheetMill v0.2.0", "Commit: " + head, "Go 1.27.2", "linux/amd64, linux/arm64, darwin/arm64",
		"sha256sum --check --ignore-missing SHA256SUMS", "not signed or notarized", "Changes since v0.1.0", "- feat: third (", "- fix: fourth ("} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("notes lack %q:\n%s", want, r.stdout)
		}
	}
	for _, unwanted := range []string{"- fix: second", "First release", "v0.1.5-rc1", "Changes since v0.0.9"} {
		if strings.Contains(r.stdout, unwanted) {
			t.Errorf("notes must not contain %q:\n%s", unwanted, r.stdout)
		}
	}
	lower := strings.ToLower(r.stdout)
	for _, word := range []string{"exciting", "amazing", "powerful", "blazing", "!"} {
		if strings.Contains(lower, word) {
			t.Errorf("notes contain marketing wording %q", word)
		}
	}

	first := newRepo(t, "feat: only")
	git(t, first, "tag", "-a", "-m", "r", "v0.1.0")
	r = sh(t, first, nil, "release-notes.sh", "v0.1.0", git(t, first, "rev-parse", "HEAD"), "1.27.2")
	if r.code != 0 || !strings.Contains(r.stdout, "First release.") || strings.Contains(r.stdout, "- feat: only") {
		t.Errorf("first release notes wrong: %+v", r)
	}
}

// ciExec runs a command and returns its stdout and stderr together; err is non-nil on a non-zero exit.
func ciExec(t *testing.T, dir string, env []string, name string, args ...string) (string, error) {
	t.Helper()
	r := run(t, dir, env, name, args...)
	if r.code != 0 {
		return r.stdout + r.stderr, fmt.Errorf("exit status %d", r.code)
	}
	return r.stdout + r.stderr, nil
}

const ciSHA = "b6ee4d79b6ee4d79b6ee4d79b6ee4d79b6ee4d79"
const otherSHA = "0123456789012345678901234567890123456789"

type ciRun struct{ sha, conclusion, event string }

func ciRuns(rs ...ciRun) string {
	var parts []string
	for _, r := range rs {
		parts = append(parts, `{"head_sha":"`+r.sha+`","conclusion":"`+r.conclusion+`","event":"`+r.event+`"}`)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func ciRunsN(n int, r ciRun) []ciRun {
	out := make([]ciRun, n)
	for i := range out {
		out[i] = r
	}
	return out
}

// ciCase describes one fake API situation. headModes and listModes are consumed one entry
// per head_sha query or per list query ("ok", "empty" for a zero-result answer, "fail" for a
// gh error); once used up the fake answers "ok". ignoreHead models a server that ignores
// the head_sha, event and status filters and returns unfiltered runs.
type ciCase struct {
	name                 string
	runs                 string
	headModes, listModes []string
	ignoreHead           bool
	ok                   bool
	heads, lists, sleeps int
	msg                  string
}

func fakeGH(t *testing.T, c ciCase, extraEnv ...string) (env []string, files func(string) []string) {
	t.Helper()
	need(t, "jq", "sed", "tr")
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("runs.json", c.runs)
	write("head", strings.Join(c.headModes, "\n")+"\n")
	write("list", strings.Join(c.listModes, "\n")+"\n")
	ignore := "0"
	if c.ignoreHead {
		ignore = "1"
	}
	write("ignorehead", ignore)
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	// The fake models the Actions API: it checks the path, applies the head_sha, event and
	// status filters of the query to the runs, pages them and answers the caller's --jq
	// filter with the real jq over {"total_count", "workflow_runs"}.
	fake := `#!/bin/sh
d=` + dir + `
[ "$1" = api ] && [ "$3" = --jq ] || exit 9
url=$2
filter=$4
echo "$url" >> "$d/calls"
path=${url%%\?*}
query=${url#*\?}
if [ "$path" != "repos/o/r/actions/workflows/check.yml/runs" ]; then echo "gh: Not Found (HTTP 404)" >&2; exit 1; fi
param() { printf '%s' "$query" | tr '&' '\n' | sed -n "s/^$1=//p"; }
case $url in
*head_sha=*) q=head ;;
*) q=list ;;
esac
mode=$(sed -n 1p "$d/$q")
sed 1d "$d/$q" > "$d/$q.next" && mv "$d/$q.next" "$d/$q"
case ${mode:-ok} in
fail) echo "gh: HTTP 502 (Bad Gateway)" >&2; exit 1 ;;
empty) body='{"total_count":0,"workflow_runs":[]}' ;;
*)
	body=$(jq -c --arg head "$(param head_sha)" --arg event "$(param event)" --arg status "$(param status)" \
		--argjson per "$(param per_page | grep . || echo 30)" --argjson page "$(param page | grep . || echo 1)" \
		--arg ignore "$(cat "$d/ignorehead")" '
		[.[] | select(($ignore == "1" or $head == "" or .head_sha == $head)
			and ($ignore == "1" or $event == "" or .event == $event)
			and ($ignore == "1" or $status == "" or ($status == "success" and .conclusion == "success")))] as $m
		| {total_count: ($m | length), workflow_runs: $m[($page - 1) * $per : $page * $per]}' "$d/runs.json")
	;;
esac
printf '%s' "$body" | jq -r "$filter"
`
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	// A fake sleep that records its argument keeps the tests fast and shows the pause.
	if err := os.WriteFile(filepath.Join(bin, "sleep"), []byte("#!/bin/sh\necho \"$1\" >> '"+dir+"/sleeps'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	files = func(name string) []string {
		raw, _ := os.ReadFile(filepath.Join(dir, name))
		return strings.Fields(string(raw))
	}
	env = append([]string{"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH"), "REPO=o/r"}, extraEnv...)
	return env, files
}

func ciGreen(t *testing.T, scriptPath string, env []string) (string, error) {
	t.Helper()
	return ciExec(t, repoRoot(t), env, "sh", scriptPath, "ci-green", ciSHA)
}

func ciCases() []ciCase {
	hit := ciRuns(ciRun{ciSHA, "success", "push"})
	other := ciRuns(ciRun{otherSHA, "success", "push"}, ciRun{otherSHA, "success", "push"})
	five := func(m string) []string { return []string{m, m, m, m, m} }
	withHit := func(n int) string {
		rs := ciRunsN(n, ciRun{otherSHA, "success", "push"})
		return ciRuns(append(rs, ciRun{ciSHA, "success", "push"})...)
	}
	const none = "no successful CI run on push for " + ciSHA
	const failing = "kept failing"
	return []ciCase{
		{name: "first answer is enough", runs: hit, ok: true, heads: 1},
		{name: "0,0,2 succeeds on the third", runs: hit, headModes: []string{"empty", "empty", "ok"}, ok: true, heads: 3, sleeps: 2},
		{name: "five empty answers use the fallback", runs: hit, headModes: five("empty"), ok: true, heads: 5, lists: 1, sleeps: 4},
		{name: "five gh errors use the fallback", runs: hit, headModes: five("fail"), ok: true, heads: 5, lists: 1, sleeps: 4},
		{name: "nothing found", runs: other, ok: false, heads: 5, lists: 1, sleeps: 4, msg: none},
		{name: "no runs at all", runs: "[]", ok: false, heads: 5, lists: 1, sleeps: 4, msg: none},
		{name: "server ignores head_sha, returns other commits", runs: other, ignoreHead: true, ok: false, heads: 5, lists: 1, sleeps: 4, msg: none},
		{name: "run of the commit did not succeed", runs: ciRuns(ciRun{ciSHA, "failure", "push"}), ignoreHead: true, ok: false, heads: 5, lists: 1, sleeps: 4, msg: none},
		{name: "run of the commit is not a push run", runs: ciRuns(ciRun{ciSHA, "success", "pull_request"}), ignoreHead: true, ok: false, heads: 5, lists: 1, sleeps: 4, msg: none},
		{name: "every call fails", runs: hit, headModes: five("fail"), listModes: five("fail"), ok: false, heads: 5, lists: 5, sleeps: 8, msg: failing},
		{name: "the API fails once the head query is empty", runs: hit, headModes: five("empty"), listModes: five("fail"), ok: false, heads: 5, lists: 5, sleeps: 8, msg: none},
		{name: "a failed list page is retried", runs: hit, headModes: five("empty"), listModes: []string{"fail", "ok"}, ok: true, heads: 5, lists: 2, sleeps: 5},
		{name: "fallback reads status=success pages only", runs: ciRuns(append(ciRunsN(450, ciRun{otherSHA, "failure", "push"}), ciRun{ciSHA, "success", "push"})...), headModes: five("empty"), ok: true, heads: 5, lists: 1, sleeps: 4},
		{name: "commit within the newest 500 is found on page 5", runs: withHit(450), headModes: five("empty"), ok: true, heads: 5, lists: 5, sleeps: 4},
		{name: "commit beyond the newest 500 is not found", runs: withHit(550), headModes: five("empty"), ok: false, heads: 5, lists: 5, sleeps: 4, msg: none},
	}
}

// ciProblems runs one case against the script at path (CI_SLEEP unset) and lists what deviates.
func ciProblems(t *testing.T, path string, c ciCase) []string {
	t.Helper()
	env, files := fakeGH(t, c)
	out, err := ciGreen(t, path, env)
	var bad []string
	if (err == nil) != c.ok {
		bad = append(bad, "exit: "+fmt.Sprint(err)+" "+out)
	}
	var heads, lists int
	for _, u := range files("calls") {
		if strings.Contains(u, "head_sha=") {
			heads++
			if !strings.Contains(u, "head_sha="+ciSHA+"&") {
				bad = append(bad, "head query for another commit: "+u)
			}
		} else {
			lists++
		}
		if !strings.Contains(u, "event=push") || !strings.Contains(u, "status=success") || !strings.HasPrefix(u, "repos/o/r/actions/workflows/check.yml/runs?") {
			bad = append(bad, "query lacks event=push, status=success or the check.yml path: "+u)
		}
	}
	if heads != c.heads || lists != c.lists {
		bad = append(bad, fmt.Sprintf("%d head and %d list queries, want %d and %d", heads, lists, c.heads, c.lists))
	}
	sleeps := files("sleeps")
	if len(sleeps) != c.sleeps {
		bad = append(bad, fmt.Sprintf("%d pauses, want %d", len(sleeps), c.sleeps))
	}
	for _, sl := range sleeps {
		if sl != "2" {
			bad = append(bad, "pause of "+sl+" seconds, want 2")
		}
	}
	if c.msg != "" && !strings.Contains(out, c.msg) {
		bad = append(bad, "message lacks "+c.msg+": "+out)
	}
	if !c.ok && c.msg == none(c) && strings.Contains(out, "5 attempts") == false {
		bad = append(bad, "message does not say how many attempts were made: "+out)
	}
	return bad
}

func none(c ciCase) string { return "no successful CI run on push for " + ciSHA }

func TestCIGreen(t *testing.T) {
	for _, c := range ciCases() {
		for _, bad := range ciProblems(t, script(t, "release-verify.sh"), c) {
			t.Errorf("%s: %s", c.name, bad)
		}
	}
}

func TestCIGreenRunsUnderDashAndBash(t *testing.T) {
	for _, shell := range []string{"dash", "bash"} {
		if _, err := exec.LookPath(shell); err != nil {
			t.Logf("SKIPPED, NOT PASSED: %s is not installed", shell)
			continue
		}
		env, _ := fakeGH(t, ciCases()[1])
		if out, err := ciExec(t, repoRoot(t), env, shell, script(t, "release-verify.sh"), "ci-green", ciSHA); err != nil {
			t.Errorf("%s: %v\n%s", shell, err, out)
		}
	}
}

func TestCIGreenRejectsABadCommit(t *testing.T) {
	env, _ := fakeGH(t, ciCases()[0])
	for _, bad := range []string{"", "main", ciSHA[:39], ciSHA + "0", "B6EE4D79B6EE4D79B6EE4D79B6EE4D79B6EE4D79"} {
		if _, err := ciExec(t, repoRoot(t), env, "sh", script(t, "release-verify.sh"), "ci-green", bad); err == nil {
			t.Errorf("commit %q accepted", bad)
		}
	}
}

// Each mutant weakens the script once; some case of TestCIGreen must deviate.
func TestCIGreenMutantsAreCaught(t *testing.T) {
	raw, err := os.ReadFile(script(t, "release-verify.sh"))
	if err != nil {
		t.Fatal(err)
	}
	headQ := `"$runs?head_sha=$sha&event=push&status=success"`
	listQ := `"$runs?event=push&status=success&per_page=100&page=$p"`
	mutants := []struct{ name, from, to string }{
		{"no retry", "CI_ATTEMPTS:-5", "CI_ATTEMPTS:-1"},
		{"retry count 0", "CI_ATTEMPTS:-5", "CI_ATTEMPTS:-0"},
		{"no fallback", "pages=${CI_PAGES:-5}", "pages=0"},
		{"fallback uncapped", "pages=${CI_PAGES:-5}", "pages=${CI_PAGES:-50}"},
		{"head filter without head_sha", `select(.head_sha == "'$sha'" and .conclusion == "success" and .event == "push")] | length'`, `select(.conclusion == "success" and .event == "push")] | length'`},
		{"head filter without conclusion", `select(.head_sha == "'$sha'" and .conclusion == "success" and .event == "push")] | length'`, `select(.head_sha == "'$sha'" and .event == "push")] | length'`},
		{"head filter conclusion inverted", `select(.head_sha == "'$sha'" and .conclusion == "success" and .event == "push")] | length'`, `select(.head_sha == "'$sha'" and .conclusion != "success" and .event == "push")] | length'`},
		{"head filter without event", `select(.head_sha == "'$sha'" and .conclusion == "success" and .event == "push")] | length'`, `select(.head_sha == "'$sha'" and .conclusion == "success")] | length'`},
		{"head count is total_count + 1", `hf='[.workflow_runs[] | select(.head_sha == "'$sha'" and .conclusion == "success" and .event == "push")] | length'`, `hf='.total_count + 1'`},
		{"fallback filter without head_sha", `map(select(.head_sha == "'$sha'" and .conclusion`, `map(select(.conclusion`},
		{"fallback filter conclusion inverted", `map(select(.head_sha == "'$sha'" and .conclusion == "success"`, `map(select(.head_sha == "'$sha'" and .conclusion != "success"`},
		{"fallback filter without conclusion", `and .conclusion == "success" and .event == "push")) | length)"'`, `and .event == "push")) | length)"'`},
		{"fallback filter without event", `and .conclusion == "success" and .event == "push")) | length)"'`, `and .conclusion == "success")) | length)"'`},
		{"head query without event=push", headQ, `"$runs?head_sha=$sha&status=success"`},
		{"head query without status=success", headQ, `"$runs?head_sha=$sha&event=push"`},
		{"head query without head_sha", headQ, `"$runs?event=push&status=success"`},
		{"fallback query without event=push", listQ, `"$runs?status=success&per_page=100&page=$p"`},
		{"fallback query without status=success", listQ, `"$runs?event=push&per_page=100&page=$p"`},
		{"other workflow path", `runs="repos/$REPO/actions/workflows/check.yml/runs"`, `runs="repos/$REPO/actions/runs"`},
		{"gh failure counts as success in the head query", `if n=$(gh api "$runs?head_sha`, `if n=1; gh api "$runs?head_sha`},
		{"gh failure counts as success in the fallback", `[ -n "$r" ] || break`, `[ -n "$r" ] || r="1 1"`},
		{"no pause in the head loop", `if [ "$i" -lt "$attempts" ]; then sleep "$sleep_s"; fi`, `:`},
		{"no pause in the page loop", `if [ "$t" -lt "$attempts" ]; then sleep "$sleep_s"; fi`, `:`},
		{"pause default changed", "CI_SLEEP:-2", "CI_SLEEP:-0"},
		{"API failure message lost", `if [ "$api_ok" = 0 ]; then`, `if false; then`},
		{"API failure never noted", "api_ok=1\n\t\t\tif [ \"$n\"", "if [ \"$n\""},
		{"list page error not retried", `[ "$t" -lt "$attempts" ]; then sleep "$sleep_s"; fi
		done`, `[ "$t" -lt "$attempts" ]; then sleep "$sleep_s"; fi
		break
		done`},
	}
	for _, m := range mutants {
		if !strings.Contains(string(raw), m.from) {
			t.Errorf("HARNESS ERROR, NOT A KILL: %q: text not found: %q", m.name, m.from)
			continue
		}
		path := filepath.Join(t.TempDir(), "release-verify.sh")
		if err := os.WriteFile(path, []byte(strings.Replace(string(raw), m.from, m.to, 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		killed := false
		for _, c := range ciCases() {
			if len(ciProblems(t, path, c)) > 0 {
				killed = true
				break
			}
		}
		if !killed {
			t.Errorf("mutant %q survived", m.name)
		}
	}
}
