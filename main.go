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

// Model
type Model struct {
	keymap      keymap
	help        help.Model
	focus       stopwatch.Model
	rest        stopwatch.Model
	writing     bool
	editing     bool
	quiting     bool
	todos       []string
	stricken    map[int]struct{}
	cursor      int
	textInput   textinput.Model
	altScreen   bool
	suspending  bool
	suspendTime time.Time
	focusOffset time.Duration
	restOffset  time.Duration
	width       int
}

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
	toggleAltScreen key.Binding
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
			k.toggleAltScreen,
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

// initialModel
func initialModel() Model {

	ti := textinput.New()
	ti.Placeholder = "todo..."
	ti.CharLimit = 256
	ti.Width = 50

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
		toggleAltScreen: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle alt screen"),
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
		todos:     []string{},
		textInput: ti,
		stricken:  make(map[int]struct{}),
		altScreen: true,
		width:     80,
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
		} else {
			return m, m.rest.Stop()
		}
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
			m.textInput.SetValue(m.todos[m.cursor])
			return m, m.textInput.Focus()
		}
	case key.Matches(msg, m.keymap.remove):
		if len(m.todos) > 0 {
			m.todos = append(m.todos[:m.cursor], m.todos[m.cursor+1:]...)

			// All of the keys greater than the cursor in stricken must be
			// decremented by 1
			updatedStricken := make(map[int]struct{})
			for k, v := range m.stricken {
				if k < m.cursor {
					updatedStricken[k] = v
				} else {
					updatedStricken[k-1] = v
				}
			}
			m.stricken = updatedStricken

			if m.cursor > len(m.todos)-1 && m.cursor > 0 {
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

func (m Model) handleWritingMode(msg tea.KeyMsg) (Model, tea.Cmd) {
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
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var keyCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		todoWidth := msg.Width - 10
		if todoWidth < 20 {
			todoWidth = 20
		}
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
		todoWidth := m.width - 10
		if todoWidth < 20 {
			todoWidth = 20
		}

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

			if _, ok := m.stricken[i]; ok {
				todoItemStyle = strikethroughStyle
			}

			s += fmt.Sprintf("%s %s\n", cursorMark, todoItemStyle.Render(todo))
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
	// return m.focus.Init()
	return nil
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Something broke: ", err)
		os.Exit(1)
	}
}
