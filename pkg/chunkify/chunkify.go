package chunkify

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/flags"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type ChunkifyCommand struct {
	Id              string
	Config          *config.Config
	Input           string
	Output          string
	Source          chunkify.Source
	Job             chunkify.Job
	Format          string
	JobFormatParams chunkify.JobCreateFormatParams
	Transcoders     *int64
	TranscoderVcpu  *int64
	Progress        *Progress
}

// Common video flags
var (
	width        *int64
	height       *int64
	framerate    *float64
	gop          *int64
	channels     *int64
	maxrate      *int64
	bufsize      *int64
	pixfmt       *string
	disableAudio *bool
	disableVideo *bool
	duration     *int64
	seek         *int64
)

// h264, h265 and av1 flags
var (
	crf        *int64
	preset     *string
	profilev   *string
	level      *int64
	x264KeyInt *int64
	x265KeyInt *int64
)

var chunkifyCmd = ChunkifyCommand{}

// BindFlags attaches root-level flags used by the root command
func BindFlags(rcmd *cobra.Command, cfg *config.Config) {
	chunkifyCmd = ChunkifyCommand{Config: cfg, Progress: NewProgress()}

	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Input, "input", "", "Video file or URL to process")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Output, "output", "", "Output file or directory")
	flags.StringVar(rcmd.Flags(), &chunkifyCmd.Format, "format", "mp4_h264", "chunkify format: mp4_h264, mp4_h265 or mp4_av1")
	flags.Int64VarPtr(rcmd.Flags(), &chunkifyCmd.Transcoders, "transcoders", 0, "chunkify transcoder quantity: Transcoders")
	flags.Int64VarPtr(rcmd.Flags(), &chunkifyCmd.TranscoderVcpu, "vcpu", 8, "chunkify transcoder vCPU: 4, 8 or 16")

	// format settings
	flags.Int64VarPtr(rcmd.Flags(), &width, "width", 0, "ffmpeg config: Width")
	flags.Int64VarPtr(rcmd.Flags(), &height, "height", 0, "ffmpeg config: Height")
	flags.Float64VarPtr(rcmd.Flags(), &framerate, "framerate", 0, "ffmpeg config: Framerate")
	flags.Int64VarPtr(rcmd.Flags(), &gop, "gop", 0, "ffmpeg config: Gop")
	flags.Int64VarPtr(rcmd.Flags(), &channels, "channels", 0, "ffmpeg config: Channels")
	flags.Int64VarPtr(rcmd.Flags(), &maxrate, "maxrate", 0, "ffmpeg config: Maxrate")
	flags.Int64VarPtr(rcmd.Flags(), &bufsize, "bufsize", 0, "ffmpeg config: Bufsize")
	flags.StringVarPtr(rcmd.Flags(), &pixfmt, "pixfmt", "", "ffmpeg config: PixFmt")
	flags.BoolVarPtr(rcmd.Flags(), &disableAudio, "an", false, "ffmpeg config: DisableAudio")
	flags.BoolVarPtr(rcmd.Flags(), &disableVideo, "vn", false, "ffmpeg config: DisableVideo")
	flags.Int64VarPtr(rcmd.Flags(), &duration, "duration", 0, "ffmpeg config: Duration")
	flags.Int64VarPtr(rcmd.Flags(), &seek, "seek", 0, "ffmpeg config: Seek")

	flags.Int64VarPtr(rcmd.Flags(), &crf, "crf", 0, "ffmpeg config: Crf")
	flags.StringVarPtr(rcmd.Flags(), &preset, "preset", "", "ffmpeg config: Preset")
	flags.StringVarPtr(rcmd.Flags(), &profilev, "profilev", "", "ffmpeg config: Profilev")
	flags.Int64VarPtr(rcmd.Flags(), &level, "level", 0, "ffmpeg config: Level")
	flags.Int64VarPtr(rcmd.Flags(), &x264KeyInt, "x264keyint", 0, "ffmpeg config: X264KeyInt")
	flags.Int64VarPtr(rcmd.Flags(), &x265KeyInt, "x265keyint", 0, "ffmpeg config: X265KeyInt")
}

