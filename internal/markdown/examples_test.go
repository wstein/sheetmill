// SPDX-License-Identifier: EUPL-1.2
package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// Every demo sheet must carry a one-sentence description so its library card is
// never empty, and the sheet must stay valid.
func TestExampleSheetsHaveDescriptions(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "examples", "sheets", "*.md"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no example sheets found: %v", err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			d, ds := Parse(src)
			if len(ds) > 0 || d == nil {
				t.Fatalf("invalid example: %v", ds)
			}
			desc, _ := d.Metadata["description"].(string)
			if strings.TrimSpace(desc) == "" {
				t.Fatal("description is missing or empty")
			}
			if n := utf8.RuneCountInString(desc); n < 20 || n > 200 || strings.ContainsAny(desc, "\r\n") {
				t.Fatalf("description should be one line of 20-200 characters, got %d", n)
			}
		})
	}
}
