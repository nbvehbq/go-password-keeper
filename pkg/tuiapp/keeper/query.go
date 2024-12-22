package keeper

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	resourceType = []string{"All", "LoginPassword", "File", "BankCard"}
)

func RenderQueryWidget() *tview.Flex {
	form := tview.
		NewForm().
		AddDropDown("Select type:", resourceType, 0, nil).
		SetFieldBackgroundColor(tcell.ColorGray)

	form.
		AddButton("Query", QueryCallback(form)).
		SetButtonTextColor(tcell.ColorLightGoldenrodYellow)

	queryWidget := tview.
		NewFlex().
		AddItem(form, 0, 1, true)

	queryWidget.SetBorder(true).SetTitle("[green]Query")

	return queryWidget
}

func QueryCallback(form *tview.Form) func() {
	return func() {
		// TODO: query database
	}
}
