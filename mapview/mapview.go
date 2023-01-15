package mapview

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eliukblau/pixterm/pkg/ansimage"
	sm "github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"
)

type MapRender string
type MapCoordinates struct {
	Lat float64
	Lng float64
	Err error
}

type NominatimResponse []struct {
	PlaceID     int    `json:"place_id"`
	License     string `json:"license"`
	OSMType     string `json:"osm_type"`
	OSMID       int    `json:"osm_id"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

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
	loc       string
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
	m.loc = ""
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

func (m *Model) SetLocation(loc string, zoom int) {
	m.loc = loc
	m.zoom = zoom
	m.applyToOSM()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	if !m.initialized {
		m.setInitialValues()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		var hit = false
		switch {

		case key.Matches(msg, m.KeyMap.Up):
			m.lat += 0.05
			hit = true

		case key.Matches(msg, m.KeyMap.Right):
			m.lng += 0.05
			hit = true

		case key.Matches(msg, m.KeyMap.Down):
			m.lat -= 0.05
			hit = true

		case key.Matches(msg, m.KeyMap.Left):
			m.lng -= 0.05
			hit = true

		case key.Matches(msg, m.KeyMap.ZoomIn):
			m.zoom += 1
			hit = true

		case key.Matches(msg, m.KeyMap.ZoomOut):
			m.zoom -= 1
			hit = true

		}
		if hit {
			m.applyToOSM()
			cmds = append(cmds, m.render(m.Width, m.Height))
			return m, tea.Batch(cmds...)
		}

	case MapRender:
		m.maprender = string(msg)
		return m, nil

	case MapCoordinates:
		if msg.Err != nil {
			m.maprender = msg.Err.Error()
		} else {
			m.lat = msg.Lat
			m.lng = msg.Lng
			m.applyToOSM()
		}
		return m, m.render(m.Width, m.Height)

	}

	if m.initialized && m.loc != "" {
		cmds = append(cmds, m.lookup(m.loc))
		m.loc = ""
		return m, tea.Batch(cmds...)
	}

	if m.initialized && m.loc == "" && m.maprender == "" {
		cmds = append(cmds, m.render(m.Width, m.Height))
	}
	return m, tea.Batch(cmds...)
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

func (m *Model) lookup(address string) tea.Cmd {
	return func() tea.Msg {
		u := fmt.Sprintf(
			"https://nominatim.openstreetmap.org/search?q=%s&format=json&polygon=1&addressdetails=1",
			url.QueryEscape(address),
		)

		resp, err := http.Get(u)
		if err != nil {
			return MapCoordinates{Err: err}
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return MapCoordinates{Err: errors.New(string(body))}
		}

		var data NominatimResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return MapCoordinates{Err: err}
		}

		if len(data) == 0 {
			return MapCoordinates{Err: errors.New("Location not found")}
		}

		lat, err := strconv.ParseFloat(data[0].Lat, 64)
		if err != nil {
			return MapCoordinates{Err: err}
		}
		lon, err := strconv.ParseFloat(data[0].Lon, 64)
		if err != nil {
			return MapCoordinates{Err: err}
		}

		return MapCoordinates{
			Lat: lat,
			Lng: lon,
		}
	}
}

func (m Model) View() string {
	return m.maprender
}
