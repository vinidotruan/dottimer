package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
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
	mainStyle           = lipgloss.NewStyle().MarginLeft(2)
	focusedButton       = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton       = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
	timeoutMinutes      = 0
	timeoutSeconds      = 0
)

type model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	Quitting   bool
	Timer      bool
}

func initialModel() model {
	m := model{
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
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	if !m.Timer {
		return updateInputViewKeyHandler(msg, m)
	}
	return updateTimer(msg, m)

}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var s string
	if m.Quitting {
		return "\nTe vejo em breve.\n\n"
	}

	if !m.Timer {
		s = inputView(m)
	} else {
		s = timerView(m)
	}

	return mainStyle.Render("\n" + s + "\n\n")
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}

func updateInputViewKeyHandler(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		return updateTimer(msg, m)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				if i, err := strconv.ParseInt(m.inputs[0].Value(), 10, 64); err == nil {
					timeoutMinutes = int(i)
					timeoutSeconds = 0
				}

				// quando o usuario pressionar enter eu preciso mudar a view.
				m.Timer = true
				return m, tea.Tick(time.Second, func(_ time.Time) tea.Msg {
					return timer.TickMsg{ID: int(time.Now().Unix()), Timeout: timeoutSeconds != 0 && timeoutSeconds != 0}
				})
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

var lastTime time.Time

func updateTimer(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	if lastTime.IsZero() {
		lastTime = time.Now()
	}

	now := time.Now()
	elapsedTime := now.Sub(lastTime).Seconds()
	lastTime = now

	if elapsedTime >= 1 {
		timeoutSeconds -= int(elapsedTime)
		if timeoutSeconds < 0 {
			timeoutMinutes -= 1
			timeoutSeconds = 59
		}
	}
	return m, tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return timer.TickMsg{ID: int(time.Now().Unix()), Timeout: timeoutSeconds != 0 && timeoutSeconds != 0}
	})
}

func inputView(m model) string {
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

func timerView(m model) string {
	currentTimer := fmt.Sprintf("%02d:%02d", timeoutMinutes, timeoutSeconds)
	os.WriteFile("sprint.txt", []byte(currentTimer), 0644)
	return currentTimer
}
