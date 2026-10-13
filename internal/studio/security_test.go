// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func TestReadAuthorization(t *testing.T) {
	gate, err := newGate("127.0.0.1:3210")
	if err != nil {
		t.Fatal(err)
	}
	next := gate.wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("PRIVATE")) }))
	request := func(path string, cookie *http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "http://127.0.0.1:3210"+path, nil)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		for k, v := range headers {
			if k == "Host" {
				r.Host = v
			} else {
				r.Header.Set(k, v)
			}
		}
		w := httptest.NewRecorder()
		next.ServeHTTP(w, r)
		return w
	}
	for _, p := range []string{"/", "/assets/app.js", "/studio/state", "/studio/events", "/studio/bootstrap?token=wrong"} {
		if request(p, nil, nil).Code != 403 {
			t.Fatal("unauthenticated read", p)
		}
	}
	boot := request("/studio/bootstrap?token="+gate.secret, nil, map[string]string{"Sec-Fetch-Mode": "navigate", "Sec-Fetch-Dest": "document"})
	if boot.Code != 303 {
		t.Fatal(boot.Code)
	}
	cookie := boot.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("weak cookie")
	}
	if request("/studio/bootstrap?token="+gate.secret, nil, nil).Code != 403 {
		t.Fatal("bootstrap replay")
	}
	for _, h := range []map[string]string{{"Origin": "null"}, {"Origin": "https://evil.invalid"}, {"Host": "localhost:3210"}, {"X-Forwarded-Host": "evil.invalid"}, {"Forwarded": "host=evil.invalid"}, {"Sec-Fetch-Site": "cross-site"}, {"Sec-Fetch-Dest": "iframe"}} {
		if request("/studio/state", cookie, h).Code != 403 {
			t.Fatal("cross-origin read", h)
		}
	}
	if request("/studio/state", cookie, map[string]string{"Origin": "http://127.0.0.1:3210", "Sec-Fetch-Site": "same-origin"}).Code != 200 {
		t.Fatal("authorized refused")
	}
}

func TestTwoPreviewSessionsKeepIndependentCookies(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	gates := []*gate{}
	for _, host := range []string{"127.0.0.1:3210", "127.0.0.1:3211"} {
		g, _ := newGate(host)
		gates = append(gates, g)
		r := httptest.NewRequest("GET", g.origin+"/studio/bootstrap?token="+g.secret, nil)
		w := httptest.NewRecorder()
		g.wrap(http.NotFoundHandler()).ServeHTTP(w, r)
		jar.SetCookies(r.URL, w.Result().Cookies())
	}
	for _, g := range gates {
		r := httptest.NewRequest("GET", g.origin+"/studio/state", nil)
		for _, cookie := range jar.Cookies(r.URL) {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		g.wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })).ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatal("second preview displaced first authorization", w.Code)
		}
	}
}

func TestGateMethodsRoutesAndHeaders(t *testing.T) {
	g, _ := newGate("127.0.0.1:3210")
	handler := g.wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("private")) }))
	for _, method := range []string{"POST", "PUT", "PATCH", "DELETE", "OPTIONS", "TRACE", "CONNECT"} {
		r := httptest.NewRequest(method, g.origin+"/studio/state", nil)
		r.Host = g.host
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != 405 {
			t.Fatal(method, w.Code)
		}
	}
	for _, path := range []string{"/%2e%2e/private", "/one/../private", "/one//private", "/one%2fprivate"} {
		r := httptest.NewRequest("GET", g.origin+path, nil)
		r.AddCookie(&http.Cookie{Name: g.cookieName, Value: g.cookie})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != 404 {
			t.Fatal("ambiguous path accepted", path, w.Code)
		}
	}
	r := httptest.NewRequest("HEAD", g.origin+"/studio/state", nil)
	r.AddCookie(&http.Cookie{Name: g.cookieName, Value: g.cookie})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	for _, header := range []string{"Content-Security-Policy", "Referrer-Policy", "Cache-Control", "X-Content-Type-Options", "X-Frame-Options"} {
		if w.Header().Get(header) == "" {
			t.Fatal("missing boundary header", header)
		}
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("CORS widened")
	}
}
