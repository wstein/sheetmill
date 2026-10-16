// SPDX-License-Identifier: EUPL-1.2

// Package pagesworkflow holds tests only: they pin the GitHub Pages workflow and
// run its verification script (.github/scripts/verify-pages.sh) against synthetic
// and real trees (specs/018-layout-paths). It has no non-test code on purpose.
package pagesworkflow

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	origin = "https://wstein.github.io/sheetmill/"
	hash   = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

// repoRoot is the repository root, two levels above this package.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// runScript runs the verifier through sh. A missing sh is a skip, never a pass.
func runScript(t *testing.T, tree string) (int, string) {
	t.Helper()
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("SKIPPED, NOT PASSED: sh is not available to run verify-pages.sh")
	}
	cmd := exec.Command(sh, filepath.Join(repoRoot(t), ".github", "scripts", "verify-pages.sh"), tree)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err = cmd.Run()
	if err == nil {
		return 0, out.String()
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), out.String()
	}
	t.Fatal(err)
	return -1, ""
}

type layout struct {
	dir     string // directory below the public root; empty for the manual
	classic bool
	cards   bool
	links   []string // header-link targets
}

var (
	editorial = layout{dir: "demo", links: []string{"/sheetmill/demo-compact/", "/sheetmill/demo-cards/"}}
	compact   = layout{dir: "demo-compact", classic: true, links: []string{"/sheetmill/demo/", "/sheetmill/demo-cards/"}}
	cards     = layout{dir: "demo-cards", classic: true, cards: true, links: []string{"/sheetmill/demo/", "/sheetmill/demo-compact/"}}
	manual    = layout{links: []string{"/sheetmill/demo/"}}
)

func (l layout) base() string {
	if l.dir == "" {
		return "/sheetmill/"
	}
	return "/sheetmill/" + l.dir + "/"
}

func (l layout) siteURL() string {
	if l.dir == "" {
		return origin
	}
	return origin + l.dir + "/"
}

// page renders one generated-looking HTML page. Only the home page carries the
// hero links; the manual's Example library link is rendered like a header link.
func (l layout) page(slug string, home bool) string {
	var b strings.Builder
	url := l.siteURL() + slug
	b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8">`)
	fmt.Fprintf(&b, `<link rel="canonical" href="%s"><meta property="og:url" content="%s"><meta property="og:image" content="%sbrand/social-preview.png">`, url, url, l.siteURL())
	fmt.Fprintf(&b, `<link rel="stylesheet" href="%sassets/site.%s.css">`, l.base(), hash)
	if l.classic {
		fmt.Fprintf(&b, `<link rel="stylesheet" href="%sassets/classic.%s.css">`, l.base(), hash)
	}
	if l.cards {
		fmt.Fprintf(&b, `<link rel="stylesheet" href="%sassets/cards.%s.css">`, l.base(), hash)
	}
	fmt.Fprintf(&b, `<script src="%sassets/app.%s.js"></script></head>`, l.base(), hash)
	switch {
	case l.cards:
		b.WriteString(`<body class="classic reference-cards">`)
	case l.classic:
		b.WriteString(`<body class="classic">`)
	default:
		b.WriteString("<body>\n")
	}
	if l.classic {
		fmt.Fprintf(&b, `<header class="topbar"><a class="topbar-logo" href="%s"><img src="%sbrand/logo.svg" alt="SheetMill"></a><a class="topbar-link" href="%s">Library</a>`, l.base(), l.base(), l.base())
		for _, link := range l.links {
			fmt.Fprintf(&b, `<a class="topbar-link" href="%s">Layout %s</a>`, link, link)
		}
		b.WriteString(`</header>`)
	} else {
		fmt.Fprintf(&b, `<header class="site-header"><a href="%s"><img src="%sbrand/logo.svg" alt="SheetMill"></a>`, l.base(), l.base())
		b.WriteString(`<nav class="header-links" aria-label="Related sites">`)
		for _, link := range l.links {
			fmt.Fprintf(&b, `<a href="%s">Layout %s</a>`, link, link)
		}
		b.WriteString("</nav></header>")
	}
	if home {
		b.WriteString(`<p class="hero-links">`)
		for _, link := range l.links {
			fmt.Fprintf(&b, `<a href="%s">Layout %s</a>`, link, link)
		}
		b.WriteString(`</p>`)
	}
	fmt.Fprintf(&b, `<main><a href="%s">Home</a></main></body></html>`, l.base())
	return b.String()
}

