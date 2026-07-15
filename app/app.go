package app

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	activeDb    *sql.DB
	activeConn  *Connection
	connections []Connection

	tviewApp   *tview.Application
	statusBar  *tview.TextView
	connList   *tview.List
	schemaTree *tview.TreeView
	tableView  *tview.Table
	leftFlex   *tview.Flex
	pages      *tview.Pages

	focusIndex    int
	currentTable  string
	currentSchema string
}

func NewApp() *App {
	a := &App{
		tviewApp: tview.NewApplication(),
	}
	a.connections = localConnections()
	return a
}

func (a *App) saveConfigConnection(form *tview.Form) {
	_, dbTypeStr := form.GetFormItemByLabel(MANAGEMENT).(*tview.DropDown).GetCurrentOption()
	name := form.GetFormItemByLabel(NAME).(*tview.InputField).GetText()
	port := form.GetFormItemByLabel(PORT).(*tview.InputField).GetText()
	host := form.GetFormItemByLabel(HOST).(*tview.InputField).GetText()
	dbname := form.GetFormItemByLabel(DB_NAME).(*tview.InputField).GetText()
	user := form.GetFormItemByLabel(USER).(*tview.InputField).GetText()
	password := form.GetFormItemByLabel(PASSWORD).(*tview.InputField).GetText()

	conn := Connection{
		Name:         name,
		Type:         DBType(dbTypeStr),
		Host:         host,
		Port:         port,
		DatabaseName: dbname,
		User:         user,
		Password:     password,
	}
	a.connections = append(a.connections, conn)
	saveConnections(a.connections)
	a.rebuildConnList()
	a.removeAddConn()

}

func (a *App) rebuildConnList() {
	a.connList.Clear()
	for _, c := range a.connections {
		icon := "🐘"
		a.connList.AddItem(icon+" "+c.DisplayName(), "", 0, func() {
			a.connectTo(&c)
		})
	}
	if len(a.connections) == 0 {
		a.connList.AddItem("[gray]Sin conexiones - Espacio para agregar[-]", "", 0, nil)
	}
}

func (a *App) removeAddConn() {
	a.pages.RemovePage("modal")
	a.pages.SwitchToPage("main")
}

func (a *App) showConfirmDialog(message string, onConfirm func()) {

	modal := tview.NewModal().SetText(message).AddButtons([]string{"Eliminar", "Cancelar"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm")
			if buttonIndex == 0 {
				onConfirm()
			}
		})
	modal.SetBorderColor(tcell.ColorRed)
	a.pages.AddPage("confirm", modal, false, true)
	a.tviewApp.SetFocus(modal)
}

func (a *App) showAddConnectionModal() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Nuev Conexión ").SetTitleColor(tcell.ColorAqua)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetFieldBackgroundColor(tcell.ColorDarkSlateGray)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorAqua)
	form.SetButtonBackgroundColor(tcell.ColorDarkCyan)

	form.AddDropDown(MANAGEMENT, []string{"postgres"}, 0, nil)
	form.AddInputField(NAME, "", 30, nil, nil)
	form.AddInputField(HOST, "localhost", 30, nil, nil)
	form.AddInputField(PORT, "5432", 6, nil, nil)
	form.AddInputField(DB_NAME, "", 30, nil, nil)
	form.AddInputField(USER, "", 30, nil, nil)
	form.AddPasswordField(PASSWORD, "", 30, '*', nil)

	form.AddButton("Guardar", func() {
		a.saveConfigConnection(form)
	})

	form.AddButton("Cancelar", a.removeAddConn)
	form.SetCancelFunc(a.removeAddConn)

	modal := tview.NewFlex().AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 22, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true)
	a.pages.AddPage("modal", modal, true, true)
	a.tviewApp.SetFocus(form)
}

func (a *App) updateBorders() {
	a.connList.SetBorderColor(tcell.ColorDarkCyan)
	a.schemaTree.SetBorderColor(tcell.ColorDarkCyan)
	a.tableView.SetBorderColor(tcell.ColorDarkCyan)

	switch a.focusIndex {
	case 0:
		a.connList.SetBorderColor(tcell.ColorYellow)
	case 1:
		a.schemaTree.SetBorderColor(tcell.ColorYellow)
	case 2:
		a.tableView.SetBorderColor(tcell.ColorYellow)
	}
}

