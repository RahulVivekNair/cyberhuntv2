package database

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Group struct {
	ID             int
	Name           string
	Pathway        string
	CurrentClueIdx int
	Completed      bool
	EndTime        *time.Time
	Password       string
}

type Clue struct {
	ID      int
	Pathway string
	Index   int
	Content string
	QRCode  string
}

type GameSettings struct {
	ID          int
	TotalClues  int
	StartTime   *time.Time
	GameStarted bool
	GameEnded   bool
}

type Admin struct {
	ID       int
	Name     string
	Password string
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		pathway TEXT NOT NULL,
		current_clue_idx INTEGER DEFAULT 0,
		completed BOOLEAN DEFAULT FALSE,
		end_time DATETIME,
		password TEXT NOT NULL
	);`

	createCluesTable = `
	CREATE TABLE IF NOT EXISTS clues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pathway TEXT NOT NULL,
		index_num INTEGER NOT NULL,
		content TEXT NOT NULL,
		qrcode TEXT NOT NULL,
		UNIQUE(pathway, index_num)
	);`

	createGameSettingsTable = `
	CREATE TABLE IF NOT EXISTS game_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		total_clues INTEGER DEFAULT 1,
		start_time DATETIME,
		game_started BOOLEAN DEFAULT FALSE,
		game_ended BOOLEAN DEFAULT FALSE
	);`

	createAdminsTable = `
	CREATE TABLE IF NOT EXISTS admins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
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
	_, err := db.Exec(`INSERT INTO admins (name, password) VALUES (?, ?)`, "admin", "admin")
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
	_, err := db.Exec(`INSERT INTO game_settings (total_clues) VALUES (?)`, 1)
	return err
}
