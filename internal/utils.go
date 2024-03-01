package fsdatabase

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func createTable(db *sql.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tokenName TEXT UNIQUE,
			revokedStatus BOOLEAN
		);`
	_, err := db.Exec(createTableSQL)
	return err
}

func (db *Database) InsertToken(refreshtoken string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	revokedStatus := false //by default its false
	stmt, err := db.Prepare("INSERT INTO tokens (tokenName, revokedStatus) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(refreshtoken, revokedStatus)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) IsTokenRevoked(refreshtoken string) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM tokens WHERE tokenName = ? and revokedStatus = false LIMIT 1)"
	err := db.QueryRow(query, refreshtoken).Scan(&exists)
	if err != nil {
		return false, err
	}
	//if token exist and not revoked then the IsTokenRevoked will return FALSE
	return !exists, nil
}
