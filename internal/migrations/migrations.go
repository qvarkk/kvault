package migrations

import (
	"database/sql"
	"qvarkk/kvault/logger"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

func RunMigrations(db *sql.DB, dbName string, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		dbName, driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	} else if err == migrate.ErrNoChange {
		logger.Logger.Info("no migrations to do")
	} else {
		logger.Logger.Info("migrations applied successfully")
	}

	return nil
}
