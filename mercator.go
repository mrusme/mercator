package main

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrusme/mercator/mapview"
)

type model struct {
	mv mapview.Model
}

type tickMsg time.Time

func main() {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func NewModel() model {
	m := model{}
	m.mv = mapview.New(200, 120)
	m.mv.SetLocation("New York", 15)
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mt := msg.(type) {
	case tea.KeyMsg:
		switch mt.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.mv, cmd = m.mv.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.mv.View()
}
