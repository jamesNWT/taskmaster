package main

import (
	"fmt"
	"time"
	"github.com/charmbracelet/lipgloss"
)


// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)
)

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

func (m Model) getTodoWidth() int {
	width := m.width - todoWidthMargin
	if width < minTodoWidth {
		return minTodoWidth
	}
	return width
}
