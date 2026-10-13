// SPDX-License-Identifier: EUPL-1.2
package model

import (
	"encoding/json"
	"github.com/wstein/sheetmill/internal/diagnostic"
)

// Document is the versioned semantic contract. It contains no host paths or clocks.
type Document struct {
	Schema   int            `json:"schema"`
	Metadata map[string]any `json:"metadata"`
	Source   string         `json:"source"`
	Preamble []Node         `json:"preamble"`
	Sections []Section      `json:"sections"`
}
type Node struct {
	Kind       string          `json:"kind"`
	Span       diagnostic.Span `json:"span"`
	Text       string          `json:"text,omitempty"`
	Properties map[string]any  `json:"properties,omitempty"`
	Children   []Node          `json:"children,omitempty"`
}
type Layout struct {
	Columns int `json:"columns,omitempty"`
	Rows    int `json:"rows,omitempty"`
}
type Section struct {
	ID      string          `json:"id"`
	Title   string          `json:"title"`
	Span    diagnostic.Span `json:"span"`
	Heading Node            `json:"heading"`
	// Columns is the optional {.cols-N} grid hint (1–6); zero means the default.
	Columns int    `json:"columns,omitempty"`
	Content []Node `json:"content"`
	Cards   []Card `json:"cards"`
}
type Card struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Span          diagnostic.Span `json:"span"`
	Heading       Node            `json:"heading"`
	Layout        Layout          `json:"layout"`
	Classes       []string        `json:"classes"`
	Compatibility []string        `json:"compatibility"`
	Content       []Node          `json:"content"`
}

func (d *Document) JSON() ([]byte, error) {
	b, e := json.MarshalIndent(d, "", "  ")
	if e != nil {
		return nil, e
	}
	return append(b, '\n'), nil
}
