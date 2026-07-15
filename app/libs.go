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

	LOADING_MODAL  string = "loading_modal"
	CONFIRM_MODAL  string = "confirm_modal"
	ADD_CONN_MODAL string = "add_conn_modal"
	MAIN_PAGE      string = "main_page"

	BTN_TEXT_SAVE       string = "Guardar"
	BTN_TEXT_CANCEL     string = "Cancelar"
	TABLE_VIEW_TITLE    string = " Datos "
	CONN_LIST_TITLE     string = " Conexiones "
	NO_CONN_TEXT        string = " Sin conexiones "
	SCHEMA_TABLES_TITLE string = " Schema / Tablas "
	TEXT_NEW_CONN       string = " [yellow]Espacio[-] Nueva conexión"
	TEXT_CHANGE_FOCUS   string = " [yellow]Espacio[-] Cambiar Foco"
	TEXT_CONNECT        string = " [yellow]Espacio[-] Conectar/Ver Tabla"
	TEXT_DELETE         string = " [yellow]Espacio[-] Eliminar"
	TEXT_QUIT           string = " [yellow]Espacio[-] Salir"
)
