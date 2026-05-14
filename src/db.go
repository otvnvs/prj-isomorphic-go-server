package main

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// sqlDB is the shared *sql.DB handle. Opened by openDB() in db_native.go or
// db_wasm.go before initDB() is called.
var sqlDB *sql.DB

// initDB creates the schema.
func initDB() {
	log.Println("initDB: begin")
	if _, err := sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL
		)
	`); err != nil {
		log.Fatalf("initDB: CREATE TABLE: %v", err)
	}
	log.Println("initDB: ready")
}

// dbAddUser inserts a new user with a bcrypt-hashed password.
// Returns false if the username is already taken.
func dbAddUser(username, password string) (bool, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	res, err := sqlDB.Exec(
		"INSERT OR IGNORE INTO users (username, password) VALUES (?, ?)",
		username, string(hash),
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// dbAuthenticate checks whether the password matches the stored bcrypt hash.
func dbAuthenticate(username, password string) (bool, error) {
	var hash string
	err := sqlDB.QueryRow(
		"SELECT password FROM users WHERE username = ?",
		username,
	).Scan(&hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false, nil
	}
	return true, nil
}
