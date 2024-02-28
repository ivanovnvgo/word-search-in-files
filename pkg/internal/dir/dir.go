package dir

import (
	"fmt"
	"io/fs"
)

// FilesFS generates a list of attached files
func FilesFS(fsys fs.FS, dir string) ([]string, error) {
	if dir == "" {
		dir = "."
	}
	var fileNames []string
	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileNames = append(fileNames, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't read directory %s, err: %v", dir, err)
	}
	return fileNames, nil
}
