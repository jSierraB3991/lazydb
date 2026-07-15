package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type TableEntry struct {
	schema string
	table  string
}

func (a *App) deleteConnection(idx int) {
	if idx < 0 || idx >= len(a.connections) {
		return
	}
	name := a.connections[idx].DisplayName()
	a.showConfirmDialog(fmt.Sprintf("¿Eliminar conexión '%s'?", name), func() {
		a.connections = append(a.connections[:idx], a.connections[idx+1:]...)
		saveConnections(a.connections)
		a.rebuildConnList()
		a.setStatus(fmt.Sprintf("[yellow]Conexión '%s' eliminada[[-]", name))
	})
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
		fmt.Sprintf("¿Eliminar file %d de %s %s?", row, a.currentSchema, a.currentTable),
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

func saveConnections(conns []Connection) {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.MarshalIndent(conns, "", "  ")
	os.WriteFile(path, data, 0600)
}
