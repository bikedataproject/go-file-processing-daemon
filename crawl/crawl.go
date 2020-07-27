package crawl

import (
	"os"
	"path/filepath"
	"strings"
)

// WalkDirectory : Function that loops through the files of a directory
func WalkDirectory(filedir string, filetype string) (result []string, err error) {
	err = filepath.Walk(filedir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.Contains(strings.ToLower(path), strings.ToLower(filetype)) {
			result = append(result, path)
		}
		return nil
	})
	return
}
