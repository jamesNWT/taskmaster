package main

import "github.com/charmbracelet/bubbles/key"

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

func newKeymap() keymap {

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
	
	return keymap
}
