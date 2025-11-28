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
	keymap    keymap
	help      help.Model
	focus     stopwatch.Model
	rest      stopwatch.Model
	writing   bool
	editing   bool
	quiting   bool
	todos     []string
	stricken  map[int]struct{}
	cursor    int
	textInput textinput.Model
	altScreen bool
}	

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
	toggleAltScreen  key.Binding
	help             key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.help, k.quit}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.create, k.edit, k.remove, k.strikeThrough, k.up, k.down},
		{k.firstStart, k.switchTimer, k.toggleAltScreen, k.help, k.quit},
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
			key.WithHelp("space", "start timer"),
		),
		switchTimer: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "switch stopwatch"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		toggleAltScreen: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle alt screen"),
		),
		help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
		),
		// bindings related to todo list items
		edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit a todo"),
		),
		create: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "create a todo"),
		),
		remove: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove a todo"),
		),
		strikeThrough: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "strike a todo "),
		),
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "move up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "move down"),
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
		stricken: make(map[int]struct{}), 
		altScreen: true,
	}
	
	return m
}
		
func (m model) handleNormalMode(msg tea.KeyMsg) (model, tea.Cmd) {

	switch {
	case key.Matches(msg, m.keymap.quit): 
		m.quiting = true
		var altScreenCmd tea.Cmd
		if m.altScreen {
			m.altScreen = false
			altScreenCmd = tea.ExitAltScreen
		}
		return m, tea.Sequence(altScreenCmd, tea.Quit)
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
	case key.Matches(msg, m.keymap.up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keymap.down):
		if m.cursor < len(m.todos) - 1 {
			m.cursor++
		}
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
			
			// All of the keys greater than the cursor in stricken must be
			// decremented by 1
			updatedStricken := make(map[int]struct{})
			for k, v := range(m.stricken) {
				if k < m.cursor {
					updatedStricken[k] = v
				} else {
					updatedStricken[k-1] = v
				}
			}
			m.stricken = updatedStricken
			
			if m.cursor > len(m.todos) - 1 && m.cursor > 0 {
				m.cursor--
			}
		}
	case key.Matches(msg, m.keymap.strikeThrough):
		
		_, ok := m.stricken[m.cursor]
		if ok {
			delete(m.stricken, m.cursor)
		} else {
			m.stricken[m.cursor] = struct{}{}
		}
	case key.Matches(msg, m.keymap.toggleAltScreen):
		if m.altScreen {
			m.altScreen = false
			return m, tea.ExitAltScreen
		} else {
			m.altScreen = true
			return m, tea.EnterAltScreen
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
			
			checked := " " // not selected
			if _, ok := m.stricken[i]; ok {
				checked = "x" // selected
			}
			s += fmt.Sprintf("%s%s %s\n", cursorMark, checked, todo)
		}
	}
	if m.writing {
		s += fmt.Sprintf(
			"\nEnter a to-do:\n\n%s\n\n%s\n",
			m.textInput.View(),
			"(press Enter to submit, Esc to quit)",
		)
	}
	if !m.quiting {
		s +=  "\n" + m.help.View(m.keymap)
	}
	return s
}

// Init
func (m model) Init() tea.Cmd {
	// return m.focus.Init()
	return nil
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); 
	err != nil {
		fmt.Println("Something broke: ", err)
		os.Exit(1)
	}
}
