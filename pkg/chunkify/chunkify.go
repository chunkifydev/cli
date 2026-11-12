package chunkify

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	_ "embed"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	ProgressUpdateInterval = 1 * time.Second
	MetadataOrigin         = "chunkify/cli"
)

type ChunkifyCommand struct {
	Id                       string
	Input                    string
	Output                   string
	Format                   string
	SourceStorageID          string
	UploadStorageID          string
	UploadStoragePath        string
	OutputStorageID          string
	TranscodingFormatParams  chunkify.TranscodingNewParamsFormatUnion
	TranscodingStorageParams chunkify.TranscodingNewParamsStorage
}

// Command represents the root notifications command and configuration
type Command struct {
	Command *cobra.Command // The root cobra command for notifications
	Config  *config.Config // Configuration for the notifications command
	App     *App           // The TUI app
}

func NewCommand(cfg *config.Config) *Command {
	app := NewApp()

	cmd := &Command{
		App:    app,
		Config: cfg,
		Command: &cobra.Command{
			Use:   "chunkify",
			Short: "Transcode videos with Chunkify CLI",
			Long: `Transcode videos with Chunkify CLI

The Chunkify CLI brings super-fast video transcoding to your terminal. With a single command, you can upload local files, transcode videos using Chunkify's parallel technology, and download the processed files to your local disk.

Examples:

Make a video 1080p H264
chunkify -i video.mp4 -o video_1080p.mp4 -f mp4_h264 -s 1920x1080 --crf 21

Upload a video only and get the Source ID
chunkify -i video.mp4

Read a video already in connected storage
chunkify -i store://videos/input.mp4 --source-storage-id stor_aws_example -o output.mp4

Extract the audio track
chunkify -i video.mp4 -o audio.m4a -f m4a_aac --ab 192K

Use specific profile to use a different project
chunkify config token sk_project_token --profile your_profile
chunkify -i video.mp4 -f mp4_av1 --preset 7 -o video_1080p.mp4 --profile your_profile
`,
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := cfg.LoadStorageID(); err != nil {
					return err
				}
				if err := app.configureStorage(cfg.StorageID); err != nil {
					return err
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				app.Ctx = ctx
				app.CancelFunc = cancel
				app.Client = cfg.Client

				// Start all background work in a goroutine
				go app.executeWorkflow(app.Ctx)

				// Run TUI synchronously - this will block until the TUI exits
				app.Run()
				return nil
			},
		},
	}

	BindFlags(app, cmd.Command)
	return cmd
}

func init() {
	logFile, err := os.Create("chunkify.log")
	if err != nil {
		fmt.Println("error creating log file", err)
		return
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(logFile, nil)))
}

// executeWorkflow runs all the background work and communicates with the TUI via channels
func (app *App) executeWorkflow(ctx context.Context) {
	// Create source
	source, err := app.CreateSource(ctx)
	if err != nil {
		app.setError(err)
		return
	}
	app.Progress.Source <- source

	// No format specified, we are done
	if app.Command.Format == "" {
		time.Sleep(1 * time.Second)
		app.Progress.Status <- Completed

		return
	}
	// Create transcoding
	app.Transcoding, err = app.CreateTranscoding(ctx, source)
	if err != nil {
		app.setError(err)
		return
	}

	// Start transcoding progress monitoring
	go app.StartTranscodingProgress(ctx, app.Transcoding.ID)

	// Wait for the transcoding to finish
	select {
	case <-app.Progress.TranscodingCompleted:
	case <-ctx.Done():
		return
	}

	// Check if transcoding failed
	if app.Transcoding != nil && transcodingHasFailed(app.Transcoding.Status) {
		err := fmt.Errorf("transcoding %s", app.Transcoding.Status)
		if app.Transcoding.Error.Message != "" {
			err = fmt.Errorf("%w: %s", err, app.Transcoding.Error.Message)
		}
		app.setError(err)
		return
	}

	// Check if context was cancelled before getting files
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Download files if output is specified
	if app.Command.Output != "" {
		files := app.Transcoding.Files
		app.Progress.Files <- files
		if _, err := downloadFiles(ctx, app, files); err != nil {
			app.setError(err)
			return
		}
	}

	// Mark as completed
	// give enough time to display the completed message
	time.Sleep(1 * time.Second)
	app.Progress.Status <- Completed
}

