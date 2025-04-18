package sources

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
	cmd          *ListCmd
	ch           chan []chunkify.Source
	sourcesTable *table.Table
}

type tickMsg time.Time

func initialModel(cmd *ListCmd, ch chan []chunkify.Source) model {
	return model{
		cmd:          cmd,
		ch:           ch,
		sourcesTable: cmd.sourcesTable(),
	}
}

func listenToSourceChan(ch chan []chunkify.Source) tea.Cmd {
	return func() tea.Msg {
		jobs := <-ch
		return jobs
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), listenToSourceChan(m.ch))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	case []chunkify.Source:
		m.cmd.Data = msg
		return m, listenToSourceChan(m.ch)
	case tickMsg:
		m.sourcesTable.Data(table.NewStringData(sourcesListToRows(m.cmd.Data)...))
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
	s := m.sourcesTable.String()
	s += "\n\n"
	s += styles.Debug.Render("[Q] Exit\n")
	return s
}

func polling(r *ListCmd, ch chan []chunkify.Source) {
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
	ch := make(chan []chunkify.Source)
	go polling(r, ch)
	p := tea.NewProgram(initialModel(r, ch))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
