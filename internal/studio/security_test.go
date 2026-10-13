// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"net/http"
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
