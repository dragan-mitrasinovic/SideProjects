// Package copilot runs the GitHub Copilot CLI in programmatic mode (ADR-0003).
package copilot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Runner runs the Copilot CLI signed in with Token, a fine-grained PAT with
// only the Copilot Requests permission.
type Runner struct {
	Token string
}

// Run runs one `copilot -p prompt` in dir and copies its JSONL stdout to out
// while it runs.
func (r *Runner) Run(ctx context.Context, dir, prompt string, out io.Writer) error {
	// Without a token the CLI silently falls back to a stored login or gh's
	// token, possibly another account with broader scopes.
	if r.Token == "" {
		return errors.New("copilot: no token")
	}

	cmd := exec.CommandContext(ctx, "copilot",
		"-p", prompt,
		"--output-format=json",
		"--no-ask-user",
		"--allow-all-tools",
		"--allow-all-urls",
		"--disable-builtin-mcps",
		"--max-ai-credits=1000",
		"--no-auto-update",
	)

	var stderr bytes.Buffer

	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = &stderr

	// Own process group, so cancelling also kills the builds and dev servers
	// Copilot started, not just the CLI itself.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	// A leftover process can keep stdout open after Copilot exits, which would
	// make Run wait for it. After this delay the pipes are closed regardless.
	cmd.WaitDelay = time.Second

	// Built from scratch: an inherited GH_TOKEN or COPILOT_* variable would
	// change which account the CLI uses or how it behaves.
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"COPILOT_GITHUB_TOKEN=" + r.Token,
	}

	err := cmd.Run()

	// Kill anything Copilot started and left running, such as a dev server.
	// Process is nil if the CLI could not be started.
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	// Copilot exited 0 and only a leftover process, now killed, held stdout.
	if errors.Is(err, exec.ErrWaitDelay) {
		err = nil
	}

	// Errors such as a bad token are reported only on stderr.
	if err != nil {
		return fmt.Errorf("copilot: %w: %s", err, bytes.TrimSpace(stderr.Bytes()))
	}

	return nil
}
