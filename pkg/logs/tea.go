package logs

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	chunkify "github.com/chunkifydev/chunkify-go"
	"github.com/chunkifydev/cli/pkg/styles"
)

// tickMsg represents a tick event for periodic updates
type tickMsg time.Time

// model represents the state and data needed for the interactive logs list view
type model struct {
	cmd       *ListCmd            // Command for listing logs
	ch        chan []chunkify.Log // Channel for receiving log updates
	logsTable *table.Table        // Table for displaying logs
}

// ListenToLogsChan creates a tea.Cmd that listens for log updates on a channel
func ListenToLogsChan(ch chan []chunkify.Log) tea.Cmd {
	return func() tea.Msg {
		jobs := <-ch
		return jobs
	}
}

// tickCmd creates a tea.Cmd that sends tick events periodically
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the model and starts the update loop
func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), ListenToLogsChan(m.ch))
}

// Update handles incoming messages and updates the model state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	case []chunkify.Log:
		m.cmd.Data = msg
		return m, ListenToLogsChan(m.ch)
	case tickMsg:
		m.logsTable.Data(table.NewStringData(logsListToRows(m.cmd)...))
		return m, tickCmd()
	}
	return m, nil
}

// View renders the current state as a string
func (m model) View() string {
	s := m.logsTable.String()
	s += "\n\n"
	s += styles.Debug.Render("Logs will appear as soon as they are available.\n[Q] Exit\n")
	return s
}

// LogsPolling periodically fetches updated log data and sends it on a channel
func LogsPolling(r *ListCmd, ch chan []chunkify.Log) {
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
	ch := make(chan []chunkify.Log)
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
