package copilot

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// fakeCopilot puts a shell script named copilot with the given body first
// and only on PATH.
func fakeCopilot(t *testing.T, body string) {
	t.Helper()

	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "copilot"), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", bin)
}

func TestRunPassesArgsDirAndStdout(t *testing.T) {
	fakeCopilot(t, `printf '%s\n' "$PWD" "$@"`)

	r := &Runner{Token: "test-token"}
	dir := t.TempDir()
	var out bytes.Buffer

	if err := r.Run(context.Background(), dir, "make a todo app", &out); err != nil {
		t.Fatal(err)
	}

	want := strings.Join([]string{
		dir,
		"-p", "make a todo app",
		"--output-format=json",
		"--no-ask-user",
		"--allow-all-tools",
		"--allow-all-urls",
		"--disable-builtin-mcps",
		"--max-ai-credits=1000",
		"--no-auto-update",
	}, "\n") + "\n"

	if got := out.String(); got != want {
		t.Errorf("fake copilot saw:\n%s\nwant:\n%s", got, want)
	}
}

func TestRunPassesOnlyPathHomeAndToken(t *testing.T) {
	fakeCopilot(t, `printf '%s\n' \
		"PATH=$PATH" \
		"HOME=$HOME" \
		"COPILOT_GITHUB_TOKEN=$COPILOT_GITHUB_TOKEN" \
		"GH_TOKEN=${GH_TOKEN-unset}" \
		"COPILOT_HOME=${COPILOT_HOME-unset}"`)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GH_TOKEN", "parent-gh-token")
	t.Setenv("COPILOT_HOME", "/parent/copilot-home")

	r := &Runner{Token: "test-token"}
	var out bytes.Buffer

	if err := r.Run(context.Background(), t.TempDir(), "hi", &out); err != nil {
		t.Fatal(err)
	}

	want := strings.Join([]string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + home,
		"COPILOT_GITHUB_TOKEN=test-token",
		"GH_TOKEN=unset",
		"COPILOT_HOME=unset",
	}, "\n") + "\n"

	if got := out.String(); got != want {
		t.Errorf("fake copilot saw:\n%s\nwant:\n%s", got, want)
	}
}

func TestRunRefusesEmptyToken(t *testing.T) {
	fakeCopilot(t, `echo ran`)

	r := &Runner{}
	var out bytes.Buffer

	if err := r.Run(context.Background(), t.TempDir(), "hi", &out); err == nil {
		t.Fatal("Run with an empty token returned nil error")
	}

	if out.Len() != 0 {
		t.Errorf("fake copilot ran: %q", out.String())
	}
}

func TestRunErrorIncludesStderr(t *testing.T) {
	fakeCopilot(t, `echo 'Error: Authentication token found but could not be validated.' >&2; exit 1`)

	r := &Runner{Token: "test-token"}

	err := r.Run(context.Background(), t.TempDir(), "hi", &bytes.Buffer{})
	if err == nil {
		t.Fatal("Run returned nil error for exit status 1")
	}

	for _, want := range []string{"exit status 1", "Authentication token found but could not be validated."} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
}

func TestRunCancelKillsProcessGroup(t *testing.T) {
	fakeCopilot(t, `/usr/bin/sleep 60 >/dev/null 2>&1 &
echo $! > child.pid
wait`)

	r := &Runner{Token: "test-token"}
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)

	go func() { done <- r.Run(ctx, dir, "hi", &bytes.Buffer{}) }()

	child := waitForPID(t, filepath.Join(dir, "child.pid"))
	t.Cleanup(func() { _ = syscall.Kill(child, syscall.SIGKILL) })

	cancel()

	if err := <-done; err == nil {
		t.Error("Run returned nil error after cancel")
	}

	for deadline := time.Now().Add(2 * time.Second); alive(child); {
		if time.Now().After(deadline) {
			t.Fatalf("grandchild %d still running after cancel", child)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func TestRunDoesNotWaitForChildHoldingStdout(t *testing.T) {
	fakeCopilot(t, `/usr/bin/sleep 60 &
echo $! > child.pid`)

	r := &Runner{Token: "test-token"}
	dir := t.TempDir()
	done := make(chan error, 1)

	go func() { done <- r.Run(context.Background(), dir, "hi", &bytes.Buffer{}) }()

	child := waitForPID(t, filepath.Join(dir, "child.pid"))
	t.Cleanup(func() { _ = syscall.Kill(child, syscall.SIGKILL) })

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v, want nil: copilot exited 0", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run still waiting 5s after copilot exited")
	}
}

func TestRunKillsLeftoverProcessesAfterExit(t *testing.T) {
	fakeCopilot(t, `/usr/bin/sleep 60 >/dev/null 2>&1 &
echo $! > child.pid`)

	r := &Runner{Token: "test-token"}
	dir := t.TempDir()

	if err := r.Run(context.Background(), dir, "hi", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	child := waitForPID(t, filepath.Join(dir, "child.pid"))
	t.Cleanup(func() { _ = syscall.Kill(child, syscall.SIGKILL) })

	for deadline := time.Now().Add(2 * time.Second); alive(child); {
		if time.Now().After(deadline) {
			t.Fatalf("leftover process %d still running after Run returned", child)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// waitForPID waits until the fake has written a PID to path and returns it.
func waitForPID(t *testing.T, path string) int {
	t.Helper()

	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		b, _ := os.ReadFile(path)
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
			return pid
		}
	}

	t.Fatalf("no PID in %s", path)

	return 0
}

// alive reports whether pid is running. A zombie counts as gone: it has
// exited and only waits for its parent to reap it.
func alive(pid int) bool {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}

	state := b[bytes.LastIndexByte(b, ')')+2]

	return state != 'Z'
}
