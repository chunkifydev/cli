package chunkify

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	_ "embed"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/chunkify/hooks"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	ProgressUpdateInterval = 1 * time.Second
)

type ChunkifyCommand struct {
	Id                     string
	Input                  string
	Output                 string
	Format                 string
	JobFormatParams        chunkify.JobCreateFormatParams
	JobTranscoderParams    *chunkify.JobCreateTranscoderParams
	JobCreateStorageParams *chunkify.JobCreateStorageParams
}

// Command represents the root notifications command and configuration
type Command struct {
	Command *cobra.Command // The root cobra command for notifications
	Config  *config.Config // Configuration for the notifications command
	App     *App           // The TUI app
}

func NewCommand(cfg *config.Config) *Command {
	app := &App{}
	cmd := &Command{
		App:    app,
		Config: cfg,
		Command: &cobra.Command{
			Use:   "chunkify",
			Short: "Transcode videos with Chunkify",
			Long:  "Transcode videos with Chunkify",
			Run: func(cmd *cobra.Command, args []string) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				app.Status = Starting
				app.Progress = NewProgress()
				app.Ctx = ctx
				app.CancelFunc = cancel
				app.Done = false
				app.DownloadedFiles = map[string]chunkify.File{}
				app.Client = cfg.Client

				// Start all background work in a goroutine
				go app.executeWorkflow(app.Ctx)

				// Run TUI synchronously - this will block until the TUI exits
				app.Run()
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
	// Create job
	app.Job, err = app.CreateJob(source)
	if err != nil {
		app.setError(err)
		return
	}

	// Start job progress monitoring
	go app.StartJobProgress(ctx, app.Job.Id)

	// Wait for job completion
	select {
	case <-app.Progress.JobCompleted:
	case <-ctx.Done():
		return
	}

	// Check if job failed
	if app.Job != nil && jobHasFailed(app.Job.Status) {
		err := fmt.Errorf("job failed with status: %s: %s", app.Job.Status, app.Job.Error.Message)
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
		files, err := app.Client.JobListFiles(app.Job.Id)
		if err != nil {
			app.setError(err)
			return
		}
		app.Progress.Files <- files
		downloadedFiles, err := downloadFiles(ctx, app, files)
		if err != nil {
			app.setError(err)
			return
		}

		// Post process files if format is jpg or hls
		// this is to rename the paths inside m3u8 and vtt files to the correct name
		if app.Command.Format == string(chunkify.FormatJpg) || strings.HasPrefix(app.Command.Format, "hls") {
			if err := hooks.Process(app.Command.Format, app.Job.Id, files, downloadedFiles); err != nil {
				app.setError(err)
				return
			}
		}
	}

	// Mark as completed
	// give enough time to display the completed message
	time.Sleep(1 * time.Second)
	app.Progress.Status <- Completed
}

func downloadFiles(ctx context.Context, app *App, files []chunkify.File) ([]string, error) {
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

		filepath := filename(file, app.Command.Output)

		if err := DownloadFile(ctx, file, filepath, app.Progress.DownloadProgress); err == nil {
			app.Progress.DownloadedFiles <- file
			downloadedFiles = append(downloadedFiles, filepath)
		}
	}

	return downloadedFiles, nil
}

func jobHasFailed(status string) bool {
	return status == chunkify.JobStatusFailed || status == chunkify.JobStatusCancelled
}

func (a *App) setError(err error) {
	a.Progress.Status <- Failed
	a.Progress.Error <- err
	// give time to display the error message before tea quit
	time.Sleep(1 * time.Second)
}

func (a *App) CreateSource(ctx context.Context) (*chunkify.Source, error) {
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
		source, err := a.Client.Source(a.Command.Input)
		if err != nil {
			return nil, err
		}
		return &source, nil
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
	source, err := a.Client.SourceCreate(chunkify.SourceCreateParams{
		Url: a.Command.Input,
		Metadata: chunkify.SourceCreateParamsMetadata{
			"chunkify_execution_id": a.Command.Id,
		},
	})
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (a *App) CreateSourceFromFile(ctx context.Context) (*chunkify.Source, error) {
	a.Progress.Status <- UploadingFromFile

	file, err := os.Open(a.Command.Input)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	md5, err := fileMD5(file)
	if err != nil {
		return nil, fmt.Errorf("error calculating file md5: %s", err)
	}

	// Try to find the source by MD5, so we don't upload the same file again
	if source, err := a.GetSourceByMd5(md5); err == nil {
		return source, nil
	}

	upload, err := a.Client.UploadCreate(chunkify.UploadCreateParams{
		Metadata: chunkify.UploadCreateParamsMetadata{
			"chunkify_execution_id": a.Command.Id,
			"md5":                   md5,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating upload: %s", err)
	}

	// Reset file pointer to the beginning
	file.Seek(0, io.SeekStart)

	if err := a.Client.UploadBlobWithProgress(ctx, file, upload, a.Progress.UploadProgress); err != nil {
		return nil, fmt.Errorf("error uploading blob: %s", err)
	}

	found := false

	retry := 0
	maxRetries := 30
	for !found && retry < maxRetries {
		results, err := a.Client.SourceList(chunkify.SourceListParams{
			Metadata: map[string]string{
				"chunkify_execution_id": a.Command.Id,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error listing sources: %s", err)
		}
		for _, source := range results.Items {
			if v, ok := source.Metadata.(map[string]any); ok && v["chunkify_execution_id"] == a.Command.Id {
				found = true
				return &source, nil
			}
		}
		time.Sleep(1 * time.Second)
		retry++
	}

	return nil, fmt.Errorf("source not found")
}

func (a *App) GetSourceByMd5(md5 string) (*chunkify.Source, error) {
	sources, err := a.Client.SourceList(chunkify.SourceListParams{
		Metadata: map[string]string{
			"md5": md5,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing sources: %s", err)
	}
	for _, source := range sources.Items {
		if v, ok := source.Metadata.(map[string]any); ok && v["md5"] == md5 {
			if source.CreatedAt.Before(time.Now().Add(-12 * time.Hour)) {
				return nil, fmt.Errorf("source is too old, upload again")
			}
			return &source, nil
		}
	}
	return nil, fmt.Errorf("source not found")
}

func (a *App) CreateJob(source *chunkify.Source) (*chunkify.Job, error) {
	a.Progress.Status <- Transcoding

	job, err := a.Client.JobCreate(chunkify.JobCreateParams{
		SourceId:      source.Id,
		Format:        a.Command.JobFormatParams,
		Transcoder:    a.Command.JobTranscoderParams,
		Storage:       a.Command.JobCreateStorageParams,
		HlsManifestId: hlsManifestId,
		Metadata: chunkify.JobCreateParamsMetadata{
			"chunkify_execution_id": a.Command.Id,
		},
	})

	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (a *App) StartJobProgress(ctx context.Context, jobId string) {
	ticker := time.NewTicker(ProgressUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.Progress.JobCompleted <- true
			return
		case <-ticker.C:
			job, err := a.Client.Job(jobId)
			if err != nil {
				return
			}
			a.Job = &job
			a.Progress.JobProgress <- job

			if job.Status == chunkify.JobStatusCompleted || jobHasFailed(job.Status) {
				a.Progress.JobCompleted <- true
				break
			}

			transcoders, err := a.Client.JobListTranscoders(job.Id)
			if err != nil {
				return
			}
			a.Progress.JobTranscoders <- transcoders
		}
	}
}

func filename(file chunkify.File, output string) string {
	fileBase := strings.Replace(path.Base(output), path.Ext(output), "", 1)
	newFilename := strings.Replace(path.Base(file.Path), file.JobId, fileBase, 1)
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
