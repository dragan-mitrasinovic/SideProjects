// Package apptemplate holds the files every generated app starts from: a
// placeholder index.html and an AGENTS.md with rules for the generating agent.
package apptemplate

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed files
var files embed.FS

// Copy writes the template's files into dir, creating it if needed. It fails,
// without overwriting anything, if a template file already exists there.
func Copy(dir string) error {
	sub, err := fs.Sub(files, "files")
	if err != nil {
		return err
	}

	return os.CopyFS(dir, sub)
}
