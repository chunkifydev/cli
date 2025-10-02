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

	tea "github.com/charmbracelet/bubbletea"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/google/uuid"
)

const (
	ProgressUpdateInterval = 1 * time.Second
)

type ChunkifyCommand struct {
	Id              string
	Config          *config.Config
	Input           string
	Output          string
	Format          string
	JobFormatParams chunkify.JobCreateFormatParams
	Transcoders     *int64
	TranscoderVcpu  *int64
	Tui             *TUI
}

var chunkifyCmd = ChunkifyCommand{}

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

	tui := NewTUI()
	tui.Command = &chunkifyCmd
	tui.Ctx = ctx
	tui.CancelFunc = cancel
	chunkifyCmd.Tui = &tui

	chunkifyCmd.Tui.Progress.Status <- Starting
	chunkifyCmd.Config = cfg

	go func() {
		//fmt.Println("Starting TUI", tui)
		p := tea.NewProgram(tui)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	}()

	chunkifyCmd.Id = uuid.New().String()
	chunkifyCmd.InitJobFormatParams()

	// Create source in a goroutine so we can check for cancellation
	sourceChan := make(chan *chunkify.Source, 1)
	errChan := make(chan error, 1)

	go func() {
		source, err := chunkifyCmd.CreateSource()
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
		chunkifyCmd.Tui.Progress.Source <- source
	case err := <-errChan:
		setError(&chunkifyCmd, err)
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	}

	job, err := chunkifyCmd.CreateJob(source)
	if err != nil {
		setError(&chunkifyCmd, err)
		return err
	}
	chunkifyCmd.Tui.Job = job

	go chunkifyCmd.StartJobProgress(ctx, job.Id)

	// Wait for either job completion or context cancellation
	select {
	case <-chunkifyCmd.Tui.Progress.JobCompleted:
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	}

	if chunkifyCmd.Tui.Job != nil && (chunkifyCmd.Tui.Job.Status == chunkify.JobStatusFailed || chunkifyCmd.Tui.Job.Status == chunkify.JobStatusCancelled) {
		err := fmt.Errorf("job failed with status: %s: %s", chunkifyCmd.Tui.Job.Status, chunkifyCmd.Tui.Job.Error.Message)
		setError(&chunkifyCmd, err)
		return err
	}

	// Check if context was cancelled before getting files
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	default:
	}

	files, err := chunkifyCmd.GetFiles(job.Id)
	if err != nil {
		setError(&chunkifyCmd, err)
		return fmt.Errorf("error getting files: %s", err)
	}
	slog.Info("Files", "files", files)
	chunkifyCmd.Tui.Progress.Files <- files

	if chunkifyCmd.Output != "" {
		chunkifyCmd.Tui.Progress.Status <- Downloading

		slog.Info("Downloading files", "files", files)
		downloadedFiles := []string{}
		jobId := ""

		for _, file := range files {
			if jobId == "" {
				jobId = file.JobId
			}

			// Check if context was cancelled before each download
			select {
			case <-ctx.Done():
				return fmt.Errorf("download cancelled")
			default:
			}
			filepath := filename(file, chunkifyCmd.Output)

			if err := DownloadFile(ctx, file, filepath, chunkifyCmd.Tui.Progress.DownloadProgress); err == nil {
				chunkifyCmd.Tui.Progress.DownloadedFiles <- file
				downloadedFiles = append(downloadedFiles, filepath)
			}

		}

		// If format is jpg
		// rename all vtt cues to match the filename set in --output flag
		if chunkifyCmd.Format == string(chunkify.FormatJpg) {
			if err := postProcessVtt(downloadedFiles, jobId); err != nil {
				return fmt.Errorf("post process vtt: %w", err)
			}
		}
	}
	chunkifyCmd.Tui.Progress.Status <- Completed

	// Give the TUI time to display the completion message
	time.Sleep(1 * time.Second)

	return nil
}

func setError(c *ChunkifyCommand, err error) {
	c.Tui.Progress.Status <- Failed
	c.Tui.Progress.Error <- err
	// give time to display the error message before tea quit
	time.Sleep(1 * time.Second)
}

func (c *ChunkifyCommand) InitJobFormatParams() {
	c.JobFormatParams = chunkify.JobCreateFormatParams{}

	videoCommon := &chunkify.Video{
		Width:        width,
		Height:       height,
		Framerate:    framerate,
		Gop:          gop,
		Channels:     channels,
		Maxrate:      maxrate,
		Bufsize:      bufsize,
		DisableAudio: disableAudio,
		DisableVideo: disableVideo,
		Duration:     duration,
		Seek:         seek,
		PixFmt:       pixfmt,
	}

	switch c.Format {
	case string(chunkify.FormatMp4H264):
		h264Params := &chunkify.H264{
			Video:      videoCommon,
			Crf:        crf,
			Preset:     preset,
			Profilev:   profilev,
			Level:      level,
			X264KeyInt: x264KeyInt,
		}
		c.JobFormatParams.Mp4H264 = h264Params
	case string(chunkify.FormatMp4H265):
		h265Params := &chunkify.H265{
			Video:      videoCommon,
			Crf:        crf,
			Preset:     preset,
			Profilev:   profilev,
			Level:      level,
			X265KeyInt: x265KeyInt,
		}
		c.JobFormatParams.Mp4H265 = h265Params
	case string(chunkify.FormatMp4Av1):
		av1Params := &chunkify.Av1{
			Video:    videoCommon,
			Crf:      crf,
			Preset:   preset,
			Profilev: profilev,
			Level:    level,
		}
		c.JobFormatParams.Mp4Av1 = av1Params
	case string(chunkify.FormatJpg):
		jpgParams := &chunkify.Jpg{
			Image: &chunkify.Image{
				Width:    width,
				Height:   height,
				Interval: *interval,
				Sprite:   sprite,
			},
		}
		c.JobFormatParams.Jpg = jpgParams
	}
}

