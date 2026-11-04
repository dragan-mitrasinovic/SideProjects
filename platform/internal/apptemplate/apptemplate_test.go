package apptemplate

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyWritesTemplateFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "app")

	if err := Copy(dir); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"index.html", "AGENTS.md"} {
		// The source files on disk, independent of what was embedded.
		want, err := os.ReadFile(filepath.Join("files", name))
		if err != nil {
			t.Fatal(err)
		}

		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, want) {
			t.Errorf("%s differs from files/%s", name, name)
		}
	}
}

func TestCopyDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	existing := []byte("<p>my generated app</p>\n")
	path := filepath.Join(dir, "index.html")

	if err := os.WriteFile(path, existing, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Copy(dir); !errors.Is(err, fs.ErrExist) {
		t.Errorf("Copy into a folder with index.html returned %v, want fs.ErrExist", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, existing) {
		t.Errorf("index.html was overwritten: %q", got)
	}
}
