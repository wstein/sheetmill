// SPDX-License-Identifier: EUPL-1.2
package reference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/frontmatter"
	"github.com/wstein/sheetmill/internal/markdown"
	"github.com/wstein/sheetmill/internal/model"
)

type Provenance struct {
	Repository  string `json:"repository"`
	Path        string `json:"path"`
	Revision    string `json:"revision"`
	License     string `json:"license"`
	Attribution string `json:"attribution"`
}
type Finding struct {
	Severity string          `json:"severity"`
	Code     string          `json:"code"`
	Message  string          `json:"message"`
	Span     diagnostic.Span `json:"span"`
}
type Report struct {
	Schema          int        `json:"schema"`
	Format          string     `json:"format"`
	Provenance      Provenance `json:"provenance"`
	SourceSHA256    string     `json:"source_sha256"`
	ConvertedSHA256 string     `json:"converted_sha256"`
	Findings        []Finding  `json:"findings"`
}
type Result struct {
	Markdown []byte
	Report   Report
}

func (r Report) JSON() ([]byte, error) {
	b, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return nil, e
	}
	return append(b, '\n'), nil
}

var revisionPattern = regexp.MustCompile(`^(?:[a-f0-9]{40}|[a-f0-9]{64})$`)

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// Convert is pure: no clocks, filesystem, network or plugin/command execution.
func Convert(src []byte, p Provenance) (Result, []diagnostic.Diagnostic) {
	s := diagnostic.NewSource(src)
	failed := func(ds []diagnostic.Diagnostic) (Result, []diagnostic.Diagnostic) {
		r := Result{Report: Report{Schema: 1, Format: "fechin-reference", Provenance: p, SourceSHA256: hash(src), Findings: []Finding{}}}
		for _, d := range ds {
			r.Report.Findings = append(r.Report.Findings, Finding{Severity: "error", Code: d.Code, Message: d.Message, Span: d.Span})
		}
		return r, ds
	}
	fail := func(code, msg string) (Result, []diagnostic.Diagnostic) {
		return failed([]diagnostic.Diagnostic{s.Error(code, msg, 0, 0)})
	}
	u, err := url.Parse(p.Repository)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fail("IMP001", "repository must be an explicit HTTPS repository URL without credentials, query or fragment")
	}
	if p.Path == "" || p.Path == "." || path.IsAbs(p.Path) || path.Clean(p.Path) != p.Path || p.Path == ".." || strings.HasPrefix(p.Path, "../") || strings.Contains(p.Path, "\\") {
		return fail("IMP001", "source path must be an explicit clean relative repository path")
	}
	if !revisionPattern.MatchString(p.Revision) {
		return fail("IMP001", "revision must be a full lowercase 40- or 64-hex commit ID")
	}
	if strings.TrimSpace(p.License) == "" || strings.TrimSpace(p.Attribution) == "" {
		return fail("IMP001", "license and attribution must be explicitly supplied; no redistribution permission is inferred")
	}
	if len(src) > markdown.MaxSourceBytes {
		return fail("SRC002", "source exceeds 8 MiB")
	}
	fm, ds := frontmatter.Parse(src)
	if len(ds) > 0 {
		return failed(ds)
	}
	if _, ok := fm.Metadata["source"]; ok {
		return fail("IMP002", "existing source provenance must not be overwritten; select the original external file")
	}
	// Reject unsupported body constructs using ORIGINAL source locations.
	parsed, ds := markdown.Parse(src)
	if len(ds) > 0 {
		return failed(ds)
	}
	r := Result{Report: Report{Schema: 1, Format: "fechin-reference", Provenance: p, SourceSHA256: hash(src), Findings: []Finding{}}}
	original := map[string]any{}
	for _, key := range []string{"plugins", "background", "class", "layout", "theme"} {
		if v, ok := fm.Metadata[key]; ok {
			original[key] = v
			delete(fm.Metadata, key)
			code, msg := "IMP102", "legacy visual metadata "+key+" archived as inert source data"
			if key == "plugins" {
				code = "IMP101"
				msg = "legacy plugins archived and disabled; no plugin behavior is imported"
			}
			r.Report.Findings = append(r.Report.Findings, Finding{Severity: "warning", Code: code, Message: msg, Span: fm.Locations[key]})
		}
	}

	var loss func(model.Node)
	loss = func(n model.Node) {
		if n.Kind == "FencedCodeBlock" || n.Kind == "CodeBlock" || n.Kind == "CodeSpan" {
			return
		}
		code, msg := "", ""
		if n.Kind == "HTMLBlock" || n.Kind == "RawHTML" {
			code = "IMP103"
			msg = "raw HTML retained as inert source; a future renderer must sanitize it"
		} else if n.Kind == "Text" && strings.Contains(n.Text, "{%") {
			code = "IMP104"
			msg = "unsupported plugin directive retained as inert text"
		}
		if code != "" {
			r.Report.Findings = append(r.Report.Findings, Finding{Severity: "warning", Code: code, Message: msg, Span: n.Span})
		}
		for _, c := range n.Children {
			loss(c)
		}
	}
	for _, n := range parsed.Preamble {
		loss(n)
	}
	for _, section := range parsed.Sections {
		loss(section.Heading)
		for _, n := range section.Content {
			loss(n)
		}
		for _, card := range section.Cards {
			loss(card.Heading)
			for _, n := range card.Content {
				loss(n)
			}
		}
	}
	source := map[string]any{"format": "fechin-reference", "repository": p.Repository, "path": p.Path, "revision": p.Revision, "license": p.License, "attribution": p.Attribution, "sha256": r.Report.SourceSHA256, "original_metadata": original}
	fm.Metadata["source"] = source
	header, err := json.MarshalIndent(fm.Metadata, "", "  ")
	if err != nil {
		return fail("IMP003", "cannot serialize canonical metadata: "+err.Error())
	}
	// JSON is a strict YAML subset. Sorted JSON keys make conversion reproducible.
	r.Markdown = append([]byte("---\n"), header...)
	r.Markdown = append(r.Markdown, []byte("\n---\n")...)
	r.Markdown = append(r.Markdown, fm.Body...)
	if _, ds = markdown.Parse(r.Markdown); len(ds) > 0 {
		return fail("IMP003", "converted source failed canonical validation: "+ds[0].Message)
	}
	r.Report.ConvertedSHA256 = hash(r.Markdown)
	return r, nil
}
