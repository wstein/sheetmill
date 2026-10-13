// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"testing"
)

func TestNewestRevisionWins(t *testing.T) {
	s := &Session{state: State{Sheets: []SheetState{}}}
	for i := uint64(1); i <= 100; i++ {
		s.begin(i, Snapshot{Entries: []Entry{{Name: "one.md", Bytes: []byte(validSource)}}})
		if s.publish(i-1, Build{Files: map[string][]byte{"index.html": []byte("old")}}) {
			t.Fatal("obsolete publication")
		}
		if !s.publish(i, Build{Files: map[string][]byte{"index.html": []byte("new")}, Sheets: []SheetState{{Name: "one.md", Status: "current", Published: i}}}) {
			t.Fatal("latest refused")
		}
	}
	if s.State().Sequence != 100 || s.State().Sheets[0].Published != 100 {
		t.Fatal("wrong final revision")
	}
}
