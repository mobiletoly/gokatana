package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func ListFilesInDirSortedByFilename(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fn := filepath.Join(dir, file.Name())
			fileNames = append(fileNames, fn)
		}
	}
	return fileNames, nil
}
