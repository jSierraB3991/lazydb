package app

import (
	"encoding/json"
	"fmt"
	"os"

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

func (c Connection) DSN(baseKey string) string {
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
