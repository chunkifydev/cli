package chunkify

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/version"
	"github.com/google/uuid"
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
	completedIcon    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("▮")
	currentStepText  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render
	errorText        = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render
	statusText       = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("42")).Padding(0, 1).Bold(true).Render
	statusOrangeText = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#E7AE59")).Padding(0, 1).Bold(true).Render
	infoText         = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE")).Render
	grayText         = lipgloss.NewStyle().Foreground(lipgloss.Color("#59636e")).Render

	indent    = "  "
	dblIndent = indent + indent

	spin = spinner.New()
)

type App struct {
	Client           *chunkify.Client
	Command          *ChunkifyCommand
	Status           int
	Progress         *Progress
	Job              *chunkify.Job
	Source           *chunkify.Source
	Files            []chunkify.File
	Transcoders      []chunkify.TranscoderStatus
	UploadProgress   chunkify.UploadProgress
	DownloadProgress DownloadProgress
	DownloadedFiles  map[string]chunkify.File
	Error            error
	Done             bool
	Ctx              context.Context
	CancelFunc       context.CancelFunc
}

type Progress struct {
	Status           chan int
	JobProgress      chan chunkify.Job
	JobTranscoders   chan []chunkify.TranscoderStatus
	JobCompleted     chan bool
	UploadProgress   chan chunkify.UploadProgress
	DownloadProgress chan DownloadProgress
	Files            chan []chunkify.File
	DownloadedFiles  chan chunkify.File
	Source           chan *chunkify.Source
	Error            chan error
}

func init() {
	chunkifyBanner = strings.Replace(chunkifyBanner, "{version}", version.Version, 1)
}

func NewProgress() *Progress {
	return &Progress{
		Status:           make(chan int, 1),
		JobProgress:      make(chan chunkify.Job, 100),
		JobTranscoders:   make(chan []chunkify.TranscoderStatus, 100),
		JobCompleted:     make(chan bool, 1),
		UploadProgress:   make(chan chunkify.UploadProgress, 100),
		DownloadProgress: make(chan DownloadProgress, 100),
		DownloadedFiles:  make(chan chunkify.File, 100),
		Source:           make(chan *chunkify.Source, 1),
		Error:            make(chan error),
		Files:            make(chan []chunkify.File, 100),
	}
}

// NewApp creates a new TUI instance with the given progress tracker
func NewApp(ctx context.Context, cancelFunc context.CancelFunc, cfg *config.Config) App {
	spin.Spinner = spinner.MiniDot

	app := App{
		Client:          cfg.Client,
		Status:          Starting,
		Progress:        NewProgress(),
		Ctx:             ctx,
		CancelFunc:      cancelFunc,
		Done:            false,
		DownloadedFiles: map[string]chunkify.File{},
		Command: &ChunkifyCommand{
			Id:                  uuid.New().String(),
			Input:               chunkifyCmd.Input,
			Output:              chunkifyCmd.Output,
			Format:              chunkifyCmd.Format,
			JobFormatParams:     chunkifyCmd.JobFormatParams,
			JobTranscoderParams: chunkifyCmd.JobTranscoderParams,
		},
	}

	return app
}

func (t App) Init() tea.Cmd {
	return tea.Batch(tickCmd(), spin.Tick)
}

func (t App) Run() {
	p := tea.NewProgram(t)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
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

func (t App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			fmt.Println("\nQuitting...")
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
		t, shouldQuit := t.checkChannels()
		if shouldQuit {
			return t, tea.Quit
		}
		return t, tickCmd()
	default:
		var cmd tea.Cmd
		spin, cmd = spin.Update(msg)
		return t, tea.Batch(cmd)
	}

	return t, nil
}

