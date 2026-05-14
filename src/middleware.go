package main

import (
	"log"
	"net/http"
	"strings"
)

// logger wraps a handler and logs the HTTP method and path for every request.
func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("--> %s %s", r.Method, r.URL.Path)
		next(w, r)
	}
}

// For protected routes, rather than repeating the auth check in every handler, extract a requireAuth middleware into middleware.go:
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
        if token == "" {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
            return
        }
        if _, ok := store.LookupSession(token); !ok {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
            return
        }
        next(w, r)
    }
}
