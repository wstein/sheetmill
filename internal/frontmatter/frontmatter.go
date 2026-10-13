// SPDX-License-Identifier: EUPL-1.2
package frontmatter

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/wstein/sheetmill/internal/diagnostic"
	"github.com/wstein/sheetmill/internal/validate"
	"go.yaml.in/yaml/v3"
)

type Result struct {
	Metadata   map[string]any
	Body       []byte
	BodyOffset int
	Locations  map[string]diagnostic.Span
}

var yamlLine = regexp.MustCompile(`line (\d+)`)

func Parse(b []byte) (Result, []diagnostic.Diagnostic) {
	s := diagnostic.NewSource(b)
	bad := func(code, msg string, start, end int) (Result, []diagnostic.Diagnostic) {
		return Result{}, []diagnostic.Diagnostic{s.Error(code, msg, start, end)}
	}
	if !utf8.Valid(b) {
		i := 0
		for i < len(b) {
			r, n := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError && n == 1 {
				break
			}
			i += n
		}
		return bad("SRC001", "source must be UTF-8", i, i+1)
	}
	first := bytes.IndexByte(b, '\n')
	if first < 0 || strings.TrimSuffix(string(b[:first]), "\r") != "---" {
		return bad("FM001", "expected leading YAML frontmatter delimiter ---", 0, 0)
	}
	start := first + 1
	end := -1
	body := len(b)
	for i := start; i < len(b); {
		j := bytes.IndexByte(b[i:], '\n')
		next := len(b)
		if j >= 0 {
			next = i + j + 1
		}
		line := strings.TrimSuffix(strings.TrimSuffix(string(b[i:next]), "\n"), "\r")
		if line == "---" {
			end = i
			body = next
			break
		}
		i = next
	}
	if end < 0 {
		return bad("FM002", "frontmatter needs a closing --- delimiter", 0, first)
	}
	dec := yaml.NewDecoder(bytes.NewReader(b[start:end]))
	var root yaml.Node
	if err := dec.Decode(&root); err != nil {
		line := 1
		if m := yamlLine.FindStringSubmatch(err.Error()); len(m) > 1 {
			line, _ = strconv.Atoi(m[1])
		}
		off := min(s.Offset(line+1, 1), end)
		return bad("FM003", "invalid YAML: "+err.Error(), off, off)
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		return bad("FM003", "frontmatter must contain exactly one YAML document", start, end)
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return bad("FM003", "frontmatter must be a mapping", start, end)
	}
	var ds []diagnostic.Diagnostic
	locations := map[string]diagnostic.Span{}
	var value func(*yaml.Node, string, int) any
	value = func(n *yaml.Node, path string, depth int) any {
		off := s.Offset(n.Line+1, n.Column)
		span := s.Span(off, off+len(n.Value))
		locations[path] = span
		add := func(code, msg string) { ds = append(ds, diagnostic.Diagnostic{Code: code, Message: msg, Span: span}) }
		if depth > 64 {
			add("FM005", "YAML nesting exceeds 64 levels")
			return nil
		}
		if n.Anchor != "" || n.Kind == yaml.AliasNode {
			add("FM005", "YAML anchors and aliases are unsupported")
			return nil
		}
		switch n.Kind {
		case yaml.MappingNode:
			m := map[string]any{}
			seen := map[string]bool{}
			for i := 0; i < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				keyoff := s.Offset(k.Line+1, k.Column)
				if k.Kind != yaml.ScalarNode || k.Tag != "!!str" || k.Anchor != "" {
					ds = append(ds, s.Error("FM005", "mapping keys must be plain strings; merge keys are unsupported", keyoff, keyoff+len(k.Value)))
					continue
				}
				if seen[k.Value] {
					ds = append(ds, s.Error("FM004", "duplicate YAML key: "+k.Value, keyoff, keyoff+len(k.Value)))
					continue
				}
				seen[k.Value] = true
				child := k.Value
				if path != "" {
					child = path + "." + child
				}
				m[k.Value] = value(v, child, depth+1)
				locations[child] = s.Span(keyoff, keyoff+len(k.Value))
			}
			return m
		case yaml.SequenceNode:
			a := make([]any, 0, len(n.Content))
			for _, v := range n.Content {
				a = append(a, value(v, path, depth+1))
			}
			return a
		case yaml.ScalarNode:
			switch n.Tag {
			case "!!str", "!!timestamp":
				return n.Value
			case "!!null":
				return nil
			case "!!bool":
				var v bool
				if err := n.Decode(&v); err == nil {
					return v
				}
			case "!!int":
				var v int64
				if err := n.Decode(&v); err == nil {
					return v
				}
			case "!!float":
				var v float64
				if err := n.Decode(&v); err == nil && !math.IsInf(v, 0) && !math.IsNaN(v) {
					return v
				}
			}
		}
		add("FM005", fmt.Sprintf("unsupported YAML value or tag %q; use JSON-compatible values", n.Tag))
		return nil
	}
	metadata, _ := value(root.Content[0], "", 0).(map[string]any)
	if len(ds) == 0 {
		ds = validate.Metadata(metadata, locations)
	}
	if len(ds) > 0 {
		return Result{}, ds
	}
	return Result{Metadata: metadata, Body: b[body:], BodyOffset: body, Locations: locations}, nil
}
