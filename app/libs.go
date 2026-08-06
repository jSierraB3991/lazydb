package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DBType string

const (
	DBPostgres  DBType = "postgres"
	DBMySQL     DBType = "mysql"
	DBSQLServer DBType = "sqlserver"
	DBMongoDB   DBType = "mongodb"
)

func (d DBType) String() string {
	return string(d)
}

func (d DBType) GetSelect(schema, table string) string {
	switch d {
	case DBPostgres:
		return fmt.Sprintf(`SELECT * FROM "%s"."%s"`, schema, table)
	case DBMySQL:
		return fmt.Sprintf("SELECT * FROM `%s`", table)
	case DBSQLServer:
		return fmt.Sprintf("SELECT * FROM [%s].[%s]", schema, table)
	default:
		return fmt.Sprintf(`SELECT * FROM "%s"."%s"`, schema, table)
	}
}

const (
	NAME       string = "Nombre (Opcional)"
	MANAGEMENT string = "Gestor"
	HOST       string = "host"
	PORT       string = "Puerto"
	DB_NAME    string = "Base de Datos"
	USER       string = "Usuario"    //#gosec no sec
	PASSWORD   string = "Contraseña" //#gosec no sec
	ALLOW_SSL  string = "Permitir Ssl"

	LOADING_MODAL           string = "loading_modal"
	CONFIRM_MODAL           string = "confirm_modal"
	ADD_CONN_MODAL          string = "add_conn_modal"
	MAIN_PAGE               string = "main_page"
	UPDATE_SELECT_ROW_MODAL string = "update_select_row_modal"

	BTN_TEXT_SAVE       string = "Guardar"
	BTN_TEXT_CANCEL     string = "Cancelar"
	BTN_TEXT_PING       string = "Ping"
	TABLE_VIEW_TITLE    string = " Datos "
	CONN_LIST_TITLE     string = " Conexiones "
	NO_CONN_TEXT        string = " Sin conexiones "
	SCHEMA_TABLES_TITLE string = " Schema / Tablas "
	INPUT_FILTER_TITLE  string = " Búscar tabla "
	TEXT_NEW_CONN       string = "[yellow]Ctrl+a[-] Nueva conexión"
	TEXT_CHANGE_FOCUS   string = "[yellow]Tab[-] Cambiar Foco"
	TEXT_CONNECT        string = "[yellow]Enter[-] Conectar DB/Ver datos de la tabla/Editar Fila"
	TEXT_DELETE         string = "[yellow]Delete[-] Eliminar"
	TEXT_QUIT           string = "[yellow]Ctrl+c[-] Salir"
	TEXT_DISCONNECT     string = "[yellow]Ctrl+d[-] Desconectar"
	TEXT_CREATE_DB      string = "[yellow]Ctrl+b[-] Crear Base de Datos (Necesitas una conexión activa)"
	COLUMN_ID_GENERIC   string = "id"
	BASE_KEY_STRING     string = "BASE_64_KEY"
)

var (
	EMOJIREGEX = regexp.MustCompile(`^[\p{So}\p{Sk}\p{Mn}\x{fe0f}\n\r\t]+`)
)

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazydb", "connections.json")
}

func (a *App) setStatus(msg string) {
	mainStatus := fmt.Sprintf(" %s  %s  %s  %s  %s %s %s",
		TEXT_CREATE_DB, TEXT_NEW_CONN, TEXT_CHANGE_FOCUS, TEXT_CONNECT, TEXT_DELETE, TEXT_QUIT, TEXT_DISCONNECT)
	status := fmt.Sprintf("[green]λ[-]%s %s", msg, mainStatus)
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

// Función auxiliar para buscar recursivamente al padre de un nodo en tview
func findParent(current, target *tview.TreeNode) *tview.TreeNode {
	for _, child := range current.GetChildren() {
		if child == target {
			return current
		}
		// Seguir buscando en profundidad si el hijo tiene sub-nodos
		if found := findParent(child, target); found != nil {
			return found
		}
	}
	return nil
}

func getDbIdx(dbType DBType, dbAvailable []string) int {
	if dbType == "" {
		return 0
	}
	for i, db := range dbAvailable {
		if db == dbType.String() {
			return i
		}
	}
	return 0
}
func (a *App) getFormToConnect(data Connection) *tview.Form {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Nueva Conexión ").SetTitleColor(tcell.ColorAqua)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetFieldBackgroundColor(tcell.ColorDarkSlateGray)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorAqua)
	form.SetButtonBackgroundColor(tcell.ColorDarkCyan)

	dbAvailable := []string{DBPostgres.String(), DBMySQL.String()}
	form.AddDropDown(MANAGEMENT, dbAvailable, getDbIdx(data.Type, dbAvailable), nil)
	form.AddInputField(NAME, data.Name, 30, nil, nil)
	form.AddInputField(HOST, data.Host, 30, nil, nil)
	form.AddInputField(PORT, data.Port, 6, nil, nil)
	form.AddInputField(DB_NAME, data.DatabaseName, 30, nil, nil)
	form.AddInputField(USER, data.User, 30, nil, nil)
	form.AddPasswordField(PASSWORD, "", 30, '*', nil)
	form.AddCheckbox(ALLOW_SSL, data.AllowSsl, nil)

	form.AddButton(BTN_TEXT_PING, func() {
		_, labelDriver := form.GetFormItemByLabel(MANAGEMENT).(*tview.DropDown).GetCurrentOption()
		nameDb := form.GetFormItemByLabel(NAME).(*tview.InputField).GetText()
		a.setStatus(fmt.Sprintf("[yellow]Ping to type: %s database: %s[-]", labelDriver, nameDb))
		port := form.GetFormItemByLabel(PORT).(*tview.InputField).GetText()
		host := form.GetFormItemByLabel(HOST).(*tview.InputField).GetText()
		dbname := form.GetFormItemByLabel(DB_NAME).(*tview.InputField).GetText()
		user := form.GetFormItemByLabel(USER).(*tview.InputField).GetText()
		allowSsl := form.GetFormItemByLabel(ALLOW_SSL).(*tview.Checkbox).IsChecked()
		password := form.GetFormItemByLabel(PASSWORD).(*tview.InputField).GetText()

		conn := Connection{
			Name:         nameDb,
			Type:         DBType(labelDriver),
			Host:         host,
			Port:         port,
			DatabaseName: dbname,
			User:         user,
			Password:     password,
			IsEncrypted:  false,
			AllowSsl:     allowSsl,
		}

		db, err := sql.Open(labelDriver, conn.DSNUnEncrypt())
		if err != nil {
			a.setStatus(fmt.Sprintf("[red]Error: al tratar de conectar %s %v [-] ", labelDriver, err))
			return
		}
		if err := db.Ping(); err != nil {
			a.setStatus(fmt.Sprintf("[red]Error al hacer Ping %v[-]", err))
			a.CloseDb()
			return
		}
		a.setStatus(fmt.Sprintf("[green]Ping exitoso a la base de datos: %s[-]", dbname))
	})
	return form
}
