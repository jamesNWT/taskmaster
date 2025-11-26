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
	rest     stopwatch.Model
	editing  bool
	quiting  bool
	todo     []string
}	

// keymap
type keymap struct {
	switchMode key.Binding
	quit  key.Binding
}

// initialModel

func initialModel() model {
	keymap := keymap{
		switchMode: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp(" ", "switchMode"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	m := model{
		keymap: keymap,
		help: help.New(),
		focus: stopwatch.NewWithInterval(time.Millisecond),
		// rest: stopwatch.New(),
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
				m.quiting = true
				return m, tea.Quit
			case key.Matches(msg, m.keymap.switchMode):
				return m, m.switchMode
		}
	}
	m.focus, cmd = m.focus.Update(msg)
	return m, cmd
}

func (m model) switchMode() tea.Msg {
	m.focus.Toggle()
	m.rest.Toggle()
	return struct{}{}
}
	
// View
func (m model) View() string {
	s := fmt.Sprintf("Focus time: %s\n", m.focus.View())
	s += fmt.Sprintf("Break time: %s\n", m.rest.View())
	return s
}

// Init
func (m model) Init() tea.Cmd {
	// return m.focus.Init()
	return nil
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("It didn't work: ", err)
		os.Exit(1)
	}
}
