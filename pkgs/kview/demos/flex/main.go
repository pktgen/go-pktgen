// Demo code for the Flex primitive.
package main

import (
	"github.com/pktgen/go-pktgen/pkgs/kview"
)

func demoBox(title string) *kview.Box {
	b := kview.NewBox()
	b.SetBorder(true)
	b.SetTitle(title)
	return b
}

func main() {
	app := kview.NewApplication()
	defer app.HandlePanic()

	app.EnableMouse(true)

	subFlex := kview.NewFlex()
	subFlex.SetDirection(kview.FlexRow)
	subFlex.AddItem(demoBox("Top"), 0, 1, false)
	subFlex.AddItem(demoBox("Middle (3 x height of Top)"), 0, 3, false)
	subFlex.AddItem(demoBox("Bottom (5 rows)"), 5, 1, false)

	flex := kview.NewFlex()
	flex.AddItem(demoBox("Left (1/2 x width of Top)"), 0, 1, false)
	flex.AddItem(subFlex, 0, 2, false)
	flex.AddItem(demoBox("Right (20 cols)"), 20, 1, false)

	app.SetRoot(flex, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
