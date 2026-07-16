package main

import (
	"fmt"
	"os"

	toolApp "github.com/jsierrab3991/lazydb/app"
)

func main() {
	app := toolApp.NewApp()
	app.BuildUI()

	if err := app.Run(); err != nil {

		fmt.Fprintf(os.Stderr, "Error run app: %v\n", err)
		os.Exit(1)
	}
}
