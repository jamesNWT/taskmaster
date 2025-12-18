package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Todo struct {
	text string
	stricken bool
}

// Model
type Model struct {
	keymap      keymap
	help        help.Model
	focus       stopwatch.Model
	rest        stopwatch.Model
	writing     bool
	editing     bool
	quiting     bool
	todos       []Todo
	cursor      int
	textInput   textinput.Model
	altScreen   bool
	suspending  bool
	suspendTime time.Time
	focusOffset time.Duration
	restOffset  time.Duration
	width       int
}

const (
	todoWidthMargin = 10
	minTodoWidth    = 20
	defaultWidth    = 80
	maxCharLimit    = 256
)

// initialModel
func initialModel() Model {

	ti := textinput.New()
	ti.Placeholder = "todo..."
	ti.CharLimit = maxCharLimit
	ti.Width = defaultWidth

	m := Model{
		keymap:    newKeymap(),
		help:      help.New(),
		focus:     stopwatch.NewWithInterval(time.Millisecond),
		rest:      stopwatch.NewWithInterval(time.Millisecond),
		todos:     []Todo{},
		textInput: ti,
		altScreen: true,
		width:     defaultWidth,
	}

	return m
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {

	switch {
	case key.Matches(msg, m.keymap.quit):
		m.quiting = true
		var altScreenCmd tea.Cmd
		if m.altScreen {
			m.altScreen = false
			altScreenCmd = tea.ExitAltScreen
		}
		return m, tea.Sequence(altScreenCmd, tea.Quit)
	case key.Matches(msg, m.keymap.suspend):
		m.suspending = true
		m.suspendTime = time.Now()
		return m, tea.Suspend
	case key.Matches(msg, m.keymap.firstStart):
		m.keymap.firstStart.SetEnabled(false)
		m.keymap.switchWatch.SetEnabled(true)
		m.keymap.pauseWatches.SetEnabled(true)
		return m, m.focus.Start()
	case key.Matches(msg, m.keymap.pauseWatches):
		m.keymap.firstStart.SetEnabled(true)
		m.keymap.switchWatch.SetEnabled(false)
		m.keymap.pauseWatches.SetEnabled(false)
		if m.focus.Running() {
			return m, m.focus.Stop()
		}
		return m, m.rest.Stop()
	case key.Matches(msg, m.keymap.switchWatch):
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
		if m.cursor < len(m.todos)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keymap.create):
		m.writing = true
		return m, m.textInput.Focus()
	case key.Matches(msg, m.keymap.edit):
		if len(m.todos) > 0 {
			m.writing = true
			m.editing = true
			m.textInput.SetValue(m.todos[m.cursor].text)
			return m, m.textInput.Focus()
		}
	case key.Matches(msg, m.keymap.remove):
		if len(m.todos) > 0 {
			m.todos = append(m.todos[:m.cursor], m.todos[m.cursor+1:]...)

			if m.cursor > len(m.todos)-1 && m.cursor > 0 {
				m.cursor--
			}
		}
	case key.Matches(msg, m.keymap.strikeThrough):
		m.todos[m.cursor].stricken = !m.todos[m.cursor].stricken
	}

	return m, nil
}

func (m Model) handleWritingMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.editing {
			m.todos[m.cursor].text = m.textInput.Value()
			m.editing = false
		} else {
			m.todos = append(m.todos, Todo{m.textInput.Value(), false})
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
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var keyCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		todoWidth := m.getTodoWidth()		
		m.textInput.Width = todoWidth

	case tea.KeyMsg:
		if m.writing {
			m, keyCmd = m.handleWritingMode(msg)
		} else {
			m, keyCmd = m.handleNormalMode(msg)
		}
	case tea.ResumeMsg:
		m.suspending = false
		if !m.suspendTime.IsZero() {
			if m.focus.Running() {
				m.focusOffset += time.Since(m.suspendTime)
			} else if m.rest.Running() {
				m.restOffset += time.Since(m.suspendTime)
			}
		}
		return m, nil
	}

	var focusCmd, restCmd tea.Cmd

	m.focus, focusCmd = m.focus.Update(msg)
	m.rest, restCmd = m.rest.Update(msg)
	return m, tea.Batch(focusCmd, restCmd, keyCmd)
}

// Init
func (m Model) Init() tea.Cmd {
	return nil
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
