package main

import (
	"log"
	"net/http"
)

// logger wraps a handler and logs the HTTP method and path for every request.
func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("--> %s %s", r.Method, r.URL.Path)
		next(w, r)
	}
}
