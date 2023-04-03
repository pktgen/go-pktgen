// Demo code for the DropDown primitive.
package main

import "github.com/pktgen/go-pktgen/pkgs/kview"

func main() {
	app := kview.NewApplication()
	defer app.HandlePanic()

	app.EnableMouse(true)

	dropdown := kview.NewDropDown()
	dropdown.SetLabel("Select an option (hit Enter): ")
	dropdown.SetOptions(nil,
		kview.NewDropDownOption("First"),
		kview.NewDropDownOption("Second"),
		kview.NewDropDownOption("Third"),
		kview.NewDropDownOption("Fourth"),
		kview.NewDropDownOption("Fifth"))

	app.SetRoot(dropdown, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
