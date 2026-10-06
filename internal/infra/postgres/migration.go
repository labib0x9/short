package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/labib0x9/short/config"
)

func SetupDatabase(ctx context.Context, cnf *config.PostgreSQL) error {
	conn := NewPostgresSuperConn(ctx, cnf)
	defer conn.Close(ctx)

	var roleExists bool
	checkRoleQuery := `SELECT EXISTS(SELECT FROM pg_roles WHERE rolname = $1)`
	if err := conn.QueryRow(ctx, checkRoleQuery, cnf.User).Scan(&roleExists); err != nil {
		return fmt.Errorf("check role exists: %w", err)
	}

	if !roleExists {
		escapedPass := strings.ReplaceAll(cnf.Pass, "'", "''")
		createRoleSQL := fmt.Sprintf(
			`CREATE ROLE %s WITH LOGIN PASSWORD '%s'`,
			pgx.Identifier{cnf.User}.Sanitize(),
			escapedPass,
		)
		if _, err := conn.Exec(ctx, createRoleSQL); err != nil {
			return fmt.Errorf("create role: %w", err)
		}
	}

	var exists bool
	query := `SELECT EXISTS(SELECT FROM pg_database WHERE datname = $1)`
	if err := conn.QueryRow(ctx, query, cnf.DatabaseName).Scan(&exists); err != nil {
		return fmt.Errorf("check db exists: %w", err)
	}

	if !exists {
		createDBSQL := fmt.Sprintf(
			`CREATE DATABASE %s OWNER %s`,
			pgx.Identifier{cnf.DatabaseName}.Sanitize(),
			pgx.Identifier{cnf.User}.Sanitize(),
		)
		if _, err := conn.Exec(ctx, createDBSQL); err != nil {
			return fmt.Errorf("create database: %w", err)
		}
	}

	grantSQL := fmt.Sprintf(
		`GRANT ALL PRIVILEGES ON DATABASE %s TO %s`,
		pgx.Identifier{cnf.DatabaseName}.Sanitize(),
		pgx.Identifier{cnf.User}.Sanitize(),
	)
	if _, err := conn.Exec(ctx, grantSQL); err != nil {
		return fmt.Errorf("grant privileges: %w", err)
	}

	slog.Info("Database setup complete, run migration to create tables")
	return nil
}

func newMigrator(cnf *config.PostgreSQL) (*migrate.Migrate, func(), error) {
	dbSource := strings.Replace(newConnectionString(cnf), "postgres://", "pgx5://", 1)

	m, err := migrate.New("file://migrations", dbSource)
	if err != nil {
		return nil, nil, fmt.Errorf("init migrator: %w", err)
	}

	cleanup := func() {
		m.Close()
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
