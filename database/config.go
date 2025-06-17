package database

import (
	"fmt"
)

type Config struct {
	Driver   string
	Addr     string
	User     string
	Password string
	DBName   string
}

func (conf *Config) GetPgConnString(defaultConn bool) string {
	dbName := conf.DBName
	if defaultConn {
		dbName = ""
	}

	return fmt.Sprintf(
		"%s://%s:%s@%s/%s?sslmode=disable", conf.Driver, conf.User, conf.Password, conf.Addr, dbName,
	)
}

func (conf *Config) GetRedisConnString() string {
	return fmt.Sprintf(
		"%s://:%s@%s/%s", conf.Driver, conf.Password, conf.Addr, conf.DBName,
	)
}