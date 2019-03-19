// Package util implements utility methods
package util

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// GenerateDocs is a wrapper around the Hugo binary in the container and instructs the binary to generate the site
func GenerateDocs(hugoDir string) error {
	log.Print("Regeneratig Hugo content")
	os.RemoveAll(filepath.Join(hugoDir, "public"))
	cmd := exec.Command("sh", "-c", "hugo")
	cmd.Dir = hugoDir
	output, err := cmd.CombinedOutput()
	log.Print(string(output))

	return err
}
