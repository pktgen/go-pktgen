// Demo code for the Frame primitive.
package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"
)

func main() {
	app := kview.NewApplication()
	defer app.HandlePanic()

	app.EnableMouse(true)

	box := kview.NewBox()
	box.SetBackgroundColor(tcell.ColorBlue.TrueColor())

	frame := kview.NewFrame(box)
	frame.SetBorders(2, 2, 2, 2, 4, 4)
	frame.AddText("Header left", true, kview.AlignLeft, tcell.ColorWhite.TrueColor())
	frame.AddText("Header middle", true, kview.AlignCenter, tcell.ColorWhite.TrueColor())
	frame.AddText("Header right", true, kview.AlignRight, tcell.ColorWhite.TrueColor())
	frame.AddText("Header second middle", true, kview.AlignCenter, tcell.ColorRed.TrueColor())
	frame.AddText("Footer middle", false, kview.AlignCenter, tcell.ColorGreen.TrueColor())
	frame.AddText("Footer second middle", false, kview.AlignCenter, tcell.ColorGreen.TrueColor())

	app.SetRoot(frame, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
