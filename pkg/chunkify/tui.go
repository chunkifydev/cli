package chunkify

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"

	_ "embed"
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

//go:embed chunkify.txt
var chunkifyBanner string

var (
	checkMark     = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	currentStep   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render
	completedStep = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render
	info          = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#555555"}).Render
)

type TUI struct {
	Command          *ChunkifyCommand
	Status           int
	Progress         *Progress
	Job              *chunkify.Job
	Source           *chunkify.Source
	Transcoders      []chunkify.TranscoderStatus
	UploadProgress   chunkify.UploadProgressChannel
	DownloadProgress DownloadProgress
	Error            error
	Done             bool
	Ctx              context.Context
	CancelFunc       context.CancelFunc

	Spinner         spinner.Model
	OverallProgress progress.Model
}

func (t TUI) Init() tea.Cmd {
	return tea.Batch(tickCmd(), t.Spinner.Tick)
}

type Progress struct {
	Status           chan int
	JobProgress      chan chunkify.Job
	JobTranscoders   chan []chunkify.TranscoderStatus
	JobCompleted     chan bool
	UploadProgress   chan chunkify.UploadProgressChannel
	DownloadProgress chan DownloadProgress
	Source           chan *chunkify.Source
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
		JobCompleted:     make(chan bool, 1),
		UploadProgress:   make(chan chunkify.UploadProgressChannel, 100),
		DownloadProgress: make(chan DownloadProgress, 100),
		Source:           make(chan *chunkify.Source, 1),
		Error:            make(chan error),
	}
}

// NewTUI creates a new TUI instance with the given progress tracker
func NewTUI() TUI {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return TUI{
		Status:          Starting,
		Progress:        NewProgress(),
		Done:            false,
		Spinner:         s,
		OverallProgress: progress.New(progress.WithSolidFill("#16a249")),
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
			fmt.Println("Quitting...")
			t.CancelFunc()
			// Send JobCompleted to unblock the main goroutine
			select {
			case t.Progress.JobCompleted <- true:
			default:
			}
			return t, tea.Quit
		}
	case tickMsg:
		// Check for updates from channels (non-blocking)
		t = t.checkChannels()
		return t, tickCmd()
	default:
		var scmd tea.Cmd
		t.Spinner, scmd = t.Spinner.Update(msg)
		// if t.Job != nil {
		// 	//pcmd := t.OverallProgress.SetPercent(t.Job.Progress)
		// 	return t, tea.Batch(scmd, pcmd)
		// }
		return t, tea.Batch(scmd)
	}

	return t, nil
}

// checkChannels performs non-blocking reads from all channels
func (t TUI) checkChannels() TUI {
	// Check job progress
	select {
	case job, ok := <-t.Progress.JobProgress:
		if ok {
			t.Job = &job
		}
	default:
	}

	// Check transcoder status
	select {
	case transcoders, ok := <-t.Progress.JobTranscoders:
		if ok {
			t.Transcoders = transcoders
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

	// Check for source updates
	select {
	case source, ok := <-t.Progress.Source:
		if ok {
			t.Source = source
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

	select {
	case <-t.Ctx.Done():
		t.Status = Cancelled
		t.Done = true
	default:
	}

	return t
}

func (t TUI) View() string {
	var view string
	indent := "  " // 2 spaces for consistent indentation
	overallProgress := 0.0

	view += fmt.Sprintf("\n%s\n\n", chunkifyBanner)

	// Display error if any
	if t.Error != nil {
		view += fmt.Sprintf("Error: %s\n", t.Error)
	}

	if t.Status == UploadingFromFile {
		overallProgress = t.UploadProgress.Progress

		// Display upload progress
		view += currentStep(fmt.Sprintf("%s%sUploading %s %.1f%% (%.1f MB/s, ETA: %s)\n",
			indent,
			t.Spinner.View(),
			t.Command.Input,
			t.UploadProgress.Progress,
			t.UploadProgress.Speed/(1024*1024),
			t.UploadProgress.Eta.Round(time.Second)))
	}
	if t.Status == UploadingFromUrl {
		// Display upload progress
		view += fmt.Sprintf("%s%s Uploaded from %s\n", indent, checkMark.String(), t.Command.Input)
	}

	if t.Source != nil {
		overallProgress = 100.0
		view += fmt.Sprintf("%s%s Uploaded to Chunkify\n", indent, checkMark.String())
		//view += fmt.Sprintf("Source: %s %dx%d\n", t.Source.VideoCodec, t.Source.Width, t.Source.Height)
	}

	// Display job progress
	if t.Status >= Transcoding && t.Job != nil {
		overallProgress = t.Job.Progress

		if t.Job.Status == chunkify.JobStatusCompleted {
			view += fmt.Sprintf("%s%s Video transcoded\n", indent, checkMark.String())
		} else {
			if t.Job.Status == chunkify.JobStatusQueued {
				view += currentStep(fmt.Sprintf("%s%sTranscoding to %s - %s\n", indent, t.Spinner.View(), t.Command.Format, t.Job.Status))
			} else {
				totalFps := 0.0
				totalSpeed := 0.0
				view += currentStep(fmt.Sprintf("%s%sTranscoding to %s - %s (%.1f%%)\n", indent, t.Spinner.View(), t.Command.Format, t.Job.Status, t.Job.Progress))
				for _, transcoder := range t.Transcoders {
					view += fmt.Sprintf("%s  [%d] %.1f%%\n", indent, transcoder.ChunkNumber, transcoder.Progress)
					totalFps += transcoder.Fps
					totalSpeed += transcoder.Speed
				}
				view += info(fmt.Sprintf("\n%sTotal FPS: %.0f, Total Speed: %.1fx\n", indent, totalFps, totalSpeed))
			}
		}
	}

	// Display download progress
	if t.Status >= Downloading {
		overallProgress = t.DownloadProgress.Progress

		if t.DownloadProgress.Progress == 100 {
			view += fmt.Sprintf("%s%s %s written (%s)\n", indent, checkMark.String(), t.Command.Output, formatter.Size(t.DownloadProgress.WrittenBytes))
		} else {
			view += currentStep(fmt.Sprintf("%s%sDownloading video %.1f%% (%.1f MB/s, ETA: %s)\n",
				indent,
				t.Spinner.View(),
				t.DownloadProgress.Progress,
				t.DownloadProgress.Speed/(1024*1024),
				t.DownloadProgress.Eta.Round(time.Second)))
		}
	}

	if !t.Done {
		view += fmt.Sprintf("\n\n%s %s\n", indent, t.OverallProgress.ViewAs(overallProgress/100.0))
	}

	// Display completion message
	if t.Done {
		//view += fmt.Sprintf("\n%s Process completed!\n", checkMark.String())
	}

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
