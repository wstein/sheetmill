// SPDX-License-Identifier: EUPL-1.2
package diagnostic

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

type Position struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}
type Span struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}
type Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Span    Span   `json:"span"`
}

func (d Diagnostic) Error() string {
	return fmt.Sprintf("%d:%d: %s: %s", d.Span.Start.Line, d.Span.Start.Column, d.Code, d.Message)
}

// Source indexes original UTF-8 bytes. Offsets are zero-based, columns count runes.
type Source struct {
	Bytes  []byte
	starts []int
}

func NewSource(b []byte) *Source {
	s := &Source{Bytes: b, starts: []int{0}}
	for i, c := range b {
		if c == '\n' {
			s.starts = append(s.starts, i+1)
		}
	}
	return s
}
func (s *Source) Position(offset int) Position {
	offset = max(0, min(offset, len(s.Bytes)))
	i := sort.Search(len(s.starts), func(i int) bool { return s.starts[i] > offset }) - 1
	return Position{offset, i + 1, utf8.RuneCount(s.Bytes[s.starts[i]:offset]) + 1}
}
func (s *Source) Span(start, end int) Span {
	return Span{s.Position(start), s.Position(max(start, end))}
}
func (s *Source) Offset(line, column int) int {
	line = max(1, min(line, len(s.starts)))
	i := s.starts[line-1]
	for n := 1; n < column && i < len(s.Bytes) && s.Bytes[i] != '\n'; n++ {
		_, size := utf8.DecodeRune(s.Bytes[i:])
		i += size
	}
	return i
}
func (s *Source) Error(code, msg string, start, end int) Diagnostic {
	return Diagnostic{code, msg, s.Span(start, end)}
}
