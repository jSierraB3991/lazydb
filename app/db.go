package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	_ "github.com/lib/pq"
)

type TableEntry struct {
	schema string
	table  string
}

func (a *App) loadSchemas() {
	if a.activeDb == nil {
		return
	}
	a.tableView.Clear()
	a.tableView.SetTitle(TABLE_VIEW_TITLE)

	rows, err := a.activeDb.Query(`
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY table_schema, table_name
        `)
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error cargando schemas %v[-]", err))
		return
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return
	}

	schemaMap := make(map[string][]string)
	var schemas []string
	for rows.Next() {
		var schema, table string
		err := rows.Scan(&schema, &table)
		if err == nil {
			if _, ok := schemaMap[schema]; !ok {
				schemas = append(schemas, schema)
			}
			schemaMap[schema] = append(schemaMap[schema], table)
		} else {
			a.setStatus(fmt.Sprintf("[red] Error get schema and tables %v[-]", err))
		}
	}
	root := tview.NewTreeNode(fmt.Sprintf("📦 %s", a.activeConn.DisplayName())).SetColor(tcell.ColorAqua)
	a.schemaTree.SetRoot(root).SetCurrentNode(root)

	for _, schema := range schemas {
		schemaNode := tview.NewTreeNode(fmt.Sprintf("📁 %s", schema)).SetColor(tcell.ColorYellow).
			SetSelectable(true).SetExpanded(true)
		for _, table := range schemaMap[schema] {
			tableNode := tview.NewTreeNode(fmt.Sprintf("🛋️ %s", table)).SetColor(tcell.ColorWhite).
				SetReference(fmt.Sprintf("%s.%s", schema, table)).SetSelectable(true)
			schemaNode.AddChild(tableNode)
		}
		root.AddChild(schemaNode)
	}
	a.setStatus(fmt.Sprintf("Conectado a la base de datos: %s", a.activeConn.DisplayName()))

}

func (a App) CloseDb() {
	if err := a.activeDb.Close(); err != nil {
		a.setStatus(fmt.Sprintf("[red]Error close connection: %v[-]", err))
	}
}

func (a *App) connectTo(conn *Connection) {
	a.showLoadingDialog(fmt.Sprintf("Conectando a %s...", conn.DisplayName()))
	if a.activeDb != nil {
		a.CloseDb()
		a.activeDb = nil
	}

	go func() {
		db, err := sql.Open("postgres", conn.DSN())
		a.tviewApp.QueueUpdateDraw(func() {
			defer a.hideLoadingDialog()
			if err != nil {
				a.setStatus(fmt.Sprintf("[red]Error: al tratar de conectar %v[-]", err))
				return
			}

			if err := db.Ping(); err != nil {
				a.setStatus(fmt.Sprintf("[red]Error al hacer Ping %v[-]", err))
				a.CloseDb()
				return
			}
			a.activeDb = db
			a.activeConn = conn
			a.loadSchemas()
			a.cycleFocus(1)
		})
	}()

}

func (a *App) loadTableData(schema string, table string) {
	if a.activeDb == nil {
		return
	}

	a.tableView.Clear()

	query := fmt.Sprintf(`SELECT * FROM "%s"."%s"`, schema, table)
	rows, err := a.activeDb.Query(query)
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error al leer la tabla: %v[-], %s", err, query))
		return
	}

	defer rows.Close()
	if err := rows.Err(); err != nil {
		return
	}
	cols, err := rows.Columns()
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error al leer las columnas %v[-]", err))
		return
	}

	//headers
	for i, col := range cols {
		cell := tview.NewTableCell(col).SetTextColor(tcell.ColorYellow).
			SetBackgroundColor(tcell.ColorDarkSlateGray).
			SetAttributes(tcell.AttrBold)
		a.tableView.SetCell(0, i, cell)
	}

	rowIdx := 1
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			a.setStatus(fmt.Sprintf("[red]Error scan ptrs %v[-]", err))
			continue
		}
		for i, val := range vals {
			text := ""
			if val != nil {
				text = fmt.Sprintf("%v", val)
			}
			cell := tview.NewTableCell(text).SetExpansion(1).SetTextColor(tcell.ColorWhite)
			a.tableView.SetCell(rowIdx, i, cell)
		}
		rowIdx++
	}
	a.setStatus(fmt.Sprintf("[green]%s.%s - %d filas[-]", schema, table, rowIdx-1))
	a.focusIndex = 2
	a.tviewApp.SetFocus(a.tableView)
	a.updateBorders()
}

func (a *App) deleteConnection(idx int) {
	if idx < 0 || idx >= len(a.connections) {
		return
	}
	name := a.connections[idx].DisplayName()
	a.showConfirmDialog(fmt.Sprintf("¿Eliminar conexión '%s'?", name), func() {
		a.connections = append(a.connections[:idx], a.connections[idx+1:]...)
		saveConnections(a.setStatus, a.connections)
		a.rebuildConnList()
		a.setStatus(fmt.Sprintf("[yellow]Conexión '%s' eliminada[[-]", name))
	})
}

func (a *App) yankRows(count int) {
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
		a.setStatus("[green]Copiado al porta papeles[-]")
	}
}

func (a *App) deleteSelectedRow() {
	if a.activeDb == nil || a.currentTable == "" {
		return
	}

	row, _ := a.tableView.GetSelection()
	if row == 0 {
		return
	}

	cols := a.tableView.GetColumnCount()
	colNames := make([]string, cols)
	for i := 0; i < cols; i++ {
		colNames[i] = a.tableView.GetCell(0, i).Text
	}

	conditions := []string{}
	args := []interface{}{}
	argIdx := 1
	for i, name := range colNames {
		val := a.tableView.GetCell(row, i).Text
		if val == "" {
			conditions = append(conditions, fmt.Sprintf(`"%s" IS NULL`, name))
		} else {
			conditions = append(conditions, fmt.Sprintf(`"%s" = $%d`, name, argIdx))
			args = append(args, val)
			argIdx++
		}
	}

	a.showConfirmDialog(
		fmt.Sprintf("¿Eliminar fila %d de %s %s?", row, a.currentSchema, a.currentTable),
		func() {
			query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE %s`, a.currentSchema, a.currentTable,
				strings.Join(conditions, " AND"))
			_, err := a.activeDb.Exec(query, args...)
			if err != nil {
				a.setStatus(fmt.Sprintf("[red]Error eliminado: %v[-]", err))
				return
			}
			a.tableView.RemoveRow(row)
			a.setStatus(fmt.Sprintf("[green]Fila eliminada de %s %s[-]", a.currentSchema, a.currentTable))
		})

}

func saveConnections(setStatus func(msg string), conns []Connection) {
	path := configPath()
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		setStatus(fmt.Sprintf("[red]Error verify folder of connection %v[-]", err))
	}
	data, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		setStatus(fmt.Sprintf("[red]Error convirtiendo la conexión en json %v[-]", err))
	}
	err = os.WriteFile(path, data, 0600)
	if err != nil {
		setStatus(fmt.Sprintf("[red]Error saving connection %v[-]", err))
	}
}
