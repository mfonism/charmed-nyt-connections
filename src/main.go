package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func main() {
	program := tea.NewProgram(initialModel())
	if _, err := program.Run(); err != nil {
		fmt.Printf("Whoops, something went wrong: %v", err)
		os.Exit(1)
	}
}

type Model struct {
	board [][]string
}

func initialModel() Model {
	return Model{
		board: [][]string{
			{
				"Roc",
				"Ruby",
				"Crystal",
				"Python",
			},
			{
				"Rails",
				"Django",
				"Phoenix",
				"Servant",
			},
			{
				"Elm",
				"Haskell",
				"Agda",
				"Miranda",
			},
			{
				"Zen",
				"Quokka",
				"Raven",
				"Kraken",
			},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderRow(true).
		BorderColumn(true).
		Rows(m.board...).
		StyleFunc(
			func(row, col int) lipgloss.Style {
				return lipgloss.NewStyle().
					Height(3).
					Width(12).
					Align(lipgloss.Center).
					Padding(1)
			},
		)

	return t.Render()
}
