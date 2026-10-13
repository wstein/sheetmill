// SPDX-License-Identifier: EUPL-1.2
package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestServeUsage(t *testing.T) {
	for _, args := range [][]string{{"serve", "--help"}, {"serve", "one", "two"}, {"serve", "--port", "-1"}, {"serve", "--hugo", "/missing/smill-hugo"}} {
		var out, err bytes.Buffer
		status := Run(args, &out, &err)
		if len(args) == 2 && args[1] == "--help" {
			if status != 0 || !strings.Contains(out.String(), "--hugo") {
				t.Fatal(status, out.String(), err.String())
			}
		} else if status != 2 {
			t.Fatal(args, status)
		}
	}
}
