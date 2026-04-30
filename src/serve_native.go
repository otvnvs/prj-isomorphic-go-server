//go:build !js || !wasm

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const dbFile = "a.db"

func startServer() {
	loadFromDisk()
	fmt.Println("Starting native server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func loadFromDisk() {
	data, err := os.ReadFile(dbFile)
	if err != nil {
		return
	}
	var persisted map[string]string
	if err := json.Unmarshal(data, &persisted); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not parse %s: %v\n", dbFile, err)
		return
	}
	store.LoadUsers(persisted)
}

func persistUser(username, password string) {
	snap := store.Snapshot()
	data, _ := json.MarshalIndent(snap, "", "  ")
	os.WriteFile(dbFile, data, 0644)
}
