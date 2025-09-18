package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}
	if err := ensureDefaultAdmin(db); err != nil {
		return nil, err
	}
	if err := ensureDefaultGameSettings(db); err != nil {
		return nil, err
	}

	return db, nil
}

const (
	createGroupsTable = `
	CREATE TABLE IF NOT EXISTS groups (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		pathway TEXT NOT NULL,
		current_clue_idx INTEGER NOT NULL DEFAULT 0,
		completed BOOLEAN NOT NULL DEFAULT FALSE,
		end_time TIMESTAMPTZ,
		password TEXT NOT NULL
	);`

	createCluesTable = `
	CREATE TABLE IF NOT EXISTS clues (
		id SERIAL PRIMARY KEY,
		pathway TEXT NOT NULL,
		index_num INTEGER NOT NULL,
		content TEXT NOT NULL,
		qrcode TEXT NOT NULL,
		UNIQUE(pathway, index_num)
	);`

	createGameSettingsTable = `
	CREATE TABLE IF NOT EXISTS game_settings (
		id SERIAL PRIMARY KEY,
		total_clues INTEGER DEFAULT 1,
		start_time TIMESTAMPTZ,
		game_started BOOLEAN DEFAULT FALSE,
		game_ended BOOLEAN DEFAULT FALSE
	);`

	createAdminsTable = `
	CREATE TABLE IF NOT EXISTS admins (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`
)

func createTables(db *sql.DB) error {
	stmts := []string{
		createGroupsTable,
		createCluesTable,
		createGameSettingsTable,
		createAdminsTable,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultAdmin(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.Exec(`INSERT INTO admins (name, password) VALUES ($1, $2)`, "admin", "admin")
	return err
}

func ensureDefaultGameSettings(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM game_settings").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.Exec(`INSERT INTO game_settings (total_clues) VALUES ($1)`, 1)
	return err
}
