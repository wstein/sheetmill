// SPDX-License-Identifier: EUPL-1.2
package hugo

type palette struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var palettes = []palette{
	{"graphite-coral", "Graphite & Coral"}, {"paper-ink", "Paper & Ink"},
	{"midnight-periwinkle", "Midnight & Periwinkle"}, {"monochrome", "Monochrome"},
	{"teal-copper", "Teal & Copper"}, {"dracula", "Dracula"},
}

// PaletteNames returns supported appearance palettes independently of layout.
func PaletteNames() []string {
	out := make([]string, len(palettes))
	for i, p := range palettes {
		out[i] = p.ID
	}
	return out
}
