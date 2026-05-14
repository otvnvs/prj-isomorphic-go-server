//go:build !js || !wasm

package main

import (
	"fmt"
	"net/http"
)

func startServer() {
	openDB()
	initDB()
	fmt.Println("Starting native server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
