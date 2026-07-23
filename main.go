package main

import (
	"fmt"
	"os"

	eliotlibs "github.com/jSierraB3991/jsierra-libs"
	"github.com/joho/godotenv"
	toolApp "github.com/jsierrab3991/lazydb/app"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("No se encontro un archivo .env %s \n", toolApp.BASE_KEY_STRING)
	}

	baseKey, err := eliotlibs.GetDataOfEnviromentRequired(toolApp.BASE_KEY_STRING)
	if err != nil {
		fmt.Printf("La aplicación requiere la variable %s \n", toolApp.BASE_KEY_STRING)
		os.Exit(1)
	}

	if err := eliotlibs.ValidateKey(baseKey); err != nil {
		fmt.Printf("La key '%s' no es una key base64 \n", baseKey)
		os.Exit(1)
	}
	app := toolApp.NewApp(baseKey)
	app.BuildUI()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error run app: %v\n", err)
		os.Exit(1)
	}
}
