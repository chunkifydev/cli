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

// model represents the application state
type model struct {
	cmd          *ListCmd               // Command for listing sources
	ch           chan []chunkify.Source // Channel for receiving source updates
	sourcesTable *table.Table           // Table for displaying sources
}

// tickMsg represents a tick event for UI updates
type tickMsg time.Time

// initialModel creates and returns the initial application model
func initialModel(cmd *ListCmd, ch chan []chunkify.Source) model {
	return model{
		cmd:          cmd,
		ch:           ch,
		sourcesTable: cmd.sourcesTable(),
	}
}

// listenToSourceChan creates a tea.Cmd that listens for source updates on the channel
func listenToSourceChan(ch chan []chunkify.Source) tea.Cmd {
	return func() tea.Msg {
		jobs := <-ch
		return jobs
	}
}

// Init initializes the model and returns the initial commands to run
func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), listenToSourceChan(m.ch))
}

// Update handles incoming messages and updates the model accordingly
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

// tickCmd creates a command that sends tick messages at regular intervals
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// View renders the current model state as a string
func (m model) View() string {
	s := m.sourcesTable.String()
	s += "\n\n"
	s += styles.Debug.Render("[Q] Exit\n")
	return s
}

// polling continuously fetches source updates and sends them through the channel
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

// StartPolling initializes and runs the interactive source polling interface
func StartPolling(r *ListCmd) {
	ch := make(chan []chunkify.Source)
	go polling(r, ch)
	p := tea.NewProgram(initialModel(r, ch))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
