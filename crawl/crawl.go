package crawl

import (
	"os"
	"path/filepath"
)

// WalkDirectory : Function that loops through the files of a directory
func WalkDirectory(filedir string) (result []string, err error) {
	err = filepath.Walk(filedir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	return
}
