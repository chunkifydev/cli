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

// configureStorage applies only the saved CLI storage override. With no config,
// IDs and paths stay as supplied and the API resolves omitted IDs.
func (a *App) configureStorage(storageID string) error {
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
		outputPath, err := a.storageObjectPath(storageID, a.Command.JobCreateStorageParams.Path.Value, "jobs", a.jobOutputFilename(), "--output-storage-path")
		if err != nil {
			return err
		}
		a.Command.JobCreateStorageParams = chunkify.JobNewParamsStorage{ID: chunkify.String(storageID)}
		if outputPath != "" {
			a.Command.JobCreateStorageParams.Path = chunkify.String(outputPath)
		}
	}
	return nil
}

func (a *App) jobOutputFilename() string {
	if a.Command.Output != "" {
		return filepath.Base(a.Command.Output)
	}
	ext := ".mp4"
	switch a.Command.Format {
	case FormatWebmVp9:
		ext = ".webm"
	case FormatHlsH264, FormatHlsH265, FormatHlsAv1:
		ext = ".m3u8"
	case FormatJpg:
		ext = ".jpg"
	}
	return "output" + ext
}
