package chunkify

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	chunkify "github.com/chunkifydev/chunkify-go"
)

func (a *App) validateSourceInput() error {
	if !strings.HasPrefix(a.Command.Input, "store://") {
		if a.Command.SourceStorageID != "" {
			return fmt.Errorf("--source-storage-id requires a store:// input")
		}
		return nil
	}
	if a.Command.UploadStorageID != "" || a.Command.UploadStoragePath != "" {
		return fmt.Errorf("store:// reads an existing object; use --source-storage-id instead of upload storage flags")
	}
	_, err := a.sourceStorageParams()
	return err
}

func (a *App) sourceStorageParams() (chunkify.SourceNewParamsStorage, error) {
	// This is an exact object key, not a URL: preserve slashes, escapes, and
	// characters such as ? and # without decoding or path normalization.
	key, storedSource := strings.CutPrefix(a.Command.Input, "store://")
	if !storedSource || len(key) == 0 || len(key) > 1024 || !utf8.ValidString(key) {
		return chunkify.SourceNewParamsStorage{}, fmt.Errorf("store:// input requires an object key of 1 to 1024 UTF-8 bytes")
	}
	if strings.HasPrefix(a.Command.SourceStorageID, "stor_chunkify_") {
		return chunkify.SourceNewParamsStorage{}, fmt.Errorf("store:// requires external storage; choose --source-storage-id with an external storage ID")
	}
	params := chunkify.SourceNewParamsStorage{Path: key}
	if a.Command.SourceStorageID != "" {
		params.ID = chunkify.String(a.Command.SourceStorageID)
	}
	return params, nil
}

// CreateSourceFromStorage references an existing object without creating an upload.
func (a *App) CreateSourceFromStorage(ctx context.Context) (*chunkify.Source, error) {
	storage, err := a.sourceStorageParams()
	if err != nil {
		return nil, err
	}
	a.Progress.Status <- ReadingFromStorage
	source, err := a.Client.Sources.New(ctx, chunkify.SourceNewParams{
		Storage: storage,
		Metadata: map[string]string{
			"origin":           MetadataOrigin,
			"cli_execution_id": a.Command.Id,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating source from storage: %w", err)
	}
	return source, nil
}
