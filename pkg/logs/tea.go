package logs

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/level63/cli/pkg/api"
	"github.com/level63/cli/pkg/styles"
)

type tickMsg time.Time

type model struct {
	cmd       *ListCmd
	ch        chan []api.Log
	logsTable *table.Table
}

func ListenToLogsChan(ch chan []api.Log) tea.Cmd {
	return func() tea.Msg {
		jobs := <-ch
		return jobs
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), ListenToLogsChan(m.ch))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	case []api.Log:
		m.cmd.Data = msg
		return m, ListenToLogsChan(m.ch)
	case tickMsg:
		m.logsTable.Data(table.NewStringData(logsListToRows(m.cmd)...))
		return m, tickCmd()
	}
	return m, nil
}

func (m model) View() string {
	s := m.logsTable.String()
	s += "\n\n"
	s += styles.Debug.Render("Logs will appear as soon as they are available.\n[Q] Exit\n")
	return s
}

func LogsPolling(r *ListCmd, ch chan []api.Log) {
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

func StartTailing(r *ListCmd) {
	ch := make(chan []api.Log)
	go LogsPolling(r, ch)

	m := model{
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
