//go:build js && wasm

package main

import (
	"encoding/json"
	"log"
	"syscall/js"
	"time"

	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

func startServer() {
	loadFromIndexedDB()
	wasmhttp.Serve(nil)
	select {}
}

func loadFromIndexedDB() {
	for js.Global().Get("loadUsersFromDB").IsUndefined() {
		time.Sleep(100 * time.Millisecond)
	}
	promise := js.Global().Call("loadUsersFromDB")
	successCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		var stored []struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &stored); err != nil {
			log.Println("loadFromIndexedDB: unmarshal error:", err)
			return nil
		}
		imported := make(map[string]string, len(stored))
		for _, u := range stored {
			imported[u.Username] = u.Password
		}
		store.LoadUsers(imported)
		log.Printf("✅ Loaded %d users from IndexedDB", len(stored))
		return nil
	})
	promise.Call("then", successCb)
}

func persistUser(username, password string) {
	user := map[string]string{"username": username, "password": password}
	b, _ := json.Marshal(user)
	js.Global().Call("saveUserToDB", string(b))
}