func downloadFiles(ctx context.Context, app *App, files []chunkify.TranscodingFile) ([]string, error) {
	app.Progress.Status <- Downloading

	slog.Info("Downloading files", "files", files)
	downloadedFiles := []string{}

	for _, file := range files {
		// Check if context was cancelled before each download
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("download cancelled")
		default:
		}

		filepath := filename(file, app.Command.Output, len(files))

		if err := DownloadFile(ctx, file, filepath, app.Progress.DownloadProgress); err != nil {
			return nil, fmt.Errorf("error downloading %s: %w", path.Base(file.Path), err)
		}
		app.Progress.DownloadedFiles <- file
		downloadedFiles = append(downloadedFiles, filepath)
	}

	return downloadedFiles, nil
}

func transcodingHasFailed(status chunkify.TranscodingStatus) bool {
	return status == chunkify.TranscodingStatusFailed ||
		status == chunkify.TranscodingStatusCancelled ||
		status == chunkify.TranscodingStatusExpired
}

func (a *App) setError(err error) {
	a.Progress.Status <- Failed
	a.Progress.Error <- err
	// give time to display the error message before tea quit
	time.Sleep(1 * time.Second)
}

func (a *App) CreateSource(ctx context.Context) (*chunkify.Source, error) {
	if strings.HasPrefix(a.Command.Input, "store://") {
		return a.CreateSourceFromStorage(ctx)
	}
	// check input if it's a valid file or URL
	if strings.HasPrefix(a.Command.Input, "https://") || strings.HasPrefix(a.Command.Input, "http://") {
		// create source directly from URL
		source, err := a.CreateSourceFromUrl()
		if err != nil {
			return nil, err
		}

		return source, nil
	}

	// the input is a source id (already uploaded)
	if strings.HasPrefix(a.Command.Input, "src_") {
		source, err := a.Client.Sources.Get(ctx, a.Command.Input)
		if err != nil {
			return nil, err
		}
		return source, nil
	}

	// it's a path file, check if it's a valid file
	if _, err := os.Stat(a.Command.Input); err != nil {
		return nil, fmt.Errorf("file not found: %s", a.Command.Input)
	}

	source, err := a.CreateSourceFromFile(ctx)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (a *App) CreateSourceFromUrl() (*chunkify.Source, error) {
	a.Progress.Status <- UploadingFromUrl
	source, err := a.Client.Sources.New(context.Background(), chunkify.SourceNewParams{
		URL: chunkify.String(a.Command.Input),
		Metadata: map[string]string{
			"origin":           MetadataOrigin,
			"cli_execution_id": a.Command.Id,
		},
	})
	if err != nil {
		return nil, err
	}
	return source, nil
}

func (a *App) CreateSourceFromFile(ctx context.Context) (*chunkify.Source, error) {
	a.Progress.Status <- UploadingFromFile

	file, err := os.Open(a.Command.Input)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err)
	}
	size := info.Size()
	if size == 0 {
		return nil, fmt.Errorf("file is empty: %s", a.Command.Input)
	}

	md5, err := fileMD5(file)
	if err != nil {
		return nil, fmt.Errorf("error calculating file md5: %s", err)
	}

	// Reuse an existing source unless a specific upload destination was requested.
	if a.Command.UploadStorageID == "" && a.Command.UploadStoragePath == "" {
		if source, err := a.GetSourceByMd5(ctx, md5); err == nil {
			return source, nil
		}
	}

	params := chunkify.UploadNewParams{
		// The API picks a single PUT or multipart transfer from the size.
		FileSizeBytes: chunkify.Int(size),
		Metadata: map[string]string{
			"origin":           MetadataOrigin,
			"cli_execution_id": a.Command.Id,
			"md5":              md5,
		},
	}
	if a.Command.UploadStorageID != "" {
		params.Storage.ID = chunkify.String(a.Command.UploadStorageID)
	}
	if a.Command.UploadStoragePath != "" {
		params.Storage.Path = chunkify.String(a.Command.UploadStoragePath)
	}
	upload, err := a.Client.Uploads.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating upload: %s", err)
	}
	if upload.CompletionURL == "" ||
		(upload.UploadMode == chunkify.UploadUploadModeMultipart && len(upload.Multipart.Parts) == 0) ||
		(upload.UploadMode != chunkify.UploadUploadModeMultipart && upload.UploadURL == "") {
		return nil, fmt.Errorf("missing upload session URLs")
	}
	completionURL, err := url.Parse(upload.CompletionURL)
	if err != nil || completionURL.Host == "" ||
		(completionURL.Scheme != "https" && completionURL.Scheme != "http") ||
		completionURL.Path == "" || strings.HasSuffix(completionURL.Path, "/") {
		return nil, fmt.Errorf("invalid upload completion URL")
	}
	token := path.Base(completionURL.Path)

	// Both the transfer and completion must finish before the session expires.
	uploadCtx := ctx
	if !upload.ExpiresAt.IsZero() {
		var cancel context.CancelFunc
		uploadCtx, cancel = context.WithDeadline(ctx, upload.ExpiresAt)
		defer cancel()
	}

	completion, err := UploadFileWithProgress(uploadCtx, file, size, upload, a.Progress.UploadProgress)
	if err != nil {
		return nil, fmt.Errorf("error uploading blob: %s", err)
	}

	// The SDK omits project authentication and retries transient completion errors
	// without repeating the file transfer.
	if err := a.Client.Uploads.Complete(uploadCtx, token, completion); err != nil {
		// SDK errors include the request URL, whose token must stay private.
		return nil, fmt.Errorf("error completing upload: %s", strings.ReplaceAll(err.Error(), token, "[redacted]"))
	}

	// Completion creates the source relationship before returning 204.
	upload, err = a.Client.Uploads.Get(ctx, upload.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving completed upload: %w", err)
	}
	if upload.SourceID == "" {
		return nil, fmt.Errorf("completed upload has no source")
	}

	source, err := a.Client.Sources.Get(ctx, upload.SourceID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving uploaded source: %w", err)
	}
	return source, nil
}

