//go:build !windows

// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"os/exec"
	"sync"
	"syscall"
	"time"
)

func ownProcess(cmd *exec.Cmd) func() {
	var mu sync.Mutex
	var timer *time.Timer
	finished := false
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		pid := cmd.Process.Pid
		mu.Lock()
		timer = time.AfterFunc(time.Second, func() {
			mu.Lock()
			defer mu.Unlock()
			if !finished {
				syscall.Kill(-pid, syscall.SIGKILL)
			}
		})
		mu.Unlock()
		return err
	}
	cmd.WaitDelay = time.Second
	return func() {
		mu.Lock()
		defer mu.Unlock()
		finished = true
		if timer != nil {
			timer.Stop()
		}
	}
}
