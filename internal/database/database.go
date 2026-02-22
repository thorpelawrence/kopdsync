package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS progress (
			device TEXT,
			device_id TEXT,
			document TEXT,
			percentage REAL,
			progress TEXT,
			timestamp TEXT,
			username TEXT,
			UNIQUE(username, document),
			FOREIGN KEY(username) REFERENCES users(username)
		);
	`)
	if err != nil {
		return err
	}

	return nil
}
