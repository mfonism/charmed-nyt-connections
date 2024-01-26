package main

import (
	"cmp"
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
	"github.com/mfonism/charmed/connections/internals/sets"
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
	revealButtonCopy      = "Reveal All"

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

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}

		defer f.Close()
	}

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
	mistakesRemaining int
}

func initialModel() Model {
	wordGroups := []WordGroup{
		newWordGroup(
			sets.New("Amazon", "Nile", "Yangtze", "Danube"),
			"Rivers of the world.",
			Green,
		),
		newWordGroup(
			sets.New("Plum", "Apple", "Orange", "Kiwi"),
			"Fruits.",
			Blue,
		),
		newWordGroup(
			sets.New("Basket", "Hand", "Base", "Foot"),
			"___ball",
			Yellow,
		),
		newWordGroup(
			sets.New("MIT", "Apache", "Mozilla", "BSD"),
			"OSI-approved Open Source licenses",
			Purple,
		),
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
			if m.mistakesRemaining > 0 {
				m.shuffleBoard()
			}
			return m, nil
		case " ":
			if m.mistakesRemaining <= 0 && m.wordGroups[len(m.wordGroups)-1].isUnrevealed() {
				m.revealRemaining()
			}
			return m, nil
		case "backspace":
			m.deselectAll()
			return m, nil
		case "enter":
			if m.mistakesRemaining > 0 {
				m.submit()
			}
			return m, nil
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			// cells on board
			if m.mistakesRemaining > 0 {
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
			}

			// shuffle
			if m.mistakesRemaining > 0 && zone.Get(shuffleButtonCopy).InBounds(msg) {
				m.shuffleBoard()
				return m, nil
			}

			// submit
			if m.mistakesRemaining > 0 && zone.Get(submitButtonCopy).InBounds(msg) {
				m.submit()
				return m, nil
			}

			// deselect-all
			if m.selectedTiles.Size() > 0 && zone.Get(deselectAllButtonCopy).InBounds(msg) {
				m.deselectAll()
				return m, nil
			}

			// reveal remaining
			if m.mistakesRemaining <= 0 && m.wordGroups[len(m.wordGroups)-1].isUnrevealed() && zone.Get(revealButtonCopy).InBounds(msg) {
				m.revealRemaining()
			}
			return m, nil
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
	if m.wordGroups[0].isUnrevealed() {
		return ""
	}

	cellBaseStyle := lipgloss.NewStyle().
		Height(2).
		Width(14).
		Bold(true).
		Align(lipgloss.Center, lipgloss.Bottom)

	rows := make([]string, 0)
	for groupIndex := range m.wordGroups {
		group := &m.wordGroups[groupIndex]
		if group.isUnrevealed() {
			// we are guaranteed to always have revealed groups coming first in the
			//  list of word groups
			break
		}

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
		case Yellow:
			rowColor = yellow
		case Green:
			rowColor = green
		case Blue:
			rowColor = blue
		case Purple:
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

		rows = append(rows, rowStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				lipgloss.JoinHorizontal(lipgloss.Center, row...),
				group.clue,
			),
		))
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

	if m.mistakesRemaining <= 0 {
		var revealButtonStyles lipgloss.Style
		if m.wordGroups[len(m.wordGroups)-1].isUnrevealed() {
			revealButtonStyles = enabledButtonStyle.Copy()
		} else {
			revealButtonStyles = disabledButtonStyle.Copy()
		}

		return zone.Mark(
			revealButtonCopy,
			revealButtonStyles.
				Width(60).
				MarginTop(2).
				Padding(0, 12).
				Background(lighterBlack).
				Render("Reveal Remaining"),
		)
	}

	var shuffleButtonStyle, deselectAllButtonStyle, submitButtonStyle lipgloss.Style

	// shuffle button
	if len(m.board) == 0 || m.mistakesRemaining <= 0 {
		shuffleButtonStyle = disabledButtonStyle.Copy()
	} else {
		shuffleButtonStyle = enabledButtonStyle.Copy()
	}

	shuffleButton := zone.Mark(
		shuffleButtonCopy,
		shuffleButtonStyle.Render(shuffleButtonCopy),
	)

	// deselect-all button
	if m.selectedTiles.Size() == 0 || m.mistakesRemaining <= 0 {
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

	// submit button
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

	for groupIndex := range m.wordGroups {
		group := &m.wordGroups[groupIndex]
		if group.isUnrevealed() && group.members.Equals(&m.selectedTiles) {
			group.makeRevealedByPlayer()
			// sort such that earliest revealed groups come first
			// and unrevealed groups come last
			slices.SortStableFunc(m.wordGroups, func(wg1, wg2 WordGroup) int {
				if wg1.isUnrevealed() && wg2.isUnrevealed() {
					return cmp.Compare(wg1.color, wg2.color)
				}

				if wg1.isUnrevealed() {
					return 1
				}

				if wg2.isUnrevealed() {
					return -1
				}

				return cmp.Compare(wg1.unix, wg2.unix)
			})

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

	// leave this out until we can animate it, because, otherwise the user won't get to
	// see that they're wrong before the selection is cleared
	// if m.mistakesRemaining <= 0 {
	// 	m.deselectAll()
	// }
}

func (m *Model) revealRemaining() {
	// log.Println(m)
	for i := len(m.wordGroups) - 1; i >= 0; i-- {
		if m.wordGroups[i].isUnrevealed() {
			m.wordGroups[i].makeRevealedByComputer()
		} else {
			// rely on the guarantee that all unrevealed groups are at the tail of the list
			break
		}
	}

	m.board = [][]string{}
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
