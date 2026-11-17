//go:build unix

package cmd

import (
	"os/exec"
	"syscall"
)

// detachFromTerminal starts the command in a new session, without a
// controlling terminal, so opening /dev/tty fails as it does in CI.
func detachFromTerminal(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
