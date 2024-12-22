package keeper

import (
	"github.com/nbvehbq/go-password-keeper/pkg/tuiapp"
	"github.com/rivo/tview"
)

func RenderDashBoardPage() *tview.Flex {
	queryWidget := RenderQueryWidget()
	tableWidget := RenderTable()

	tuiapp.KeeperTui.AddWidget(queryWidget)
	tuiapp.KeeperTui.AddWidget(tableWidget)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryWidget, 0, 1, true).
		AddItem(tableWidget, 0, 4``, false)

	return flex
}
