package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/rivo/tview"

	eliotlibs "github.com/jSierraB3991/jsierra-libs"
)

type Connection struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Type         DBType `json:"type"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DatabaseName string `json:"database"`
	IsEncrypted  bool   `json:"is_encrypted"`
	AllowSsl     bool   `json:"allow_ssl"`
}

func (c Connection) DSNPre(baseKey string) string {
	password := c.Password
	if c.IsEncrypted {
		passwordDecrypt, err := eliotlibs.Decrypt(c.Password, baseKey)
		if err == nil {
			password = passwordDecrypt
		}
	}
	sslConfig := "disable"
	if c.AllowSsl {
		sslConfig = "allow"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, password, c.DatabaseName, sslConfig)
}
func (c Connection) DSN(baseKey string) string {
	password := c.Password
	if c.IsEncrypted {
		passwordDecrypt, err := eliotlibs.Decrypt(c.Password, baseKey)
		if err == nil {
			password = passwordDecrypt
		}
	}

	var format, extra string
	switch c.Type {
	case "mysql":
		format = "%s:%s@tcp(%s:%s)/%s?tls=%s&parseTime=true"
		extra = "false"
		if c.AllowSsl {
			extra = "true"
		}
	case "sqlserver":
		format = "sqlserver://%s:%s@%s:%s?database=%s&encrypt=%s"
		extra = "disable"
		if c.AllowSsl {
			extra = "true"
		}
	default: // postgres
		format = "user=%s password=%s host=%s port=%s dbname=%s sslmode=%s"
		extra = "disable"
		if c.AllowSsl {
			extra = "allow"
		}
	}

	return fmt.Sprintf(format, c.User, password, c.Host, c.Port, c.DatabaseName, extra)
}

func (c Connection) DSNUnEncrypt() string {

	var format, extra string
	switch c.Type {
	case "mysql":
		format = "%s:%s@tcp(%s:%s)/%s?tls=%s&parseTime=true"
		extra = "false"
		if c.AllowSsl {
			extra = "true"
		}
	case "sqlserver":
		format = "sqlserver://%s:%s@%s:%s?database=%s&encrypt=%s"
		extra = "disable"
		if c.AllowSsl {
			extra = "true"
		}
	default: // postgres
		format = "user=%s password=%s host=%s port=%s dbname=%s sslmode=%s"
		extra = "disable"
		if c.AllowSsl {
			extra = "allow"
		}
	}
	return fmt.Sprintf(format, c.User, c.Password, c.Host, c.Port, c.DatabaseName, extra)
}

func (c Connection) DisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("%s@%s/%s", c.User, c.Host, c.DatabaseName)
}

func localConnections() ([]Connection, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil, err
	}

	var conns []Connection
	err = json.Unmarshal(data, &conns)
	if err != nil {
		return nil, err
	}
	return conns, nil
}

func saveConnections(baseKey string, setStatus func(msg string), conns []Connection) {
	path := configPath()
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		setStatus(fmt.Sprintf("[red]Error verify folder of connection %v[-]", err))
		return
	}

	for i := range conns {
		if conns[i].Id == "" {
			conns[i].Id = uuid.New().String()
			passwordEncript, err := eliotlibs.Encrypt(conns[i].Password, baseKey)
			if err != nil {
				setStatus(fmt.Sprintf("[red]Error encriptando pass: %v[-]", err))
				return
			}
			conns[i].Password = passwordEncript
			conns[i].IsEncrypted = true
		}
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

func (a *App) deleteConnectionOnClick(idx int) {

	name := a.connections[idx].DisplayName()
	a.connections = append(a.connections[:idx], a.connections[idx+1:]...)
	saveConnections(a.baseKey, a.setStatus, a.connections)
	a.rebuildConnList()
	if a.activeConn != nil && a.activeConn.DisplayName() == name {
		a.disconnect()
	}
	a.setStatus(fmt.Sprintf("[yellow]Conexión '%s' eliminada[-]", name))
}

func (a *App) deleteConnection(idx int) {
	if idx < 0 || idx >= len(a.connections) {
		return
	}
	name := a.connections[idx].DisplayName()
	a.showConfirmDialog(fmt.Sprintf("¿Eliminar conexión '%s'?", name), func() {
		a.deleteConnectionOnClick(idx)
	})
}

func (a *App) disconnect() {
	if a.activeDb == nil {
		a.setStatus("[yellow]No hay conexión activa[-]")
		return
	}

	name := a.activeConn.DisplayName()
	a.activeDb.Close()
	a.activeDb = nil
	a.activeConn = nil
	a.currentSchema = ""
	a.currentTable = ""
	a.schemaMap = nil
	a.filterTable.SetText("")
	a.filterRow.SetText("")
	a.filterTable.SetDisabled(true)

	// Limpiar el árbol de schemas
	root := tview.NewTreeNode("Sin conexión")
	a.schemaTree.SetRoot(root).SetCurrentNode(root)

	// Limpiar la tabla
	a.tableView.Clear()
	a.tableView.SetTitle(" Datos ")

	a.setStatus(fmt.Sprintf("[yellow]Desconectado de %s[-]", name))
}
