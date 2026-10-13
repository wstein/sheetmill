// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"github.com/wstein/sheetmill/internal/diagnostic"
	"sort"
)

type SheetState struct {
	Name        string                  `json:"name"`
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	Status      string                  `json:"status"`
	Published   uint64                  `json:"published"`
	Stale       bool                    `json:"stale"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}
type State struct {
	Version   uint64       `json:"version"`
	Sequence  uint64       `json:"sequence"`
	Published uint64       `json:"published"`
	Sheets    []SheetState `json:"sheets"`
	Error     string       `json:"error,omitempty"`
}

func (s *Session) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state := s.state
	state.Sheets = append([]SheetState{}, state.Sheets...)
	return state
}
func (s *Session) notify() {
	for _, ch := range s.subscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
func (s *Session) begin(seq uint64, snap Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Sequence = seq
	s.state.Version++
	s.state.Error = snap.Error
	known := map[string]SheetState{}
	for _, old := range s.state.Sheets {
		old.Status = "removed"
		if snap.Error != "" {
			old.Status = "pending"
		}
		old.Stale = true
		known[old.Name] = old
	}
	for _, entry := range snap.Entries {
		old := known[entry.Name]
		old.Name = entry.Name
		old.ID = routeID(entry.Name)
		old.Status = "pending"
		old.Stale = old.Published > 0
		old.Diagnostics = nil
		known[entry.Name] = old
	}
	s.state.Sheets = nil
	for _, v := range known {
		s.state.Sheets = append(s.state.Sheets, v)
	}
	sort.Slice(s.state.Sheets, func(i, j int) bool { return s.state.Sheets[i].Name < s.state.Sheets[j].Name })
	s.notify()
}
func (s *Session) publish(seq uint64, b Build) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq != s.state.Sequence {
		return false
	}
	s.state.Version++
	if b.Error != "" {
		s.state.Error = b.Error
		for i := range s.state.Sheets {
			if s.state.Sheets[i].Status != "removed" {
				s.state.Sheets[i].Status = "failed"
				s.state.Sheets[i].Stale = true
			}
		}
	} else {
		s.files = b.Files
		s.state.Error = ""
		s.state.Published = seq
		known := map[string]bool{}
		for _, v := range b.Sheets {
			known[v.Name] = true
		}
		for _, old := range s.state.Sheets {
			if !known[old.Name] {
				old.Status = "removed"
				old.Stale = true
				old.Diagnostics = nil
				b.Sheets = append(b.Sheets, old)
			}
		}
		s.state.Sheets = b.Sheets
		sort.Slice(s.state.Sheets, func(i, j int) bool { return s.state.Sheets[i].Name < s.state.Sheets[j].Name })
		s.retained = b.Retained
	}
	s.notify()
	return true
}
