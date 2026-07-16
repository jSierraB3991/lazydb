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
	a.pages.RemovePage(ADD_CONN_MODAL)
	a.pages.SwitchToPage(MAIN_PAGE)
}

func (a *App) showLoadingDialog(message string) {
	modal := tview.NewModal().SetText("⏳ " + message)
	modal.SetBorderColor(tcell.ColorDarkCyan)
	a.pages.AddPage(LOADING_MODAL, modal, false, true)
	a.tviewApp.SetFocus(modal)
}

func (a *App) hideLoadingDialog() {
	a.pages.RemovePage(LOADING_MODAL)
}

func (a *App) showConfirmDialog(message string, onConfirm func()) {

	modal := tview.NewModal().SetText(message).AddButtons([]string{"Eliminar", "Cancelar"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage(CONFIRM_MODAL)
			if buttonIndex == 0 {
				onConfirm()
			}
		})
	modal.SetBorderColor(tcell.ColorRed)
	a.pages.AddPage(CONFIRM_MODAL, modal, false, true)
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

	form.AddButton(BTN_TEXT_SAVE, func() {
		a.saveConfigConnection(form)
	})

	form.AddButton(BTN_TEXT_CANCEL, a.removeAddConn)
	form.SetCancelFunc(a.removeAddConn)

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 22, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)
	a.pages.AddPage(ADD_CONN_MODAL, modal, true, true)
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

func (a *App) cycleFocus(delta int) {
	panels := []tview.Primitive{a.connList, a.schemaTree, a.tableView}
	a.focusIndex = (a.focusIndex + delta + len(panels)) % len(panels)
	a.tviewApp.SetFocus(panels[a.focusIndex])
	a.updateBorders()
}

func (a *App) buildConnList() *tview.List {
	list := tview.NewList()
	list.SetTitle(CONN_LIST_TITLE).SetBorder(true)
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
	root := tview.NewTreeNode(NO_CONN_TEXT)
	tree := tview.NewTreeView()
	tree.SetRoot(root).SetCurrentNode(root)
	tree.SetTitle(SCHEMA_TABLES_TITLE).SetBorder(true)
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
	table.SetTitle(TABLE_VIEW_TITLE).SetBorderColor(tcell.ColorDarkCyan)
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
	statusBar.SetText(fmt.Sprintf(" %s  %s  %s  %s  %s ",
		TEXT_NEW_CONN, TEXT_CHANGE_FOCUS, TEXT_CONNECT, TEXT_DELETE, TEXT_QUIT))
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

	rootFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainFlex, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false)

	a.pages = tview.NewPages().AddPage(MAIN_PAGE, rootFlex, true, true)
	a.tviewApp.SetRoot(a.pages, true)
	a.tviewApp.SetFocus(a.connList)

	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			name, _ := a.pages.GetFrontPage()
			if name == MAIN_PAGE {
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
