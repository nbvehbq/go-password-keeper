package keeper

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/nbvehbq/go-password-keeper/pkg/tuiapp"
)

var LoginErrOut *tview.TextView

func RenderLoginPage() *tview.Flex {
	form := renderLoginForm()
	textView := renderLoginErrTextView()

	flex := tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(false).SetTitle(""), 0, 2, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(false).SetTitle(""), 0, 1, false).
			AddItem(form, 9, 3, true).
			AddItem(textView, 0, 1, false), 40, 3, true).
		AddItem(tview.NewBox().SetBorder(false).SetTitle(""), 0, 2, false)

	flex.SetBorder(true)

	// tuiapp.KeeperTui.ShowPage("login")

	return flex
}

func renderLoginForm() *tview.Form {
	// TODO: load config?

	form := tview.NewForm().
		AddInputField("   Login:", "", 20, nil, nil).
		AddInputField("Password:", "", 20, nil, nil).
		SetFieldBackgroundColor(tcell.ColorGray)

	form.AddButton("Login", LoginCallback(form)).
		AddButton("Exit", ExitCallback()).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorGray).
		SetButtonTextColor(tcell.ColorLightGoldenrodYellow)

	form.SetBorder(true).SetBorderColor(tcell.ColorWhite)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// log.Println("event: ", event.Key())
		switch event.Key() {
		case tcell.KeyEnter:
			LoginCallback(form)()
		}
		return event
	})

	return form
}

func renderLoginErrTextView() *tview.TextView {
	textView := tview.NewTextView().
		SetWrap(true).
		SetDynamicColors(true)

	LoginErrOut = textView
	return textView
}

func LoginCallback(form *tview.Form) func() {
	return func() {
		count := form.GetFormItemCount()
		for i := 0; i < count; i++ {
			log.Println(form.GetFormItem(i).GetLabel())
			log.Println(form.GetFormItem(i).(*tview.InputField).GetText())
		}
		// username := form.GetFormItem(0).(*tview.InputField).GetText()
		// password := form.GetFormItem(1).(*tview.InputField).GetText()

		// TODO: connect to db
		// save current config

		tuiapp.KeeperTui.ShowPage("dashboard")
	}
}

func ExitCallback() func() {
	return func() {
		tuiapp.KeeperTui.App.Stop()
	}
}
