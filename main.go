package main

import (
	"fmt"
	"os"
	"strconv"

	toolApp "github.com/jsierrab3991/lazydb/app"
)

// main.go
func main() {
	if portStr := os.Getenv("LAZYDB_PORT"); portStr != "" {
		if _, err := strconv.Atoi(portStr); err != nil {
			fmt.Fprintf(os.Stderr, "LAZYDB_PORT debe ser un número\n")
			os.Exit(1)
		}
	}

	app := toolApp.NewApp()
	app.BuildUI()

	if err := app.Run(); err != nil {

		fmt.Fprintf(os.Stderr, "Error run app: %v\n", err)
		os.Exit(1)
	}
}
