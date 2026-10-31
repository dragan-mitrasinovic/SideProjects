package copilot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// fakeCopilot writes an executable shell script that stands in for the CLI.
func fakeCopilot(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "copilot")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func run(t *testing.T, r *Runner, dir string) (Result, []string, error) {
	t.Helper()
	if r.Timeout == 0 {
		r.Timeout = 10 * time.Second
	}
	var lines []string
	res, err := r.Run(context.Background(), dir, "build a todo app", func(l string) { lines = append(lines, l) })
	return res, lines, err
}

func TestRunPassesArgs(t *testing.T) {
	r := &Runner{Program: fakeCopilot(t, `for a in "$@"; do echo "$a"; done`)}
	res, lines, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 || res.TimedOut {
		t.Fatalf("result = %+v", res)
	}
	want := []string{"-p", "build a todo app", "--no-ask-user", "--allow-all-tools", "--output-format=json"}
	if !slices.Equal(lines, want) {
		t.Fatalf("args = %q, want %q", lines, want)
	}
}

func TestRunStreamsLinesInOrder(t *testing.T) {
	r := &Runner{Program: fakeCopilot(t, `echo '{"n":1}'; printf '{"n":2}\r\n'; printf '{"n":3}'`)}
	_, lines, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{`{"n":1}`, `{"n":2}`, `{"n":3}`}
	if !slices.Equal(lines, want) {
		t.Fatalf("lines = %q, want %q", lines, want)
	}
}

func TestRunWorkdirAndStdin(t *testing.T) {
	dir := t.TempDir()
	// cat returns at once only if stdin is closed rather than inherited.
	r := &Runner{Program: fakeCopilot(t, `pwd; cat; echo done`)}
	_, lines, err := run(t, r, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0] != dir || lines[1] != "done" {
		t.Fatalf("lines = %q", lines)
	}
}

func TestRunEnvironment(t *testing.T) {
	t.Setenv("PLATFORM_SECRET", "leak")
	r := &Runner{Program: fakeCopilot(t, `env`), Token: "tok"}
	_, lines, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, l := range lines {
		name, _, _ := strings.Cut(l, "=")
		names = append(names, name)
	}
	slices.Sort(names)
	// The shell may add PWD, SHLVL and _ on its own.
	names = slices.DeleteFunc(names, func(n string) bool { return n == "PWD" || n == "SHLVL" || n == "_" })
	want := []string{"COPILOT_GITHUB_TOKEN", "HOME", "PATH"}
	if !slices.Equal(names, want) {
		t.Fatalf("env names = %q, want %q", names, want)
	}
	if !slices.Contains(lines, "COPILOT_GITHUB_TOKEN=tok") {
		t.Fatalf("token not passed: %q", lines)
	}
}

func TestRunFailure(t *testing.T) {
	r := &Runner{Program: fakeCopilot(t, `echo '{"type":"x"}'; echo "auth failed" >&2; exit 3`)}
	res, lines, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 3 || res.TimedOut || res.Stderr != "auth failed\n" {
		t.Fatalf("result = %+v", res)
	}
	if len(lines) != 1 {
		t.Fatalf("lines = %q", lines)
	}
}

func TestRunStderrKeepsTail(t *testing.T) {
	r := &Runner{Program: fakeCopilot(t, `head -c 20000 /dev/zero | tr '\0' a >&2; printf END >&2`)}
	res, _, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Stderr) != stderrTailSize || !strings.HasSuffix(res.Stderr, "END") {
		t.Fatalf("stderr len = %d, suffix ok = %v", len(res.Stderr), strings.HasSuffix(res.Stderr, "END"))
	}
}

func TestRunTimeoutKillsProcessGroup(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pid")
	// The background sleep stands in for a command Copilot started; it keeps
	// stdout open, so Run only returns quickly if the whole group is killed.
	r := &Runner{
		Program: fakeCopilot(t, `sleep 30 & echo $! > `+pidFile+`; echo started; wait`),
		Timeout: 500 * time.Millisecond,
	}
	start := time.Now()
	res, lines, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !res.TimedOut || res.ExitCode != -1 {
		t.Fatalf("result = %+v", res)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("Run took %v", elapsed)
	}
	if !slices.Equal(lines, []string{"started"}) {
		t.Fatalf("lines = %q", lines)
	}
	assertDead(t, pidFile)
}

func TestRunKillsLeftoverProcesses(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pid")
	// Copilot exits normally but leaves a detached child behind.
	r := &Runner{Program: fakeCopilot(t, `sleep 30 >/dev/null 2>&1 & echo $! > `+pidFile)}
	res, _, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("result = %+v", res)
	}
	assertDead(t, pidFile)
}

func TestRunExitBeforeTimeoutWithOpenStdout(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pid")
	// Copilot exits cleanly, but a leftover child keeps stdout open past the
	// timeout. That must not be reported as a timeout.
	r := &Runner{
		Program: fakeCopilot(t, `sleep 30 & echo $! > `+pidFile+`; sleep 0.2; exit 0`),
		Timeout: time.Second,
	}
	res, _, err := run(t, r, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if res.TimedOut || res.ExitCode != 0 {
		t.Fatalf("result = %+v", res)
	}
	assertDead(t, pidFile)
}

func TestRunParentCancel(t *testing.T) {
	r := &Runner{Program: fakeCopilot(t, `sleep 30`), Timeout: 10 * time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(200*time.Millisecond, cancel)
	_, err := r.Run(ctx, t.TempDir(), "p", func(string) {})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestRunMissingProgram(t *testing.T) {
	r := &Runner{Program: filepath.Join(t.TempDir(), "nope"), Timeout: time.Second}
	if _, err := r.Run(context.Background(), t.TempDir(), "p", func(string) {}); err == nil {
		t.Fatal("expected an error")
	}
}

func assertDead(t *testing.T, pidFile string) {
	t.Helper()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	// The killed process may linger briefly as a zombie until it's reaped.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if syscall.Kill(pid, 0) != nil || isZombie(pid) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("process %d still running", pid)
}

func isZombie(pid int) bool {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return false
	}
	// Format: pid (comm) state ...
	s := string(data)
	i := strings.LastIndexByte(s, ')')
	return i >= 0 && i+2 < len(s) && s[i+2] == 'Z'
}
