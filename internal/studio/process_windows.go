// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"os/exec"
	"time"
)

func ownProcess(cmd *exec.Cmd) func() { cmd.WaitDelay = time.Second; return func() {} }
