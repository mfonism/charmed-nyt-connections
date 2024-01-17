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

var (
	black                  = lipgloss.Color("#000000")
	mutedBlack             = lipgloss.Color("#161616")
	white                  = lipgloss.Color("#FFFFFF")
	mutedWhite             = lipgloss.Color("#E0E0E0")
	selectedCellBackground = lipgloss.Color("#A9A9A9")
	selectedCellForeground = lipgloss.Color("#DCDCDC")

	shuffleButtonCopy     = "Shuffle"
	deselectAllButtonCopy = "Deselect All"
	submitButtonCopy      = "Submit"
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
	board             [][]string
	selectedTiles     map[string]struct{}
	mistakesRemaining int
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
		mistakesRemaining: 4,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q":
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

						if _, isAlreadySelected := m.selectedTiles[cellData]; isAlreadySelected {
							delete(m.selectedTiles, cellData)
						} else if len(m.selectedTiles) < 4 {
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
	return zone.Scan(
		lipgloss.JoinVertical(
			lipgloss.Center,
			m.viewHeader(),
			m.viewBoard(),
			m.viewMistakesRemaining(),
			m.viewActions(),
		),
	)
}

func (m Model) viewHeader() string {
	return lipgloss.NewStyle().
		Padding(1, 0).
		Render("Create four groups of four!")
}

func (m Model) viewBoard() string {
	cellBaseStyle := lipgloss.NewStyle().
		Height(3).
		Width(12).
		Align(lipgloss.Center, lipgloss.Center)

	readyBoard := make([]string, len(m.board))

	for rowIndex, row := range m.board {
		readyRow := make([]string, len(row))
		for cellIndex, cellData := range row {
			var cellStyle lipgloss.Style
			if _, isSelected := m.selectedTiles[cellData]; isSelected {
				cellStyle = cellBaseStyle.Copy().
					Background(mutedWhite).
					Foreground(mutedBlack)
			} else {
				cellStyle = cellBaseStyle.Copy().
					Background(mutedBlack).
					Foreground(mutedWhite)
			}

			readyRow[cellIndex] = zone.Mark(cellData, cellStyle.Render(cellData))
		}

		readyBoard[rowIndex] = lipgloss.JoinHorizontal(lipgloss.Center, readyRow...)
	}

	return lipgloss.JoinVertical(lipgloss.Center, readyBoard...)
}

func (m Model) viewMistakesRemaining() string {
	return lipgloss.NewStyle().
		Padding(1, 0).
		Render(fmt.Sprintf("Mistakes remaining: %d", m.mistakesRemaining))
}

func (m Model) viewActions() string {
	buttonBaseStyle := lipgloss.NewStyle().
		Width(14).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(mutedWhite).
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(mutedWhite)

	shuffleButton := zone.Mark(
		shuffleButtonCopy,
		buttonBaseStyle.Copy().Render(shuffleButtonCopy),
	)

	deselectAllButton := zone.Mark(
		deselectAllButtonCopy,
		buttonBaseStyle.Copy().Render(deselectAllButtonCopy),
	)

	submitButton := zone.Mark(
		shuffleButtonCopy,
		buttonBaseStyle.Copy().Render(submitButtonCopy),
	)

	return lipgloss.JoinHorizontal(lipgloss.Center, shuffleButton, deselectAllButton, submitButton)
}
