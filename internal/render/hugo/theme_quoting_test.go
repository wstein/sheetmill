// SPDX-License-Identifier: EUPL-1.2

package hugo

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// The Pages verifier (.github/scripts/verify-pages.sh, specs/018 FR-7) reads
// double-quoted attributes only, so templates must never write anything else.
var unquotedURLAttribute = regexp.MustCompile(`<[a-zA-Z][^>]* (?:href|src|action)=[^"]`)

func TestUnquotedURLAttributePattern(t *testing.T) {
	for _, bad := range []string{`<a href='/x'>`, `<img src=/x.png>`, `<form method="get" action=/q>`, `<link rel="icon" href={{ .x }}>`} {
		if !unquotedURLAttribute.MatchString(bad) {
			t.Errorf("pattern misses %s", bad)
		}
	}
	for _, good := range []string{`<a href="/x">`, `<img src="{{ site.Params.base }}x.png" alt="">`, `<p>href=x in prose</p>`} {
		if unquotedURLAttribute.MatchString(good) {
			t.Errorf("pattern flags %s", good)
		}
	}
}

func TestThemeTemplatesQuoteURLAttributes(t *testing.T) {
	files := 0
	err := filepath.Walk(filepath.Join("theme", "layouts"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		files++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if m := unquotedURLAttribute.Find(data); m != nil {
			t.Errorf("%s: href, src and action must be double-quoted, found %q", path, m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("no templates scanned")
	}
}