func (l layout) sitemap(slugs ...string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, s := range slugs {
		fmt.Fprintf(&b, "<url><loc>%s%s</loc></url>\n", l.siteURL(), s)
	}
	b.WriteString("</urlset>\n")
	return b.String()
}

// validTree is a small tree that satisfies every rule: the manual plus the three demos.
func validTree() map[string]string {
	files := map[string]string{
		"llms.txt":        "llms\n",
		"agent-guide.md":  "guide\n",
		"search.json":     "[]",
		"assets/site.css": "a{}",
		"brand/logo.svg":  "<svg/>",
	}
	for _, l := range []layout{manual, editorial, compact, cards} {
		prefix := l.dir
		if prefix != "" {
			prefix += "/"
		}
		files[prefix+"index.html"] = l.page("", true)
		files[prefix+"git/index.html"] = l.page("git/", false)
		files[prefix+"sitemap.xml"] = l.sitemap("", "git/")
		files[prefix+"search.json"] = "[]"
		files[prefix+"assets/site.css"] = "a{}"
		files[prefix+"brand/logo.svg"] = "<svg/>"
	}
	// The manual keeps ordinary pages and may link to the demo homes and pages.
	files["cli/index.html"] = strings.Replace(manual.page("cli/", false), "<main>", `<main><a href="/sheetmill/demo/git/">demo git</a><a href="http://127.0.0.1:1313/">local preview</a>`, 1)
	files["sitemap.xml"] = manual.sitemap("", "git/", "cli/")
	return files
}

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, data := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func replaceIn(files map[string]string, name, old, new string) {
	data, ok := files[name]
	if !ok || !strings.Contains(data, old) {
		panic(fmt.Sprintf("test setup: %s does not contain %q", name, old))
	}
	files[name] = strings.Replace(data, old, new, 1)
}

