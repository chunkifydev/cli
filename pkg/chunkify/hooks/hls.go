package hooks

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"
)

func ProcessM3u8(downloadedFiles []string, basename string, oldManifestContent []byte) error {
	var (
		m3u8Content, manifestContent []byte
		videoBasename                string
		m3u8Path, manifestPath       string
		err                          error
	)

	for _, filepath := range downloadedFiles {
		switch path.Ext(filepath) {
		case ".m3u8":
			if strings.HasSuffix(filepath, "manifest.m3u8") {
				manifestPath = filepath
				manifestContent, err = os.ReadFile(filepath)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				if len(oldManifestContent) > 0 {
					manifestContent = mergeManifest(manifestContent, oldManifestContent)
				}
			} else {
				m3u8Path = filepath
				m3u8Content, err = os.ReadFile(filepath)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
			}
		case ".mp4":
			videoBasename = strings.Replace(path.Base(filepath), ".mp4", "", 1)
		}
	}

	m3u8Content = []byte(strings.ReplaceAll(string(m3u8Content), basename, videoBasename))
	if err := os.WriteFile(m3u8Path, m3u8Content, 0644); err != nil {
		return fmt.Errorf("write m3u8 file: %w", err)
	}
	manifestContent = []byte(strings.ReplaceAll(string(manifestContent), basename, videoBasename))
	if err := os.WriteFile(manifestPath, manifestContent, 0644); err != nil {
		return fmt.Errorf("write manifest file: %w", err)
	}
	return nil
}

func mergeManifest(manifestContent []byte, oldManifestContent []byte) []byte {
	oldManifestLines := strings.Split(string(oldManifestContent), "\n")
	currentManifestLines := strings.Split(string(manifestContent), "\n")

	for i, line := range oldManifestLines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			// locate this line in currentManifestLines
			// and replace the next line (m3u8 path) with the next line from oldManifestLines
			lineIndex := slices.Index(currentManifestLines, line)

			slog.Info("Looking for line", "line", line, "lineIndex", lineIndex)
			if lineIndex == -1 {
				continue
			}

			if len(currentManifestLines) > lineIndex+1 && len(oldManifestLines) > i+1 {
				currentManifestLines[lineIndex+1] = oldManifestLines[i+1]
			}
		}
	}
	return []byte(strings.Join(currentManifestLines, "\n"))
}
