package main

import (
	"fmt"
	"hash/maphash"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/debug"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mfonism/charmed/connections/src/sets"
)

var (
	black                         = lipgloss.Color("#000000")
	mutedBlack                    = lipgloss.Color("#161616")
	lighterBlack                  = lipgloss.Color("#202020")
	white                         = lipgloss.Color("#FFFFFF")
	mutedWhite                    = lipgloss.Color("#E0E0E0")
	disabledGrey                  = lipgloss.Color("#363636")
	disabledGreyForeground        = lipgloss.Color("#222222")
	selectedCellBackground        = lipgloss.Color("#A9A9A9")
	selectedCellForeground        = lipgloss.Color("#DCDCDC")
	alreadySelectedCellForeground = lipgloss.Color("#F60D94")
	alreadySelectedCellBackground = lipgloss.Color("#F8CCE6")

	shuffleButtonCopy     = "Shuffle"
	deselectAllButtonCopy = "Deselect All"
	submitButtonCopy      = "Submit"

	yellow = lipgloss.Color("#F9DF6D")
	green  = lipgloss.Color("#A0C35A")
	blue   = lipgloss.Color("#B0C4EF")
	purple = lipgloss.Color("#BA81C5")
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
	wordGroups        []WordGroup
	board             [][]string
	selectedTiles     sets.Set[string]
	selectionHistory  []sets.Set[string]
	revealedGroups    []WordGroup
	mistakesRemaining int
}

type WordGroup struct {
	clue    string
	members sets.Set[string]
	color   string
}

func initialModel() Model {
	wordGroups := []WordGroup{
		{
			clue:    "Rivers of the world.",
			members: sets.New("Amazon", "Nile", "Yangtze", "Danube"),
			color:   "green",
		},
		{
			clue:    "Fruits.",
			members: sets.New("Star", "Apple", "Orange", "Kiwi"),
			color:   "blue",
		},
		{
			clue:    "___ball",
			members: sets.New("Basket", "Hand", "Base", "Foot"),
			color:   "yellow",
		},
		{
			clue:    "OSI-approved Open Source licenses",
			members: sets.New("MIT", "Apache", "Mozilla", "BSD"),
			color:   "purple",
		},
	}
	selectedTiles := sets.Empty[string]()

	board := make([][]string, len(wordGroups))
	for groupIndex, group := range wordGroups {
		board[groupIndex] = make([]string, 0, group.members.Size())
		group.members.ForEach(func(word string) {
			board[groupIndex] = append(board[groupIndex], word)
		})
	}

	m := Model{
		wordGroups:        wordGroups,
		selectedTiles:     selectedTiles,
		board:             board,
		mistakesRemaining: 4,
	}
	m.shuffleBoard()

	return m
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
		case "h", "H":
			m.shuffleBoard()
			return m, nil
		case "backspace":
			m.deselectAll()
			return m, nil
		case "enter":
			m.submit()
			return m, nil
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			for _, row := range m.board {
				for _, cellData := range row {
					if zone.Get(cellData).InBounds(msg) {
						if m.selectedTiles.Contains(cellData) {
							m.selectedTiles.Remove(cellData)
						} else if m.selectedTiles.Size() < 4 {
							m.selectedTiles.Add(cellData)
						}

						return m, nil
					}
				}
			}

			if zone.Get(shuffleButtonCopy).InBounds(msg) {
				m.shuffleBoard()
				return m, nil
			}

			if zone.Get(deselectAllButtonCopy).InBounds(msg) {
				m.deselectAll()
				return m, nil
			}

			if zone.Get(submitButtonCopy).InBounds(msg) {
				m.submit()
				return m, nil
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	return zone.Scan(
		lipgloss.NewStyle().
			Padding(2, 6, 3).
			Background(lighterBlack).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Center,
					m.viewHeader(),
					m.viewRevealedGroups(),
					m.viewBoard(),
					m.viewMistakesRemaining(),
					m.viewActions(),
				),
			),
	)
}

