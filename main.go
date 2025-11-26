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
	firstStart key.Binding
	switchMode key.Binding
	quit  key.Binding
}

// initialModel

func initialModel() model {
	keymap := keymap{
		firstStart: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp(" ", "start"),
		),
		switchMode: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp(" ", "switchMode"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	keymap.switchMode.SetEnabled(false)

	m := model{
		keymap: keymap,
		help: help.New(),
		focus: stopwatch.NewWithInterval(time.Millisecond),
		rest: stopwatch.NewWithInterval(time.Millisecond),
	}
	
	return m
}
		

// Update
func (m model) Update(msg tea.Msg) (tea.Model,tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
			case key.Matches(msg, m.keymap.quit): 
				m.quiting = true
				return m, tea.Quit
			case key.Matches(msg, m.keymap.firstStart):
				m.keymap.firstStart.SetEnabled(false)
				m.keymap.switchMode.SetEnabled(true)
				return m, m.focus.Start()
			case key.Matches(msg, m.keymap.switchMode):
				restCmd := m.rest.Toggle()
				focusCmd := m.focus.Toggle()
				return m, tea.Batch(restCmd, focusCmd)
		}
	}
	var focusCmd, restCmd tea.Cmd
	m.focus, focusCmd = m.focus.Update(msg)
	m.rest, restCmd = m.rest.Update(msg)
	return m, tea.Batch(focusCmd, restCmd)
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
