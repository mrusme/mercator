package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrusme/mercator/mapview"
)

type model struct {
	mv mapview.Model
}

type tickMsg time.Time

func main() {
	args := os.Args[1:]

	m := NewModel()

	var isLatLng bool = false
	if len(args) == 2 {
		lat, err1 := strconv.ParseFloat(args[0], 64)
		lng, err2 := strconv.ParseFloat(args[1], 64)
		if err1 == nil && err2 == nil {
			isLatLng = true
			m.mv.SetLatLng(lat, lng, 15)
		}
	}
	if len(args) < 2 || len(args) > 2 && !isLatLng {
		m.mv.SetLocation(strings.Join(args, " "), 15)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func NewModel() model {
	m := model{}
	m.mv = mapview.New(80, 24)
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

	case tea.WindowSizeMsg:
		m.mv.Width = mt.Width
		m.mv.Height = mt.Height
		return m, nil

	}

	var cmd tea.Cmd
	m.mv, cmd = m.mv.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.mv.View()
}
