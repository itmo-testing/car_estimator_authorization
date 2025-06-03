package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	migrationTool 	*migrate.Migrate
}

func (m *Migrator) Init(conn *Connection, conf *Config, path string) error {
	driver, _ := postgres.WithInstance(conn.db, &postgres.Config{})
	newMigrationTool, err := migrate.NewWithDatabaseInstance(
        "file://" + path, conf.Driver, driver)

	if err != nil {
		fmt.Println("Error occured trying to init migration tool - ", err)
		return err
	}

	m.migrationTool = newMigrationTool
	return nil
}

func (m *Migrator) Apply() error {
	err := m.migrationTool.Up()

	if err == migrate.ErrNoChange {
		fmt.Println("Nothing to apply")
		return nil
	}

	if err != nil {
		fmt.Println("Can't apply migrations - ", err)
		return err
	}

	fmt.Println("Migration(s) applied successfully")
	return nil
}

func (m *Migrator) RollBack(steps int) error {
	err := m.migrationTool.Steps(-steps)

	if err == migrate.ErrNoChange {
		fmt.Println("Nothing to rollback")
		return nil
	}

	if err != nil {
		fmt.Println("Rollback failed - ", err)
		return err
	}

	fmt.Println("Rollback complete")
	return nil
}
