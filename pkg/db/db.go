package db

import (
	"database/sql"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	// Get the directory where this db.go file is located
	_, filename, _, _ := runtime.Caller(0)
	dbDir := filepath.Dir(filename)

	// Construct database path in the same directory as db.go
	dbPath := filepath.Join(dbDir, "regatta.db")

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create tables if they don't exist
	return createTables()
}

func createTables() error {
	createTables := `
	CREATE TABLE IF NOT EXISTS regattas (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		location TEXT NOT NULL,
		status TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS teams (
		id TEXT PRIMARY KEY,
		regatta_id TEXT NOT NULL,
		name TEXT NOT NULL,
		FOREIGN KEY (regatta_id) REFERENCES regattas(id)
	);

	CREATE TABLE IF NOT EXISTS races (
		id TEXT PRIMARY KEY,
		regatta_id TEXT NOT NULL,
		start_time TIMESTAMP,
		end_time TIMESTAMP,
		status TEXT NOT NULL,
		FOREIGN KEY (regatta_id) REFERENCES regattas(id)
	);

	CREATE TABLE IF NOT EXISTS race_results (
		id TEXT PRIMARY KEY,
		regatta_id TEXT NOT NULL,
		team_id TEXT NOT NULL,
		race_number INTEGER NOT NULL,
		position INTEGER NOT NULL,
		points INTEGER NOT NULL,
		FOREIGN KEY (regatta_id) REFERENCES regattas(id),
		FOREIGN KEY (team_id) REFERENCES teams(id)
	);`

	_, err := DB.Exec(createTables)
	return err
}
