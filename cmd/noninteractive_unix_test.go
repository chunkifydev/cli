//go:build unix

package cmd

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"syscall"
	"testing"
)

func TestJSONCancelledBySignal(t *testing.T) {
	cmd := startCLI(t, "-i", "src_hang", "-f", "mp4_h264", "--json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(stdout)
	// The first progress line shows the program is running and handles signals.
	first, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("no progress before the signal: %v", err)
	}
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	rest, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	code := exitStatus(t, cmd.Wait())

	if code != 130 {
		t.Fatalf("exit code %d, want 130; stderr: %s", code, stderr.String())
	}
	lines := jsonLines(t, first+string(rest))
	if last := lines[len(lines)-1]; last["status"] != "Cancelled" {
		t.Fatalf("last line = %v, want Cancelled", last)
	}
	var report map[string]string
	if err := json.Unmarshal([]byte(stderr.String()), &report); err != nil || report["error"] != "cancelled" {
		t.Fatalf(`stderr = %q, want {"error":"cancelled"}`, stderr.String())
	}
}
