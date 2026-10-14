// SPDX-License-Identifier: EUPL-1.2
package frontmatter

import "testing"

func TestSharedCommandMetadataPositions(t *testing.T) {
	src := []byte("---\ntitle: Commands\ncheatsheet:\n  cards:\n    run:\n      params:\n        file:\n          type: invalid\n---\n## Commands\n")
	r, ds := Parse(src)
	if len(ds) != 1 || ds[0].Code != "CMD001" || ds[0].Span.Start.Line != 8 || ds[0].Span.Start.Column != 11 || r.Metadata != nil {
		t.Fatalf("bad metadata diagnostic: %+v result=%+v", ds, r)
	}
}
