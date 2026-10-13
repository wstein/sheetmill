// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

//go:embed web
var web embed.FS

func (s *Session) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/studio/state":
			w.Header().Set("Content-Type", "application/json")
			if r.Method != "HEAD" {
				json.NewEncoder(w).Encode(s.State())
			}
			return
		case "/studio/events":
			s.events(w, r)
			return
		case "/studio/client.js", "/studio/status.css":
			name := strings.TrimPrefix(r.URL.Path, "/studio/")
			data, _ := web.ReadFile("web/" + name)
			w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(name)))
			if r.Method != "HEAD" {
				w.Write(data)
			}
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" || strings.HasSuffix(name, "/") {
			name += "index.html"
		}
		s.mu.RLock()
		data, exists := s.files[name]
		for _, sheet := range s.state.Sheets {
			if sheet.ID == strings.Split(name, "/")[0] && sheet.Status == "removed" {
				exists = false
			}
		}
		s.mu.RUnlock()
		status := http.StatusOK
		if !exists && strings.HasSuffix(name, "/index.html") {
			data = []byte(`<!doctype html><html><head><meta charset="utf-8"><title>SheetMill Studio — unavailable</title></head><body><main><h1>Reference unavailable</h1><p>This reference was removed, is invalid or is not selected.</p><a href="/">Reference library</a></main></body></html>`)
			exists = true
			status = http.StatusNotFound
		}
		if !exists {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(name)))
		if strings.HasSuffix(name, ".html") {
			injection := []byte(`<link rel="stylesheet" href="/studio/status.css"><script src="/studio/client.js" defer></script>`)
			data = bytes.Replace(data, []byte("</head>"), append(injection, []byte("</head>")...), 1)
		}
		w.WriteHeader(status)
		if r.Method != "HEAD" {
			w.Write(data)
		}
	})
}
func (s *Session) events(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Events require GET", 405)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unavailable", 500)
		return
	}
	s.mu.Lock()
	if len(s.subscribers) >= 8 {
		s.mu.Unlock()
		http.Error(w, "Preview stream limit reached", http.StatusTooManyRequests)
		return
	}
	id := s.nextSubscriber
	s.nextSubscriber++
	ch := make(chan struct{}, 1)
	s.subscribers[id] = ch
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.subscribers, id); s.mu.Unlock() }()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("X-Accel-Buffering", "no")
	send := func() error {
		http.NewResponseController(w).SetWriteDeadline(time.Now().Add(3 * time.Second))
		_, err := fmt.Fprintf(w, "data: %d\n\n", s.State().Version)
		flusher.Flush()
		return err
	}
	if send() != nil {
		return
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-r.Context().Done():
			return
		case <-ch:
			if send() != nil {
				return
			}
		case <-ticker.C:
			if send() != nil {
				return
			}
		}
	}
}
