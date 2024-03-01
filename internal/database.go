package fsdatabase

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the SQLite database
type Database struct {
	*sql.DB
	mu sync.Mutex
}

// NewDatabase creates and initializes a new Database instance
func NewDatabase(dbName string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}

	// Ensure the tokens table exists
	if err := createTable(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Database{DB: db}, nil
}

func (db *Database) Close() {
	db.DB.Close()
}
