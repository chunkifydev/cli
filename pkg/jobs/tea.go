package jobs

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/styles"
)

type model struct {
	cmd       *ListCmd
	ch        chan []chunkify.Job
	jobsTable *table.Table
}

type tickMsg time.Time

func listenToJobChan(ch chan []chunkify.Job) tea.Cmd {
	return func() tea.Msg {
		jobs := <-ch
		return jobs
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), listenToJobChan(m.ch))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	case []chunkify.Job:
		m.cmd.Data = msg
		return m, listenToJobChan(m.ch)
	case tickMsg:
		m.jobsTable.Data(table.NewStringData(jobsListToRows(m.cmd.Data)...))
		return m, tickCmd()
	}
	return m, nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) View() string {
	s := m.jobsTable.String()
	s += "\n\n"
	s += styles.Debug.Render("[Q] Exit\n")
	return s
}

func polling(r *ListCmd, ch chan []chunkify.Job) {
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

func StartPolling(r *ListCmd) {
	ch := make(chan []chunkify.Job)
	go polling(r, ch)

	m := model{
		cmd:       r,
		ch:        ch,
		jobsTable: r.jobsTable(),
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
