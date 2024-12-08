package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"time"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	driverName   = "pgx"
	schemaName   = "backend"
	connLifetime = 10 * time.Second
)

func New(dsn string) (*sqlx.DB, error) {
	conn, err := sqlx.Connect(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx.Connect: %w", err)
	}

	conn.SetConnMaxLifetime(connLifetime)

	return conn, nil
}

func RunMigrations(db *sql.DB, migrationsPath string) error {
	if _, err := db.Exec("create schema if not exists backend"); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable:       "",
		MigrationsTableQuoted: false,
		MultiStatementEnabled: false,
		DatabaseName:          "",
		SchemaName:            schemaName,
		StatementTimeout:      0,
		MultiStatementMaxSize: 0,
	})
	if err != nil {
		return fmt.Errorf("postgres.WithInstance: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(migrationsPath, driverName, driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance: %w", err)
	}

	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration.Up: %w", err)
	}

	return nil
}
