package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
)

type DBType string

const (
	DBPostgres DBType = "postgres"
)

const (
	NAME       string = "Nombre (Opcional)"
	MANAGEMENT string = "Gestor"
	HOST       string = "host"
	PORT       string = "Puerto"
	DB_NAME    string = "Base de Datos"
	USER       string = "Usuario"    //#gosec no sec
	PASSWORD   string = "Contraseña" //#gosec no sec
	ALLOW_SSL  string = "Permitir Ssl"

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
	TEXT_CHANGE_FOCUS   string = " [yellow]Tab[-] Cambiar Foco"
	TEXT_CONNECT        string = " [yellow]Enter[-] Conectar/Ver Tabla"
	TEXT_DELETE         string = " [yellow]Delete[-] Eliminar"
	TEXT_QUIT           string = " [yellow]Ctrl+C[-] Salir"
	TEXT_DISCONNECT     string = " [yellow]Ctrl+D[-] Desconectar"
	TEXT_CREATE_DB      string = " [yellow]Ctrl+B[-] Crear Base de Datos (Necesitas una conexión activa)"
	COLUMN_ID_GENERIC   string = "id"
	BASE_KEY_STRING     string = "BASE_64_KEY"
)

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazydb", "connections.json")
}

func (a *App) setStatus(msg string) {
	mainStatus := fmt.Sprintf(" %s  %s  %s  %s  %s %s %s",
		TEXT_CREATE_DB, TEXT_NEW_CONN, TEXT_CHANGE_FOCUS, TEXT_CONNECT, TEXT_DELETE, TEXT_QUIT, TEXT_DISCONNECT)
	status := fmt.Sprintf("[green]λ[-] %s %s  [gray]|[-]  [yellow]Ctrl+C[-]: Salir", msg, mainStatus)
	a.statusBar.SetText(status)
}

func (a *App) copySelectRow(count int) {
	row, _ := a.tableView.GetSelection()
	totalRows := a.tableView.GetRowCount() - 1

	end := row + count - 1
	if end > totalRows {
		end = totalRows
	}

	cols := a.tableView.GetColumnCount()
	headers := make([]string, cols)
	for i := 0; i < cols; i++ {
		headers[i] = a.tableView.GetCell(0, i).Text
	}

	var result []map[string]string
	for r := row; r <= end; r++ {
		rowMap := map[string]string{}
		for i := 0; i < cols; i++ {
			rowMap[headers[i]] = a.tableView.GetCell(r, i).Text
		}
		result = append(result, rowMap)
	}
	err := copyToClipboard(result)
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error Tratando de pasarlo al clipboard %s[-]", err))
	} else {
		fila := fmt.Sprintf("Copiada la fila %v", row)
		if (row - end) > 1 {
			fila = fmt.Sprintf("Copiadas las filas de: %v a la: %v", row, end)
		}
		a.setStatus(fmt.Sprintf("[green] %s al portapapeles[-]", fila))
	}
}

func copyToClipboard(dataToCopy []map[string]string) error {
	var data interface{}
	if len(dataToCopy) == 1 {
		data = dataToCopy[0]
	} else {
		data = dataToCopy
	}

	jsonBytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	return clipboard.WriteAll(string(jsonBytes))
}