func TestVerifyPagesTrees(t *testing.T) {
	cases := []struct {
		name string
		edit func(f map[string]string)
		rule int  // the rule the script must report
		pass bool // the tree must be accepted
	}{
		{name: "valid tree", pass: true},
		{name: "loopback prose link is exempt", pass: true, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", "local preview", "http://127.0.0.1:1313/ preview")
		}},

		// Rule 0: the tag-anchored quoting check.
		{name: "single-quoted href", rule: 0, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/git/index.html", `<a href="/sheetmill/demo-cards/">Home</a>`, `<a href='/sheetmill/demo-cards/'>Home</a>`)
		}},
		{name: "unquoted src", rule: 0, edit: func(f map[string]string) {
			replaceIn(f, "git/index.html", `<img src="/sheetmill/brand/logo.svg"`, `<img src=/sheetmill/brand/logo.svg`)
		}},

		{name: "uppercase single-quoted HREF", rule: 0, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", "<main>", "<main><A HREF='/sheetmill/demo/'>x</A>")
		}},
		{name: "tab before an unquoted href", rule: 0, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", "<main>", "<main><a\thref=/sheetmill/demo/>x</a>")
		}},
		{name: "unquoted href at line end", rule: 0, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", "<main>", "<main><a href=\n>x</a>")
		}},

		// Rule 1: top-level entries.
		{name: "missing demo directory", rule: 1, edit: func(f map[string]string) {
			for name := range f {
				if strings.HasPrefix(name, "demo-cards/") {
					delete(f, name)
				}
			}
		}},
		{name: "missing llms.txt", rule: 1, edit: func(f map[string]string) { delete(f, "llms.txt") }},
		{name: "missing agent guide", rule: 1, edit: func(f map[string]string) { delete(f, "agent-guide.md") }},
		{name: "manual slug demonstration", rule: 1, edit: func(f map[string]string) {
			f["demonstration/index.html"] = manual.page("demonstration/", false)
		}},
		{name: "extra demo2 entry", rule: 1, edit: func(f map[string]string) { f["demo2/index.html"] = manual.page("demo2/", false) }},

		// Rule 2: every site is complete.
		{name: "demo without search.json", rule: 2, edit: func(f map[string]string) { delete(f, "demo-compact/search.json") }},
		{name: "demo without sitemap", rule: 2, edit: func(f map[string]string) { delete(f, "demo/sitemap.xml") }},
		{name: "demo without assets", rule: 2, edit: func(f map[string]string) { delete(f, "demo-cards/assets/site.css") }},
		{name: "demo without home", rule: 2, edit: func(f map[string]string) { delete(f, "demo/index.html") }},

		// Rule 3: base paths.
		{name: "demo links a sibling page", rule: 3, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<a href="/sheetmill/demo/">Home</a>`, `<a href="/sheetmill/demo-compact/git/">Home</a>`)
		}},
		{name: "demo links the manual", rule: 3, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/git/index.html", `<a href="/sheetmill/demo-compact/">Home</a>`, `<a href="/sheetmill/cli/">Home</a>`)
		}},
		{name: "demo asset on a wrong base", rule: 3, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/index.html", `href="/sheetmill/demo-cards/assets/cards.`, `href="/sheetmill/demo/assets/cards.`)
		}},
		{name: "manual links a demonstration slug", rule: 3, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", `<a href="/sheetmill/demo/git/">`, `<a href="/sheetmill/demonstration/">`)
		}},
		{name: "manual links a demo without slash", rule: 3, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", `<a href="/sheetmill/demo/git/">`, `<a href="/sheetmill/demo">`)
		}},
		{name: "manual links outside its base", rule: 3, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", `<a href="/sheetmill/demo/git/">`, `<a href="/other/">`)
		}},

		// Rule 4: absolute URLs.
		{name: "variant canonical points at the default", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/git/index.html", `rel="canonical" href="`+origin+`demo-compact/git/"`, `rel="canonical" href="`+origin+`demo/git/"`)
		}},
		{name: "og:url of another demo", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/index.html", `property="og:url" content="`+origin+`demo-cards/"`, `property="og:url" content="`+origin+`demo/"`)
		}},
		{name: "og:image of another demo", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "demo/index.html", `content="`+origin+`demo/brand/social-preview.png"`, `content="`+origin+`demo-cards/brand/social-preview.png"`)
		}},
		{name: "demo sitemap lists another demo", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/sitemap.xml", origin+"demo-compact/git/", origin+"demo/git/")
		}},
		{name: "manual sitemap lists a demo", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "sitemap.xml", "<loc>"+origin+"cli/</loc>", "<loc>"+origin+"demo/git/</loc>")
		}},
		{name: "manual sitemap off host", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "sitemap.xml", "<loc>"+origin+"cli/</loc>", "<loc>https://example.org/cli/</loc>")
		}},
		{name: "example.invalid leak", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<main>`, `<main>https://example.invalid/`)
		}},
		{name: "relative canonical", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/git/index.html", `rel="canonical" href="`+origin+`demo-cards/git/"`, `rel="canonical" href="/sheetmill/demo-cards/git/"`)
		}},

		{name: "page without a canonical link", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/git/index.html", `<link rel="canonical" href="`+origin+`demo-compact/git/">`, ``)
		}},
		{name: "page with two canonical links", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "git/index.html", `<meta charset="utf-8">`, `<meta charset="utf-8"><link rel="canonical" href="`+origin+`git/">`)
		}},
		{name: "manual page without a canonical link", rule: 4, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", `<link rel="canonical" href="`+origin+`cli/">`, ``)
		}},

		// Rule 5: layout markers.
		{name: "compact lacks the classic body token", rule: 5, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/git/index.html", `<body class="classic">`, `<body>`)
		}},
		{name: "cards lacks the cards stylesheet", rule: 5, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/index.html", `<link rel="stylesheet" href="/sheetmill/demo-cards/assets/cards.`+hash+`.css">`, ``)
		}},
		{name: "cards lacks the reference-cards token", rule: 5, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/git/index.html", `<body class="classic reference-cards">`, `<body class="classic">`)
		}},
		{name: "compact carries the cards stylesheet", rule: 5, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/index.html", `<script`, `<link rel="stylesheet" href="/sheetmill/demo-compact/assets/cards.`+hash+`.css"><script`)
		}},
		{name: "editorial carries the classic stylesheet", rule: 5, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<script`, `<link rel="stylesheet" href="/sheetmill/demo/assets/classic.`+hash+`.css"><script`)
		}},
		{name: "editorial body is classic", rule: 5, edit: func(f map[string]string) {
			replaceIn(f, "demo/index.html", "<body>", `<body class="classic">`)
		}},
		{name: "manual body is classic", rule: 5, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", "<body>", `<body class="classic">`)
		}},

		// Rule 6: page sets.
		{name: "variant misses a page", rule: 6, edit: func(f map[string]string) { delete(f, "demo-cards/git/index.html") }},
		{name: "variant has an extra page", rule: 6, edit: func(f map[string]string) {
			f["demo-compact/extra/index.html"] = compact.page("extra/", false)
		}},

		// Rule 7: header links.
		{name: "compact home links itself", rule: 7, edit: func(f map[string]string) {
			f["demo-compact/index.html"] = strings.ReplaceAll(f["demo-compact/index.html"], `<a class="topbar-link" href="/sheetmill/demo-cards/">`, `<a class="topbar-link" href="/sheetmill/demo-compact/">`)
		}},
		{name: "editorial hero link missing", rule: 7, edit: func(f map[string]string) {
			replaceIn(f, "demo/index.html", `<p class="hero-links"><a href="/sheetmill/demo-compact/">Layout /sheetmill/demo-compact/</a><a href="/sheetmill/demo-cards/">Layout /sheetmill/demo-cards/</a></p>`, `<p class="hero-links"><a href="/sheetmill/demo-compact/">Layout /sheetmill/demo-compact/</a></p>`)
			replaceIn(f, "demo/index.html", `<a href="/sheetmill/demo-cards/">Layout /sheetmill/demo-cards/</a></nav>`, `</nav>`)
		}},
		{name: "cards header links the manual", rule: 7, edit: func(f map[string]string) {
			f["demo-cards/index.html"] = strings.ReplaceAll(f["demo-cards/index.html"], `<a class="topbar-link" href="/sheetmill/demo/">`, `<a class="topbar-link" href="/sheetmill/">`)
		}},
		{name: "editorial hero links itself", rule: 7, edit: func(f map[string]string) {
			replaceIn(f, "demo/index.html", `<p class="hero-links"><a href="/sheetmill/demo-compact/">`, `<p class="hero-links"><a href="/sheetmill/demo/">`)
		}},

		{name: "editorial nav repeats a link", rule: 7, edit: func(f map[string]string) {
			replaceIn(f, "demo/index.html", `<nav class="header-links" aria-label="Related sites"><a href="/sheetmill/demo-compact/">Layout /sheetmill/demo-compact/</a>`, `<nav class="header-links" aria-label="Related sites"><a href="/sheetmill/demo-compact/">Layout /sheetmill/demo-compact/</a><a href="/sheetmill/demo-compact/">again</a>`)
		}},
		{name: "classic top bar repeats a link", rule: 7, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/index.html", `</header>`, `<a class="topbar-link" href="/sheetmill/demo/">again</a></header>`)
		}},
		{name: "manual home lost its Example library link", rule: 7, edit: func(f map[string]string) {
			f["index.html"] = strings.ReplaceAll(f["index.html"], `"/sheetmill/demo/"`, `"/sheetmill/cli/"`)
		}},

		// Rule 8: protocol-relative and insecure URLs.
		{name: "protocol-relative src", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<main>`, `<main><img src="//cdn.example.org/x.png">`)
		}},
		{name: "protocol-relative href", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", `<main>`, `<main><a href="//example.org/">x</a>`)
		}},
		{name: "http script", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/git/index.html", `<main>`, `<main><script src="http://example.org/x.js"></script>`)
		}},
		{name: "http stylesheet link", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", `<main>`, `<main><link rel="stylesheet" href="http://example.org/x.css">`)
		}},
		{name: "foreign https script", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<main>`, `<main><script src="https://cdn.example.org/x.js"></script>`)
		}},
		{name: "foreign https stylesheet link", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "cli/index.html", `<main>`, `<main><link rel="stylesheet" href="https://cdn.example.org/x.css">`)
		}},
		{name: "foreign https form action", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo-compact/git/index.html", `<main>`, `<main><form action="https://example.org/q"></form>`)
		}},
		{name: "foreign srcset", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/git/index.html", `<main>`, `<main><img srcset="https://cdn.example.org/x.png 2x" alt="">`)
		}},
		{name: "protocol-relative srcset", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<main>`, `<main><img srcset="//cdn.example.org/x.png 2x" alt="">`)
		}},
		{name: "uppercase HTTP script", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<main>`, `<main><SCRIPT SRC="HTTP://example.org/x.js"></SCRIPT>`)
		}},
		{name: "base element", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<meta charset="utf-8">`, `<meta charset="utf-8"><base href="/">`)
		}},
		{name: "own-site https script is accepted", pass: true, edit: func(f map[string]string) {
			replaceIn(f, "demo/git/index.html", `<main>`, `<main><img src="`+origin+`demo/brand/logo.svg" alt="">`)
		}},
		{name: "http form action", rule: 8, edit: func(f map[string]string) {
			replaceIn(f, "demo-cards/git/index.html", `<main>`, `<main><form action="http://example.org/q"></form>`)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := validTree()
			if tc.edit != nil {
				tc.edit(files)
			}
			code, out := runScript(t, writeTree(t, files))
			if tc.pass {
				if code != 0 {
					t.Fatalf("valid tree rejected (exit %d):\n%s", code, out)
				}
				return
			}
			if code != 1 {
				t.Fatalf("exit %d, want 1:\n%s", code, out)
			}
			if want := fmt.Sprintf("FAIL rule %d:", tc.rule); !strings.Contains(out, want) {
				t.Fatalf("output lacks %q:\n%s", want, out)
			}
		})
	}
}

// Roots outside [A-Za-z0-9/._-] could be mis-split or vacuously pass, so they are usage errors.
func TestVerifyPagesRejectsUnsafeRoots(t *testing.T) {
	for _, name := range []string{"with space", "pipe|dir", "glob*dir", "quote'dir"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, name)
			if err := os.Mkdir(root, 0o755); err != nil {
				t.Fatal(err)
			}
			code, out := runScript(t, root)
			if code != 2 {
				t.Fatalf("exit %d, want 2:\n%s", code, out)
			}
		})
	}
}

func TestVerifyPagesUsage(t *testing.T) {
	code, _ := runScript(t, filepath.Join(t.TempDir(), "missing"))
	if code != 2 {
		t.Fatalf("a missing directory must be a usage error, exit %d", code)
	}
}
