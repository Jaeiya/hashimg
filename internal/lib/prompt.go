package lib

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func getModel() model {
	return model{
		choices:  []string{"Image Rename", "Test 2", "Test 3", "Test 4", "Test 5"},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}

		case "right", "l":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	s := "Select what to rename\n\n"

	s += fmt.Sprintf("%s %s\n", ">", m.choices[m.cursor])
	return s
}
