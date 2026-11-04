package apptemplate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dragan-mitrasinovic/SideProjects/platform/internal/copilot"
)

// TestSmoke generates a real app from the template with the real Copilot CLI.
// It spends real AI credits, so it only runs when asked to:
//
//	set -a; . /root/.config/sideprojects/copilot.env; set +a
//	COPILOT_SMOKE=1 go test -run Smoke -v ./internal/apptemplate
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

	// Outside any git repo, so the CLI loads only the template's AGENTS.md.
	dir := filepath.Join(t.TempDir(), "app")

	if err := Copy(dir); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	r := &copilot.Runner{Token: token}

	if err := r.Run(ctx, dir, "A counter app with + and - buttons.", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	placeholder, err := os.ReadFile(filepath.Join("files", "index.html"))
	if err != nil {
		t.Fatal(err)
	}

	page, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(page, placeholder) {
		t.Error("index.html is still the placeholder")
	}

	htmlFiles, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		t.Fatal(err)
	}

	if len(htmlFiles) != 1 {
		t.Errorf("HTML files in app folder = %v, want only index.html", htmlFiles)
	}
}
