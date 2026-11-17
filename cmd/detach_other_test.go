//go:build !unix

package cmd

import "os/exec"

// detachFromTerminal has nothing to do: the process already has no console
// input when stdin is redirected.
func detachFromTerminal(*exec.Cmd) {}
