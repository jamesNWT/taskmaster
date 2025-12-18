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
	"github.com/charmbracelet/lipgloss"
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

// keymap
type keymap struct {
	edit          key.Binding
	create        key.Binding
	remove        key.Binding
	strikeThrough key.Binding

	firstStart      key.Binding
	switchWatch     key.Binding
	pauseWatches    key.Binding
	quit            key.Binding
	up              key.Binding
	down            key.Binding
	help            key.Binding
	suspend         key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.help, k.quit}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.create, k.edit, k.remove, k.strikeThrough, k.up, k.down},
		{k.firstStart,
			k.switchWatch,
			k.help,
			k.quit,
			k.suspend,
			k.pauseWatches},
	}
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)
)

func (m Model) getTodoWidth() int {
	width := m.width - todoWidthMargin
	if width < minTodoWidth {
		return minTodoWidth
	}
	return width
}

// initialModel
func initialModel() Model {

	ti := textinput.New()
	ti.Placeholder = "todo..."
	ti.CharLimit = maxCharLimit
	ti.Width = defaultWidth

	keymap := keymap{
		firstStart: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "start focus stopwatch"),
		),
		switchWatch: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "switch stopwatch"),
		),
		pauseWatches: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "stop stopwatches"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		suspend: key.NewBinding(
			key.WithKeys("ctrl+z"),
			key.WithHelp(
				"ctrl+z",
				"suspend program (timers will be updated upon resume)"),
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

	keymap.switchWatch.SetEnabled(false)
	keymap.pauseWatches.SetEnabled(false)

	m := Model{
		keymap:    keymap,
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

func formatDuration(d time.Duration) string {
	// Round to tenths of a second
	d = d.Round(100 * time.Millisecond)

	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	tenths := d / (100 * time.Millisecond)

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%d", h, m, s, tenths)
	} else if m > 0 {
		return fmt.Sprintf("%d:%02d.%d", m, s, tenths)
	} else {
		return fmt.Sprintf("%d.%d", s, tenths)
	}
}

// View
func (m Model) View() string {
	s := headerStyle.Render("Focus time:") + " " + formatDuration(m.focus.Elapsed()+m.focusOffset) + "\n"
	s += headerStyle.Render("Break time:") +" " + formatDuration(m.rest.Elapsed()+m.restOffset) + "\n"

	if len(m.todos) > 0 {
		todoWidth := m.getTodoWidth()
		todoStyle := lipgloss.NewStyle().Width(todoWidth)
		strikethroughStyle := lipgloss.NewStyle().
			Width(todoWidth).
			Strikethrough(true).
			Foreground(lipgloss.Color("240"))

		s += "\n" + headerStyle.Render("To do list:") + "\n"
		for i, todo := range m.todos {
			cursorMark := " "
			if i == m.cursor {
				cursorMark = cursorStyle.Render(">")
			}

			todoItemStyle := todoStyle

			if todo.stricken {
				todoItemStyle = strikethroughStyle
			}

			s += fmt.Sprintf("%s %s\n", cursorMark, todoItemStyle.Render(todo.text))
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
		s += "\n" + m.help.View(m.keymap)
	}
	return s
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
