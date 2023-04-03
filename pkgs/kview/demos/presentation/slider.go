package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/pktgen/go-pktgen/pkgs/kview"
)

const sliderCode = `[green]package[white] main

[green]import[white] (
    [red]"fmt"[white]

    [red]"github.com/gdamore/tcell/v2"[white]
    [red]"github.com/pktgen/go-pktgen/pkgs/kview"[white]
)

[green]func[white] [yellow]main[white]() {
    slider := kview.[yellow]NewSlider[white]()
    slider.[yellow]SetLabel[white]([red]"Volume:   0%"[white])
    slider.[yellow][yellow]SetChangedFunc[white]([yellow]func[white](key tcell.Key) {
        label := fmt.[yellow]Sprintf[white]("Volume: %3d%%", value)
        slider.[yellow]SetLabel[white](label)
    })
    slider.[yellow][yellow]SetDoneFunc[white]([yellow]func[white](key tcell.Key) {
        [yellow]nextSlide[white]()
    })
    app := kview.[yellow]NewApplication[white]()
    app.[yellow]SetRoot[white](slider, true)
    app.[yellow]Run[white]()
}`

// Slider demonstrates the Slider.
func Slider(nextSlide func()) (title string, info string, content kview.Primitive) {
	slider := kview.NewSlider()
	slider.SetLabel("Volume:   0%")
	slider.SetChangedFunc(func(value int) {
		slider.SetLabel(fmt.Sprintf("Volume: %3d%%", value))
	})
	slider.SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	return "Slider", sliderInfo, Code(slider, 30, 1, sliderCode)
}
