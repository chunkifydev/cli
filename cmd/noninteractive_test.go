package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/zalando/go-keyring"
)

const helperEnv = "CHUNKIFY_TEST_HELPER"

// TestHelperProcess runs the CLI with the arguments after "--". startCLI runs it
// in a subprocess without a terminal.
func TestHelperProcess(t *testing.T) {
	if os.Getenv(helperEnv) == "" {
		t.Skip("helper process")
	}
	if os.Getenv("CHUNKIFY_TEST_NO_KEYRING") != "" {
		keyring.MockInitWithError(errors.New("no session bus"))
	} else {
		keyring.MockInit()
	}
	os.Args = append([]string{"chunkify"}, flag.Args()...)
	Execute()
	os.Exit(0)
}

type cliResult struct {
	exitCode int
	stdout   string
	stderr   string
}

// startCLI starts the CLI against a fake API, with stdin closed and no
// controlling terminal, like CI, cron or Docker without -t. The fake API knows
// src_test, never answers for src_hang, and returns 404 for anything else.
func startCLI(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	return startCLIWithEnv(t, nil, args...)
}

// startCLIWithEnv is startCLI with extra environment variables, which override
// the defaults.
func startCLIWithEnv(t *testing.T, env []string, args ...string) *exec.Cmd {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/sources/src_test":
			fmt.Fprint(w, `{"data":{"id":"src_test","duration":10,"size":1000}}`)
		case "/v1/sources/src_hang":
			<-r.Context().Done()
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"status":"error","error":{"message":"source not found"}}`)
		}
	}))
	t.Cleanup(server.Close)

	home := t.TempDir()
	// The CLI checks os.Args for --json before parsing flags, so pass the
	// arguments on the command line rather than in the environment.
	// Kill the CLI if it hangs, so a regression fails here instead of at the
	// global test timeout.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	cmd := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestHelperProcess$", "--"}, args...)...)
	cmd.Env = append(os.Environ(),
		helperEnv+"=1",
		"CHUNKIFY_TOKEN=sk_project_test",
		"CHUNKIFY_ENDPOINT="+server.URL+"/v1",
		// Keep the debug log out of the real cache directory.
		"HOME="+home, "XDG_CACHE_HOME="+home, "LocalAppData="+home,
	)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdin = nil // reads from the null device
	detachFromTerminal(cmd)
	// Cleanups run last-registered first: stop the CLI before closing the server.
	t.Cleanup(cancel)
	t.Cleanup(func() {
		if cmd.Process != nil && cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	return cmd
}

// runCLI runs the CLI to completion; see startCLI.
func runCLI(t *testing.T, args ...string) cliResult {
	t.Helper()
	return runCLIWithEnv(t, nil, args...)
}

// runCLIWithEnv runs the CLI to completion; see startCLIWithEnv.
func runCLIWithEnv(t *testing.T, env []string, args ...string) cliResult {
	t.Helper()
	cmd := startCLIWithEnv(t, env, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return cliResult{exitCode: exitStatus(t, cmd.Run()), stdout: stdout.String(), stderr: stderr.String()}
}

// exitStatus returns the exit code from the error of a finished command.
func exitStatus(t *testing.T, err error) int {
	t.Helper()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return 0
}

// jsonLines decodes stdout, which must hold only JSON progress lines.
func jsonLines(t *testing.T, stdout string) []map[string]any {
	t.Helper()
	var lines []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		var line map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatalf("stdout line is not JSON: %q\nstdout:\n%s", scanner.Text(), stdout)
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		t.Fatal("no JSON output")
	}
	return lines
}

func TestJSONWithoutTerminal(t *testing.T) {
	for _, flag := range []string{"--json", "--json=true"} {
		t.Run(flag, func(t *testing.T) {
			result := runCLI(t, "-i", "src_test", flag)
			if result.exitCode != 0 {
				t.Fatalf("exit code %d, stdout:\n%s\nstderr:\n%s", result.exitCode, result.stdout, result.stderr)
			}
			lines := jsonLines(t, result.stdout) // also fails on the banner
			last := lines[len(lines)-1]
			if last["status"] != "Completed" || last["progress"] != 100.0 {
				t.Fatalf("last line = %v, want Completed with progress 100", last)
			}
		})
	}
}

// Linux servers, CI and containers often have no keyring; CHUNKIFY_TOKEN is enough.
func TestJSONWithoutKeyring(t *testing.T) {
	result := runCLIWithEnv(t, []string{"CHUNKIFY_TEST_NO_KEYRING=1"}, "-i", "src_test", "--json")
	if result.exitCode != 0 {
		t.Fatalf("exit code %d, stderr:\n%s", result.exitCode, result.stderr)
	}
	if last := jsonLines(t, result.stdout); last[len(last)-1]["status"] != "Completed" {
		t.Fatalf("last line = %v, want Completed", last[len(last)-1])
	}
	if report := stderrJSON(t, result.stderr); !strings.Contains(report["warning"], "CHUNKIFY_STORAGE_ID") {
		t.Fatalf("stderr must warn about the skipped storage-id config, got %q", result.stderr)
	}
}

// Errors before the run starts are JSON too, and stdout stays empty.
func TestJSONErrorsBeforeRun(t *testing.T) {
	for _, tt := range []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{name: "invalid flag value", args: []string{"-i", "src_test", "-f", "not_a_format", "--json"}, want: "not_a_format"},
		{name: "missing token", env: []string{"CHUNKIFY_TOKEN="}, args: []string{"-i", "src_test", "--json"}, want: "CHUNKIFY_TOKEN"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := runCLIWithEnv(t, tt.env, tt.args...)
			if result.exitCode != 1 {
				t.Fatalf("exit code %d, want 1", result.exitCode)
			}
			if result.stdout != "" {
				t.Fatalf("stdout must stay empty, got %q", result.stdout)
			}
			if report := stderrJSON(t, result.stderr); !strings.Contains(report["error"], tt.want) {
				t.Fatalf("error must mention %q, got %q", tt.want, report["error"])
			}
		})
	}
}

// stderrJSON decodes stderr, which must hold one JSON object.
func stderrJSON(t *testing.T, stderr string) map[string]string {
	t.Helper()
	var report map[string]string
	if err := json.Unmarshal([]byte(stderr), &report); err != nil {
		t.Fatalf("stderr is not one JSON object: %q", stderr)
	}
	return report
}

func TestFailedRunExitsNonZero(t *testing.T) {
	result := runCLI(t, "-i", "src_doesnotexist", "-o", "out.mp4", "-f", "mp4_h264", "--json")
	if result.exitCode == 0 {
		t.Fatalf("failed run exited 0, stdout:\n%s", result.stdout)
	}
	lines := jsonLines(t, result.stdout)
	if last := lines[len(lines)-1]; last["status"] != "Failed" {
		t.Fatalf("last line = %v, want Failed", last)
	}
	if report := stderrJSON(t, result.stderr); !strings.Contains(report["error"], "source not found") {
		t.Fatalf("stderr must report the error, got %q", result.stderr)
	}
}
