// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if condition() {
			return
		}
		select {
		case <-deadline.C:
			t.Fatal("state did not converge")
		case <-tick.C:
		}
	}
}
func TestSessionSavedChangesAndCleanup(t *testing.T) {
	path := fixture(t, "one.md", validSource)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, e := Start(ctx, Options{Input: filepath.Dir(path), Hugo: testTool(t)})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	root := s.root
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: time.Second}
	r, e := client.Get(s.BootstrapURL())
	if e != nil {
		t.Fatal(e)
	}
	r.Body.Close()
	if r.StatusCode != 200 {
		t.Fatal(r.StatusCode)
	}
	port, _ := strconv.Atoi(strings.Split(s.listener.Addr().String(), ":")[1])
	if _, e = Start(ctx, Options{Input: path, Hugo: testTool(t), Port: port}); e == nil {
		t.Fatal("occupied port accepted")
	}
	original := s.State().Sequence
	os.WriteFile(path, []byte("invalid"), 0600)
	eventually(t, func() bool {
		v := s.State()
		return v.Sequence > original && v.Sheets[0].Status == "invalid" && v.Sheets[0].Stale
	})
	os.WriteFile(path, []byte(validSource+"repaired"), 0600)
	eventually(t, func() bool { return s.State().Sheets[0].Status == "current" })
	peer := filepath.Join(filepath.Dir(path), "peer.md")
	os.WriteFile(peer, []byte(validSource), 0600)
	eventually(t, func() bool { return len(s.State().Sheets) == 2 && s.State().Sheets[1].Status == "current" })
	os.Remove(path)
	eventually(t, func() bool { return s.State().Sheets[0].Status == "removed" })
	r, e = client.Get(s.gate.origin + "/one/")
	if e != nil {
		t.Fatal(e)
	}
	r.Body.Close()
	if r.StatusCode != 404 {
		t.Fatal("removed route still served")
	}
	cancel()
	start := time.Now()
	if e = s.Close(); e != nil {
		t.Fatal(e)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("shutdown exceeded 5s")
	}
	if _, e = os.Stat(root); !os.IsNotExist(e) {
		t.Fatal("owned state remains")
	}
	if _, e = client.Get(s.gate.origin); e == nil {
		t.Fatal("listener still open")
	}
	if _, e = os.Stat(peer); e != nil {
		t.Fatal("source removed")
	}
}
func TestInvalidStartupAndActualHugo(t *testing.T) {
	path := fixture(t, "one.md", "invalid")
	s, e := Start(context.Background(), Options{Input: path, Hugo: testTool(t)})
	if e != nil {
		t.Fatal(e)
	}
	if s.State().Sheets[0].Status != "invalid" {
		t.Fatal("invalid source not diagnosed")
	}
	s.Close()
	tool := os.Getenv("SHEETMILL_HUGO")
	if tool == "" {
		t.Skip("installed pinned Hugo is required in browser/CI integration")
	}
	os.WriteFile(path, []byte(validSource), 0600)
	s, e = Start(context.Background(), Options{Input: path, Hugo: tool})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if !strings.Contains(string(s.files["one/index.html"]), "hello") {
		t.Fatal("actual Hugo omitted source")
	}
}

func TestStreamCeilingAndShutdown(t *testing.T) {
	path := fixture(t, "one.md", validSource)
	s, e := Start(context.Background(), Options{Input: path, Hugo: testTool(t)})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	r, e := client.Get(s.BootstrapURL())
	if e != nil {
		t.Fatal(e)
	}
	r.Body.Close()
	var streams []*http.Response
	for i := 0; i < 8; i++ {
		r, e := client.Get(s.gate.origin + "/studio/events")
		if e != nil {
			t.Fatal(e)
		}
		if r.StatusCode != 200 {
			t.Fatal(r.StatusCode)
		}
		streams = append(streams, r)
	}
	defer func() {
		for _, r := range streams {
			r.Body.Close()
		}
	}()
	r, e = client.Get(s.gate.origin + "/studio/events")
	if e != nil {
		t.Fatal(e)
	}
	r.Body.Close()
	if r.StatusCode != 429 {
		t.Fatal("stream limit", r.StatusCode)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestListenerFailureIsFatalAndCleansOwnedState(t *testing.T) {
	path := fixture(t, "one.md", validSource)
	s, e := Start(context.Background(), Options{Input: path, Hugo: testTool(t)})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	s.listener.Close()
	if e = s.Wait(); e == nil {
		t.Fatal("listener failure hidden")
	}
	if e = s.Close(); e != nil {
		t.Fatal(e)
	}
}
