package main

import (
	"fmt"
	"mc-portable-launcher/src/exp"
	"mc-portable-launcher/src/launcherApp"
)

func main() {

	exp.CheckToken()

	fmt.Println("Hello")

	app := launcherApp.NewLauncherApp()
	app.Initialize()

	app.Run()

}