func (c *ChunkifyCommand) CreateSource() (*chunkify.Source, error) {
	// check input if it's a valid file or URL
	if strings.HasPrefix(c.Input, "https://") || strings.HasPrefix(c.Input, "http://") {
		// create source directly from URL
		source, err := c.CreateSourceFromUrl()
		if err != nil {
			return nil, err
		}

		return source, nil
	}

	// it's a path file, check if it's a valid file
	if _, err := os.Stat(c.Input); err != nil {
		return nil, fmt.Errorf("file not found: %s", c.Input)
	}

	// fmt.Println("Creating source directly from file")
	source, err := c.CreateSourceFromFile()
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (c *ChunkifyCommand) CreateSourceFromUrl() (*chunkify.Source, error) {
	c.Tui.Progress.Status <- UploadingFromUrl
	source, err := c.Config.Client.SourceCreate(chunkify.SourceCreateParams{
		Url: c.Input,
		Metadata: chunkify.SourceCreateParamsMetadata{
			"chunkify_execution_id": c.Id,
		},
	})
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (c *ChunkifyCommand) CreateSourceFromFile() (*chunkify.Source, error) {
	c.Tui.Progress.Status <- UploadingFromFile
	upload, err := c.Config.Client.UploadCreate(chunkify.UploadCreateParams{
		Metadata: chunkify.UploadCreateParamsMetadata{
			"chunkify_execution_id": c.Id,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating upload: %s", err)
	}
	file, err := os.Open(c.Input)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	if err := c.Config.Client.UploadBlobWithProgressAndContext(c.Tui.Ctx, file, upload, c.Tui.Progress.UploadProgress); err != nil {
		return nil, fmt.Errorf("error uploading blob: %s", err)
	}

	found := false

	retry := 0
	maxRetries := 30
	for !found && retry < maxRetries {
		results, err := c.Config.Client.SourceList(chunkify.SourceListParams{
			Metadata: map[string]string{
				"chunkify_execution_id": c.Id,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error listing sources: %s", err)
		}
		for _, source := range results.Items {
			// fmt.Printf("metadata: %#+v\n", source.Metadata)
			if v, ok := source.Metadata.(map[string]any); ok && v["chunkify_execution_id"] == c.Id {
				found = true
				return &source, nil
			}
		}
		time.Sleep(1 * time.Second)
		retry++
	}

	return nil, fmt.Errorf("source not found")
}

func (c *ChunkifyCommand) CreateJob(source *chunkify.Source) (*chunkify.Job, error) {
	c.Tui.Progress.Status <- Transcoding
	t := &chunkify.JobCreateTranscoderParams{}

	if c.Transcoders != nil && *c.Transcoders > 0 {
		t.Quantity = *c.Transcoders
		t.Type = fmt.Sprintf("%dvCPU", *c.TranscoderVcpu)
	}

	job, err := c.Config.Client.JobCreate(chunkify.JobCreateParams{
		SourceId:   source.Id,
		Format:     c.JobFormatParams,
		Transcoder: t,
		Metadata: chunkify.JobCreateParamsMetadata{
			"chunkify_execution_id": c.Id,
		},
	})

	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *ChunkifyCommand) GetJobProgress(jobId string) (chunkify.Job, error) {
	job, err := c.Config.Client.Job(jobId)
	if err != nil {
		return chunkify.Job{}, fmt.Errorf("error getting job: %s", err)
	}
	return job, nil
}

func (c *ChunkifyCommand) GetJobTranscoders(jobId string) ([]chunkify.TranscoderStatus, error) {
	transcoders, err := c.Config.Client.JobListTranscoders(jobId)
	if err != nil {
		return nil, fmt.Errorf("error getting job: %s", err)
	}
	return transcoders, nil
}

func (c *ChunkifyCommand) StartJobProgress(ctx context.Context, jobId string) {
	ticker := time.NewTicker(ProgressUpdateInterval)
	defer ticker.Stop()

	for {

		select {
		case <-ctx.Done():
			c.Tui.Progress.JobCompleted <- true
			return
		case <-ticker.C:
			job, err := c.GetJobProgress(jobId)
			if err != nil {
				return
			}
			c.Tui.Job = &job

			c.Tui.Progress.JobProgress <- job
			if job.Status == chunkify.JobStatusCompleted || job.Status == chunkify.JobStatusFailed || job.Status == chunkify.JobStatusCancelled {
				c.Tui.Progress.JobCompleted <- true
				break
			}

			transcoders, err := c.GetJobTranscoders(job.Id)
			if err != nil {
				return
			}
			c.Tui.Progress.JobTranscoders <- transcoders
		}
	}
}

func (c *ChunkifyCommand) GetFiles(jobId string) ([]chunkify.File, error) {
	files, err := c.Config.Client.JobListFiles(jobId)
	if err != nil {
		return nil, fmt.Errorf("error getting files: %s", err)
	}
	return files, nil
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
