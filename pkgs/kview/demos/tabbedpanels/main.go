// Demo code for the TabbedPanels primitive.
package main

import (
	"fmt"

	"github.com/pktgen/go-pktgen/pkgs/kview"
)

const panelCount = 5

func main() {
	app := kview.NewApplication()
	defer app.HandlePanic()

	app.EnableMouse(true)

	panels := kview.NewTabbedPanels()
	for panel := 0; panel < panelCount; panel++ {
		func(panel int) {
			form := kview.NewForm()
			form.SetBorder(true)
			form.SetTitle(fmt.Sprintf("This is tab %d. Choose another tab.", panel+1))
			form.AddButton("Next", func() {
				panels.SetCurrentTab(fmt.Sprintf("panel-%d", (panel+1)%panelCount))
			})
			form.AddButton("Quit", func() {
				app.Stop()
			})
			form.SetCancelFunc(func() {
				app.Stop()
			})

			panels.AddTab(fmt.Sprintf("panel-%d", panel), fmt.Sprintf("Panel #%d", panel), form)
		}(panel)
	}
	//panels.SetTabSwitcherAfterContent(true)
	panels.SetTabSwitcherVertical(true)

	app.SetRoot(panels, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
