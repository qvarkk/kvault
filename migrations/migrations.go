package migrations

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationsFS embed.FS

type Migrator struct {
	m *migrate.Migrate
}

func NewMigrator(db *sql.DB, dbName string) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", d, dbName, driver)
	if err != nil {
		return nil, err
	}

	return &Migrator{m: m}, nil
}

func (mg *Migrator) Up() error {
	err := mg.m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}

func (mg *Migrator) Down() error {
	return mg.m.Down()
}

func (mg *Migrator) Steps(n int) error {
	return mg.m.Steps(n)
}

func (mg *Migrator) Force(version int) error {
	return mg.m.Force(version)
}

func (mg *Migrator) Version() (uint, bool, error) {
	return mg.m.Version()
}
