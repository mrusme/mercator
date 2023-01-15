package mapview

import (
	"image/color"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eliukblau/pixterm/pkg/ansimage"
	sm "github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"
)

type MapRender string

type KeyMap struct {
	Up      key.Binding
	Right   key.Binding
	Down    key.Binding
	Left    key.Binding
	ZoomIn  key.Binding
	ZoomOut key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("↑/l", "right"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("↑/h", "left"),
		),
		ZoomIn: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "plus"),
		),
		ZoomOut: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "minus"),
		),
	}
}

type Model struct {
	Width  int
	Height int
	KeyMap KeyMap

	Style lipgloss.Style

	initialized bool

	osm       *sm.Context
	lat       float64
	lng       float64
	zoom      int
	maprender string
}

func New(width, height int) (m Model) {
	m.Width = width
	m.Height = height
	m.setInitialValues()
	return m
}

func (m *Model) setInitialValues() {
	m.KeyMap = DefaultKeyMap()
	m.osm = sm.NewContext()
	m.osm.SetSize(400, 400)
	m.zoom = 15
	m.lat = 25.0782266
	m.lng = -77.3383438
	m.applyToOSM()
	m.initialized = true
}

func (m *Model) applyToOSM() {
	m.osm.SetCenter(s2.LatLngFromDegrees(m.lat, m.lng))
	m.osm.SetZoom(m.zoom)
}

func (m *Model) SetLatLng(lat float64, lng float64, zoom int) {
	m.lat = lat
	m.lng = lng
	m.zoom = zoom
	m.applyToOSM()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd = nil

	if !m.initialized {
		m.setInitialValues()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, m.KeyMap.Up):
			m.lat += 0.05

		case key.Matches(msg, m.KeyMap.Right):
			m.lng += 0.05

		case key.Matches(msg, m.KeyMap.Down):
			m.lat -= 0.05

		case key.Matches(msg, m.KeyMap.Left):
			m.lng -= 0.05

		case key.Matches(msg, m.KeyMap.ZoomIn):
			m.zoom += 1

		case key.Matches(msg, m.KeyMap.ZoomOut):
			m.zoom -= 1

		}
		m.applyToOSM()
		cmd = m.render(m.Width, m.Height)
		return m, cmd

	case MapRender:
		m.maprender = string(msg)
		return m, nil
	}

	if m.initialized && m.maprender == "" {
		cmd = m.render(m.Width, m.Height)
	}
	return m, cmd
}

func (m *Model) render(width, height int) tea.Cmd {
	return func() tea.Msg {
		img, err := m.osm.Render()
		if err != nil {
			return MapRender(err.Error())
		}

		ascii, err := ansimage.NewScaledFromImage(
			img,
			height,
			width,
			color.Transparent,
			ansimage.ScaleModeFill,
			ansimage.NoDithering,
		)
		if err != nil {
			return MapRender(err.Error())
		}

		return MapRender(ascii.RenderExt(false, false))
	}
}

func (m Model) View() string {
	return m.maprender
}