func (a *App) loadSchemas() {
	if a.activeDb == nil {
		return
	}
	a.tableView.Clear()
	a.tableView.SetTitle("Datos")

	rows, err := a.activeDb.Query(`
	    SELECT table_schema, table_name
	    FROM information_schema.tables
	    WHERE table_schema NOT IN ('pg_catalog','information_schema')
	    ORDER BY table_schema, table_name   
	`)
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error cargando schemas: %v[-]", err))
		return
	}
	defer rows.Close()
	schemaMap := make(map[string][]string)
	var schemas []string
	for rows.Next() {
		var schema, table string
		rows.Scan(&schema, &table)
		if _, ok := schemaMap[schema]; !ok {
			schemas = append(schemas, schema)
		}
		schemaMap[schema] = append(schemaMap[schema], table)
	}
	root := tview.NewTreeNode(fmt.Sprintf("📦 %s", a.activeConn.DisplayName())).SetColor(tcell.ColorAqua)
	a.schemaTree.SetRoot(root).SetCurrentNode(root)

	for _, schema := range schemas {
		schemaNode := tview.NewTreeNode("📁 " + schema).SetColor(tcell.ColorYellow).
			SetSelectable(true).SetExpanded(true)
		for _, table := range schemaMap[schema] {
			tableNode := tview.NewTreeNode(" 🗃️ " + table).SetColor(tcell.ColorWhite).
				SetReference(schema + "." + table).SetSelectable(true)
			schemaNode.AddChild(tableNode)
		}
		root.AddChild(schemaNode)
	}
}

func (a *App) loadTableData(schema string, table string) {
	if a.activeDb == nil {
		return
	}

	a.tableView.Clear()
	a.tableView.SetTitle(fmt.Sprintf(" %s %s ", schema, table))

	query := fmt.Sprintf(`SELECT * FROM "%s"."%s" LIMIT 500`, schema, table)
	rows, err := a.activeDb.Query(query)
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error al leer la tabla: %v[-]", err))
		return
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error al leer las columnas: %v[-]", err))
		return
	}

	//headers
	for i, col := range cols {
		cell := tview.NewTableCell(col).SetTextColor(tcell.ColorYellow).
			SetBackgroundColor(tcell.ColorDarkSlateGray).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1)
		a.tableView.SetCell(0, i, cell)
	}

	//Data
	rowIdx := 1
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {
		rows.Scan(ptrs...)
		for i, val := range vals {
			text := ""
			if val != nil {
				text = fmt.Sprintf("%v", val)
			}
			cell := tview.NewTableCell(text).SetExpansion(1).
				SetTextColor(tcell.ColorWhite)
			a.tableView.SetCell(rowIdx, i, cell)
		}
		rowIdx++
	}

	a.setStatus(fmt.Sprintf("[green]%s.%s - %d filas[-]", schema, table, rowIdx-1))
	a.focusIndex = 2
	a.tviewApp.SetFocus(a.tableView)
	a.updateBorders()
}

func (a *App) connectTo(conn *Connection) {
	if a.activeDb != nil {
		a.activeDb.Close()
		a.activeDb = nil
	}
	db, err := sql.Open("postgres", conn.DSN())
	if err != nil {
		a.setStatus(fmt.Sprintf("[red]Error: %v[-]", err))
		return
	}
	if err := db.Ping(); err != nil {
		a.setStatus(fmt.Sprintf("[red]No se pudo conectar %v[-]", err))
		db.Close()
		return
	}
	a.activeDb = db
	a.activeConn = conn
	a.setStatus(fmt.Sprintf("[green]Conectado a: %s[-]", conn.DisplayName()))
	a.loadSchemas()

	a.focusIndex = 1
	a.tviewApp.SetFocus(a.schemaTree)
	a.updateBorders()
}

func (a *App) cycleFocus(delta int) {
	panels := []tview.Primitive{a.connList, a.schemaTree, a.tableView}
	a.focusIndex = (a.focusIndex + delta + len(panels)) % len(panels)
	a.tviewApp.SetFocus(panels[a.focusIndex])
	a.updateBorders()
}

