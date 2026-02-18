package migrations

import (
	"database/sql"
	"embed"
	"qvarkk/kvault/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB, dbName string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, dbName, driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	} else if err != migrate.ErrNoChange {
		logger.Logger.Info("migrations applied successfully")
	}

	return nil
}
