package main

import (
	"github.com/andlabs/ui"
	"github.com/bertrandd/updater/middlewar"
	"log"
	"time"
	"fmt"
)

func main() {

	_, err := middlewar.GetVersion()
	if err != nil {
		log.Panic(err)
	}

	err = ui.Main(func() {
		//name := ui.NewEntry()
		button := ui.NewButton("Download latest version")
		progress:= ui.NewLabel("")
		box := ui.NewVerticalBox()
		box.Append(button, false)
		box.Append(progress, false)
		window := ui.NewWindow("MiddleWar", 200, 100, false)
		window.SetChild(box)
		button.OnClicked(func(*ui.Button) {

			button.Disable()
			button.SetText("Downloading...")
			var percent float64
			go middlewar.Update(&percent)
			go printProgress(progress, &percent)
			//greeting.SetText("Hello, " + name.Text() + "!")
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
	if err != nil {
		panic(err)
	}
}

func printProgress(label *ui.Label, percent *float64) {
	for {
		label.SetText(fmt.Sprintf("%d", int64(*percent))+"%")
		if *percent == 100.0 {
			break
		}
		time.Sleep(time.Second/2)
	}

}