func (a *App) buildConnList() *tview.List {
	list := tview.NewList()
	list.SetTitle(" Conexiones ").SetBorder(true)
	list.SetTitleColor(tcell.ColorAqua)
	list.SetBorderColor(tcell.ColorDarkCyan)
	list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)

	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)

	for _, c := range a.connections {
		icon := "🐘"
		list.AddItem(icon+" "+c.DisplayName(), "", 0, func() { a.connectTo(&c) })
	}

	if len(a.connections) == 0 {
		list.AddItem("[gray]Sin conexiones - Espacio para agregar[-]", "", 0, nil)
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			a.cycleFocus(1)
			return nil
		case tcell.KeyBacktab:
			a.cycleFocus(-1)
			return nil
		case tcell.KeyDelete, tcell.KeyBackspace2:
			idx := list.GetCurrentItem()
			if idx >= 0 && idx < len(a.connections) {
				a.deleteConnection(idx)
			}
			return nil
		}
		return event
	})

	return list
}

func (a *App) buildSchemaTree() *tview.TreeView {
	root := tview.NewTreeNode("Sin Conexión")
	tree := tview.NewTreeView()
	tree.SetRoot(root).SetCurrentNode(root)
	tree.SetTitle(" Schema / Tablas ").SetBorder(true)
	tree.SetTitleColor(tcell.ColorAqua)
	tree.SetBorderColor(tcell.ColorDarkCyan)
	tree.SetGraphicsColor(tcell.ColorDarkCyan)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		ref := node.GetReference()
		if ref == nil {
			return
		}

		switch v := ref.(type) {
		case string:
			parts := strings.SplitN(v, ".", 2)
			if len(parts) == 2 {
				a.currentSchema = parts[0]
				a.currentTable = parts[1]
				a.loadTableData(parts[0], parts[1])
			}
		}
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			a.cycleFocus(1)
			return nil
		case tcell.KeyBacktab:
			a.cycleFocus(-1)
			return nil
		}
		return event
	})

	return tree
}

func (a *App) buildTableView() *tview.Table {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetTitle(" Datos ").SetBorderColor(tcell.ColorDarkCyan)
	table.SetTitleColor(tcell.ColorDarkCyan)
	table.SetFixed(0, 1)
	table.SetSelectable(true, false)
	table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorWhite))

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			a.cycleFocus(1)
			return nil
		case tcell.KeyBacktab:
			a.cycleFocus(-1)
			return nil
		case tcell.KeyDelete:
			a.deleteSelectedRow()
			return nil
		}
		return event
	})

	return table
}

func (a *App) buildStatusBar() *tview.TextView {
	statusBar := tview.NewTextView()
	statusBar.SetDynamicColors(true)
	statusBar.SetText(" [yellow]Espacio[-] Nueva conexión  [yellow]Tab[-] Cambiar Foco  [yellow]Enter[-] Conectar /Ver tabla  [yellow]Del[-] Eliminar  [yellow]Ctrl+c[-] Salir")
	statusBar.SetBackgroundColor(tcell.ColorDarkSlateGray)
	return statusBar
}

func (a *App) BuildUI() {
	a.connList = a.buildConnList()
	a.schemaTree = a.buildSchemaTree()
	a.tableView = a.buildTableView()
	a.statusBar = a.buildStatusBar()

	a.leftFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.connList, 0, 1, true).
		AddItem(a.schemaTree, 0, 2, false)

	mainFlex := tview.NewFlex().
		AddItem(a.leftFlex, 0, 1, true).
		AddItem(a.tableView, 0, 3, false)

	rootFlex := tview.NewFlex().
		AddItem(mainFlex, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false)

	a.pages = tview.NewPages().AddPage("main", rootFlex, true, true)
	a.tviewApp.SetRoot(a.pages, true)
	a.tviewApp.SetFocus(a.connList)

	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			name, _ := a.pages.GetFrontPage()
			if name == "main" {
				a.showAddConnectionModal()
			}
		}
		return event
	})
}

func (a *App) Run() error {
	a.updateBorders()
	return a.tviewApp.Run()
}