func (m Model) viewHeader() string {
	return lipgloss.NewStyle().
		Width(62).
		MarginTop(2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Create four groups of four!")
}

func (m Model) viewRevealedGroups() string {
	if len(m.revealedGroups) == 0 {
		return ""
	}

	cellBaseStyle := lipgloss.NewStyle().
		Height(2).
		Width(14).
		Bold(true).
		Align(lipgloss.Center, lipgloss.Bottom)

	rows := make([]string, len(m.revealedGroups))
	for groupIndex, group := range m.revealedGroups {
		row := make([]string, 0, group.members.Size())

		// sort the revealed words before adding them to display to ensure they're
		// shown in the same order with every screen refresh
		revealedWords := make([]string, 0, group.members.Size())
		group.members.ForEach(func(word string) {
			revealedWords = append(revealedWords, word)
		})
		slices.Sort(revealedWords)

		var rowColor lipgloss.Color
		switch group.color {
		case "yellow":
			rowColor = yellow
		case "green":
			rowColor = green
		case "blue":
			rowColor = blue
		case "purple":
			rowColor = purple
		}

		for _, revealedData := range revealedWords {
			row = append(
				row,
				cellBaseStyle.Copy().
					Background(rowColor).
					Foreground(black).
					Render(revealedData),
			)
		}

		var rowStyle = lipgloss.NewStyle().
			Background(rowColor).
			Padding(0, 3, 1)
		if groupIndex != 0 {
			rowStyle = rowStyle.Copy().MarginTop(1)
		}

		rows[groupIndex] = rowStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				lipgloss.JoinHorizontal(lipgloss.Center, row...),
				group.clue,
			),
		)
	}

	return lipgloss.NewStyle().
		MarginTop(2).
		Render(lipgloss.JoinVertical(lipgloss.Center, rows...))
}

func (m Model) viewBoard() string {
	if len(m.board) == 0 {
		return ""
	}

	cellBaseStyle := lipgloss.NewStyle().
		Height(3).
		Width(14).
		MarginBackground(lighterBlack).
		Align(lipgloss.Center, lipgloss.Center)

	cellMarginTopVal := 1
	cellMarginLeftVal := 2

	selectionIsCompleteButAlreadySeen := m.selectedTiles.Size() == len(m.board[0]) && m.selectedTilesInHistory()

	readyBoard := make([]string, len(m.board))

	for rowIndex, row := range m.board {
		readyRow := make([]string, len(row))
		for cellIndex, cellData := range row {

			cellStyle := cellBaseStyle.Copy()
			// every cell that's not on the topmost row should have a top margin
			if rowIndex != 0 {
				cellStyle.MarginTop(cellMarginTopVal)
			}

			// every cell that's not on the leftmost column should have a left margin
			if cellIndex != 0 {
				cellStyle.MarginLeft(cellMarginLeftVal)
			}

			if m.selectedTiles.Contains(cellData) {
				if selectionIsCompleteButAlreadySeen {
					cellStyle.Background(alreadySelectedCellBackground).Foreground(alreadySelectedCellForeground)
				} else {
					cellStyle.Background(mutedWhite).Foreground(mutedBlack)
				}
			} else {
				cellStyle.Background(mutedBlack).Foreground(mutedWhite)
			}

			readyRow[cellIndex] = zone.Mark(cellData, cellStyle.Render(cellData))
		}

		readyBoard[rowIndex] = lipgloss.JoinHorizontal(lipgloss.Center, readyRow...)
	}

	return lipgloss.NewStyle().
		MarginTop(2).
		Render(lipgloss.JoinVertical(lipgloss.Center, readyBoard...))
}

func (m Model) viewMistakesRemaining() string {
	return lipgloss.NewStyle().
		MarginTop(2).
		Align(lipgloss.Center, lipgloss.Bottom).
		Render(fmt.Sprintf("Mistakes remaining: %d", m.mistakesRemaining))
}

