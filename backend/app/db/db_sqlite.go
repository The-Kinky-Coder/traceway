//go:build !pgch

package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/tracewayapp/traceway/backend/app/config"
	"github.com/tracewayapp/lit/v2"
	_ "modernc.org/sqlite"
)

func Init() error {
	cfg := config.Config
	if cfg.DBType == "sqlite" {
		return initSQLite()
	}
	return initPostgres()
}

func initSQLite() error {
	path := config.Config.SQLitePath
	if path == "" {
		path = "./traceway.db"
	}

	mainDB, err := openSQLite(path, true)
	if err != nil {
		return err
	}
	DB = mainDB
	Driver = lit.SQLite
	config.Logf("SQLite database opened at %s", path)

	telemetryPath := strings.TrimSuffix(path, ".db") + "_telemetry.db"
	if path == ":memory:" {
		telemetryPath = ":memory:"
	}
	telDB, err := openSQLite(telemetryPath, false)
	if err != nil {
		return err
	}
	TelemetryDB = telDB
	config.Logf("SQLite telemetry database opened at %s", telemetryPath)

	return nil
}

func openSQLite(path string, foreignKeys bool) (*sql.DB, error) {
	d, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite at %s: %w", path, err)
	}
	if err := d.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite at %s: %w", path, err)
	}
	if foreignKeys {
		if _, err := d.Exec("PRAGMA foreign_keys = ON"); err != nil {
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	}
	if _, err := d.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}
	d.SetMaxOpenConns(1)
	return d, nil
}
