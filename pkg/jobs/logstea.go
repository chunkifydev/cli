package jobs

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/styles"
)

type logModel struct {
	cmd       *LogsListCmd
	ch        chan []api.Log
	logsTable *table.Table
}

func listenToLogsChan(ch chan []api.Log) tea.Cmd {
	return func() tea.Msg {
		jobs := <-ch
		return jobs
	}
}

func (m logModel) Init() tea.Cmd {
	return tea.Batch(tickCmd(), listenToLogsChan(m.ch))
}

func (m logModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	case []api.Log:
		m.cmd.Data = msg
		return m, listenToLogsChan(m.ch)
	case tickMsg:
		m.logsTable.Data(table.NewStringData(logsListToRows(m.cmd.Data, m.cmd.Service, m.cmd.Levels, m.cmd.NoProgress)...))
		return m, tickCmd()
	}
	return m, nil
}

func (m logModel) View() string {
	s := m.logsTable.String()
	s += "\n\n"
	s += styles.Debug.Render("Logs will appear as soon as they are available.\nPress q to quit.\n")
	return s
}

func logsPolling(r *LogsListCmd, ch chan []api.Log) {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()

	for range t.C {
		if err := r.Execute(); err != nil {
			printError(err)
			return
		}
		ch <- r.Data
	}
}

func StartTailing(r *LogsListCmd) {
	ch := make(chan []api.Log)
	go logsPolling(r, ch)

	m := logModel{
		cmd:       r,
		ch:        ch,
		logsTable: r.logsTable(),
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
