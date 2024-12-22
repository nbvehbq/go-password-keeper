package tuiapp

import "github.com/rivo/tview"

var KeeperTui *Application

type Application struct {
	Pages   *tview.Pages
	App     *tview.Application
	Widgets []tview.Primitive
}

func init() {
	KeeperTui = &Application{
		App:     tview.NewApplication(),
		Pages:   tview.NewPages(),
		Widgets: make([]tview.Primitive, 0),
	}
}

func (a *Application) GetPage() *tview.Pages {
	return a.Pages
}

func (a *Application) AddPage(name string, item tview.Primitive) {
	a.Pages.AddPage(name, item, true, false)

}

func (a *Application) ShowPage(name string) {
	a.Pages.SwitchToPage(name)
}

func (a *Application) AddWidget(w tview.Primitive) {
	a.Widgets = append(a.Widgets, w)
}

func (a *Application) GetCurrentFocus() tview.Primitive {
	return a.App.GetFocus()
}

func (a *Application) SetNextFocus() {
	wiget := a.GetCurrentFocus()
	a.App.SetFocus(a.NextWigets(wiget))
}

func (a *Application) PreviousWidgets(curent tview.Primitive) tview.Primitive {
	for i, w := range a.Widgets {
		if w == curent {
			if i-1 >= 0 {
				return a.Widgets[i-1]
			}
			return a.Widgets[len(a.Widgets)-1]
		}
	}
	return a.Widgets[0]
}

func (a *Application) NextWigets(curent tview.Primitive) tview.Primitive {
	for i, w := range a.Widgets {
		if w == curent {
			if i+1 < len(a.Widgets) {
				return a.Widgets[i+1]
			}
			return a.Widgets[0]
		}
	}
	return a.Widgets[0]
}
