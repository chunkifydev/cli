package chunkify

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/google/uuid"
)

// storageObjectPath leaves managed storage paths to the API and preserves
// explicit external paths. Generated paths are stable within one CLI execution.
func (a *App) storageObjectPath(storageID, explicitPath, category, filename, flag string) (string, error) {
	if strings.HasPrefix(storageID, "stor_chunkify_") {
		if explicitPath != "" {
			return "", fmt.Errorf("%s cannot be used with Chunkify-managed storage", flag)
		}
		return "", nil
	}
	if explicitPath != "" {
		return explicitPath, nil
	}
	if a.Command.Id == "" {
		a.Command.Id = uuid.New().String()
	}
	return path.Join("chunkify-cli", category, a.Command.Id, filepath.Base(filename)), nil
}

// configureStorage applies the saved storage override, then checks that the
// output object path ends in the format's extension, which the API requires.
func (a *App) configureStorage(storageID string) error {
	if err := a.applyStorageConfig(storageID); err != nil {
		return err
	}
	return a.validateOutputPathExtension()
}

// applyStorageConfig applies only the saved CLI storage override. With no
// config, IDs and paths stay as supplied and the API resolves omitted IDs.
func (a *App) applyStorageConfig(storageID string) error {
	storedSource := strings.HasPrefix(a.Command.Input, "store://")
	if storedSource {
		if a.Command.SourceStorageID == "" {
			a.Command.SourceStorageID = storageID
		}
		if _, err := a.sourceStorageParams(); err != nil {
			return err
		}
	}
	if storageID == "" {
		return nil
	}
	if !storedSource {
		a.Command.UploadStorageID = storageID
	}
	a.Command.OutputStorageID = storageID
	input := a.Command.Input
	if !storedSource && !strings.HasPrefix(input, "src_") && !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		uploadPath, err := a.storageObjectPath(storageID, a.Command.UploadStoragePath, "sources", input, "--upload-storage-path")
		if err != nil {
			return err
		}
		a.Command.UploadStoragePath = uploadPath
	}
	if a.Command.Format != "" {
		outputPath, err := a.storageObjectPath(storageID, a.Command.TranscodingStorageParams.Path.Value, "transcodings", a.outputFilename(), "--output-storage-path")
		if err != nil {
			return err
		}
		a.Command.TranscodingStorageParams = chunkify.TranscodingCreateStorageParams{ID: chunkify.String(storageID)}
		if outputPath != "" {
			a.Command.TranscodingStorageParams.Path = chunkify.String(outputPath)
		}
	}
	return nil
}

// validateOutputPathExtension rejects an output object path whose extension
// doesn't match the format, for example -f m4a_aac with -o result.mp4 on
// external storage. Without an object path (local output only, or
// Chunkify-managed storage), the API names the file and nothing is checked.
func (a *App) validateOutputPathExtension() error {
	objectPath := a.Command.TranscodingStorageParams.Path.Value
	if a.Command.Format == "" || objectPath == "" {
		return nil
	}
	want := formatExtension(a.Command.Format)
	if !strings.EqualFold(path.Ext(objectPath), want) {
		return fmt.Errorf("output object path %q must end in %s for %s; change --output or --output-storage-path", objectPath, want, a.Command.Format)
	}
	return nil
}

// formatExtension returns the file extension the API expects for a format.
func formatExtension(format string) string {
	switch format {
	case FormatWebmVp9:
		return ".webm"
	case FormatM4aAac:
		return ".m4a"
	default:
		return ".mp4"
	}
}

func (a *App) outputFilename() string {
	if a.Command.Output != "" {
		return filepath.Base(a.Command.Output)
	}
	return "output" + formatExtension(a.Command.Format)
}
