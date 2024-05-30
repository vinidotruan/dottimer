package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type model struct {
  focusIndex int 
  inputs []textinput.Model
  cursorMode cursor.Mode
}

func initialModel() model {
  m := model {
    inputs: make([]textinput.Model, 1),
  } 

  var t textinput.Model
  for i := range m.inputs {
    t = textinput.New()
    t.Cursor.Style = cursorStyle
    t.CharLimit = 2

    switch i {
    case 0: 
      t.Placeholder = "Tempo em minutos"
      t.Focus()
      t.PromptStyle = focusedStyle
      t.TextStyle = focusedStyle
    }

    m.inputs[i] = t
  }

  return m
}

func (m model) Init() tea.Cmd {
  return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "esc":
      return m, tea.Quit
   
    case "tab", "shift+tab", "enter", "up", "down":
    s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
    }
  } 

  cmd := m.updateInputs(msg)

  return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
  cmds := make([]tea.Cmd, len(m.inputs))

  for i := range m.inputs {
    m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
  }

  return tea.Batch(cmds...)
}

func (m model) View() string {
  var builder strings.Builder

  for index := range m.inputs {
    builder.WriteString(m.inputs[index].View())
    if index < len(m.inputs)-1 {
      builder.WriteRune('\n')
    }
  }

  button := &blurredButton

  if m.focusIndex == len(m.inputs) {
    button = &focusedButton
  }
  
  fmt.Fprintf(&builder, "\n\n%s\n\n", *button)

  return builder.String()

}

func main() {
  if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}