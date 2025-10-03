package chunkify

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	_ "embed"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
)

const (
	ProgressUpdateInterval = 1 * time.Second
)

type ChunkifyCommand struct {
	Id                  string
	Input               string
	Output              string
	Format              string
	JobFormatParams     chunkify.JobCreateFormatParams
	JobTranscoderParams *chunkify.JobCreateTranscoderParams
}

func init() {
	logFile, err := os.Create("chunkify.log")
	if err != nil {
		fmt.Println("error creating log file", err)
		return
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(logFile, nil)))
}

func Execute(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApp(ctx, cancel, cfg)
	go app.Run()

	// Create source in a goroutine so we can check for cancellation
	sourceChan := make(chan *chunkify.Source, 1)
	errChan := make(chan error, 1)

	go func() {
		source, err := app.CreateSource(ctx)
		if err != nil {
			errChan <- err
			return
		}
		sourceChan <- source
	}()

	// Wait for either source creation or context cancellation
	var source *chunkify.Source
	select {
	case source = <-sourceChan:
		app.Progress.Source <- source
	case err := <-errChan:
		app.setError(err)
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	}

	var err error
	app.Job, err = app.CreateJob(source)
	if err != nil {
		app.setError(err)
		return err
	}

	go app.StartJobProgress(ctx, app.Job.Id)

	// Wait for either job completion or context cancellation
	select {
	case <-app.Progress.JobCompleted:
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	}

	if app.Job != nil && jobHasFailed(app.Job.Status) {
		err := fmt.Errorf("job failed with status: %s: %s", app.Job.Status, app.Job.Error.Message)
		app.setError(err)
		return err
	}

	// Check if context was cancelled before getting files
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	default:
	}

	if app.Command.Output != "" {
		files, err := app.Client.JobListFiles(app.Job.Id)
		if err != nil {
			app.setError(err)
			return fmt.Errorf("error getting files: %s", err)
		}
		app.Progress.Files <- files
		if err := downloadFiles(ctx, &app, files); err != nil {
			app.setError(err)
			return fmt.Errorf("error downloading files: %s", err)
		}
	}

	app.Progress.Status <- Completed
	// Give the TUI time to display the completion message
	time.Sleep(1 * time.Second)

	return nil
}

func downloadFiles(ctx context.Context, app *App, files []chunkify.File) error {
	app.Progress.Status <- Downloading

	slog.Info("Downloading files", "files", files)
	downloadedFiles := []string{}

	for _, file := range files {
		// Check if context was cancelled before each download
		select {
		case <-ctx.Done():
			return fmt.Errorf("download cancelled")
		default:
		}

		filepath := filename(file, app.Command.Output)
		if err := DownloadFile(ctx, file, filepath, app.Progress.DownloadProgress); err == nil {
			app.Progress.DownloadedFiles <- file
			downloadedFiles = append(downloadedFiles, filepath)
		}

	}

	// If format is jpg
	// rename all vtt cues to match the filename set in --output flag
	if app.Command.Format == string(chunkify.FormatJpg) {
		if err := postProcessVtt(downloadedFiles, app.Job.Id); err != nil {
			return fmt.Errorf("post process vtt: %w", err)
		}
	} else if strings.HasPrefix(app.Command.Format, "hls") {
		if err := postProcessM3u8(downloadedFiles, app.Job.Id); err != nil {
			return fmt.Errorf("post process m3u8: %w", err)
		}
	}

	return nil
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
	upload, err := a.Client.UploadCreate(chunkify.UploadCreateParams{
		Metadata: chunkify.UploadCreateParamsMetadata{
			"chunkify_execution_id": a.Command.Id,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating upload: %s", err)
	}
	file, err := os.Open(a.Command.Input)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

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

func (a *App) CreateJob(source *chunkify.Source) (*chunkify.Job, error) {
	a.Progress.Status <- Transcoding

	job, err := a.Client.JobCreate(chunkify.JobCreateParams{
		SourceId:      source.Id,
		Format:        a.Command.JobFormatParams,
		Transcoder:    a.Command.JobTranscoderParams,
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

func postProcessVtt(downloadedFiles []string, jobId string) error {
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
			parts := strings.Split(path.Base(filepath), "-")
			if len(parts) >= 2 {
				imageBasename = strings.Join(parts[0:len(parts)-1], "-")
			}
		}
	}

	vttContent = []byte(strings.ReplaceAll(string(vttContent), jobId, imageBasename))
	if err := os.WriteFile(vttPath, vttContent, 0644); err != nil {
		return fmt.Errorf("write vtt file: %w", err)
	}
	return nil
}

func postProcessM3u8(downloadedFiles []string, jobId string) error {
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

	m3u8Content = []byte(strings.ReplaceAll(string(m3u8Content), jobId, videoBasename))
	if err := os.WriteFile(m3u8Path, m3u8Content, 0644); err != nil {
		return fmt.Errorf("write m3u8 file: %w", err)
	}
	manifestContent = []byte(strings.ReplaceAll(string(manifestContent), jobId, videoBasename))
	if err := os.WriteFile(manifestPath, manifestContent, 0644); err != nil {
		return fmt.Errorf("write manifest file: %w", err)
	}
	return nil
}
