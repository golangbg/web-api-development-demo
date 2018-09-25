package database

import (
	"github.com/golangbg/web-api-development/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

func (db *DB) SaveUser(user models.User, password string) (models.User, error) {
	if password != "" {
		// Passwords need to be stored encrypted in the database
		// We can hash the password with the bcrypt package (https://godoc.org/golang.org/x/crypto/bcrypt#GenerateFromPassword)
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return user, err
		}

		// The hashed password is stored as a string
		user.Password = string(hash)
	}

	// Prepare the query
	q := `INSERT OR REPLACE INTO users(username, name, password)
	values(?, ?, ?)`
	stmt, err := db.conn.Prepare(q)
	if err != nil {
		// Preparing the query went wrong, so we'll return an empty post and the error
		return user, err
	}
	// Make sure stmt gets closed
	defer stmt.Close()

	// Ececute the query
	res, err := stmt.Exec(user.Username, user.Name, user.Password)
	if err != nil {
		// Execution went wrong, so we'll return an empty post and the error
		return models.User{}, err
	}

	// Update the user with the provided ID if a new record was inserted
	if id, err := res.LastInsertId(); err == nil {
		user.ID = id
	}
	// Everything went well, let's return the user and nil for the error
	return user, nil
}

// GetUserByUsername gets a user by the username
func (db *DB) GetUserByUsername(username string) (user models.User, err error) {
	// Prepare the query
	q := "SELECT * FROM users WHERE username=?"
	stmt, err := db.conn.Prepare(q)
	if err != nil {
		// Preparing the query went wrong, so we'll return an empty user and the error
		return user, err
	}
	// Make sure stmt gets closed
	defer stmt.Close()

	// Get the user
	if err := stmt.QueryRow(username).Scan(&user.ID, &user.Username, &user.Name, &user.Password); err != nil {
		return user, err
	}

	return user, err
}
