package copilot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestSmoke runs the real Copilot CLI once. It spends real AI credits, so it
// only runs when asked to:
//
//	set -a; . /root/.config/sideprojects/copilot.env; set +a
//	COPILOT_SMOKE=1 go test -run Smoke -v ./internal/copilot
func TestSmoke(t *testing.T) {
	if os.Getenv("COPILOT_SMOKE") != "1" {
		t.Skip("set COPILOT_SMOKE=1 to run against the real Copilot CLI")
	}

	token := os.Getenv("COPILOT_GITHUB_TOKEN")
	if token == "" {
		t.Fatal("COPILOT_GITHUB_TOKEN is not set")
	}

	// A throwaway HOME keeps the CLI's sessions and logs out of the real one.
	t.Setenv("HOME", t.TempDir())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	r := &Runner{Token: token}
	var out bytes.Buffer

	err := r.Run(ctx, t.TempDir(), "Use the web_fetch tool to fetch https://example.com and reply with only the page title.", &out)
	if err != nil {
		t.Fatal(err)
	}

	var last event
	fetches := 0
	scanner := bufio.NewScanner(&out)
	scanner.Buffer(nil, 1<<20)

	for scanner.Scan() {
		var e event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("not a JSON line: %s", scanner.Bytes())
		}

		if e.Type == "tool.execution_start" && e.Data.ToolName == "web_fetch" {
			fetches++
		}

		if e.Type == "tool.execution_complete" && !e.Data.Success {
			t.Errorf("a tool call failed: %s", scanner.Bytes())
		}

		last = e
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	if fetches == 0 {
		t.Error("Copilot never called web_fetch")
	}

	if last.Type != "result" || last.ExitCode != 0 {
		t.Errorf("last event = %+v, want type result with exitCode 0", last)
	}
}

// event holds the few JSONL fields the smoke test checks.
type event struct {
	Type     string `json:"type"`
	ExitCode int    `json:"exitCode"`
	Data     struct {
		ToolName string `json:"toolName"`
		Success  bool   `json:"success"`
	} `json:"data"`
}
