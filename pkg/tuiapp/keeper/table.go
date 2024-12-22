package keeper

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var TableRecords *tview.Table

func RenderTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(true).
		SetSeparator('|').
		SetFixed(1, 0).
		Select(0, 0)

	table.SetDoneFunc(
		func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				table.SetSelectable(true, false)
			}
		})

	table.SetSelectedFunc(func(row int, column int) {
		for i := 0; i < table.GetRowCount(); i++ {
			table.GetCell(row, i).SetTextColor(tcell.ColorGreen)
		}
		// table.GetCell(row, column).SetTextColor(tcell.ColorGreen)
		table.SetSelectable(false, false)
	})

	table.SetBorder(true).SetTitle("[green]Row list")

	TableRecords = table

	lorem := strings.Split("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.", " ")
	cols, rows := 10, 40
	word := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			color := tcell.ColorWhite
			if r == 0 {
				color = tcell.ColorYellow
			}
			SetCell(r, c, lorem[word], color)
			word = (word + 1) % len(lorem)
		}
	}

	return TableRecords
}

func SetCell(row int, column int, text string, color tcell.Color) {
	cell := tview.NewTableCell(text).
		SetTextColor(color).
		SetAlign(tview.AlignCenter)
	TableRecords.SetCell(row, column, cell)
}

func ClearTableRecords() {
	TableRecords.Clear()
}
