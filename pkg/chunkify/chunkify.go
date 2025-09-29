package chunkify

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
	Source          chunkify.Source
	Job             chunkify.Job
	Format          string
	JobFormatParams chunkify.JobCreateFormatParams
	Transcoders     *int64
	TranscoderVcpu  *int64
	Progress        *Progress
}

var chunkifyCmd = ChunkifyCommand{}

func Execute(cfg *config.Config) error {
	chunkifyCmd.Progress.Status <- Starting
	chunkifyCmd.Config = cfg
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go chunkifyCmd.Progress.Render()

	chunkifyCmd.Id = uuid.New().String()

	source, err := chunkifyCmd.CreateSource()
	if err != nil {
		chunkifyCmd.Progress.Status <- Failed
		chunkifyCmd.Progress.Error <- err
		return fmt.Errorf("error creating source: %s", err)
	}
	fmt.Println("Source created:", source)

	chunkifyCmd.InitJobFormatParams()
	fmt.Printf("format: %#+v\n", chunkifyCmd.JobFormatParams)

	job, err := chunkifyCmd.CreateJob()
	if err != nil {
		chunkifyCmd.Progress.Status <- Failed
		chunkifyCmd.Progress.Error <- err
		return fmt.Errorf("error creating job: %s", err)
	}
	chunkifyCmd.Job = job
	fmt.Println("Job created:", job)

	go chunkifyCmd.StartJobProgress()

	<-chunkifyCmd.Progress.JobCompleted
	fmt.Println("Job completed with status:", chunkifyCmd.Job.Status)

	if chunkifyCmd.Job.Status == chunkify.JobStatusFailed || chunkifyCmd.Job.Status == chunkify.JobStatusCancelled {
		err := fmt.Errorf("job failed with status: %s: %s", chunkifyCmd.Job.Status, chunkifyCmd.Job.Error.Message)
		chunkifyCmd.Progress.Status <- Failed
		chunkifyCmd.Progress.Error <- err
		return err
	}

	files, err := chunkifyCmd.GetFiles()
	if err != nil {
		chunkifyCmd.Progress.Status <- Failed
		chunkifyCmd.Progress.Error <- err
		return fmt.Errorf("error getting files: %s", err)
	}
	fmt.Printf("Files: %#+v\n", files)

	if chunkifyCmd.Output != "" {
		chunkifyCmd.Progress.Status <- Downloading
		for _, file := range files {
			fmt.Printf("Downloading file: %s\n", file.Url)
			DownloadFile(ctx, file.Url, chunkifyCmd.Output, chunkifyCmd.Progress.DownloadProgress)
		}
	}
	chunkifyCmd.Progress.Status <- Completed

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
	c.Progress.Status <- UploadingFromUrl
	source, err := c.Config.Client.SourceCreate(chunkify.SourceCreateParams{
		Url: c.Input,
		Metadata: chunkify.SourceCreateParamsMetadata{
			"chunkify_execution_id": c.Id,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating source: %s", err)
	}
	return &source, nil
}

func (c *ChunkifyCommand) CreateSourceFromFile() (*chunkify.Source, error) {
	c.Progress.Status <- UploadingFromFile
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

	if err := c.Config.Client.UploadBlobWithProgress(file, upload, c.Progress.UploadProgress); err != nil {
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
			fmt.Printf("metadata: %#+v\n", source.Metadata)
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

func (c *ChunkifyCommand) CreateJob() (chunkify.Job, error) {
	c.Progress.Status <- Transcoding
	t := &chunkify.JobCreateTranscoderParams{}

	if c.Transcoders != nil && *c.Transcoders > 0 {
		t.Quantity = *c.Transcoders
		t.Type = fmt.Sprintf("%dvCPU", *c.TranscoderVcpu)
	}

	job, err := c.Config.Client.JobCreate(chunkify.JobCreateParams{
		SourceId:   c.Source.Id,
		Format:     c.JobFormatParams,
		Transcoder: t,
		Metadata: chunkify.JobCreateParamsMetadata{
			"chunkify_execution_id": c.Id,
		},
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
	ticker := time.NewTicker(ProgressUpdateInterval)
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

// func DownloadFile(file chunkify.File, output string) error {
// 	resp, err := http.Get(file.Url)
// 	if err != nil {
// 		return fmt.Errorf("error downloading file: %s", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return fmt.Errorf("error reading file: %s", err)
// 	}

// 	err = os.WriteFile(output, body, 0644)
// 	if err != nil {
// 		return fmt.Errorf("error writing file: %s", err)
// 	}

// 	fmt.Printf("File downloaded to: %s\n", output)

// 	return nil
// }