type Progress struct {
	JobProgress    chan chunkify.Job
	JobTranscoders chan []chunkify.TranscoderStatus
	JobCompleted   chan bool
	UploadProgress chan chunkify.UploadProgressChannel
	Error          chan error
}

func NewProgress() *Progress {
	return &Progress{
		JobProgress:    make(chan chunkify.Job, 100),
		JobTranscoders: make(chan []chunkify.TranscoderStatus, 100),
		JobCompleted:   make(chan bool),
		UploadProgress: make(chan chunkify.UploadProgressChannel),
		Error:          make(chan error),
	}
}

func (p *Progress) Render() {
	for {
		select {
		case job, ok := <-p.JobProgress:
			if ok {
				fmt.Printf("Job progress: %s (%f%%)\n", job.Status, job.Progress)
			}
			if job.Status == chunkify.JobStatusCompleted || job.Status == chunkify.JobStatusFailed || job.Status == chunkify.JobStatusCancelled {
				return
			}
		case transcoders, ok := <-p.JobTranscoders:
			if ok {
				fmt.Printf("Job transcoders: %#+v\n", transcoders)
			}
		case uploadProgress, ok := <-p.UploadProgress:
			if ok {
				fmt.Printf("Upload progress: %#+v\n", uploadProgress)
			}
		case err := <-p.Error:
			fmt.Printf("Error: %s\n", err)
			p.JobCompleted <- true
			return
		}
	}
}

func Execute() error {
	go chunkifyCmd.Progress.Render()

	chunkifyCmd.Id = uuid.New().String()

	source, err := chunkifyCmd.CreateSource()
	if err != nil {
		return fmt.Errorf("error creating source: %s", err)
	}
	fmt.Println("Source created:", source)

	chunkifyCmd.InitJobFormatParams()
	fmt.Printf("format: %#+v\n", chunkifyCmd.JobFormatParams)

	job, err := chunkifyCmd.CreateJob()
	if err != nil {
		return fmt.Errorf("error creating job: %s", err)
	}
	chunkifyCmd.Job = job
	fmt.Println("Job created:", job)

	go chunkifyCmd.StartJobProgress()

	<-chunkifyCmd.Progress.JobCompleted
	fmt.Println("Job completed")

	files, err := chunkifyCmd.GetFiles()
	if err != nil {
		return fmt.Errorf("error getting files: %s", err)
	}
	fmt.Printf("Files: %#+v\n", files)

	if chunkifyCmd.Output != "" {
		for _, file := range files {
			fmt.Printf("Downloading file: %s\n", file.Url)
			DownloadFile(file, chunkifyCmd.Output)
		}
	}

	return nil
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
	case "mp4_h264":
		h264Params := &chunkify.H264{
			Video:      videoCommon,
			Crf:        crf,
			Preset:     preset,
			Profilev:   profilev,
			Level:      level,
			X264KeyInt: x264KeyInt,
		}
		c.JobFormatParams.Mp4H264 = h264Params
	case "mp4_h265":
		h265Params := &chunkify.H265{
			Video:      videoCommon,
			Crf:        crf,
			Preset:     preset,
			Profilev:   profilev,
			Level:      level,
			X265KeyInt: x265KeyInt,
		}
		c.JobFormatParams.Mp4H265 = h265Params
	case "mp4_av1":
		av1Params := &chunkify.Av1{
			Video:    videoCommon,
			Crf:      crf,
			Preset:   preset,
			Profilev: profilev,
			Level:    level,
		}
		c.JobFormatParams.Mp4Av1 = av1Params
	}
}

func (c *ChunkifyCommand) CreateSource() (*chunkify.Source, error) {
	// check input if it's a valid file or URL
	if strings.HasPrefix(c.Input, "https://") || strings.HasPrefix(c.Input, "http://") {
		// create source directly from URL
		source, err := c.CreateSourceFromUrl()
		if err != nil {
			return nil, fmt.Errorf("error creating source: %s", err)
		}
		c.Source = *source
		return source, nil
	}

	// it's a path file, check if it's a valid file
	if _, err := os.Stat(c.Input); err != nil {
		return nil, fmt.Errorf("file not found: %s", c.Input)
	}

	fmt.Println("Creating source directly from file")
	source, err := c.CreateSourceFromFile()
	if err != nil {
		return nil, fmt.Errorf("error creating source: %s", err)
	}
	c.Source = *source
	return source, nil
}

