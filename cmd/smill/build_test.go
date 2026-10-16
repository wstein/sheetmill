// SPDX-License-Identifier: EUPL-1.2
package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.yaml.in/yaml/v3"
)

type workflow struct {
	Jobs map[string]job `yaml:"jobs"`
}

type job struct {
	Strategy strategy `yaml:"strategy"`
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

func TestWorkflowDefinesBuildMatrix(t *testing.T) {
	workflowPath := filepath.Join("..", "..", ".github", "workflows", "check.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read check.yml: %v", err)
	}

	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("failed to parse check.yml: %v", err)
	}

	buildJob, ok := wf.Jobs["build"]
	if !ok {
		t.Fatalf("check.yml must define a 'build' job")
	}

	wantPlatforms := map[string]bool{
		"linux/amd64":  true,
		"linux/arm64":  true,
		"darwin/amd64": true,
		"darwin/arm64": true,
	}

	for _, p := range buildJob.Strategy.Matrix.Include {
		key := p.OS + "/" + p.Arch
		delete(wantPlatforms, key)
	}

	if len(wantPlatforms) > 0 {
		t.Fatalf("check.yml missing required build targets: %v", wantPlatforms)
	}
}

func TestCrossCompileTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cross-compile test in short mode")
	}

	targets := []struct {
		os   string
		arch string
	}{
		{os: "linux", arch: "amd64"},
		{os: "linux", arch: "arm64"},
		{os: "darwin", arch: "amd64"},
		{os: "darwin", arch: "arm64"},
	}

	tempDir := t.TempDir()

	for _, tc := range targets {
		t.Run(tc.os+"_"+tc.arch, func(t *testing.T) {
			outBin := filepath.Join(tempDir, "smill-"+tc.os+"-"+tc.arch)
			cmd := exec.Command("go", "build", "-trimpath", "-o", outBin, ".")
			cmd.Env = append(os.Environ(),
				"CGO_ENABLED=0",
				"GOOS="+tc.os,
				"GOARCH="+tc.arch,
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("cross-compilation for %s/%s failed: %v\nOutput: %s", tc.os, tc.arch, err, string(out))
			}
			info, err := os.Stat(outBin)
			if err != nil {
				t.Fatalf("output binary not found: %v", err)
			}
			if info.Size() == 0 {
				t.Fatalf("output binary is empty")
			}
		})
	}
}
