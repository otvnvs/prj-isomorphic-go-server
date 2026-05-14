//go:build !js || !wasm

package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "a.db"

// openDB opens the SQLite database file for the native build and assigns sqlDB.
func openDB() {
	var err error
	sqlDB, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("openDB: %v", err)
	}
}