func (a *App) GetSourceByMd5(ctx context.Context, md5 string) (*chunkify.Source, error) {
	sources, err := a.Client.Sources.List(ctx, chunkify.SourceListParams{
		Metadata: [][]string{
			{"md5", md5},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing sources: %s", err)
	}
	for _, source := range sources.Data {
		if source.Metadata != nil {
			if v, ok := source.Metadata["md5"]; ok && v == md5 {
				if source.CreatedAt.Before(time.Now().Add(-12 * time.Hour)) {
					return nil, fmt.Errorf("source is too old, upload again")
				}
				return &source, nil
			}
		}
	}
	return nil, fmt.Errorf("source not found")
}

func (a *App) CreateTranscoding(ctx context.Context, source *chunkify.Source) (*chunkify.Transcoding, error) {
	a.Progress.Status <- Transcoding

	return a.Client.Transcodings.New(ctx, chunkify.TranscodingNewParams{
		SourceID: source.ID,
		Format:   a.Command.TranscodingFormatParams,
		Storage:  a.Command.TranscodingStorageParams,
		Metadata: map[string]string{
			"origin":           MetadataOrigin,
			"cli_execution_id": a.Command.Id,
		},
	})
}

func (a *App) StartTranscodingProgress(ctx context.Context, transcodingID string) {
	ticker := time.NewTicker(ProgressUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.Progress.TranscodingCompleted <- true
			return
		case <-ticker.C:
			transcoding, err := a.Client.Transcodings.Get(ctx, transcodingID)
			if err != nil {
				slog.Error("Error retrieving transcoding", "id", transcodingID, "error", err)
				continue
			}
			a.Transcoding = transcoding
			a.Progress.TranscodingProgress <- *transcoding

			if transcoding.Status == chunkify.TranscodingStatusReady || transcodingHasFailed(transcoding.Status) {
				a.Progress.TranscodingCompleted <- true
				return
			}
		}
	}
}

// filename returns the local path for a transcoded file. A single output is
// saved at the requested path; multiple outputs replace the transcoding ID in
// their remote names with the requested base name.
func filename(file chunkify.TranscodingFile, output string, count int) string {
	if count <= 1 {
		return output
	}
	fileBase := strings.Replace(path.Base(output), path.Ext(output), "", 1)
	newFilename := strings.Replace(path.Base(file.Path), file.TranscodingID, fileBase, 1)
	return path.Join(path.Dir(output), newFilename)
}

func fileMD5(file *os.File) (string, error) {
	w := md5.New()
	_, err := io.Copy(w, file)
	if err != nil {
		return "", fmt.Errorf("md5 copy: %w", err)
	}

	rawHash := w.Sum(nil)
	return hex.EncodeToString(rawHash), nil
}
