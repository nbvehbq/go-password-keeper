package main

import (
	"log"

	"github.com/rivo/tview"
	"github.com/spf13/cobra"

	"github.com/nbvehbq/go-password-keeper/pkg/tuiapp"
	"github.com/nbvehbq/go-password-keeper/pkg/tuiapp/keeper"
)

var keeperCmd = &cobra.Command{
	Use:   "keeper",
	Short: "password keeper client",
	Long:  "password keeper client",
	Run: func(cmd *cobra.Command, args []string) {
		keeper.Init()

		layout := tview.NewFlex().
			AddItem(tuiapp.KeeperTui.Pages, 0, 1, true)

		if err := tuiapp.KeeperTui.App.SetRoot(layout, true).
			EnableMouse(true).
			Run(); err != nil {
			log.Fatal(err, "run application")
		}
	},
}

func main() {
	if err := keeperCmd.Execute(); err != nil {
		log.Fatal(err, "execute command")
	}
}
