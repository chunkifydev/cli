package hooks

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/chunkifydev/chunkify-go"
)

const ManifestFileName = "manifest.m3u8"

func Process(format string, defaultBasename string, files []chunkify.File, downloadedFiles []string) error {
	// For HLS, we need to merge the previous manifest.m3u8
	var oldManifestContent []byte
	var basename string

	for _, file := range files {
		// we get a sample filename to rename everything with the correct name
		if basename == "" && path.Ext(file.Path) != ".m3u8" && path.Ext(file.Path) != ".vtt" {
			basename = path.Base(file.Path)
			break
		}
	}

	slog.Info("Downloaded files", "basename", basename)
	if basename == "" {
		basename = defaultBasename
	}

	if path.Ext(basename) == ".jpg" {
		parts := strings.Split(basename, "-")
		basename = strings.Join(parts[0:len(parts)-1], "-")
	}

	basename = strings.Replace(basename, path.Ext(basename), "", 1)

	for _, filepath := range downloadedFiles {
		// check if we have a manifest.m3u8 already
		// if so we will merge it with the new manifest.m3u8
		if strings.HasSuffix(filepath, ManifestFileName) {
			if _, err := os.Stat(filepath); err == nil {
				manifestContent, err := os.ReadFile(filepath)
				if err != nil {
					return fmt.Errorf("read manifest file: %w", err)
				}
				oldManifestContent = manifestContent
			}
		}
	}

	// If format is jpg
	// rename all vtt cues to match the filename set in --output flag
	if format == string(chunkify.FormatJpg) {
		if err := ProcessVtt(downloadedFiles, basename); err != nil {
			return fmt.Errorf("post process vtt: %w", err)
		}
	} else if strings.HasPrefix(format, "hls") {
		if err := ProcessM3u8(downloadedFiles, basename, oldManifestContent); err != nil {
			return fmt.Errorf("post process m3u8: %w", err)
		}
	}

	return nil
}
