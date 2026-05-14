package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

const (
	maxUsernameLen = 64
	maxPasswordLen = 128
)

func signup(w http.ResponseWriter, r *http.Request) {
	var creds struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if creds.Username == "" || creds.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password required"})
		return
	}
	if len(creds.Username) > maxUsernameLen || len(creds.Password) > maxPasswordLen {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username or password too long"})
		return
	}
	ok, err := dbAddUser(creds.Username, creds.Password)
	if err != nil {
		log.Printf("signup: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	if !ok {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user exists"})
		return
	}
	log.Printf("👤 Created user: %s", creds.Username)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Account created!"})
}

func login(w http.ResponseWriter, r *http.Request) {
	var creds struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	ok, err := dbAuthenticate(creds.Username, creds.Password)
	if err != nil {
		log.Printf("login: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}
	token := store.CreateSession(creds.Username)
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "message": "Welcome back!"})
}

func checkAuth(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == "" {
		writeJSON(w, http.StatusOK, map[string]any{"logged_in": false})
		return
	}
	user, ok := store.LookupSession(token)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"logged_in": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"logged_in": true, "user": user})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func sampleNotRequiringToken(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "sampleNotRequiringToken"})
}

func sampleRequiringToken(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "sampleRequiringToken"})
}
