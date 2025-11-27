package main

import (
	"fmt"
	"time"
	"os"
	// "strings"

	"github.com/charmbracelet/bubbles/textinput"
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
	writing  bool
	editing  bool
	quiting  bool
	todos    []string
	cursor   int
	textInput textinput.Model
}	

/*
Things the user needs to be able to do in normal mode:
- Start the timers once
- Toggle the timers
- Show help
- Enter "writing mode" to start creating a to-do
- Remove a todo
- Strikethrough a todo
- Enter writing mode to start editing a to-do
- Toggle fullscreen
- Scroll through todos with some UI showing which one their cursor is on.
- quit application

Things user needs to be able to in "writing mode"
- Type normally into the textInput
- Submit the textinput to return to normal mode.
- exit writing mode without submiting the textInput
- quit application
*/

// keymap
type keymap struct {
	edit             key.Binding
	create           key.Binding
	remove           key.Binding
	strikeThrough    key.Binding

	firstStart       key.Binding
	switchTimer      key.Binding
	quit             key.Binding
	up               key.Binding
	down             key.Binding
	toggleFullScreen key.Binding
	help             key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.help, k.quit}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.edit, k.create, k.remove, k.strikeThrough},
		{k.firstStart, k.switchTimer, k.help, k.quit},
	}
}

// initialModel
func initialModel() model {

	ti := textinput.New()
	ti.Placeholder = "todo..."
	ti.CharLimit = 256
	ti.Width = 50

	keymap := keymap{
		firstStart: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "start"),
		),
		switchTimer: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "switch stopwatch"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		toggleFullScreen: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fullscreen"),
		),
		help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
		),
		// bindings related to todo list items
		edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		create: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "create"),
		),
		remove: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove"),
		),
		strikeThrough: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "strikeThrough"),
		),
		up: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "down"),
		),
	}

	keymap.switchTimer.SetEnabled(false)

	m := model{
		keymap: keymap,
		help: help.New(),
		focus: stopwatch.NewWithInterval(time.Millisecond),
		rest: stopwatch.NewWithInterval(time.Millisecond),
		todos: []string{},
		textInput: ti,
	}
	
	return m
}
		
func (m model) handleNormalMode(msg tea.KeyMsg) (model, tea.Cmd) {

	switch {
	case key.Matches(msg, m.keymap.quit): 
		m.quiting = true
		return m, tea.Quit
	case key.Matches(msg, m.keymap.firstStart):
		m.keymap.firstStart.SetEnabled(false)
		m.keymap.switchTimer.SetEnabled(true)
		return m, m.focus.Start()
	case key.Matches(msg, m.keymap.switchTimer):
		restCmd := m.rest.Toggle()
		focusCmd := m.focus.Toggle()
		return m, tea.Batch(restCmd, focusCmd)
	case key.Matches(msg, m.keymap.help):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil
	case key.Matches(msg, m.keymap.up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case key.Matches(msg, m.keymap.down):
		if m.cursor < len(m.todos) - 1 {
			m.cursor++
		}
		return m, nil
	case key.Matches(msg, m.keymap.create):
		m.writing = true
		return m, m.textInput.Focus()
	case key.Matches(msg, m.keymap.edit):
		if len(m.todos) > 0 {
			m.writing = true
			m.editing = true
			m.textInput.SetValue(m.todos[m.cursor])
			return m, m.textInput.Focus()
		}
	case key.Matches(msg, m.keymap.remove):
		if len(m.todos) > 0 {
			m.todos = append(m.todos[:m.cursor], m.todos[m.cursor+1:]...)
			if m.cursor > len(m.todos) - 1 && m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m model) handleWritingMode(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.editing {
			m.todos[m.cursor] = m.textInput.Value()
			m.editing = false
		} else {
			m.todos = append(m.todos, m.textInput.Value())
		}
		m.textInput.SetValue("")
		m.writing = false
		return m, nil
	case "esc":
		m.textInput.SetValue("")
		m.writing = false
		return m, nil
	}
	var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)
    return m, cmd
}

// Update
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var keyCmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.writing {
			m, keyCmd = m.handleWritingMode(msg)
		} else { 
			m, keyCmd = m.handleNormalMode(msg)
		}
	}

	var focusCmd, restCmd tea.Cmd 

	m.focus, focusCmd = m.focus.Update(msg)
	m.rest, restCmd = m.rest.Update(msg)
	return m, tea.Batch(focusCmd, restCmd, keyCmd)
}

// View
func (m model) View() string {
	s := fmt.Sprintf("Focus time: %s\n", m.focus.View())
	s += fmt.Sprintf("Break time: %s\n", m.rest.View())
	if len(m.todos) > 0 {
		s += fmt.Sprintf("\nTo do list:\n")
		for i, todo := range(m.todos) {
			cursorMark := " "
			if i == m.cursor {
				cursorMark = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursorMark, todo)
		}
	}
	if m.writing {
		s += fmt.Sprintf(
			"\nEnter a to-do:\n\n%s\n\n%s\n",
			m.textInput.View(),
			"(press Enter to submit, Esc to quit)",
		)
	}
	s +=  "\n" + m.help.View(m.keymap)
	return s
}

// Init
func (m model) Init() tea.Cmd {
	// return m.focus.Init()
	return nil
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Something broke: ", err)
		os.Exit(1)
	}
}
