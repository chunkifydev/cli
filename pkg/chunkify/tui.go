package chunkify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/formatter"
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

var (
	completedIcon    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("▮")
	currentStepText  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render
	errorText        = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render
	statusText       = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("42")).Padding(0, 1).Bold(true).Render
	statusOrangeText = lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(lipgloss.Color("#E7AE59")).Padding(0, 1).Bold(true).Render
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

	// Will only output in JSON format
	// See JSONView()
	JSON           bool
	LastJSONOutput time.Time
}

func NewApp() *App {
	return &App{
		Status:          Starting,
		Progress:        NewProgress(),
		DownloadedFiles: map[string]chunkify.File{},
		LastJSONOutput:  time.Now(),
	}
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

func (t App) Init() tea.Cmd {
	spin.Spinner = spinner.MiniDot
	return tea.Batch(tickCmd(), spin.Tick)
}

func (t App) Run() {
	var p *tea.Program
	if t.JSON {
		// Disable Bubble Tea renderer in JSON mode to avoid whitespace artifacts
		p = tea.NewProgram(t, tea.WithoutRenderer())
	} else {
		p = tea.NewProgram(t)
	}
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

		// if JSON mode is enabled, print the JSON to the terminal every second
		if t.JSON {
			if shouldQuit || t.Done || time.Since(t.LastJSONOutput) >= time.Second {
				t.LastJSONOutput = time.Now()
				fmt.Println(t.JSONView())
			}
		}

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

	return t, t.Done
}

type JSONOutput struct {
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	Fps      float64 `json:"fps"`
	Speed    string  `json:"speed"`
	OutTime  int64   `json:"out_time"`
	Eta      string  `json:"eta"`
}

func (t App) JSONView() string {
	fps := 0.0
	speed := 0.0
	outTime := int64(0)
	progress := 0.0
	eta := "N/A"

	speedStr := ""

	switch t.Status {
	case UploadingFromFile:
		progress = t.UploadProgress.Progress
		speedStr = formatter.Bitrate(int64(t.UploadProgress.Speed))
		eta = formatter.Duration(int64(t.UploadProgress.Eta.Round(time.Second).Seconds()))
	case UploadingFromUrl:
		progress = 100.0
		speedStr = "N/A"
		eta = formatter.Duration(int64(t.UploadProgress.Eta.Round(time.Second).Seconds()))
	case Downloading:
		progress = t.DownloadProgress.Progress
		speedStr = formatter.Bitrate(int64(t.DownloadProgress.Speed))
		eta = formatter.Duration(int64(t.DownloadProgress.Eta.Round(time.Second).Seconds()))
	case Transcoding:
		eta = "N/A"
		if t.Job != nil {

			for _, transcoder := range t.Transcoders {
				fps += transcoder.Fps
				speed += transcoder.Speed
				outTime += transcoder.OutTime
			}

			speedStr = fmt.Sprintf("%.1fx", speed)
			progress = t.Job.Progress
		}
	}

	out := JSONOutput{
		Status:   t.getStatusString(),
		Progress: progress,
		Fps:      fps,
		Speed:    speedStr,
		OutTime:  outTime,
		Eta:      eta,
	}
	j, _ := json.Marshal(out)
	return string(j)
}

func (t App) View() string {
	if t.JSON {
		// No side-effects in View when JSON mode is on; Update handles printing
		return ""
	}
	var view string
	statusInfo := ""

	var v string
	v, statusInfo = t.uploadView()
	view += v

	if t.Source != nil {
		v, statusInfo = t.sourceView()
		view += v
	}

	// Display job progress
	if t.Command.Format != "" && t.Status >= Transcoding && t.Job != nil {
		view += "\n"
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
		view += fmt.Sprintf("\n\n%s%s %s\n", indent, statusText(t.getStatusString()), statusInfo)
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
		statusInfo = fmt.Sprintf("%.1f%% (%s, ETA: %s)", t.UploadProgress.Progress, formatter.Bitrate(int64(t.UploadProgress.Speed)), t.UploadProgress.Eta.Round(time.Second))

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
	view += "\n"
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
		view += progressBar(transcoder.Status, transcoder.Progress, progressBarWidth)

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

	view := fmt.Sprintf("\n%s────────────────────────────────────────────────\n\n", indent)
	// if format is not set, we show the source ID
	if t.Command.Format == "" && t.Source != nil {
		view += fmt.Sprintf("%sSource ID: %s\n\n", indent, t.Source.Id)
		return view
	}

	if t.Job != nil {
		view += fmt.Sprintf("%sJob ID: %s\n", indent, t.Job.Id)
		view += fmt.Sprintf("%sSource ID: %s\n", indent, t.Job.SourceId)

		if t.Job.HlsManifestId != nil {
			view += fmt.Sprintf("\n%sHLS Manifest: %s", indent, *t.Job.HlsManifestId)
		}

		view += fmt.Sprintf("%sFormat: %s ", indent, t.Command.Format)
		view += formatConfig(t.Job.Format.Config)
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
		return "Queued"
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
