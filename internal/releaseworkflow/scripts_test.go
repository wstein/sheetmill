// SPDX-License-Identifier: EUPL-1.2

// Package releaseworkflow tests the release pipeline of specs/019-release-workflow:
// the workflow structure (workflow_test.go) and the scripts it calls (this file).
// It has no non-test code.
package releaseworkflow

import (
	"bytes"
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
	for name, body := range map[string]string{"LICENSE": "licence\n", "README.md": "readme\n", "examples/sheets/a.md": "x\n"} {
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
