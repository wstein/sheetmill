// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"context"
	"time"
)

type revision struct {
	seq      uint64
	snapshot Snapshot
}

func (s *Session) watch(ctx context.Context, initial Snapshot) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	previous := initial.hash()
	seq := uint64(1)
	var latest Snapshot
	var first, last time.Time
	pending := false
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			snapshot := s.selection.Read()
			h := snapshot.hash()
			if h != previous {
				previous = h
				seq++
				latest = snapshot
				last = now
				if !pending {
					first = now
				}
				pending = true
				s.begin(seq, snapshot)
			}
			if pending && (now.Sub(last) >= 100*time.Millisecond || now.Sub(first) >= 500*time.Millisecond) {
				r := revision{seq, latest}
				select {
				case s.jobs <- r:
				default:
					select {
					case <-s.jobs:
					default:
					}
					select {
					case s.jobs <- r:
					case <-ctx.Done():
						return
					}
				}
				pending = false
			}
		}
	}
}
