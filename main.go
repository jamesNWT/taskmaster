package main

import (
	"fmt"
	"time"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// model
type model struct {
	keymap   keymap
	help     help.Model
	focus    stopwatch.Model
	// rest     stopwatch.Model
	// focusing bool
	// editing  bool
	// todo     []string
}	

// keymap
type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
}

// initialModel

func initialModel() model {
	keymap := keymap{
		start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
		stop: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stop"),
		),
		reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	keymap.start.SetEnabled(false)

	m := model{
		keymap: keymap,
		help: help.New(),
		focus: stopwatch.NewWithInterval(time.Millisecond),
	}
	return m
}
		

// Update
func (m model) Update(msg tea.Msg) (tea.Model,tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
			case key.Matches(msg, m.keymap.quit): 
				return m, tea.Quit
		}
	}
	m.focus, cmd = m.focus.Update(msg)
	return m, cmd
}

// View
func (m model) View() string {
	s := m.focus.View() + "\n"

	return s
}

// Init
func (m model) Init() tea.Cmd {
	return m.focus.Init()
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("It didn't work: ", err)
		os.Exit(1)
	}
}
