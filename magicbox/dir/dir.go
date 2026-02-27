// Package dir provides a directory utility.
package dir

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandHomeDir expands the ~ symbol in the path to the user's home directory
func ExpandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
