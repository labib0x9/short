package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labib0x9/short/config"
	_ "github.com/lib/pq"
)

func SetupDatabase(ctx context.Context, cnf *config.PostgreSQL) error {
	conn := NewPostgresSuperConn(ctx, cnf)
	defer conn.Close(ctx)

	createRoleSQL := fmt.Sprintf(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '%s') THEN
				CREATE ROLE %s WITH LOGIN PASSWORD '%s';
			END IF;
		END
		$$;
	`, cnf.User, cnf.User, cnf.Pass)

	if _, err := conn.Exec(ctx, createRoleSQL); err != nil {
		return fmt.Errorf("create role: %w", err)
	}

	var exists bool
	query := `SELECT EXISTS(SELECT FROM pg_database WHERE datname = $1)`
	if err := conn.QueryRow(ctx, query, cnf.DatabaseName).Scan(&exists); err != nil {
		return fmt.Errorf("check db exists: %w", err)
	}

	if !exists {
		createDBSQL := fmt.Sprintf(`CREATE DATABASE %s OWNER %s`, cnf.DatabaseName, cnf.User)
		if _, err := conn.Exec(ctx, createDBSQL); err != nil {
			return fmt.Errorf("create database: %w", err)
		}
	}

	grantSQL := fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s`, cnf.DatabaseName, cnf.User)
	if _, err := conn.Exec(ctx, grantSQL); err != nil {
		return fmt.Errorf("grant privileges: %w", err)
	}

	slog.Info("Database setup complete, run migration to create tables")
	return nil
}

func newMigrator(cnf *config.PostgreSQL) (*migrate.Migrate, func(), error) {
	dbSource := newConnectionString(cnf)

	appDB, err := sql.Open("postgres", dbSource)
	if err != nil {
		return nil, nil, fmt.Errorf("open db for migration: %w", err)
	}

	driver, err := postgres.WithInstance(appDB, &postgres.Config{})
	if err != nil {
		appDB.Close()
		return nil, nil, fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		appDB.Close()
		return nil, nil, fmt.Errorf("init migrator: %w", err)
	}

	cleanup := func() {
		m.Close()
		appDB.Close()
	}

	return m, cleanup, nil
}

func Run(ctx context.Context, cnf *config.PostgreSQL) error {
	m, cleanup, err := newMigrator(cnf)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}

// Rollback runs down migrations. If steps == 0, it rolls back all migrations.
func Rollback(ctx context.Context, cnf *config.PostgreSQL, steps int) error {
	m, cleanup, err := newMigrator(cnf)
	if err != nil {
		return err
	}
	defer cleanup()

	if steps == 0 {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration down: %w", err)
		}
	} else {
		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration rollback (%d steps): %w", steps, err)
		}
	}

	slog.Info("Migrations rollback complete")
	return nil
}
