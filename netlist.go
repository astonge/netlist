package main

import (
	"fmt"
	"strings"
	"time"
	"strconv"
	"os"
	"github.com/bastjan/netstat"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func formatPort(port int) string {
	if port == 0 {
		return "*"
	}
	return strconv.Itoa(port)
}

func emptyCells(slots int) []table.Row {
	results := make([]table.Row, 0, slots)
	for i := 0; i <= slots; i++ {
		results = append(results, []string{
			"---",
			"---.---.---.---",
			"---.---.---.---",
			"<---->",
		})
	}
	return results
}

func getConnections() []table.Row {
	connections, _ := netstat.TCP.Connections()
	results := make([]table.Row, 0, len(connections))
	for _, conn := range connections {
		if conn.State.String() == stateMode || stateMode == "ANY" {
			results = append(results, []string{
				fmt.Sprintf("%s", strings.ToUpper(conn.Protocol.Name)),
				fmt.Sprintf("%s:%s", conn.IP, formatPort(conn.Port)),
				fmt.Sprintf("%s:%s", conn.RemoteIP, formatPort(conn.RemotePort)),
				conn.State.String(),
			})
		}
	}
	results = append(results, emptyCells(25)...)

	return results
}

var stateMode = "LISTEN"
var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#CE6B66"))

type model struct {
	table table.Model
}

type TickMsg time.Time

// Send a message every second.
func tickEvery() tea.Cmd {
    return tea.Every(time.Millisecond * 5, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

func (m model) Init() tea.Cmd { return tickEvery() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case TickMsg:
		m.table.SetRows(getConnections())
		return m, tickEvery()
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			stateMode = "LISTEN"
		case "2":
			stateMode = "ESTABLISHED"
		case "3":
			stateMode = "ANY"
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func main() {
	columns := []table.Column{
		{Title: "Proto", Width: 10},
		{Title: "Local", Width: 25},
		{Title: "Remote", Width: 25},
		{Title: "State", Width: 25},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#CE6B66")).
		BorderBottom(true).
		Bold(true)
	
	s.Selected = s.Selected.Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#72EAEA")).Bold(true)
	t.SetStyles(s)

	m := model{t}	
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}