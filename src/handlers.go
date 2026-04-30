package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example/assets"
)
/*
func serveHome(w http.ResponseWriter, r *http.Request) {
	data, err := assets.FS.ReadFile("index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}
*/

// serveAsset serves any file from the embedded assets.FS by its filename.
// The Content-Type is detected automatically from the file extension.
func serveAsset(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := assets.FS.ReadFile(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	var creds struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if !store.AddUser(creds.Username, creds.Password) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user exists"})
		return
	}
	persistUser(creds.Username, creds.Password)
	log.Printf("👤 Created user: %s", creds.Username)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Account created!"})
}

func login(w http.ResponseWriter, r *http.Request) {
	var creds struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	token, ok := store.Authenticate(creds.Username, creds.Password)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "message": "Welcome back!"})
}

func checkAuth(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	user, ok := store.LookupSession(token)
	w.Header().Set("Content-Type", "application/json")
	if ok && token != "" {
		writeJSON(w, http.StatusOK, map[string]any{"logged_in": true, "user": user})
	} else {
		writeJSON(w, http.StatusOK, map[string]any{"logged_in": false})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
