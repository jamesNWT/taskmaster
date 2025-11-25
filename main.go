package main

import (
	"fmt"
	"time"
	"os"

	"github.com/charmbracelet/bubbles/stopwatch"
	// "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// model
type model struct {
	focus    stopwatch.Model
	// rest     stopwatch.Model
	// focusing bool
	// editing  bool
	// todo     []string
}	

// Update

func (m model) Update(msg tea.Msg) (tea.Model,tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
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
	m := model{
		focus: stopwatch.NewWithInterval(time.Millisecond),
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("It didn't work: ", err)
		os.Exit(1)
	}
}
