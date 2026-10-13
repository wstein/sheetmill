// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A fixed test executable implements only the trusted tool protocol. It never
// evaluates source. Real Hugo/browser verification is separate and required in CI.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("hugo v0.167.0 test")
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--source" {
		args := os.Args[1:]
		if len(args) != 8 || args[2] != "--destination" || args[4] != "--clock" || args[5] != "2000-01-01T00:00:00Z" || args[6] != "--noBuildLock" || args[7] != "--quiet" {
			os.Exit(10)
		}
		for _, e := range os.Environ() {
			if strings.HasPrefix(e, "HUGO_EVIL=") || strings.HasPrefix(e, "HTTP_PROXY=") {
				os.Exit(11)
			}
		}
		source, destination := args[1], args[3]
		os.MkdirAll(destination, 0700)
		os.WriteFile(filepath.Join(destination, "index.html"), []byte("<!doctype html><html><head></head><body>Generated</body></html>"), 0600)
		input, _ := os.ReadFile(filepath.Join(source, "data/sheets/one.json"))
		if len(input) > 0 {
			os.MkdirAll(filepath.Join(destination, "one"), 0700)
			os.WriteFile(filepath.Join(destination, "one/index.html"), input, 0600)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func testTool(t *testing.T) string {
	t.Helper()
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return path
}
func TestCompileRetainsOnlyLastValidIRAndValidPeers(t *testing.T) {
	root := t.TempDir()
	tool := testTool(t)
	t.Setenv("HUGO_EVIL", "unexpected")
	t.Setenv("HTTP_PROXY", "http://unexpected.invalid")
	first := compile(context.Background(), tool, root, 1, Snapshot{Entries: []Entry{{Name: "one.md", Bytes: []byte(validSource)}}}, nil, 0)
	if first.Error != "" || first.Sheets[0].Status != "current" {
		t.Fatalf("%+v", first)
	}
	second := compile(context.Background(), tool, root, 2, Snapshot{Entries: []Entry{{Name: "one.md", Bytes: []byte("bad")}, {Name: "peer.md", Bytes: []byte(validSource)}}}, first.Retained, 0)
	if second.Error != "" || second.Sheets[0].Status != "invalid" || !second.Sheets[0].Stale || second.Sheets[0].Published != 1 || second.Sheets[1].Status != "current" {
		t.Fatalf("%+v", second)
	}
	for _, d := range second.Sheets[0].Diagnostics {
		if d.Code == "" || d.Span.Start.Line < 1 {
			t.Fatal("lost compiler diagnostic", d)
		}
	}
	collision := compile(context.Background(), tool, root, 3, Snapshot{Entries: []Entry{{Name: "one!.md", Bytes: []byte(validSource)}, {Name: "one.md", Bytes: []byte(validSource)}, {Name: "studio.md", Bytes: []byte(validSource)}}}, first.Retained, 0)
	for _, s := range collision.Sheets {
		if s.Status != "invalid" || s.Stale {
			t.Fatal("collision retained route", s)
		}
	}
	if got := compile(context.Background(), tool, root, 4, Snapshot{Entries: []Entry{{Name: "one.md", Bytes: []byte(validSource)}}}, nil, maxGenerated); got.Error == "" {
		t.Fatal("generated budget accepted")
	}
}
func TestPublisherEnvironmentArgumentsAndCancellation(t *testing.T) {
	root := t.TempDir()
	env := toolEnv(root)
	data, _ := json.Marshal(env)
	if strings.Contains(string(data), "HUGO_EVIL") || strings.Contains(string(data), os.Getenv("HOME")+"\"") {
		t.Fatal("inherited user config")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if runTool(ctx, testTool(t), []string{"version"}, root, true, maxGenerated) == nil {
		t.Fatal("cancelled tool started")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("cancel not bounded")
	}
	log := &boundedLog{}
	if _, e := log.Write(make([]byte, maxLog+1)); e == nil || log.text.Len() > maxLog {
		t.Fatal("unbounded output")
	}
	path := filepath.Join(root, "oversized")
	f, _ := os.Create(path)
	f.Truncate(maxGenerated + 1)
	f.Close()
	if _, e := treeSize(root, maxGenerated); e == nil {
		t.Fatal("oversized generation accepted")
	}
}
