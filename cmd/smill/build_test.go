// SPDX-License-Identifier: EUPL-1.2
package main_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type workflow struct {
	Jobs map[string]job `yaml:"jobs"`
}

type job struct {
	Strategy strategy `yaml:"strategy"`
	Steps    []step   `yaml:"steps"`
}

type strategy struct {
	Matrix matrix `yaml:"matrix"`
}

type matrix struct {
	Include []platform `yaml:"include"`
}

type platform struct {
	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type step struct {
	Uses string            `yaml:"uses"`
	Run  string            `yaml:"run"`
	With map[string]any    `yaml:"with"`
	Env  map[string]string `yaml:"env"`
}

func buildJob(t *testing.T) job {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "check.yml"))
	if err != nil {
		t.Fatalf("failed to read check.yml: %v", err)
	}
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("failed to parse check.yml: %v", err)
	}
	j, ok := wf.Jobs["build"]
	if !ok {
		t.Fatalf("check.yml must define a 'build' job")
	}
	return j
}

// The release platforms are linux amd64 and arm64 and darwin arm64. The set is
// exact so that an end-of-life or out-of-scope target cannot creep back in.
func TestWorkflowBuildMatrixIsExactlyTheReleasePlatforms(t *testing.T) {
	var got []string
	for _, p := range buildJob(t).Strategy.Matrix.Include {
		got = append(got, p.OS+"/"+p.Arch)
	}
	sort.Strings(got)
	want := []string{"darwin/arm64", "linux/amd64", "linux/arm64"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("build matrix = %v, want exactly %v", got, want)
	}
}

// The build job only checks that the code compiles for each target. It must not
// publish artifacts, must stamp the version like a release build, and must not
// share the module and build cache with the core job.
func TestWorkflowBuildJobIsAVerificationOnly(t *testing.T) {
	j := buildJob(t)
	var built, cacheOff bool
	for _, s := range j.Steps {
		if strings.HasPrefix(s.Uses, "actions/upload-artifact") || strings.HasPrefix(s.Uses, "actions/attest") {
			t.Errorf("build job must not publish artifacts: %s", s.Uses)
		}
		if strings.HasPrefix(s.Uses, "actions/setup-go") {
			if v, ok := s.With["cache"].(bool); ok && !v {
				cacheOff = true
			}
		}
		if strings.Contains(s.Run, "go build") {
			built = true
			if !strings.Contains(s.Run, "-o /dev/null") {
				t.Errorf("build must write to /dev/null: %q", s.Run)
			}
			if !strings.Contains(s.Run, "-X github.com/wstein/sheetmill/internal/cli.Version=") {
				t.Errorf("build must exercise the version stamp: %q", s.Run)
			}
			if s.Env["GOOS"] == "" || s.Env["GOARCH"] == "" || s.Env["CGO_ENABLED"] != "0" {
				t.Errorf("build must set GOOS, GOARCH and CGO_ENABLED=0: %v", s.Env)
			}
		}
	}
	if !built {
		t.Error("build job has no go build step")
	}
	if !cacheOff {
		t.Error("build job must set cache: false on setup-go so it cannot displace core's cache")
	}
}
