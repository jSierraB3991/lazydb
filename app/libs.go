package app

import (
	"fmt"
	"os"
	"path/filepath"
)

type DBType string

const (
	DBPostgres DBType = "postgres"
)

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazydb", "connections.json")
}

func (a *App) setStatus(msg string) {
	status := fmt.Sprintf("[green]λ[-] %s  [gray]|[-]  [yellow]Ctrl+C[-]: Salir", msg)
	a.statusBar.SetText(status)
}

const (
	NAME       string = "Nombre (Opcional)"
	MANAGEMENT string = "Gestor"
	HOST       string = "host"
	PORT       string = "Puerto"
	DB_NAME    string = "Base de Datos"
	USER       string = "Usuario"
	PASSWORD   string = "Contraseña"
)
