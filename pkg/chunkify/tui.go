package chunkify

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/version"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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
	completedIconStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("▮")
	currentStepStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render
	errorStyle                  = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render
	statusStyle                 = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("42")).Padding(0, 1).Bold(true).Render
	statusFormatStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#E7AE59")).Padding(0, 1).Bold(true).Render
	completedStepStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render
	infoStyle                   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE")).Render
	pendingTranscoderChunkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#59636e")).Render

	indent = "  "
)

type TUI struct {
	Command          *ChunkifyCommand
	Status           int
	Progress         *Progress
	Job              *chunkify.Job
	Source           *chunkify.Source
	Files            []chunkify.File
	Transcoders      []chunkify.TranscoderStatus
	UploadProgress   chunkify.UploadProgressChannel
	DownloadProgress DownloadProgress
	Error            error
	Done             bool
	Ctx              context.Context
	CancelFunc       context.CancelFunc

	Spinner spinner.Model
}

func init() {
	chunkifyBanner = strings.Replace(chunkifyBanner, "{version}", version.Version, 1)
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
	File         chunkify.File
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
	s.Spinner = spinner.MiniDot
	return TUI{
		Status:   Starting,
		Progress: NewProgress(),
		Done:     false,
		Spinner:  s,
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
	statusInfo := ""
	view = fmt.Sprintf("\n%s\n\n", chunkifyBanner)

	var v string
	v, statusInfo = t.uploadView()
	view += v

	if t.Source != nil {
		v, statusInfo = t.sourceView()
		view += v
	}

	// Display job progress
	if t.Status >= Transcoding && t.Job != nil {
		v, statusInfo = t.transcodingView()
		view += v
	}

	// Display download progress
	if t.Job != nil && t.Job.Status == chunkify.JobStatusCompleted && t.Status >= Downloading {
		v, statusInfo = t.downloadView()
		view += v
	}

	if t.Error != nil {
		view += t.errorView()
		view += "\n"
		return view
	}

	if !t.Done {
		view += fmt.Sprintf("\n\n%s%s %s\n", indent, statusStyle(t.getStatusString()), infoStyle(statusInfo))
	}

	return view
}

func (t TUI) errorView() string {
	var view string
	if apiErr, ok := t.Error.(chunkify.ApiError); ok {
		view = fmt.Sprintf("%s%s", indent, errorStyle(apiErr.Message))
	} else {
		view = fmt.Sprintf("%s%s", indent, errorStyle(t.Error.Error()))
	}
	view += "\n"
	return view
}

func (t TUI) uploadView() (string, string) {
	view := ""
	statusInfo := ""

	if t.Status == UploadingFromFile {
		statusInfo = fmt.Sprintf("%.1f%% (%.1f MB/s, ETA: %s)", t.UploadProgress.Progress, t.UploadProgress.Speed/(1024*1024), t.UploadProgress.Eta.Round(time.Second))

		// Display upload progress
		view = currentStepStyle(fmt.Sprintf("%s%s Uploading %s",
			indent,
			t.Spinner.View(),
			t.Command.Input))
		view += "\n"
	}
	if t.Status == UploadingFromUrl {
		statusInfo = t.Command.Input
		view += fmt.Sprintf("%s%s Uploaded from URL\n", indent, completedIconStyle.String())
	}

	return view, statusInfo
}

func (t TUI) sourceView() (string, string) {
	view := fmt.Sprintf("%s%s %s uploaded\n", indent, completedIconStyle.String(), path.Base(t.Command.Input))
	view += infoStyle(fmt.Sprintf("%sDuration: %s Size: %s Video: %s, %dx%d, %s, %.2ffps", indent+indent, formatter.Duration(t.Source.Duration), formatter.Size(t.Source.Size), t.Source.VideoCodec, t.Source.Width, t.Source.Height, formatter.Bitrate(t.Source.VideoBitrate), t.Source.VideoFramerate))
	view += "\n\n"
	return view, t.Command.Format
}

func (t TUI) transcodingView() (string, string) {
	view := ""

	completedTranscoders := 0
	totalTranscoders := len(t.Transcoders)
	totalFps := 0.0
	totalSpeed := 0.0
	totalOutTime := int64(0)

	for _, transcoder := range t.Transcoders {
		if transcoder.Status == chunkify.TranscoderStatusCompleted {
			completedTranscoders++
		}
	}

	if t.Job.Status == chunkify.JobStatusCompleted {
		view += fmt.Sprintf("%s%s Video transcoded\n", indent, completedIconStyle.String())
	} else {
		view += currentStepStyle(fmt.Sprintf("%s%s Transcoding video (%d x %s)", indent, t.Spinner.View(), totalTranscoders, t.Job.Transcoder.Type))
		view += "\n"
	}

	// start of progress bar
	view += indent + indent
	counter := 0

	maxCharsPerLine := 60
	progressBarWidth := 10
	transcodersPerLine := maxCharsPerLine / progressBarWidth

	if totalTranscoders > 0 && totalTranscoders/transcodersPerLine < 1 {
		progressBarWidth = maxCharsPerLine / totalTranscoders
		transcodersPerLine = maxCharsPerLine / progressBarWidth
	}

	for _, transcoder := range t.Transcoders {
		bar := progressBar(transcoder.Status, transcoder.Progress, progressBarWidth)

		//view += fmt.Sprintf("%s%s%s %s %.0f%%\n", indent, indent, chunkNumStr, bar, transcoder.Progress)
		view += bar

		totalFps += transcoder.Fps
		totalSpeed += transcoder.Speed
		totalOutTime += transcoder.OutTime

		// 6 transcoders per line
		if len(t.Transcoders) > transcodersPerLine {
			if counter == transcodersPerLine-1 {
				view += "\n" + indent + indent
				counter = 0
			} else {
				counter++
			}
		}
	}
	view += "\n"

	statusInfo := fmt.Sprintf("%s %.f%%, FPS: %.0f, Speed: %.1fx, OutTime: %s", statusFormatStyle(t.Command.Format), t.Job.Progress, totalFps, totalSpeed, formatter.Duration(totalOutTime))
	return view, statusInfo
}

func (t TUI) downloadView() (string, string) {
	view := ""
	statusInfo := ""

	view += "\n"
	if t.DownloadProgress.Progress == 100 {
		view += fmt.Sprintf("%s%s %s written (%s)\n\n", indent, completedIconStyle.String(), t.Command.Output, formatter.Size(t.DownloadProgress.WrittenBytes))
	} else {
		statusInfo = fmt.Sprintf("%.1f%% (%.1f MB/s, ETA: %s)", t.DownloadProgress.Progress, t.DownloadProgress.Speed/(1024*1024), t.DownloadProgress.Eta.Round(time.Second))
		view += currentStepStyle(fmt.Sprintf("%s%s Saving video",
			indent,
			t.Spinner.View()))
		view += "\n"
	}
	return view, statusInfo
}

func progressBar(status string, progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	filled := int((progress/100.0)*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "▮"
		} else {
			if status == chunkify.TranscoderStatusPending {
				bar += pendingTranscoderChunkStyle("▯")
			} else {
				bar += "▯"
			}
		}
	}
	bar += ""
	return bar
}

// getStatusString returns a human-readable status string
func (t TUI) getStatusString() string {
	switch t.Status {
	case Status:
		return "Initializing"
	case Starting:
		return "Starting"
	case UploadingFromUrl:
		return "Uploading"
	case UploadingFromFile:
		return "Uploading"
	case Transcoding:
		if t.Job != nil {
			return cases.Title(language.English).String(t.Job.Status)
		}
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
