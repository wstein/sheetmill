// SPDX-License-Identifier: EUPL-1.2
package hugo

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

type tagPage struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Sheets []page `json:"sheets"`
}

// tagCatalog preserves exact, case-sensitive labels. A readable ASCII stem plus
// a digest makes punctuation/Unicode labels safe and independently stable.
func tagCatalog(pages []page, base string) ([]tagPage, error) {
	groups := map[string][]page{}
	for _, p := range pages {
		p.Body, p.Related = "", nil
		seen := map[string]bool{}
		for _, tag := range p.Tags {
			if !seen[tag] {
				groups[tag] = append(groups[tag], p)
				seen[tag] = true
			}
		}
	}
	labels := make([]string, 0, len(groups))
	for label := range groups {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	result := make([]tagPage, 0, len(labels))
	ids := map[string]string{}
	for _, label := range labels {
		var stem strings.Builder
		for _, r := range strings.ToLower(label) {
			if stem.Len() >= 32 {
				break
			}
			if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
				stem.WriteRune(r)
			} else if stem.Len() > 0 && !strings.HasSuffix(stem.String(), "-") {
				stem.WriteByte('-')
			}
		}
		name := strings.Trim(stem.String(), "-")
		if name == "" {
			name = "tag"
		}
		sum := sha256.Sum256([]byte(label))
		id := fmt.Sprintf("%s-%x", name, sum[:6])
		if prior, exists := ids[id]; exists && prior != label {
			return nil, fmt.Errorf("tag route collision between %q and %q", prior, label)
		}
		ids[id] = label
		result = append(result, tagPage{id, label, base + "tags/" + id + "/", groups[label]})
	}
	return result, nil
}