func (c *ChunkifyCommand) CreateSourceFromUrl() (*chunkify.Source, error) {
	source, err := c.Config.Client.SourceCreate(chunkify.SourceCreateParams{
		Url: c.Input,
		Metadata: chunkify.SourceCreateParamsMetadata{
			"cli_execution_id": c.Id,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating source: %s", err)
	}
	return &source, nil
}

func (c *ChunkifyCommand) CreateSourceFromFile() (*chunkify.Source, error) {
	upload, err := c.Config.Client.UploadCreate(chunkify.UploadCreateParams{
		Metadata: chunkify.UploadCreateParamsMetadata{
			"cli_execution_id": c.Id,
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

	if err := c.Config.Client.UploadBlobWithProgress(file, upload, c.Progress.UploadProgress); err != nil {
		return nil, fmt.Errorf("error uploading blob: %s", err)
	}

	found := false

	retry := 0
	maxRetries := 30
	for !found && retry < maxRetries {
		results, err := c.Config.Client.SourceList(chunkify.SourceListParams{
			Metadata: map[string]string{
				"cli_execution_id": c.Id,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error listing sources: %s", err)
		}
		for _, source := range results.Items {
			fmt.Printf("metadata: %#+v\n", source.Metadata)
			if v, ok := source.Metadata.(map[string]any); ok && v["cli_execution_id"] == c.Id {
				found = true
				return &source, nil
			}
		}
		time.Sleep(1 * time.Second)
		retry++
	}

	return nil, fmt.Errorf("source not found")
}

func (c *ChunkifyCommand) CreateJob() (chunkify.Job, error) {
	t := &chunkify.JobCreateTranscoderParams{}

	if c.Transcoders != nil && *c.Transcoders > 0 {
		t.Quantity = *c.Transcoders
		t.Type = fmt.Sprintf("%dvCPU", *c.TranscoderVcpu)
	}

	job, err := c.Config.Client.JobCreate(chunkify.JobCreateParams{
		SourceId:   c.Source.Id,
		Format:     c.JobFormatParams,
		Transcoder: t,
	})

	if err != nil {
		return chunkify.Job{}, fmt.Errorf("error creating job: %s", err)
	}
	return job, nil
}

func (c *ChunkifyCommand) GetJobProgress() (chunkify.Job, error) {
	job, err := c.Config.Client.Job(c.Job.Id)
	if err != nil {
		return chunkify.Job{}, fmt.Errorf("error getting job: %s", err)
	}
	return job, nil
}

func (c *ChunkifyCommand) GetJobTranscoders() ([]chunkify.TranscoderStatus, error) {
	transcoders, err := c.Config.Client.JobListTranscoders(c.Job.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting job: %s", err)
	}
	return transcoders, nil
}

func (c *ChunkifyCommand) StartJobProgress() {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		job, err := c.GetJobProgress()
		if err != nil {
			return
		}
		c.Job = job

		c.Progress.JobProgress <- job
		if job.Status == chunkify.JobStatusCompleted || job.Status == chunkify.JobStatusFailed || job.Status == chunkify.JobStatusCancelled {
			c.Progress.JobCompleted <- true
			break
		}

		transcoders, err := c.GetJobTranscoders()
		if err != nil {
			return
		}
		c.Progress.JobTranscoders <- transcoders
	}
}

func (c *ChunkifyCommand) GetFiles() ([]chunkify.File, error) {
	files, err := c.Config.Client.JobListFiles(c.Job.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting files: %s", err)
	}
	return files, nil
}

func DownloadFile(file chunkify.File, output string) error {
	resp, err := http.Get(file.Url)
	if err != nil {
		return fmt.Errorf("error downloading file: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading file: %s", err)
	}

	err = os.WriteFile(output, body, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %s", err)
	}

	fmt.Printf("File downloaded to: %s\n", output)

	return nil
}
