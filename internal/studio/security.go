// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
)

type gate struct {
	mu                                       sync.Mutex
	host, origin, secret, cookie, cookieName string
	consumed                                 bool
}

func randomSecret() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
func newGate(host string) (*gate, error) {
	secret, e := randomSecret()
	if e != nil {
		return nil, e
	}
	cookie, e := randomSecret()
	if e != nil {
		return nil, e
	}
	return &gate{host: host, origin: "http://" + host, secret: secret, cookie: cookie, cookieName: "smill_" + strings.ReplaceAll(host, ":", "_")}, nil
}
func same(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
func (g *gate) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Cache-Control", "no-store")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; font-src 'self'; img-src 'self'; connect-src 'self'; base-uri 'none'; object-src 'none'; form-action 'none'; frame-ancestors 'none'")
		deny := func() { http.Error(w, "Local preview access refused", http.StatusForbidden) }
		if r.Host != g.host || len(r.Header.Values("Origin")) > 1 || r.Header.Get("Forwarded") != "" || r.Header.Get("X-Forwarded-Host") != "" || r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Forwarded-Proto") != "" {
			deny()
			return
		}
		origin := r.Header.Get("Origin")
		site := r.Header.Get("Sec-Fetch-Site")
		dest := r.Header.Get("Sec-Fetch-Dest")
		if origin != "" && origin != g.origin || site != "" && site != "same-origin" && site != "none" || dest != "" && dest != "empty" && dest != "document" && dest != "script" && dest != "style" && dest != "font" && dest != "image" {
			deny()
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Read-only preview", 405)
			return
		}
		// Reject ambiguous/traversing encoded routes before the immutable file map lookup.
		if strings.Contains(r.URL.EscapedPath(), "%") || strings.Contains(r.URL.Path, "\\") || strings.Contains(r.URL.Path, "//") {
			http.NotFound(w, r)
			return
		}
		for _, part := range strings.Split(r.URL.Path, "/") {
			if part == "." || part == ".." {
				http.NotFound(w, r)
				return
			}
		}
		if r.URL.Path == "/studio/bootstrap" {
			if r.Method != "GET" || dest != "" && dest != "document" || r.Header.Get("Sec-Fetch-Mode") != "" && r.Header.Get("Sec-Fetch-Mode") != "navigate" {
				deny()
				return
			}
			g.mu.Lock()
			ok := !g.consumed && same(r.URL.Query().Get("token"), g.secret)
			if ok {
				g.consumed = true
			}
			g.mu.Unlock()
			if !ok {
				deny()
				return
			}
			http.SetCookie(w, &http.Cookie{Name: g.cookieName, Value: g.cookie, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		cookies := 0
		authorized := false
		for _, c := range r.Cookies() {
			if c.Name == g.cookieName {
				cookies++
				authorized = same(c.Value, g.cookie)
			}
		}
		if !authorized || cookies != 1 {
			deny()
			return
		}
		next.ServeHTTP(w, r)
	})
}
