package chunkify

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	chunkify "github.com/chunkifydev/chunkify-go"
)

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

type TUI struct {
	Status   int
	Progress *Progress
}

func (t TUI) Init() tea.Cmd {
	return tea.Batch(tickCmd())
}

type Progress struct {
	Status           chan int
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

// tickMsg represents a tick event for periodic updates
type tickMsg time.Time

// tickCmd creates a tea.Cmd that sends tick events periodically
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (t TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return t, tea.Quit
		}
	case tickMsg:
		return t, tickCmd()
	}

	return t, nil
}

func (t TUI) View() string {
	done := false
	for {
		select {
		case job, ok := <-t.Progress.JobProgress:
			if ok {
				fmt.Printf("Job progress: %s (%f%%)\n", job.Status, job.Progress)
			}
		case transcoders, ok := <-t.Progress.JobTranscoders:
			if ok {
				for _, transcoder := range transcoders {
					fmt.Printf("[%d] %f%%\n", transcoder.ChunkNumber, transcoder.Progress)
				}

			}
		case uploadProgress, ok := <-t.Progress.UploadProgress:
			if ok {
				fmt.Printf("Upload progress: %f%%: %#+v\n", uploadProgress.Progress, uploadProgress.Eta.Seconds())
			}
		case downloadProgress, ok := <-t.Progress.DownloadProgress:
			if ok {
				fmt.Printf("Download progress: %f%%\n", downloadProgress.Progress)
			}
		case status, ok := <-t.Progress.Status:
			if ok {
				fmt.Printf("Status: %d\n", status)
				if status == Completed || status == Failed || status == Cancelled {
					done = true
				}
			}
		case err := <-t.Progress.Error:
			fmt.Printf("Error: %s\n", err)
			t.Progress.JobCompleted <- true
		}
		if done {
			fmt.Println("Done")
			return ""
		}
	}
	return ""
}
