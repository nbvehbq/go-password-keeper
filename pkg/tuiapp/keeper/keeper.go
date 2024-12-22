package keeper

import (
	"github.com/nbvehbq/go-password-keeper/pkg/tuiapp"
)

func Init() {
	tuiapp.KeeperTui.AddPage("login", RenderLoginPage())
	tuiapp.KeeperTui.AddPage("dashboard", RenderDashBoardPage())

	// first enter into login page
	tuiapp.KeeperTui.ShowPage("login")
}
