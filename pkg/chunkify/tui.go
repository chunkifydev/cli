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
	Status           int
	Progress         *Progress
	JobProgress      chunkify.Job
	JobTranscoders   []chunkify.TranscoderStatus
	UploadProgress   chunkify.UploadProgressChannel
	DownloadProgress DownloadProgress
	Error            error
	Done             bool
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

// NewTUI creates a new TUI instance with the given progress tracker
func NewTUI(progress *Progress) TUI {
	return TUI{
		Status:   Status,
		Progress: progress,
		Done:     false,
	}
}

// tickMsg represents a tick event for periodic updates
type tickMsg time.Time

// tickCmd creates a tea.Cmd that sends tick events periodically
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*1, func(t time.Time) tea.Msg {
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
		// Check for updates from channels (non-blocking)
		t = t.checkChannels()
		return t, tickCmd()
	}

	return t, nil
}

// checkChannels performs non-blocking reads from all channels
func (t TUI) checkChannels() TUI {
	// Check job progress
	select {
	case job, ok := <-t.Progress.JobProgress:
		if ok {
			t.JobProgress = job
		}
	default:
	}

	// Check transcoder status
	select {
	case transcoders, ok := <-t.Progress.JobTranscoders:
		if ok {
			t.JobTranscoders = transcoders
		}
	default:
	}

	// Check upload progress
	select {
	case uploadProgress, ok := <-t.Progress.UploadProgress:
		if ok {
			t.UploadProgress = uploadProgress
		}
	default:
	}

	// Check download progress
	select {
	case downloadProgress, ok := <-t.Progress.DownloadProgress:
		if ok {
			t.DownloadProgress = downloadProgress
		}
	default:
	}

	// Check status
	select {
	case status, ok := <-t.Progress.Status:
		if ok {
			t.Status = status
			if status == Completed || status == Failed || status == Cancelled {
				t.Done = true
			}
		}
	default:
	}

	// Check for errors
	select {
	case err := <-t.Progress.Error:
		t.Error = err
		t.Progress.JobCompleted <- true
	default:
	}

	return t
}

func (t TUI) View() string {
	var view string

	// Display current status
	view += fmt.Sprintf("Status: %s\n", t.getStatusString())

	// Display error if any
	if t.Error != nil {
		view += fmt.Sprintf("Error: %s\n", t.Error)
	}

	// Display upload progress
	view += fmt.Sprintf("Upload: %.1f%% (ETA: %.0fs) %#+v\n",
		t.UploadProgress.Progress, t.UploadProgress.Eta.Seconds(), t.UploadProgress)
	// if t.Status == UploadingFromUrl || t.Status == UploadingFromFile {

	// }

	// Display job progress
	if t.Status >= Transcoding {
		view += fmt.Sprintf("Job: %s (%.1f%%)\n", t.JobProgress.Status, t.JobProgress.Progress)
		view += "Transcoders:\n"
		for _, transcoder := range t.JobTranscoders {
			view += fmt.Sprintf("  [%d] %.1f%%\n", transcoder.ChunkNumber, transcoder.Progress)
		}
	}

	// Display download progress
	if t.Status >= Downloading {
		view += fmt.Sprintf("Download: %.1f%% (%.1f MB/s, ETA: %s)\n",
			t.DownloadProgress.Progress,
			t.DownloadProgress.Speed/(1024*1024),
			t.DownloadProgress.Eta.Round(time.Second))
	}

	// Display completion message
	if t.Done {
		view += "\n✅ Process completed!\n"
	}

	view += "\nPress 'q' or Ctrl+C to quit\n"

	return view
}

// getStatusString returns a human-readable status string
func (t TUI) getStatusString() string {
	switch t.Status {
	case Status:
		return "Initializing"
	case Starting:
		return "Starting"
	case UploadingFromUrl:
		return "Uploading from URL"
	case UploadingFromFile:
		return "Uploading from file"
	case Transcoding:
		return "Transcoding"
	case Downloading:
		return "Downloading"
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	case Cancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}