// checkChannels performs non-blocking reads from all channels
// Returns the updated app and whether the TUI should quit
func (t App) checkChannels() (App, bool) {
	// Check job progress
	select {
	case job, ok := <-t.Progress.JobProgress:
		if ok {
			t.Job = &job
		}
	case transcoders, ok := <-t.Progress.JobTranscoders:
		if ok {
			t.Transcoders = transcoders
		}
	case uploadProgress, ok := <-t.Progress.UploadProgress:
		if ok {
			t.UploadProgress = uploadProgress
		}
	case downloadProgress, ok := <-t.Progress.DownloadProgress:
		if ok {
			t.DownloadProgress = downloadProgress
		}
	case downloadedFile, ok := <-t.Progress.DownloadedFiles:
		if ok {
			t.DownloadedFiles[downloadedFile.Id] = downloadedFile
		}
	case files, ok := <-t.Progress.Files:
		if ok {
			t.Files = files
		}
	case status, ok := <-t.Progress.Status:
		if ok {
			t.Status = status
			if status == Completed || status == Failed || status == Cancelled {
				t.Done = true
			}
		}
	case source, ok := <-t.Progress.Source:
		if ok {
			t.Source = source
		}
	case err := <-t.Progress.Error:
		t.Error = err
		t.Progress.JobCompleted <- true
	case <-t.Ctx.Done():
		t.Done = true
	default:
	}

	if t.Done {
		time.Sleep(1 * time.Second)
		return t, true
	}

	return t, t.Done
}

func (t App) View() string {
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
	if t.Command.Output != "" && t.Job != nil && t.Job.Status == chunkify.JobStatusCompleted && t.Status >= Downloading {
		v, statusInfo = t.downloadView()
		view += v
	}

	if t.Error != nil {
		view += t.errorView()
		view += "\n"
		return view
	}

	if !t.Done {
		view += fmt.Sprintf("\n\n%s%s %s\n", indent, statusText(t.getStatusString()), infoText(statusInfo))
	} else {
		view += t.summaryView()
	}

	return view
}

func (t App) errorView() string {
	var view string
	if apiErr, ok := t.Error.(chunkify.ApiError); ok {
		view = fmt.Sprintf("%s%s", indent, errorText(apiErr.Message))
	} else {
		view = fmt.Sprintf("%s%s", indent, errorText(t.Error.Error()))
	}
	view += "\n"
	return view
}

func (t App) uploadView() (string, string) {
	view := ""
	statusInfo := ""

	if t.Status == UploadingFromFile {
		statusInfo = fmt.Sprintf("%.1f%% (%.1f MB/s, ETA: %s)", t.UploadProgress.Progress, t.UploadProgress.Speed/(1024*1024), t.UploadProgress.Eta.Round(time.Second))

		// Display upload progress
		view = currentStepText(fmt.Sprintf("%s%s Uploading %s",
			indent,
			spin.View(),
			t.Command.Input))
		view += "\n"
	}
	if t.Status == UploadingFromUrl {
		statusInfo = t.Command.Input
		view += fmt.Sprintf("%s%s Uploaded from URL\n", indent, completedIcon.String())
	}

	return view, statusInfo
}

func (t App) sourceView() (string, string) {
	view := fmt.Sprintf("%s%s Source: %s\n", indent, completedIcon.String(), sourceName(t.Command.Input))
	view += fmt.Sprintf("%sDuration: %s Size: %s Video: %s, %dx%d, %s, %.2ffps", indent+indent, formatter.Duration(t.Source.Duration), formatter.Size(t.Source.Size), t.Source.VideoCodec, t.Source.Width, t.Source.Height, formatter.Bitrate(t.Source.VideoBitrate), t.Source.VideoFramerate)
	view += "\n\n"
	return view, t.Command.Format
}

func sourceName(input string) string {
	name := ""
	if u, err := url.Parse(input); err == nil {
		name += u.Path
		if u.Host != "" {
			name += fmt.Sprintf(" (%s)", u.Host)
		}

		return name
	}
	return path.Base(input)
}

