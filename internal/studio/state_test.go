// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"fmt"
	"net/http/httptest"
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

func TestRenameDoesNotMaskCurrentRouteAndTombstonesAreBounded(t *testing.T) {
	s := &Session{state: State{Sheets: []SheetState{{Name: "one.md", ID: "one", Status: "current", Published: 1}}}, files: map[string][]byte{}}
	s.begin(2, Snapshot{Entries: []Entry{{Name: "one!.md", Bytes: []byte(validSource)}}})
	s.publish(2, Build{Sheets: []SheetState{{Name: "one!.md", ID: "one", Status: "current", Published: 2}}, Files: map[string][]byte{"one/index.html": []byte("current")}})
	r := httptest.NewRequest("GET", "http://127.0.0.1/one/", nil)
	w := httptest.NewRecorder()
	s.handler().ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "current" {
		t.Fatal("historical tombstone masked current output", w.Code)
	}
	for i := uint64(3); i < 203; i++ {
		name := fmt.Sprintf("file-%d.md", i)
		id := routeID(name)
		s.begin(i, Snapshot{Entries: []Entry{{Name: name}}})
		s.publish(i, Build{Sheets: []SheetState{{Name: name, ID: id, Status: "current", Published: i}}})
		if len(s.State().Sheets) > 2 {
			t.Fatal("unbounded removed history")
		}
	}
}
