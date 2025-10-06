package hooks

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func ProcessVtt(downloadedFiles []string, basename string) error {
	var vttContent []byte
	var imageBasename string
	var vttPath string
	var err error

	for _, filepath := range downloadedFiles {
		switch path.Ext(filepath) {
		case ".vtt":
			vttPath = filepath
			vttContent, err = os.ReadFile(filepath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
		case ".jpg":
			if imageBasename != "" {
				continue
			}
			parts := strings.Split(path.Base(filepath), "-")
			if len(parts) >= 2 {
				imageBasename = strings.Join(parts[0:len(parts)-1], "-")
			}
		}
	}

	if imageBasename == "" {
		return fmt.Errorf("image basename not found")
	}

	vttContent = []byte(strings.ReplaceAll(string(vttContent), basename, imageBasename))
	if err := os.WriteFile(vttPath, vttContent, 0644); err != nil {
		return fmt.Errorf("write vtt file: %w", err)
	}
	return nil
}
