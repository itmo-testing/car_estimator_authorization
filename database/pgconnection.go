package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Connection struct {
	db *sql.DB
}

type Config struct {
	Driver string
	Addr string
	User string
	Password string
	DBName string
}

func (conf *Config) getConnString(defaultConn bool) string {
	dbName := conf.DBName
	if defaultConn {
		dbName = ""
	}
	return fmt.Sprintf(
		"%s://%s@%s:%s/%s?sslmode=disable", conf.Driver, conf.User, conf.Password, conf.Addr, dbName,
	)
}

func (c *Connection) Init(conf *Config) error {
	if err := CreateDBIfNotExists(conf); err != nil {
		return err
	}

	db, err := sql.Open(conf.Driver, conf.getConnString(false))
	if err != nil {
		fmt.Println("invalid connection arguments:", err)
		return err	
	}

	if err := db.Ping(); err != nil {
		fmt.Println("database connection failed:", err)
		return err
	}	
	
	c.db = db
	return nil
}

func (c *Connection) Close() {
	if c.db != nil {
		c.db.Close()
	}
}

func CreateDBIfNotExists(conf *Config) error {
	defaultConn, err := sql.Open(conf.Driver, conf.getConnString(true))
	if err != nil {
		fmt.Println("invalid default connection arguments:", err)
		return err
	}

	defer defaultConn.Close()
	var exists bool

	err = defaultConn.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", conf.DBName).Scan(&exists)
	if err != nil {
		fmt.Println("database existense check query failed:", err)
		return err
	}

	if !exists {
		_, err := defaultConn.Exec("CREATE DATABASE $1", conf.DBName)
		if err != nil {
			fmt.Println("database creation failed:", err)
			return err
		}
	}

	return nil
}
