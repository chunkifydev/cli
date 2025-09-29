package chunkify

import (
	"fmt"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
)

type Progress struct {
	JobProgress      chan chunkify.Job
	JobTranscoders   chan []chunkify.TranscoderStatus
	JobCompleted     chan bool
	UploadProgress   chan chunkify.UploadProgressChannel
	DownloadProgress chan DownloadProgress
	Error            chan error
}

type DownloadProgress struct {
	Progress     float64
	TotalBytes   int64
	BytesWritten int64
	Eta          time.Duration
	Speed        float64
}

func NewProgress() *Progress {
	return &Progress{
		JobProgress:      make(chan chunkify.Job, 100),
		JobTranscoders:   make(chan []chunkify.TranscoderStatus, 100),
		JobCompleted:     make(chan bool),
		UploadProgress:   make(chan chunkify.UploadProgressChannel),
		DownloadProgress: make(chan DownloadProgress, 100),
		Error:            make(chan error),
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
				for _, transcoder := range transcoders {
					fmt.Printf("[%d] %f%%\n", transcoder.ChunkNumber, transcoder.Progress)
				}

			}
		case uploadProgress, ok := <-p.UploadProgress:
			if ok {
				fmt.Printf("Upload progress: %f%%\n", uploadProgress.Progress)
			}
		case downloadProgress, ok := <-p.DownloadProgress:
			if ok {
				fmt.Printf("Download progress: %f%%\n", downloadProgress.Progress)
			}
		case err := <-p.Error:
			fmt.Printf("Error: %s\n", err)
			p.JobCompleted <- true
			return
		}
	}
}
