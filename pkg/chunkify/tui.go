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
	ReadingFromStorage
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
	Transcoding      *chunkify.Transcoding
	Source           *chunkify.Source
	Files            []chunkify.TranscodingFile
	UploadProgress   UploadProgress
	DownloadProgress DownloadProgress
	DownloadedFiles  map[string]chunkify.TranscodingFile
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
		DownloadedFiles: map[string]chunkify.TranscodingFile{},
		LastJSONOutput:  time.Now(),
	}
}

type Progress struct {
	Status               chan int
	TranscodingProgress  chan chunkify.Transcoding
	TranscodingCompleted chan bool
	UploadProgress       chan UploadProgress
	DownloadProgress     chan DownloadProgress
	Files                chan []chunkify.TranscodingFile
	DownloadedFiles      chan chunkify.TranscodingFile
	Source               chan *chunkify.Source
	Error                chan error
}

func NewProgress() *Progress {
	return &Progress{
		Status:               make(chan int, 1),
		TranscodingProgress:  make(chan chunkify.Transcoding, 100),
		TranscodingCompleted: make(chan bool, 1),
		UploadProgress:       make(chan UploadProgress, 100),
		DownloadProgress:     make(chan DownloadProgress, 100),
		DownloadedFiles:      make(chan chunkify.TranscodingFile, 100),
		Source:               make(chan *chunkify.Source, 1),
		Error:                make(chan error),
		Files:                make(chan []chunkify.TranscodingFile, 100),
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
			// Send TranscodingCompleted to unblock the main goroutine
			select {
			case t.Progress.TranscodingCompleted <- true:
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
	case transcoding, ok := <-t.Progress.TranscodingProgress:
		if ok {
			t.Transcoding = &transcoding
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
			t.DownloadedFiles[downloadedFile.ID] = downloadedFile
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
		select {
		case t.Progress.TranscodingCompleted <- true:
		default:
		}
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
		speedStr = "N/A"
		if t.Transcoding != nil {
			progress = t.Transcoding.Progress
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

	// Display transcoding progress
	if t.Command.Format != "" && t.Status >= Transcoding && t.Transcoding != nil {
		view += "\n"
		v, statusInfo = t.transcodingView()
		view += v
	}

	// Display download progress
	if t.Command.Output != "" && t.Transcoding != nil && t.Transcoding.Status == chunkify.TranscodingStatusReady && t.Status >= Downloading {
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
	//if apiErr, ok := t.Error.(chunkify.ChunkifyError); ok {
	//	view = fmt.Sprintf("%s%s", indent, errorText(apiErr.Message))
	//	} else {
	view = fmt.Sprintf("%s%s", indent, errorText(t.Error.Error()))
	//	}
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
	if t.Status == ReadingFromStorage {
		statusInfo = t.Command.Input
		view = currentStepText(fmt.Sprintf("%s%s Reading source from storage\n", indent, spin.View()))
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

	if t.Transcoding.Status == chunkify.TranscodingStatusReady {
		view += fmt.Sprintf("%s%s Transcoding completed\n", indent, completedIcon.String())
	} else {
		view += currentStepText(fmt.Sprintf("%s%s Transcoding", indent, spin.View()))
		view += " " + formatter.TimeDiff(t.Transcoding.CreatedAt, time.Now())
		view += "\n"
	}

	view += dblIndent + progressBar(string(t.Transcoding.ProcessingStage), t.Transcoding.Progress, 60) + "\n"

	statusInfo := statusOrangeText(t.Command.Format)
	statusInfo += fmt.Sprintf(" %.f%%", t.Transcoding.Progress)
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
		_, downloaded = t.DownloadedFiles[file.ID]
		if !downloaded {
			currentFile = i - 1
		}

		if currentFile < 0 {
			currentFile = 0
		}
		if len(t.Files) <= 10 {
			info := fmt.Sprintf("%s (%s)", filename(file, t.Command.Output, len(t.Files)), formatter.Size(file.Size))
			if downloaded {
				view += dblIndent + "- " + info
			} else {
				view += grayText(dblIndent + "> " + info)
			}
			view += "\n"
		}
	}

	if len(t.Files) > 10 {
		view += fmt.Sprintf("%s%d/%d %s", dblIndent, len(t.DownloadedFiles), len(t.Files), filename(t.Files[currentFile], t.Command.Output, len(t.Files)))
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
			if status == "queued" || status == "starting" {
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
	view := fmt.Sprintf("\n%s────────────────────────────────────────────────\n\n", indent)
	// if format is not set, we show the source ID
	if t.Command.Format == "" && t.Source != nil {
		view += fmt.Sprintf("%sSource ID: %s\n\n", indent, t.Source.ID)
		return view
	}

	if t.Transcoding != nil {
		view += fmt.Sprintf("%sTranscoding ID: %s\n", indent, t.Transcoding.ID)
		view += fmt.Sprintf("%sSource ID: %s\n", indent, t.Transcoding.SourceID)
		view += fmt.Sprintf("%sFormat: %s ", indent, t.Command.Format)
		view += formatConfig(t.Transcoding.Format)
		view += "\n"

		if !t.Transcoding.ReadyAt.IsZero() {
			view += fmt.Sprintf("%sTranscoding time: %s\n", indent, formatter.TimeDiff(t.Transcoding.CreatedAt, t.Transcoding.ReadyAt))
		}
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
	case ReadingFromStorage:
		return "Reading source"
	case Transcoding:
		if t.Transcoding != nil {
			status := string(t.Transcoding.Status)
			if t.Transcoding.Status == chunkify.TranscodingStatusPreparing && t.Transcoding.ProcessingStage != "" {
				status = string(t.Transcoding.ProcessingStage)
			}
			return cases.Title(language.English).String(status)
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

func formatConfig(config chunkify.TranscodingFormat) string {
	jsonBytes, _ := json.Marshal(config)
	configMap := make(map[string]any)
	err := json.Unmarshal(jsonBytes, &configMap)
	if err != nil {
		return ""
	}
	view := ""
	if len(configMap) > 0 {
		keys := []string{}
		for k := range configMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			val := configMap[k]
			switch v := val.(type) {
			case string:
				if v != "" {
					view += fmt.Sprintf("%s=%s ", k, val)
				}
			case int, int64:
				if val.(int) > 0 {
					view += fmt.Sprintf("%s=%d ", k, val)
				}
			case float64:
				if v > 0 {
					view += fmt.Sprintf("%s=%.2f ", k, val)
				}
			case bool:
				if v {
					view += fmt.Sprintf("%s=%t ", k, val)
				}
			}
		}
	}

	return view
}
