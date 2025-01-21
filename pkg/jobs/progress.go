package jobs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/chunkifydev/cli/pkg/api"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/logs"
	"github.com/chunkifydev/cli/pkg/styles"
	"github.com/spf13/cobra"
)

var progressBarColor = "#7ce4a1"

type progressModel struct {
	cmd                *logs.ListCmd
	ch                 chan []api.Log
	transcoderProgress map[string]api.LogAttrs
	progressBars       map[string]progress.Model
}

func (m *progressModel) Init() tea.Cmd {
	m.progressBars = map[string]progress.Model{}
	return tea.Batch(tickCmd(), logs.ListenToLogsChan(m.ch))
}

func (m *progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	case []api.Log:
		m.cmd.Data = msg
		return m, logs.ListenToLogsChan(m.ch)
	case tickMsg:
		m.getLastProgressLines(m.cmd.Data)
		return m, tickCmd()
	}
	return m, nil
}

func (m *progressModel) getLastProgressLines(logs []api.Log) {
	transcoders := map[string]api.LogAttrs{}

	for i := len(logs) - 1; i >= 0; i-- {
		log := logs[i]
		if log.Msg == "Progress" {
			if _, ok := transcoders[log.Service]; ok {
				continue
			}
			transcoders[log.Service] = log.Attributes

			pr := progress.New(progress.WithSolidFill(progressBarColor), progress.WithWidth(30))

			m.progressBars[log.Service] = pr
		}
	}

	m.transcoderProgress = transcoders
}

func (m *progressModel) View() string {
	s := ""
	transcoders := make([]api.LogAttrs, 50)

	for k, v := range m.transcoderProgress {
		trName := strings.Split(k, "#")
		if len(trName) == 2 {
			index, err := strconv.Atoi(trName[1])
			if err != nil {
				continue
			}
			transcoders[index-1] = v
		}
	}

	count := 0
	var (
		overallProgress float64
		frame           float64
		fps             float64
		totalSize       float64
		outTime         float64
		speed           float64
	)

	for i, tr := range transcoders {
		if _, ok := tr["progress"].(float64); !ok {
			break
		}

		count++

		s += fmt.Sprintf("transcoder#%2d", i+1) + " "
		s += m.progressBars[fmt.Sprintf("transcoder#%d", i+1)].ViewAs(tr["progress"].(float64)/100) + " "
		s += styles.Debug.Render("frame=") + fmt.Sprintf("%6.f", tr["frame"]) + " "
		s += styles.Debug.Render("fps=") + fmt.Sprintf("%4.f", tr["fps"]) + " "
		s += styles.Debug.Render("Lsize=") + fmt.Sprintf("%8.fkB", tr["total_size"].(float64)/1024) + " "
		s += styles.Debug.Render("time=") + fmt.Sprintf("%v", formatter.Duration(int64(tr["out_time"].(float64)))) + " "
		s += styles.Debug.Render("bitrate=") + fmt.Sprintf("%v", tr["bitrate"]) + " "
		s += styles.Debug.Render("speed=") + fmt.Sprintf("%.2fx", tr["speed"])
		s += "\n\n"

		overallProgress += tr["progress"].(float64)
		frame += tr["frame"].(float64)
		fps += tr["fps"].(float64)
		totalSize += tr["total_size"].(float64)
		outTime += tr["out_time"].(float64)
		speed += tr["speed"].(float64)
	}

	pr := progress.New(progress.WithSolidFill(progressBarColor), progress.WithWidth(30))
	s += "total         "
	s += pr.ViewAs(overallProgress/float64(count)/100) + " "
	s += styles.Debug.Render("frame=") + fmt.Sprintf("%6.f", frame) + " "
	s += styles.Debug.Render("fps=") + fmt.Sprintf("%4.f", fps) + " "
	s += styles.Debug.Render("Lsize=") + fmt.Sprintf("%8.fkB", totalSize/1024) + " "
	s += styles.Debug.Render("time=") + fmt.Sprintf("%v", formatter.Duration(int64(outTime))) + " "
	s += styles.Debug.Render("bitrate=") + "             "
	s += styles.Debug.Render("speed=") + fmt.Sprintf("%.2fx", speed)
	s += "\n\n"

	s += styles.Debug.Render("[Q] Exit\n")
	return s
}

func StartTranscoderProgressTailing(r *logs.ListCmd) {
	ch := make(chan []api.Log)
	go logs.LogsPolling(r, ch)

	m := progressModel{
		cmd:                r,
		ch:                 ch,
		transcoderProgress: map[string]api.LogAttrs{},
	}

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func newTranscoderProgressCmd() *cobra.Command {
	req := logs.ListCmd{Service: "transcoder"}

	cmd := &cobra.Command{
		Use:   "progress job-id",
		Short: "progress mode",
		Long:  `progress mode`,
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			req.Id = args[0]
			if err := req.Execute(); err != nil {
				printError(err)
				return
			}

			StartTranscoderProgressTailing(&req)
		},
	}

	return cmd
}
