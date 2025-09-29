package chunkify

import (
	"fmt"
	"time"

	chunkify "github.com/chunkifydev/chunkify-go"
)

type Progress struct {
	Status           chan int
	JobProgress      chan chunkify.Job
	JobTranscoders   chan []chunkify.TranscoderStatus
	JobCompleted     chan bool
	UploadProgress   chan chunkify.UploadProgressChannel
	DownloadProgress chan DownloadProgress
	Error            chan error
}

const (
	Status = iota
	Starting
	UploadingFromUrl
	UploadingFromFile
	Transcoding
	Downloading
	Completed
	Failed
	Cancelled
)

type DownloadProgress struct {
	Progress     float64
	TotalBytes   int64
	WrittenBytes int64
	Eta          time.Duration
	Speed        float64 // bytes/sec
}

func NewProgress() *Progress {
	return &Progress{
		Status:           make(chan int, 1),
		JobProgress:      make(chan chunkify.Job, 100),
		JobTranscoders:   make(chan []chunkify.TranscoderStatus, 100),
		JobCompleted:     make(chan bool),
		UploadProgress:   make(chan chunkify.UploadProgressChannel),
		DownloadProgress: make(chan DownloadProgress, 100),
		Error:            make(chan error),
	}
}

func (p *Progress) Render() {
	done := false
	for {
		select {
		case job, ok := <-p.JobProgress:
			if ok {
				fmt.Printf("Job progress: %s (%f%%)\n", job.Status, job.Progress)
			}
		case transcoders, ok := <-p.JobTranscoders:
			if ok {
				for _, transcoder := range transcoders {
					fmt.Printf("[%d] %f%%\n", transcoder.ChunkNumber, transcoder.Progress)
				}

			}
		case uploadProgress, ok := <-p.UploadProgress:
			if ok {
				fmt.Printf("Upload progress: %f%%: %#+v\n", uploadProgress.Progress, uploadProgress.Eta.Seconds())
			}
		case downloadProgress, ok := <-p.DownloadProgress:
			if ok {
				fmt.Printf("Download progress: %f%%\n", downloadProgress.Progress)
			}
		case status, ok := <-p.Status:
			if ok {
				fmt.Printf("Status: %d\n", status)
				if status == Completed || status == Failed || status == Cancelled {
					done = true
				}
			}
		case err := <-p.Error:
			fmt.Printf("Error: %s\n", err)
			p.JobCompleted <- true
		}
		if done {
			fmt.Println("Done")
			return
		}
	}
}
