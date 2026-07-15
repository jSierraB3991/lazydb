package app

import (
	"encoding/json"
	"fmt"
	"os"
)

type Connection struct {
	Name         string `json:"name"`
	Type         DBType `json:"type"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DatabaseName string `json:"database"`
}

func (c Connection) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DatabaseName)
}

func (c Connection) DisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("%s@%s/%s", c.User, c.Host, c.DatabaseName)
}

func localConnections() []Connection {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil
	}

	var conns []Connection
	json.Unmarshal(data, &conns)
	return conns
}
