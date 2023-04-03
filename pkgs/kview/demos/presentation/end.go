package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"
)

// End shows the final slide.
func End(nextSlide func()) (title string, info string, content kview.Primitive) {
	textView := kview.NewTextView()
	textView.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	url := "https://github.com/pktgen/go-pktgen/pkgs/kview"
	fmt.Fprint(textView, url)
	return "End", "", Center(len(url), 1, textView)
}
