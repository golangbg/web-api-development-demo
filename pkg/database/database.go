package database

import (
	"database/sql"
	"fmt"

	// The driver won't be used directly, therefore we use a blank import
	_ "github.com/mattn/go-sqlite3"
)

// DB provides access to the databases
type DB struct {
	conn *sql.DB
}

// InitDB initializes the database
func (db *DB) InitDB() error {
	// Check if the server has a database connection
	if db.conn == nil {
		return fmt.Errorf("no database is set")
	}

	// Define the posts table
	posts := `CREATE TABLE IF NOT EXISTS posts(
		slug TEXT NOT NULL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT,
		created DATETIME,
		modified DATETIME
	);`

	// Create the posts table
	if _, err := db.conn.Exec(posts); err != nil {
		// Couldn't create the table, return the error
		return err
	}

	users := `CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		name TEXT,
		password TEXT NOT NULL
	);`

	// Create the users table
	if _, err := db.conn.Exec(users); err != nil {
		// Couldn't create the table, return the error
		return err
	}

	return nil
}

// CloseDB closes the database connection
func (db *DB) CloseDB() error {
	return db.conn.Close()
}

// New creates a database connection. Returns a pointer to DB or an error
func New(name string) (*DB, error) {
	// Open the database
	sqlite3, err := sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}

	// Create a DB instance
	db := DB{conn: sqlite3}

	// Initialize the database
	if err := db.InitDB(); err != nil {
		return nil, err
	}

	return &db, nil
}
