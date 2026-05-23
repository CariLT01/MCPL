package main

import (
	"fmt"
	"mc-portable-launcher/src/exp"
	"mc-portable-launcher/src/launcherApp"
	"mc-portable-launcher/src/ux"
	"os"
)

func handlePanic() {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "Error: The application encountered a fatal issue and must exit.\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n", r)

		ux.ShowError("Fatal Error Occurred", "A fatal error occurred inside the application. MCPL cannot recover.\n\nPlease report this issue.")

		os.Exit(1)
	}
}

func main() {

	defer handlePanic()

	exp.CheckToken()

	fmt.Println("Hello")

	app := launcherApp.NewLauncherApp()
	app.Initialize()

	app.Run()

}
