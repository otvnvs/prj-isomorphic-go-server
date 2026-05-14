//go:build js && wasm

package main

import (
	"database/sql"
	"log"
	"sync"
)

var registerOnce sync.Once

// openDB registers the WASM SQLite driver (once) and opens a *sql.DB backed
// by it, assigning the result to the shared sqlDB variable.
func openDB() {
	registerOnce.Do(func() {
		sql.Register("wasmSQLite", &wasmDriver{})
	})
	var err error
	sqlDB, err = sql.Open("wasmSQLite", "")
	if err != nil {
		log.Fatalf("openDB: %v", err)
	}
}