func (m Model) viewActions() string {
	buttonBaseStyle := lipgloss.NewStyle().
		Width(14).
		Align(lipgloss.Center, lipgloss.Center).
		BorderBackground(lighterBlack)

	enabledButtonStyle := buttonBaseStyle.Copy().
		Background(lighterBlack).
		Foreground(mutedWhite).
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(mutedWhite)

	disabledButtonStyle := buttonBaseStyle.Copy().
		Background(lighterBlack).
		Foreground(disabledGreyForeground).
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(disabledGrey)

	var shuffleButtonStyle, deselectAllButtonStyle, submitButtonStyle lipgloss.Style

	if len(m.board) == 0 {
		shuffleButtonStyle = disabledButtonStyle.Copy()
	} else {
		shuffleButtonStyle = enabledButtonStyle.Copy()
	}

	shuffleButton := zone.Mark(
		shuffleButtonCopy,
		shuffleButtonStyle.Render(shuffleButtonCopy),
	)

	if m.selectedTiles.Size() == 0 {
		deselectAllButtonStyle = disabledButtonStyle.Copy()
	} else {
		deselectAllButtonStyle = enabledButtonStyle.Copy()
	}

	deselectAllButton := zone.Mark(
		deselectAllButtonCopy,
		deselectAllButtonStyle.
			Margin(0, 2).
			MarginBackground(lighterBlack).
			Render(deselectAllButtonCopy),
	)

	if !m.canSubmit() {
		submitButtonStyle = disabledButtonStyle.Copy()
	} else {
		submitButtonStyle = enabledButtonStyle.Copy()
	}

	submitButton := zone.Mark(
		shuffleButtonCopy,
		submitButtonStyle.Render(submitButtonCopy),
	)

	return lipgloss.NewStyle().
		MarginTop(2).
		Padding(0, 5).
		Background(lighterBlack).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, shuffleButton, deselectAllButton, submitButton))
}

func (m *Model) shuffleBoard() {
	flattened := flatten(m.board)
	shuffle(flattened)
	m.board = unflatten(flattened, len(m.board))
}

func (m *Model) deselectAll() {
	m.selectedTiles.Clear()
}

func (m *Model) submit() {
	if m.canSubmit() {
		m.doSubmit()
	}
}

func (m Model) canSubmit() bool {
	// can submit only if all the following hold
	// * has chances left
	// * has stuff left on the board
	// * has made enough selections on the board
	// * has NOT made the same selections before
	return m.mistakesRemaining > 0 &&
		len(m.board) > 0 &&
		m.selectedTiles.Size() == len(m.board[0]) &&
		!m.selectedTilesInHistory()
}

func (m Model) selectedTilesInHistory() bool {
	for _, seenSelection := range m.selectionHistory {
		if m.selectedTiles.Equals(&seenSelection) {
			return true
		}
	}

	return false
}

func (m *Model) doSubmit() {
	m.selectionHistory = append(m.selectionHistory, m.selectedTiles.Copy())

	for _, group := range m.wordGroups {
		if group.members.Equals(&m.selectedTiles) {
			m.revealedGroups = append(m.revealedGroups, group)

			if len(m.board) <= 1 {
				m.board = [][]string{}
				m.deselectAll()
				return
			}

			// remove selected items from board
			flattened := flatten(m.board)
			flattened = slices.DeleteFunc(flattened, func(data string) bool {
				return m.selectedTiles.Contains(data)
			})

			m.board = unflatten(flattened, len(m.board)-1)

			m.deselectAll()
			return
		}
	}

	m.mistakesRemaining -= 1
}

func flatten(matrix [][]string) []string {
	if len(matrix) == 0 {
		return []string{}
	} else if len(matrix) == 1 {
		return append([]string(nil), matrix[0]...)
	}

	flattened := make([]string, len(matrix)*len(matrix[0]))

	flatIndex := 0
	for _, row := range matrix {
		for _, cellData := range row {
			flattened[flatIndex] = cellData
			flatIndex += 1
		}
	}

	return flattened
}

func shuffle(slice []string) {
	generator := rand.New(rand.NewSource(int64(new(maphash.Hash).Sum64())))
	generator.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

func unflatten(slice []string, numRows int) [][]string {
	if numRows <= 0 {
		return [][]string{}
	}

	numCols := int(math.Trunc(float64(len(slice) / numRows)))
	matrix := make([][]string, numRows)

	flatIndex := 0
	for rowIndex := range matrix {
		matrix[rowIndex] = make([]string, numCols)
		for cellIndex := range matrix[rowIndex] {
			matrix[rowIndex][cellIndex] = slice[flatIndex]
			flatIndex += 1
		}
	}

	return matrix
}
