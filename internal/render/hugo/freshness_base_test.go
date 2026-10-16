// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/wstein/sheetmill/internal/markdown"
)

// updateFreshnessBase rewrites testdata/hugo/freshness/base-hashes.json. It is a
// reviewed, deliberate action: the file pins the no-field output of the templates
// from before feature 017 (spec SC-003) and is retired once 017 has merged.
var updateFreshnessBase = flag.Bool("update-base", false, "rewrite the 017 no-freshness base hashes")

const freshnessFixtureDir = "../../../testdata/hugo/freshness"

var fingerprintedCSS = regexp.MustCompile(`assets/(site|classic|cards)\.[0-9a-f]+\.css`)

type freshnessBase struct {
	Hugo    string                       `json:"hugo"`
	Options map[string]string            `json:"options"`
	Themes  map[string]map[string]string `json:"themes"`
}

func freshnessFixtureSheets(t *testing.T) []Sheet {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(freshnessFixtureDir, "*.md"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no freshness fixtures: %v", err)
	}
	slices.Sort(paths)
	var sheets []Sheet
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		d, ds := markdown.Parse(src)
		if len(ds) > 0 {
			t.Fatalf("%s: %v", p, ds)
		}
		id := strings.TrimSuffix(filepath.Base(p), ".md")
		sheets = append(sheets, Sheet{ID: id, SourceName: id + ".md", Document: d})
	}
	return sheets
}

var freshnessBaseThemes = []string{"editorial", "classic-compact", "reference-cards"}

func freshnessBaseOptions(theme string) Options {
	return Options{Title: "Freshness base", Description: "Pinned base.", Theme: theme}
}

// pageHashes hashes every generated HTML page with fingerprinted stylesheet names
// normalised, so a template-neutral change keeps the hashes.
func pageHashes(t *testing.T, out string) map[string]string {
	t.Helper()
	hashes := map[string]string{}
	err := filepath.WalkDir(out, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".html") {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		b = fingerprintedCSS.ReplaceAll(b, []byte("assets/$1.css"))
		sum := sha256.Sum256(b)
		rel, _ := filepath.Rel(out, p)
		hashes[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hashes
}

func hugoVersion(t *testing.T, binary string) string {
	t.Helper()
	b, err := exec.Command(binary, "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.Fields(string(b))[1]
}

// TestHugoIntegrationNoFreshnessBase is the one-time 017 guard: without the
// optional fields the rendered pages equal the pre-017 templates (SC-003).
func TestHugoIntegrationNoFreshnessBase(t *testing.T) {
	binary := os.Getenv("SHEETMILL_HUGO")
	if binary == "" {
		t.Skip("SKIPPED, NOT PASSED: set SHEETMILL_HUGO to an installed pinned Hugo to compare the no-freshness base")
	}
	got := freshnessBase{
		Hugo:    hugoVersion(t, binary),
		Options: map[string]string{"title": "Freshness base", "description": "Pinned base."},
		Themes:  map[string]map[string]string{},
	}
	for _, theme := range freshnessBaseThemes {
		p, ds := Render(freshnessFixtureSheets(t), freshnessBaseOptions(theme))
		if len(ds) > 0 {
			t.Fatal(ds)
		}
		got.Themes[theme] = pageHashes(t, hugoBuild(t, binary, p))
	}
	file := filepath.Join(freshnessFixtureDir, "base-hashes.json")
	if *updateFreshnessBase {
		b, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var want freshnessBase
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if want.Hugo != got.Hugo {
		t.Fatalf("base hashes were captured with Hugo %s, running %s", want.Hugo, got.Hugo)
	}
	for _, theme := range freshnessBaseThemes {
		for page, sum := range want.Themes[theme] {
			if got.Themes[theme][page] != sum {
				t.Errorf("%s %s differs from the pre-017 base", theme, page)
			}
		}
		if len(got.Themes[theme]) != len(want.Themes[theme]) {
			t.Errorf("%s page set changed", theme)
		}
	}
}
