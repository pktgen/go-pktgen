package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"
)

const logo = `
 _           _
| | ____   _(_) _____      __
| |/ /\ \ / / |/ _ \ \ /\ / /
| - <  \ V /| |  __/\ V  V /
|_|\_\  \_/ |_|\___| \_/\_/
`

const subtitle = "Terminal-based user interface toolkit"

// Cover returns the cover page.
func Cover(nextSlide func()) (title string, info string, content kview.Primitive) {
	// What's the size of the logo?
	lines := strings.Split(logo, "\n")
	logoWidth := 0
	logoHeight := len(lines)
	for _, line := range lines {
		if len(line) > logoWidth {
			logoWidth = len(line)
		}
	}
	logoBox := kview.NewTextView()
	logoBox.SetTextColor(tcell.ColorGreen.TrueColor())
	logoBox.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	fmt.Fprint(logoBox, logo)

	// Create a frame for the subtitle and navigation infos.
	frame := kview.NewFrame(kview.NewBox())
	frame.SetBorders(0, 0, 0, 0, 0, 0)
	frame.AddText(subtitle, true, kview.AlignCenter, tcell.ColorDarkMagenta.TrueColor())

	// Create a Flex layout that centers the logo and subtitle.
	subFlex := kview.NewFlex()
	subFlex.AddItem(kview.NewBox(), 0, 1, false)
	subFlex.AddItem(logoBox, logoWidth, 1, true)
	subFlex.AddItem(kview.NewBox(), 0, 1, false)

	flex := kview.NewFlex()
	flex.SetDirection(kview.FlexRow)
	flex.AddItem(kview.NewBox(), 0, 7, false)
	flex.AddItem(subFlex, logoHeight, 1, true)
	flex.AddItem(frame, 0, 10, false)

	return "Start", appInfo, flex
}
