package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			errorMessage := fmt.Sprintf("Error reading file: %s\n%s", err, debug.Stack())
			log.Output(1, errorMessage)
		}
	}()

	// initialize global manager for BubbleZone
	zone.NewGlobal()

	program := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
	if _, err := program.Run(); err != nil {
		fmt.Printf("Whoops, something went wrong: %v", err)
		os.Exit(1)
	}
}

type Model struct {
	board         [][]string
	selectedTiles map[string]struct{}
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
	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft &&
			msg.Action == tea.MouseActionPress {

			for _, row := range m.board {
				for _, cellData := range row {
					if zone.Get(cellData).InBounds(msg) {
						if m.selectedTiles == nil {
							m.selectedTiles = make(map[string]struct{})
						}

						if len(m.selectedTiles) <= 3 {
							m.selectedTiles[cellData] = struct{}{}
						}
					}
				}
			}

			return m, nil
		}

	}
	return m, nil
}

func (m Model) View() string {
	cellBaseStyle := lipgloss.NewStyle().
		Height(3).
		Width(12).
		Align(lipgloss.Center).
		Padding(1)

	readyBoard := make([]string, len(m.board))
	for rowIndex, row := range m.board{
		readyRow := make([]string, len(row))
		for cellIndex, cellData := range row {
			readyRow[cellIndex] = zone.Mark(cellData, cellBaseStyle.Copy().Render(cellData))
			_ = rowIndex
		}

		readyBoard[rowIndex] = lipgloss.JoinHorizontal(lipgloss.Center, readyRow...)
	}

	return zone.Scan(lipgloss.JoinVertical(lipgloss.Center, readyBoard...))
}