func (t App) transcodingView() (string, string) {
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
		view += fmt.Sprintf("%s%s Transcoding completed\n", indent, completedIcon.String())
	} else {
		view += currentStepText(fmt.Sprintf("%s%s Transcoding (%d x %s)", indent, spin.View(), totalTranscoders, t.Job.Transcoder.Type))
		view += " " + formatter.TimeDiff(t.Job.StartedAt, time.Now())
		view += "\n"
	}

	// start of progress bar
	view += dblIndent
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
				view += "\n" + dblIndent
				counter = 0
			} else {
				counter++
			}
		}
	}
	view += "\n"

	statusInfo := statusOrangeText(t.Command.Format)
	if totalOutTime > 0 {
		statusInfo += fmt.Sprintf(" %.f%%, FPS: %.0f, Speed: %.1fx, OutTime: %s", t.Job.Progress, totalFps, totalSpeed, formatter.Duration(totalOutTime))
	}
	return view, statusInfo
}

func (t App) downloadView() (string, string) {
	view := ""
	statusInfo := ""

	view += "\n"
	// We downloaded all files
	if t.Status == Completed && len(t.DownloadedFiles) == len(t.Files) {
		view += fmt.Sprintf("%s%s All files saved\n", indent, completedIcon.String())
	} else {
		statusInfo = fmt.Sprintf("%.1f%% (%.1f MB/s, ETA: %s)", t.DownloadProgress.Progress, t.DownloadProgress.Speed/(1024*1024), t.DownloadProgress.Eta.Round(time.Second))
		view += currentStepText(fmt.Sprintf("%s%s Saving files",
			indent,
			spin.View(),
		))
		view += "\n"
	}

	currentFile := 0
	downloaded := false

	for i, file := range t.Files {
		_, downloaded = t.DownloadedFiles[file.Id]
		if !downloaded {
			currentFile = i - 1
		}

		if currentFile < 0 {
			currentFile = 0
		}
		if len(t.Files) <= 10 {
			info := fmt.Sprintf("%s (%s)", filename(file, t.Command.Output), formatter.Size(file.Size))
			if downloaded {
				view += dblIndent + "- " + info
			} else {
				view += grayText(dblIndent + "> " + info)
			}
			view += "\n"
		}
	}

	if len(t.Files) > 10 {
		view += fmt.Sprintf("%s%d/%d %s", dblIndent, len(t.DownloadedFiles), len(t.Files), filename(t.Files[currentFile], t.Command.Output))
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
	for i := range width {
		if i < filled {
			bar += "▮"
		} else {
			if status == chunkify.TranscoderStatusPending || status == chunkify.TranscoderStatusStarting {
				bar += grayText("▯")
			} else {
				bar += "▯"
			}
		}
	}
	bar += ""
	return bar
}

func (t App) summaryView() string {
	speed := 0.0

	for _, transcoder := range t.Transcoders {
		speed += transcoder.Speed
	}

	view := fmt.Sprintf("\n%s────────────────────────────────────────────────\n", indent)
	if t.Job != nil {
		view += fmt.Sprintf("%sJob ID: %s\n", indent, t.Job.Id)

		view += fmt.Sprintf("%sFormat: %s ", indent, t.Command.Format)
		view += formatConfig(t.Job.Format.Config)

		if t.Job.HlsManifestId != nil {
			view += fmt.Sprintf("\n%sHLS Manifest: %s", indent, *t.Job.HlsManifestId)
		}
		view += "\n"

		view += fmt.Sprintf("%sTranscoders: %d x %s\n", indent, t.Job.Transcoder.Quantity, t.Job.Transcoder.Type)
		view += fmt.Sprintf("%sSpeed: %.1fx\n", indent, speed)
		view += fmt.Sprintf("%sTranscoding time: %s\n", indent, formatter.TimeDiff(t.Job.StartedAt, t.Job.UpdatedAt))
		view += fmt.Sprintf("%sBillable time: %ds\n", indent, t.Job.BillableTime)
	}

	view += "\n"
	return view
}

// getStatusString returns a human-readable status string
func (t App) getStatusString() string {
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

func formatConfig(config map[string]any) string {
	view := ""
	if len(config) > 0 {
		keys := []string{}
		for k := range config {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			view += fmt.Sprintf("%s=%v ", k, config[k])
		}
	}

	return view
}
